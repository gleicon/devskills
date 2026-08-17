package bench

import (
	"fmt"
	"strings"
)

// Report is a completed bench invocation, renderable as PR-ready markdown
// (FR-12). It carries everything needed to re-run identically (NFR-3) and
// never a verdict on old vs new (FR-13) — interpretation belongs to the
// author and reviewer.
type Report struct {
	Skill    string
	Command  string // exact reproduction command
	Baseline bool
	OldSHA   string // empty in baseline mode
	NewSHA   string
	Groups   []HarnessReport
}

// HarnessReport is one harness's runs. Reports never compare across
// harnesses; each group stands alone.
type HarnessReport struct {
	Harness   string // display name
	Model     string // pinned model ID actually used
	Scenarios []ScenarioReport
}

// ScenarioReport is one scenario's old and new run sets on one harness.
type ScenarioReport struct {
	Name         string
	Expectations int         // planted-defect expectation count; 0 for other tiers
	Old          []RunReport // empty in baseline mode
	New          []RunReport
}

// RunReport is one harness invocation's outcome.
type RunReport struct {
	Failed  bool
	FailMsg string
	Checked bool // planted-defect scoring ran
	Hits    int
	Extras  int
	Stdout  string
	Stderr  string
	Diff    string
}

// Markdown renders the report in the pr-md format: per-harness per-run hit
// tables, aggregates, model IDs, the reproduction command, and raw
// transcripts collapsed in <details>.
func (r Report) Markdown() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Bench report: %s\n\n", r.Skill)
	fmt.Fprintf(&b, "- Reproduce: `%s`\n", r.Command)
	if r.Baseline {
		fmt.Fprintf(&b, "- Baseline mode: skill absent on the main branch; new version only, `%s` (working tree)\n", r.NewSHA)
	} else {
		fmt.Fprintf(&b, "- Versions: old `%s` (main branch), new `%s` (working tree)\n", r.OldSHA, r.NewSHA)
	}
	for _, g := range r.Groups {
		fmt.Fprintf(&b, "\n## %s — model `%s`\n", g.Harness, g.Model)
		for _, s := range g.Scenarios {
			s.render(&b, r.Baseline)
		}
	}
	return b.String()
}

func (s ScenarioReport) render(b *strings.Builder, baseline bool) {
	fmt.Fprintf(b, "\n### %s\n\n", s.Name)
	if baseline {
		fmt.Fprintf(b, "| run | new |\n|---|---|\n")
		for i, run := range s.New {
			fmt.Fprintf(b, "| %d | %s |\n", i+1, run.cell(s.Expectations))
		}
		fmt.Fprintf(b, "| **aggregate** | %s |\n", aggregateCell(s.New, s.Expectations))
	} else {
		fmt.Fprintf(b, "| run | old | new |\n|---|---|---|\n")
		for i := range max(len(s.Old), len(s.New)) {
			fmt.Fprintf(b, "| %d | %s | %s |\n", i+1, runCell(s.Old, i, s.Expectations), runCell(s.New, i, s.Expectations))
		}
		fmt.Fprintf(b, "| **aggregate** | %s | %s |\n",
			aggregateCell(s.Old, s.Expectations), aggregateCell(s.New, s.Expectations))
	}

	fmt.Fprintf(b, "\n<details>\n<summary>%s transcripts</summary>\n", s.Name)
	for i, run := range s.Old {
		run.renderTranscript(b, fmt.Sprintf("%s run %d", LabelOld, i+1))
	}
	for i, run := range s.New {
		run.renderTranscript(b, fmt.Sprintf("%s run %d", LabelNew, i+1))
	}
	fmt.Fprintf(b, "\n</details>\n")
}

func runCell(runs []RunReport, i, expectations int) string {
	if i >= len(runs) {
		return "—"
	}
	return runs[i].cell(expectations)
}

func (r RunReport) cell(expectations int) string {
	switch {
	case r.Failed:
		return "failed"
	case r.Checked:
		return fmt.Sprintf("%d/%d hits, %d extra", r.Hits, expectations, r.Extras)
	default:
		return "ok"
	}
}

// aggregateCell sums a run set: hit totals for checked runs, otherwise a
// success count. Failed runs contribute zero hits but stay in the denominator
// — a failure is never silently dropped (FR-9).
func aggregateCell(runs []RunReport, expectations int) string {
	if len(runs) == 0 {
		return "—"
	}
	hits, extras, okCount, checked := 0, 0, 0, false
	for _, r := range runs {
		if r.Failed {
			continue
		}
		okCount++
		if r.Checked {
			checked = true
			hits += r.Hits
			extras += r.Extras
		}
	}
	if checked {
		return fmt.Sprintf("%d/%d hits, %d extra", hits, expectations*len(runs), extras)
	}
	return fmt.Sprintf("%d/%d ok", okCount, len(runs))
}

func (r RunReport) renderTranscript(b *strings.Builder, label string) {
	fmt.Fprintf(b, "\n#### %s\n", label)
	if r.Failed {
		fmt.Fprintf(b, "\nrun failed: %s\n", r.FailMsg)
	}
	for _, sec := range []struct{ name, body string }{
		{"stdout", r.Stdout}, {"stderr", r.Stderr}, {"diff", r.Diff},
	} {
		if sec.body == "" {
			continue
		}
		// Four-backtick fences survive transcript bodies that contain ``` blocks.
		fmt.Fprintf(b, "\n%s:\n\n````\n%s\n````\n", sec.name, strings.TrimRight(sec.body, "\n"))
	}
}
