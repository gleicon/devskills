package bench

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Usage is one run's token spend, normalized across assistants. Output
// includes reasoning tokens. CostUSD is a list-price figure, never what a
// subscription actually bills; CostKnown is false when neither the assistant
// nor evals/bench.yaml could price the run.
type Usage struct {
	Input      int // input tokens not served from or written to the cache
	CacheRead  int
	CacheWrite int
	Output     int
	CostUSD    float64
	CostKnown  bool
}

func (u Usage) String() string {
	cost := "cost unknown"
	if u.CostKnown {
		cost = fmt.Sprintf("$%.4f", u.CostUSD)
	}
	return fmt.Sprintf("input %d, cache read %d, cache write %d, output %d, %s",
		u.Input, u.CacheRead, u.CacheWrite, u.Output, cost)
}

// claudeResult is the one object `claude -p --output-format json` prints.
type claudeResult struct {
	Type         string   `json:"type"`
	Subtype      string   `json:"subtype"`
	IsError      bool     `json:"is_error"`
	Result       string   `json:"result"`
	Errors       []string `json:"errors"`
	TotalCostUSD float64  `json:"total_cost_usd"`
	ModelUsage   map[string]struct {
		InputTokens              int `json:"inputTokens"`
		OutputTokens             int `json:"outputTokens"`
		CacheReadInputTokens     int `json:"cacheReadInputTokens"`
		CacheCreationInputTokens int `json:"cacheCreationInputTokens"`
	} `json:"modelUsage"`
}

// parseClaude splits Claude's JSON result into the final text the checker
// scores and the run's usage. Usage is non-nil whenever the result decoded,
// even when Claude reports an error — a failed run still spent tokens.
func parseClaude(stdout string) (string, *Usage, error) {
	var r claudeResult
	if err := json.Unmarshal([]byte(stdout), &r); err != nil {
		return "", nil, fmt.Errorf("claude output is not a JSON result: %w", err)
	}
	if r.Type != "result" {
		return "", nil, fmt.Errorf("claude output has type %q, want \"result\"", r.Type)
	}
	u := &Usage{CostUSD: r.TotalCostUSD, CostKnown: true}
	// The top-level usage covers the main loop only; modelUsage also
	// counts subagents.
	for _, m := range r.ModelUsage {
		u.Input += m.InputTokens
		u.CacheRead += m.CacheReadInputTokens
		u.CacheWrite += m.CacheCreationInputTokens
		u.Output += m.OutputTokens
	}
	switch {
	case r.Subtype != "success":
		return "", u, fmt.Errorf("claude %s: %s", r.Subtype, strings.Join(r.Errors, "; "))
	case r.IsError:
		return r.Result, u, fmt.Errorf("claude error: %s", r.Result)
	}
	return r.Result, u, nil
}

// codexEvent is one line of `codex exec --json`, reduced to the fields bench
// reads.
type codexEvent struct {
	Type string `json:"type"`
	Item struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"item"`
	Usage *struct {
		InputTokens           int `json:"input_tokens"`
		CachedInputTokens     int `json:"cached_input_tokens"`
		CacheWriteInputTokens int `json:"cache_write_input_tokens"`
		OutputTokens          int `json:"output_tokens"`
	} `json:"usage"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// parseCodex reads Codex's JSONL event stream: the last agent message is the
// final text, and the last turn.completed carries usage — a running total for
// the thread, so it is never summed. Codex reports no cost; price, when
// non-nil, supplies it. Usage is non-nil whenever a turn completed.
func parseCodex(stdout string, price *Price) (string, *Usage, error) {
	var (
		text   string
		u      *Usage
		failed error
	)
	for line := range strings.Lines(stdout) {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var e codexEvent
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			return "", nil, fmt.Errorf("codex output is not a JSON event stream: %w", err)
		}
		switch e.Type {
		case "item.completed":
			if e.Item.Type == "agent_message" {
				text = e.Item.Text
			}
		case "turn.completed":
			if e.Usage == nil {
				return "", nil, fmt.Errorf("codex turn.completed carries no usage")
			}
			// input_tokens already counts the cached and cache-write tokens,
			// and output_tokens the reasoning ones.
			u = &Usage{
				Input:      max(e.Usage.InputTokens-e.Usage.CachedInputTokens-e.Usage.CacheWriteInputTokens, 0),
				CacheRead:  e.Usage.CachedInputTokens,
				CacheWrite: e.Usage.CacheWriteInputTokens,
				Output:     e.Usage.OutputTokens,
			}
		case "turn.failed":
			failed = fmt.Errorf("codex turn failed: %s", e.Error.Message)
		}
	}
	if u == nil {
		return "", nil, cmp.Or(failed, errors.New("codex output has no turn.completed event"))
	}
	if price != nil {
		u.CostUSD, u.CostKnown = price.Cost(*u), true
	}
	return text, u, failed
}

// openCodeEvent is one line of `opencode run --format json`, reduced to the
// fields bench reads.
type openCodeEvent struct {
	Type string `json:"type"`
	Part struct {
		Text   string  `json:"text"`
		Cost   float64 `json:"cost"`
		Tokens struct {
			Input     float64 `json:"input"`
			Output    float64 `json:"output"`
			Reasoning float64 `json:"reasoning"`
			Cache     struct {
				Read  float64 `json:"read"`
				Write float64 `json:"write"`
			} `json:"cache"`
		} `json:"tokens"`
	} `json:"part"`
	Error struct {
		Name string `json:"name"`
		Data struct {
			Message string `json:"message"`
		} `json:"data"`
	} `json:"error"`
}

// parseOpenCode reads OpenCode's JSONL event stream. Every text part joins
// the final text, as the text mode prints them all. Tokens and cost arrive
// per step and are summed. Usage is non-nil whenever a step finished.
func parseOpenCode(stdout string) (string, *Usage, error) {
	var (
		texts  []string
		u      *Usage
		failed error
	)
	for line := range strings.Lines(stdout) {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var e openCodeEvent
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			return "", nil, fmt.Errorf("opencode output is not a JSON event stream: %w", err)
		}
		switch e.Type {
		case "text":
			if t := strings.TrimSpace(e.Part.Text); t != "" {
				texts = append(texts, t)
			}
		case "step_finish":
			if u == nil {
				u = &Usage{}
			}
			// input excludes the cache counts and output excludes
			// reasoning, unlike Claude's and Codex's.
			tk := e.Part.Tokens
			u.Input += int(tk.Input)
			u.CacheRead += int(tk.Cache.Read)
			u.CacheWrite += int(tk.Cache.Write)
			u.Output += int(tk.Output + tk.Reasoning)
			u.CostUSD += e.Part.Cost
		case "error":
			failed = fmt.Errorf("opencode error: %s", cmp.Or(e.Error.Data.Message, e.Error.Name))
		}
	}
	if u == nil {
		return "", nil, cmp.Or(failed, errors.New("opencode output has no step_finish event"))
	}
	// OpenCode reports 0 for a model it cannot price, including every model
	// behind a ChatGPT sign-in, so 0 means unknown, not free.
	u.CostKnown = u.CostUSD > 0
	return strings.Join(texts, "\n"), u, failed
}
