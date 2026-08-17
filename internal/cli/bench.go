package cli

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

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
	return cmd
}

// runBench benches the old (main-branch) and new (working-tree) versions of a
// skill against its scenarios on Claude Code, K runs each, printing every
// run's raw output and post-run diff. With the skill absent on main it runs
// baseline mode: new version only. Failed runs are reported inline (FR-9);
// the command exits non-zero only when every run failed.
func runBench(ctx context.Context, out io.Writer, root string, opts benchOptions) error {
	if opts.Runs < 1 {
		return fmt.Errorf("--runs must be at least 1, got %d", opts.Runs)
	}
	versions, err := bench.LoadVersions(root, opts.Skill)
	if err != nil {
		return err
	}
	if len(versions) == 1 {
		fmt.Fprintf(out, "baseline mode: %s is absent on the main branch, benching the new version only\n", opts.Skill)
	}
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

	runner := bench.Runner{Harness: harness.Claude, Model: model}
	total, failures := 0, 0
	for _, s := range scenarios {
		for _, v := range versions {
			for i := 1; i <= opts.Runs; i++ {
				total++
				fmt.Fprintf(out, "== %s/%s %s run %d/%d (%s, model %s)\n",
					opts.Skill, s.Name, v.Label, i, opts.Runs, harness.Claude.Name(), model)
				res, err := runner.Run(ctx, s, v)
				if err != nil {
					return err
				}
				if res.Err != nil {
					failures++
					fmt.Fprintf(out, "run failed: %v\n", res.Err)
				}
				for _, sec := range []struct{ name, body string }{
					{"stdout", res.Stdout}, {"stderr", res.Stderr}, {"diff", res.Diff},
				} {
					fmt.Fprintf(out, "-- %s --\n%s\n", sec.name, sec.body)
				}
			}
		}
	}
	if failures == total {
		return fmt.Errorf("all %d runs failed", failures)
	}
	return nil
}
