package bench

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Usage is one run's token spend, normalized across assistants. Output
// includes reasoning tokens; CostUSD is the assistant's list-price figure,
// never what a subscription actually bills.
type Usage struct {
	Input      int // input tokens not served from or written to the cache
	CacheRead  int
	CacheWrite int
	Output     int
	CostUSD    float64
}

func (u Usage) String() string {
	return fmt.Sprintf("input %d, cache read %d, cache write %d, output %d, $%.4f",
		u.Input, u.CacheRead, u.CacheWrite, u.Output, u.CostUSD)
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
	u := &Usage{CostUSD: r.TotalCostUSD}
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
