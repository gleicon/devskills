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
	Tier         string
	Expectations int         // per-run hit denominator: planted defects or required elements; 0 for smoke
	Old          []RunReport // empty in baseline mode
	New          []RunReport
}

// RunReport is one harness invocation's outcome.
type RunReport struct {
	Failed  bool
	FailMsg string
	Checked bool // tier scoring ran
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
			fmt.Fprintf(b, "| %d | %s |\n", i+1, s.cell(run))
		}
		fmt.Fprintf(b, "| **aggregate** | %s |\n", s.aggregate(s.New))
	} else {
		fmt.Fprintf(b, "| run | old | new |\n|---|---|---|\n")
		for i := range max(len(s.Old), len(s.New)) {
			fmt.Fprintf(b, "| %d | %s | %s |\n", i+1, s.runCell(s.Old, i), s.runCell(s.New, i))
		}
		fmt.Fprintf(b, "| **aggregate** | %s | %s |\n", s.aggregate(s.Old), s.aggregate(s.New))
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

func (s ScenarioReport) runCell(runs []RunReport, i int) string {
	if i >= len(runs) {
		return "—"
	}
	return s.cell(runs[i])
}

func (s ScenarioReport) cell(r RunReport) string {
	switch {
	case r.Failed:
		return "failed"
	case !r.Checked:
		return "ok"
	}
	switch s.Tier {
	case TierStructural:
		return fmt.Sprintf("%d/%d elements", r.Hits, s.Expectations)
	case TierSmoke:
		if r.Hits > 0 {
			return "ok"
		}
		return "no output"
	default:
		return fmt.Sprintf("%d/%d hits, %d extra", r.Hits, s.Expectations, r.Extras)
	}
}

// aggregate sums a run set: hit totals for checked runs, otherwise a success
// count. Failed runs contribute zero hits but stay in the denominator — a
// failure is never silently dropped (FR-9).
func (s ScenarioReport) aggregate(runs []RunReport) string {
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
	if !checked {
		return fmt.Sprintf("%d/%d ok", okCount, len(runs))
	}
	switch s.Tier {
	case TierStructural:
		return fmt.Sprintf("%d/%d elements", hits, s.Expectations*len(runs))
	case TierSmoke:
		// smoke Hits is 0 or 1 per run, so the sum counts the runs that
		// exited zero with output
		return fmt.Sprintf("%d/%d ok", hits, len(runs))
	default:
		return fmt.Sprintf("%d/%d hits, %d extra", hits, s.Expectations*len(runs), extras)
	}
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
