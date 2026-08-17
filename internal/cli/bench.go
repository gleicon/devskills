package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gleicon/devskills/internal/bench"
	"github.com/gleicon/devskills/internal/harness"
)

// benchOptions are the resolved flags of one bench invocation.
type benchOptions struct {
	Skill    string
	Scenario string // empty runs all of the skill's scenarios
	Model    string // empty uses the pinned model
	Runs     int
	Format   string // "" streams raw runs; "pr-md" emits the markdown report
	Out      string // pr-md destination file; empty writes to stdout
}

func newBenchCmd() *cobra.Command {
	opts := benchOptions{}
	cmd := &cobra.Command{
		Use:   "bench <skill>",
		Short: "Benchmark a skill against its scenarios in evals/",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Skill = args[0]
			r, err := harness.NewResolver(nil)
			if err != nil {
				return err
			}
			return runBench(cmd.Context(), cmd.OutOrStdout(), r.ProjectRoot, opts)
		},
	}
	f := cmd.Flags()
	f.StringVar(&opts.Scenario, "scenario", "", "run a single scenario by name")
	f.StringVar(&opts.Model, "model", "", "override the pinned model from evals/bench.yaml")
	f.IntVar(&opts.Runs, "runs", 3, "runs per skill version per scenario")
	f.StringVar(&opts.Format, "format", "", `output format: "pr-md" for the PR-ready markdown report`)
	f.StringVar(&opts.Out, "out", "", "with --format pr-md, write the report to this file")
	return cmd
}

// runBench benches the old (main-branch) and new (working-tree) versions of a
// skill against its scenarios on Claude Code, K runs each. The default output
// streams each run raw; --format pr-md collects everything into a markdown
// report instead. With the skill absent on main it runs baseline mode: new
// version only. Failed runs are reported, never skipped (FR-9); the command
// exits non-zero only when every run failed.
func runBench(ctx context.Context, out io.Writer, root string, opts benchOptions) error {
	if opts.Runs < 1 {
		return fmt.Errorf("--runs must be at least 1, got %d", opts.Runs)
	}
	if opts.Format != "" && opts.Format != "pr-md" {
		return fmt.Errorf(`--format must be "pr-md", got %q`, opts.Format)
	}
	versions, err := bench.LoadVersions(root, opts.Skill)
	if err != nil {
		return err
	}
	baseline := len(versions) == 1
	cfg, err := bench.LoadConfig(filepath.Join(root, "evals", "bench.yaml"))
	if err != nil {
		return err
	}
	model := opts.Model
	if model == "" {
		if model, err = cfg.Model(harness.Claude); err != nil {
			return err
		}
	}

	var scenarios []*bench.Scenario
	if opts.Scenario != "" {
		s, err := bench.LoadScenario(filepath.Join(root, "evals", opts.Skill, opts.Scenario))
		if err != nil {
			return err
		}
		scenarios = []*bench.Scenario{s}
	} else if scenarios, err = bench.LoadScenarios(filepath.Join(root, "evals"), opts.Skill); err != nil {
		return err
	}

	// Streaming goes to out in raw mode; in pr-md mode out carries the report
	// (or progress lines when the report goes to --out).
	stream := out
	if opts.Format == "pr-md" && opts.Out == "" {
		stream = io.Discard
	}
	if baseline {
		fmt.Fprintf(stream, "baseline mode: %s is absent on the main branch, benching the new version only\n", opts.Skill)
	}

	runner := bench.Runner{Harness: harness.Claude, Model: model}
	group := bench.HarnessReport{Harness: harness.Claude.Name(), Model: model}
	total, failures := 0, 0
	for _, s := range scenarios {
		sr := bench.ScenarioReport{Name: s.Name, Expectations: len(s.Expectations)}
		for _, v := range versions {
			for i := 1; i <= opts.Runs; i++ {
				total++
				fmt.Fprintf(stream, "== %s/%s %s run %d/%d (%s, model %s)\n",
					opts.Skill, s.Name, v.Label, i, opts.Runs, harness.Claude.Name(), model)
				res, err := runner.Run(ctx, s, v)
				if err != nil {
					return err
				}
				rr, err := recordRun(stream, s, res, opts.Format == "")
				if err != nil {
					return err
				}
				if rr.Failed {
					failures++
				}
				if v.Label == bench.LabelOld {
					sr.Old = append(sr.Old, rr)
				} else {
					sr.New = append(sr.New, rr)
				}
			}
		}
		group.Scenarios = append(group.Scenarios, sr)
	}

	if opts.Format == "pr-md" {
		report := bench.Report{
			Skill:    opts.Skill,
			Command:  reproCommand(opts, model),
			Baseline: baseline,
			NewSHA:   versions[len(versions)-1].SHA,
			Groups:   []bench.HarnessReport{group},
		}
		if !baseline {
			report.OldSHA = versions[0].SHA
		}
		md := report.Markdown()
		if opts.Out != "" {
			if err := os.WriteFile(opts.Out, []byte(md), 0o644); err != nil {
				return err
			}
			fmt.Fprintf(out, "report written to %s\n", opts.Out)
		} else {
			fmt.Fprint(out, md)
		}
	}
	if failures == total {
		return fmt.Errorf("all %d runs failed", failures)
	}
	return nil
}

// recordRun converts one run into its report record, streaming the raw
// sections when verbose.
func recordRun(stream io.Writer, s *bench.Scenario, res bench.Result, verbose bool) (bench.RunReport, error) {
	rr := bench.RunReport{Stdout: res.Stdout, Stderr: res.Stderr, Diff: res.Diff}
	if res.Err != nil {
		rr.Failed = true
		rr.FailMsg = res.Err.Error()
		fmt.Fprintf(stream, "run failed: %v\n", res.Err)
	} else if s.Tier == bench.TierPlantedDefect {
		c, err := bench.Check(s, res)
		if err != nil {
			return bench.RunReport{}, err
		}
		rr.Checked = true
		rr.Hits = c.HitCount()
		rr.Extras = len(c.Extras)
		fmt.Fprintf(stream, "hits: %d/%d, extra findings: %d\n", rr.Hits, len(s.Expectations), rr.Extras)
	}
	if verbose {
		for _, sec := range []struct{ name, body string }{
			{"stdout", res.Stdout}, {"stderr", res.Stderr}, {"diff", res.Diff},
		} {
			fmt.Fprintf(stream, "-- %s --\n%s\n", sec.name, sec.body)
		}
	}
	return rr, nil
}

// reproCommand is the exact command line that re-runs this bench identically
// (NFR-3): the resolved model is always explicit, so a later pin change
// cannot alter a reproduction.
func reproCommand(opts benchOptions, model string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "devskills bench %s", opts.Skill)
	if opts.Scenario != "" {
		fmt.Fprintf(&b, " --scenario %s", opts.Scenario)
	}
	fmt.Fprintf(&b, " --runs %d --model %s --format pr-md", opts.Runs, model)
	return b.String()
}
