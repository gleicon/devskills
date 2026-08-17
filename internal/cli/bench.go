package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/gleicon/devskills/internal/bench"
	"github.com/gleicon/devskills/internal/harness"
)

func newBenchCmd() *cobra.Command {
	var scenario, model string
	cmd := &cobra.Command{
		Use:   "bench <skill>",
		Short: "Benchmark a skill against its scenarios in evals/",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := harness.NewResolver(nil)
			if err != nil {
				return err
			}
			return runBench(cmd.Context(), cmd.OutOrStdout(), r.ProjectRoot, args[0], scenario, model)
		},
	}
	f := cmd.Flags()
	f.StringVar(&scenario, "scenario", "", "run a single scenario by name")
	f.StringVar(&model, "model", "", "override the pinned model from evals/bench.yaml")
	return cmd
}

// runBench runs the working-tree version of a skill against its scenarios on
// Claude Code, printing each run's raw output and post-run diff. Failed runs
// are reported inline (FR-9); the command exits non-zero only when every run
// failed.
func runBench(ctx context.Context, out io.Writer, root, skill, scenarioName, modelOverride string) error {
	content, err := os.ReadFile(filepath.Join(root, "skills", skill, "SKILL.md"))
	if err != nil {
		return fmt.Errorf("skill %q not found in working tree: %w", skill, err)
	}
	cfg, err := bench.LoadConfig(filepath.Join(root, "evals", "bench.yaml"))
	if err != nil {
		return err
	}
	model := modelOverride
	if model == "" {
		if model, err = cfg.Model(harness.Claude); err != nil {
			return err
		}
	}

	var scenarios []*bench.Scenario
	if scenarioName != "" {
		s, err := bench.LoadScenario(filepath.Join(root, "evals", skill, scenarioName))
		if err != nil {
			return err
		}
		scenarios = []*bench.Scenario{s}
	} else if scenarios, err = bench.LoadScenarios(filepath.Join(root, "evals"), skill); err != nil {
		return err
	}

	runner := bench.Runner{Harness: harness.Claude, Model: model}
	version := bench.SkillVersion{Name: skill, Content: content}
	failures := 0
	for _, s := range scenarios {
		fmt.Fprintf(out, "== %s/%s (%s, model %s)\n", skill, s.Name, harness.Claude.Name(), model)
		res, err := runner.Run(ctx, s, version)
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
	if failures == len(scenarios) {
		return fmt.Errorf("all %d runs failed", failures)
	}
	return nil
}
