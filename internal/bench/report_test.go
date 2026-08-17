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
					{Checked: true, Hits: 1, Extras: 0, Stdout: "removed one comment\n", Diff: "diff --git a/greet.go b/greet.go\n-// First we get the greeting\n"},
					{Failed: true, FailMsg: "timed out after 5m0s", Stderr: "signal: killed\n"},
				},
				New: []RunReport{
					{Checked: true, Hits: 3, Extras: 1, Stdout: "cleaned all three\nplus a ```code``` fence\n"},
					{Checked: true, Hits: 2, Extras: 0, Stdout: "cleaned two\n"},
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
		"| 2 | failed | 2/3 hits, 0 extra |",
		"| **aggregate** | 1/6 hits, 0 extra | 5/6 hits, 1 extra |",
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
