package bench

import (
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
			wantUsage: &Usage{Input: 107, CacheRead: 3050, CacheWrite: 400, Output: 23, CostUSD: 0.5},
		},
		{
			name:      "error subtype fails the run but keeps its spend",
			stdout:    `{"type":"result","subtype":"error_max_turns","is_error":true,"errors":["hit max turns"],"total_cost_usd":0.2,"modelUsage":{}}`,
			wantUsage: &Usage{CostUSD: 0.2},
			wantErr:   "error_max_turns: hit max turns",
		},
		{
			name:      "success flagged is_error carries the API error in result",
			stdout:    `{"type":"result","subtype":"success","is_error":true,"result":"API Error: overloaded","total_cost_usd":0,"modelUsage":{}}`,
			wantText:  "API Error: overloaded",
			wantUsage: &Usage{},
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
