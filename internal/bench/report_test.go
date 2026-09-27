package bench

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite golden files")

func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run with -update to create)", err)
	}
	if got != string(want) {
		t.Errorf("output does not match %s (run with -update to rewrite):\n%s", path, got)
	}
}

func sampleReport() Report {
	return Report{
		Skill:   "ds-deslop",
		Command: "devskills bench ds-deslop --runs 2 --model claude-sonnet-5 --format pr-md",
		OldSHA:  "1111111111111111111111111111111111111111",
		NewSHA:  "2222222222222222222222222222222222222222",
		Groups: []HarnessReport{{
			Harness: "Claude Code",
			Model:   "claude-sonnet-5",
			Scenarios: []ScenarioReport{{
				Name:         "narrated-greeting",
				Expectations: 3,
				Old: []RunReport{
					{Checked: true, Hits: 1, Extras: 0, Usage: &Usage{Input: 10, CacheRead: 900, CacheWrite: 50, Output: 30, CostUSD: 0.05, CostKnown: true}, Stdout: "removed one comment\n", Diff: "diff --git a/greet.go b/greet.go\n-// First we get the greeting\n"},
					{Failed: true, FailMsg: "timed out after 5m0s", Stderr: "signal: killed\n"},
				},
				New: []RunReport{
					{Checked: true, Hits: 3, Extras: 1, Usage: &Usage{CostUSD: 0.04, CostKnown: true}, Stdout: "cleaned all three\nplus a ```code``` fence\n"},
					{Checked: true, Hits: 2, Extras: 0, Usage: &Usage{CostUSD: 0.03, CostKnown: true}, Stdout: "cleaned two\n"},
				},
			}},
		}},
	}
}

func TestReportMarkdown(t *testing.T) {
	got := sampleReport().Markdown()
	golden(t, "report.golden.md", got)

	for _, want := range []string{
		"| run | old | new |",
		"| 2 | failed | 2/3 hits, 0 extra · $0.0300 |",
		"| **aggregate** | 1/6 hits, 0 extra | 5/6 hits, 1 extra |",
		"| **cost / success** | no success ($0.0500 spent) | $0.0700 (1/2 succeeded) |",
		"| **median cost** | $0.0500 | $0.0350 |",
		"usage: input 10, cache read 900, cache write 50, output 30, $0.0500",
		"<details>", "old run 2", "timed out after 5m0s", "````",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("markdown missing %q", want)
		}
	}
	// FR-13: interpretation belongs to humans.
	if strings.Contains(strings.ToLower(got), "verdict") ||
		strings.Contains(strings.ToLower(got), "pass") || strings.Contains(strings.ToLower(got), "improved") {
		t.Error("report must carry no verdict on old vs new")
	}
}

func TestReportMarkdownTiers(t *testing.T) {
	r := Report{
		Skill:   "ds-project-checkpoint",
		Command: "devskills bench ds-project-checkpoint --runs 2 --model claude-sonnet-5 --format pr-md",
		OldSHA:  "1111111111111111111111111111111111111111",
		NewSHA:  "2222222222222222222222222222222222222222",
		Groups: []HarnessReport{{
			Harness: "Claude Code",
			Model:   "claude-sonnet-5",
			Scenarios: []ScenarioReport{
				{
					Name: "fresh-checkpoint", Tier: TierStructural, Expectations: 4,
					Old: []RunReport{
						{Checked: true, Hits: 3, Stdout: "wrote state.md\n"},
						{Checked: true, Hits: 4, Stdout: "wrote state.md\n"},
					},
					New: []RunReport{
						{Checked: true, Hits: 4, Stdout: "wrote state.md\n"},
						{Checked: true, Hits: 4, Stdout: "wrote state.md\n"},
					},
				},
				{
					Name: "invoke", Tier: TierSmoke,
					Old: []RunReport{
						{Checked: true, Hits: 1, Stdout: "Git mode active.\n"},
						{Failed: true, FailMsg: "exit status 1", Stderr: "boom\n"},
					},
					New: []RunReport{
						{Checked: true, Hits: 1, Stdout: "Git mode active.\n"},
						{Checked: true, Hits: 0},
					},
				},
			},
		}},
	}
	got := r.Markdown()
	golden(t, "report-tiers.golden.md", got)
	if strings.Contains(got, "cost / success") {
		t.Error("cost rows must be omitted when no run reported usage")
	}

	for _, want := range []string{
		"| 1 | 3/4 elements | 4/4 elements |",
		"| **aggregate** | 7/8 elements | 8/8 elements |",
		"| 1 | ok | ok |",
		"| 2 | failed | no output |",
		"| **aggregate** | 1/2 ok | 1/2 ok |",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("markdown missing %q", want)
		}
	}
}

func TestReportMarkdownBaseline(t *testing.T) {
	r := sampleReport()
	r.Baseline = true
	r.OldSHA = ""
	r.Groups[0].Scenarios[0].Old = nil
	got := r.Markdown()
	golden(t, "report-baseline.golden.md", got)
	if !strings.Contains(got, "| run | new |") || strings.Contains(got, "| run | old | new |") {
		t.Error("baseline report must be single-column")
	}
	if !strings.Contains(got, "Baseline mode") {
		t.Error("baseline report must say so")
	}
}

func TestReportMarkdownBlocks(t *testing.T) {
	tests := []struct {
		name     string
		baseline bool
		want     []string
	}{
		{
			name: "each block names its old and new SHA",
			want: []string{
				"- Blocks, each from its version's tree:\n",
				"  - `base`: old `aaaa`, new `bbbb`\n",
				"  - `concise`: new `cccc` (absent on the main branch, so old runs go without)\n",
			},
		},
		{
			name:     "baseline names only the new SHA",
			baseline: true,
			want: []string{
				"  - `base`: new `bbbb`\n",
				"  - `concise`: new `cccc`\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := sampleReport()
			r.Baseline = tt.baseline
			r.OldBlocks = []Block{{ID: "base", SHA: "aaaa"}}
			r.NewBlocks = []Block{{ID: "base", SHA: "bbbb"}, {ID: "concise", SHA: "cccc"}}
			got := r.Markdown()
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("markdown missing %q in:\n%s", want, got)
				}
			}
		})
	}
}

func TestReportMarkdownPricedAndUnpriced(t *testing.T) {
	tests := []struct {
		name     string
		price    *Price
		usage    *Usage
		want     []string
		wantNone []string
	}{
		{
			name:  "bench.yaml price is stated with its date",
			price: &Price{Checked: "2026-09-26", Input: 2, CacheRead: 0.2, CacheWrite: 2.5, Output: 12},
			usage: &Usage{Input: 100, Output: 10, CostUSD: 0.0003, CostKnown: true},
			want: []string{
				"Cost at list price checked 2026-09-26: $2.00 input, $0.20 cache read, $2.50 cache write, $12.00 output per 1M tokens.",
				"| 1 | 1/1 hits, 0 extra · $0.0003 |",
				"| **median cost** | $0.0003 |",
			},
		},
		{
			name:     "unknown cost shows tokens only",
			usage:    &Usage{Input: 100, Output: 10},
			want:     []string{"| 1 | 1/1 hits, 0 extra |", "usage: input 100, cache read 0, cache write 0, output 10, cost unknown"},
			wantNone: []string{"Cost at", "cost / success", "median cost", "$0.0000"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Report{Skill: "ds-x", Command: "c", Baseline: true, NewSHA: "2", Groups: []HarnessReport{{
				Harness: "OpenAI Codex", Model: "gpt-5.6-terra", Price: tt.price,
				Scenarios: []ScenarioReport{{Name: "s", Expectations: 1, New: []RunReport{{Checked: true, Hits: 1, Usage: tt.usage}}}},
			}}}
			got := r.Markdown()
			for _, w := range tt.want {
				if !strings.Contains(got, w) {
					t.Errorf("markdown missing %q:\n%s", w, got)
				}
			}
			for _, w := range tt.wantNone {
				if strings.Contains(got, w) {
					t.Errorf("markdown has %q:\n%s", w, got)
				}
			}
		})
	}
}
