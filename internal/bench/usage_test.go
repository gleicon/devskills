package bench

import (
	"math"
	"strings"
	"testing"
)

func TestParseClaude(t *testing.T) {
	tests := []struct {
		name      string
		stdout    string
		wantText  string
		wantUsage *Usage
		wantErr   string
	}{
		{
			name: "success sums usage across models",
			stdout: `{"type":"result","subtype":"success","is_error":false,"result":"done","total_cost_usd":0.5,
				"usage":{"input_tokens":1,"output_tokens":1},
				"modelUsage":{
					"claude-sonnet-5":{"inputTokens":100,"outputTokens":20,"cacheReadInputTokens":3000,"cacheCreationInputTokens":400},
					"claude-haiku-4-5":{"inputTokens":7,"outputTokens":3,"cacheReadInputTokens":50,"cacheCreationInputTokens":0}}}`,
			wantText:  "done",
			wantUsage: &Usage{Input: 107, CacheRead: 3050, CacheWrite: 400, Output: 23, CostUSD: 0.5, CostKnown: true},
		},
		{
			name:      "error subtype fails the run but keeps its spend",
			stdout:    `{"type":"result","subtype":"error_max_turns","is_error":true,"errors":["hit max turns"],"total_cost_usd":0.2,"modelUsage":{}}`,
			wantUsage: &Usage{CostUSD: 0.2, CostKnown: true},
			wantErr:   "error_max_turns: hit max turns",
		},
		{
			name:      "success flagged is_error carries the API error in result",
			stdout:    `{"type":"result","subtype":"success","is_error":true,"result":"API Error: overloaded","total_cost_usd":0,"modelUsage":{}}`,
			wantText:  "API Error: overloaded",
			wantUsage: &Usage{CostKnown: true},
			wantErr:   "API Error: overloaded",
		},
		{name: "plain text", stdout: "hello", wantErr: "not a JSON result"},
		{name: "empty", stdout: "", wantErr: "not a JSON result"},
		{name: "wrong type", stdout: `{"type":"system"}`, wantErr: `type "system"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, usage, err := parseClaude(tt.stdout)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("err = %v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("err = %v, want it to contain %q", err, tt.wantErr)
			}
			if text != tt.wantText {
				t.Errorf("text = %q, want %q", text, tt.wantText)
			}
			if (usage == nil) != (tt.wantUsage == nil) || (usage != nil && *usage != *tt.wantUsage) {
				t.Errorf("usage = %+v, want %+v", usage, tt.wantUsage)
			}
		})
	}
}

func TestParseCodex(t *testing.T) {
	terra := &Price{Checked: "2026-09-26", Input: 2, CacheRead: 0.2, CacheWrite: 2.5, Output: 12}
	// Mirrors Codex's own fixture: input_tokens counts the cached and
	// cache-write tokens, output_tokens counts the reasoning ones.
	const stream = `{"type":"thread.started","thread_id":"t1"}
{"type":"turn.started"}
{"type":"item.completed","item":{"id":"item_0","type":"agent_message","text":"looking"}}
{"type":"item.completed","item":{"id":"item_1","type":"command_execution","command":"cat a.txt","aggregated_output":"secret fixture text","exit_code":0,"status":"completed"}}
{"type":"item.completed","item":{"id":"item_2","type":"agent_message","text":"found it"}}
{"type":"turn.completed","usage":{"input_tokens":1000000,"cached_input_tokens":400000,"cache_write_input_tokens":100000,"output_tokens":100000,"reasoning_output_tokens":50000}}
`
	tests := []struct {
		name      string
		stdout    string
		price     *Price
		wantText  string
		wantUsage *Usage
		wantErr   string
	}{
		{
			name:     "priced",
			stdout:   stream,
			price:    terra,
			wantText: "found it",
			// 500k×$2 + 400k×$0.20 + 100k×$2.50 + 100k×$12, per 1M
			wantUsage: &Usage{Input: 500000, CacheRead: 400000, CacheWrite: 100000, Output: 100000, CostUSD: 2.53, CostKnown: true},
		},
		{
			name:      "no price leaves cost unknown",
			stdout:    stream,
			wantText:  "found it",
			wantUsage: &Usage{Input: 500000, CacheRead: 400000, CacheWrite: 100000, Output: 100000},
		},
		{
			name:    "turn failed",
			stdout:  `{"type":"turn.started"}` + "\n" + `{"type":"turn.failed","error":{"message":"rate limited"}}` + "\n",
			wantErr: "codex turn failed: rate limited",
		},
		{name: "no turn completed", stdout: `{"type":"turn.started"}` + "\n", wantErr: "no turn.completed"},
		{name: "plain text", stdout: "done\n", wantErr: "not a JSON event stream"},
		{name: "empty", stdout: "", wantErr: "no turn.completed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, usage, err := parseCodex(tt.stdout, tt.price)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("err = %v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("err = %v, want it to contain %q", err, tt.wantErr)
			}
			if text != tt.wantText {
				t.Errorf("text = %q, want %q", text, tt.wantText)
			}
			if (usage == nil) != (tt.wantUsage == nil) {
				t.Fatalf("usage = %+v, want %+v", usage, tt.wantUsage)
			}
			if usage != nil {
				got, want := *usage, *tt.wantUsage
				if math.Abs(got.CostUSD-want.CostUSD) > 1e-9 {
					t.Errorf("cost = %v, want %v", got.CostUSD, want.CostUSD)
				}
				got.CostUSD, want.CostUSD = 0, 0
				if got != want {
					t.Errorf("usage = %+v, want %+v", got, want)
				}
			}
		})
	}
}
