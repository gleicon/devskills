package cli

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"slices"
	"strings"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"

	"github.com/gleicon/devskills/internal/harness"
	dsync "github.com/gleicon/devskills/internal/sync"
)

// installableHarnesses is the set install offers. Antigravity is added with its
// agy plugin path in a later step; today it has no directory to copy into.
var installableHarnesses = []harness.ID{harness.Claude, harness.OpenCode, harness.Codex}

type installOpts struct {
	local      bool
	all        bool
	harnessCSV string
	yes        bool
	tty        bool
}

func newInstallCmd(catalog fs.FS) *cobra.Command {
	var (
		local, all, yes, dryRun          bool
		harnessCSV                       string
		claudeDir, codexDir, opencodeDir string
	)
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install devskills' skills into your AI coding assistants",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			scope := harness.Global
			if local {
				scope = harness.Local
			}
			overrides := map[harness.ID]string{}
			for id, dir := range map[harness.ID]string{harness.Claude: claudeDir, harness.Codex: codexDir, harness.OpenCode: opencodeDir} {
				if dir != "" {
					overrides[id] = dir
				}
			}
			r, err := harness.NewResolver(overrides)
			if err != nil {
				return err
			}

			opts := installOpts{local: local, all: all, harnessCSV: harnessCSV, yes: yes, tty: isTTY()}
			ids, needPrompt, err := chooseHarnesses(opts, installableHarnesses, detectedHarnesses(r, scope))
			if err != nil {
				return err
			}
			if needPrompt {
				ids, err = promptHarnesses(installableHarnesses, detectedHarnesses(r, scope))
				if err != nil {
					return err
				}
			}
			if len(ids) == 0 {
				cmd.Println("No assistants selected — nothing to do.")
				return nil
			}
			return runInstall(cmd.OutOrStdout(), catalog, r, scope, ids, dryRun)
		},
	}
	f := cmd.Flags()
	f.BoolVar(&local, "local", false, "install into the current repo instead of globally")
	f.BoolVar(&all, "all", false, "install for every supported assistant")
	f.StringVar(&harnessCSV, "harness", "", "comma-separated assistants (claude,opencode,codex)")
	f.BoolVarP(&yes, "yes", "y", false, "accept detected defaults without prompting")
	f.BoolVar(&dryRun, "dry-run", false, "print the plan without changing anything")
	f.StringVar(&claudeDir, "claude-dir", "", "override Claude config dir (global only)")
	f.StringVar(&codexDir, "codex-dir", "", "override Codex config dir (global only)")
	f.StringVar(&opencodeDir, "opencode-dir", "", "override OpenCode config dir (global only)")
	return cmd
}

// chooseHarnesses resolves the target set with no I/O or prompting. It returns
// needPrompt=true when the caller should fall back to the interactive
// multi-select (a TTY with no explicit selection). Precedence: --all, then
// --harness, then detection. Detection is meaningful only for global; a
// non-interactive local run with no explicit selection is an error rather than
// a silent install-for-all.
func chooseHarnesses(o installOpts, universe, detected []harness.ID) (ids []harness.ID, needPrompt bool, err error) {
	if o.all {
		return universe, false, nil
	}
	if o.harnessCSV != "" {
		ids, err := parseHarnessCSV(o.harnessCSV, universe)
		return ids, false, err
	}
	if o.tty && !o.yes {
		return nil, true, nil
	}
	if o.local {
		return nil, false, errors.New("local install needs --harness or --all (there is no way to detect intent for a repo)")
	}
	return detected, false, nil
}

func parseHarnessCSV(csv string, universe []harness.ID) ([]harness.ID, error) {
	var ids []harness.ID
	for _, tok := range strings.Split(csv, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		id := harness.ID(tok)
		if !slices.Contains(universe, id) {
			return nil, fmt.Errorf("unknown or unsupported harness %q (supported: %s)", tok, joinIDs(universe))
		}
		if !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return nil, errors.New("--harness listed no assistants")
	}
	return ids, nil
}

func detectedHarnesses(r harness.Resolver, scope harness.Scope) []harness.ID {
	if scope == harness.Local {
		return nil // local has no detection signal
	}
	var out []harness.ID
	for _, id := range installableHarnesses {
		if r.Detected(id) {
			out = append(out, id)
		}
	}
	return out
}

func promptHarnesses(universe, checked []harness.ID) ([]harness.ID, error) {
	opts := make([]huh.Option[harness.ID], len(universe))
	for i, id := range universe {
		opts[i] = huh.NewOption(id.Name(), id).Selected(slices.Contains(checked, id))
	}
	var selected []harness.ID
	form := huh.NewForm(huh.NewGroup(
		huh.NewMultiSelect[harness.ID]().
			Title("Install devskills for which assistants?").
			Options(opts...).
			Value(&selected),
	))
	if err := form.Run(); err != nil {
		return nil, err
	}
	return selected, nil
}

func runInstall(out io.Writer, catalog fs.FS, r harness.Resolver, scope harness.Scope, ids []harness.ID, dryRun bool) error {
	engine := dsync.New(catalog)
	for _, id := range ids {
		t, err := buildTarget(r, id, scope)
		if err != nil {
			return err
		}
		plan, err := engine.Plan(t)
		if err != nil {
			return err
		}
		renderPlan(out, plan, scope, dryRun)
		if dryRun {
			continue
		}
		if err := engine.Apply(plan); err != nil {
			return fmt.Errorf("%s: %w", id.Name(), err)
		}
	}
	return nil
}

// buildTarget wires a harness+scope into a sync target: the skills dir, plus the
// legacy purge dir (global only) and the Codex sidecar flag.
func buildTarget(r harness.Resolver, id harness.ID, scope harness.Scope) (dsync.Target, error) {
	skillsDir, err := r.SkillsDir(id, scope)
	if err != nil {
		return dsync.Target{}, err
	}
	t := dsync.Target{Name: id.Name(), SkillsDir: skillsDir, Codex: id == harness.Codex}
	if scope == harness.Global {
		if legacy, ok := r.LegacyCommandDir(id); ok {
			t.LegacyDir = legacy
		}
	}
	return t, nil
}

func renderPlan(out io.Writer, p dsync.Plan, scope harness.Scope, dryRun bool) {
	header := lipgloss.NewStyle().Bold(true)
	dim := lipgloss.NewStyle().Faint(true)
	fmt.Fprintln(out, header.Render(fmt.Sprintf("%s (%s) → %s", p.Target.Name, scopeName(scope), p.Target.SkillsDir)))
	fmt.Fprintf(out, "  %d skills to write/update\n", len(p.Writes))
	for _, rm := range p.Removes {
		fmt.Fprintf(out, "  remove %s: %s\n", rm.Kind, rm.Path)
	}
	if dryRun {
		fmt.Fprintln(out, dim.Render("  dry run — nothing written"))
	}
	fmt.Fprintln(out)
}

func scopeName(s harness.Scope) string {
	if s == harness.Local {
		return "local"
	}
	return "global"
}

func joinIDs(ids []harness.ID) string {
	s := make([]string, len(ids))
	for i, id := range ids {
		s[i] = string(id)
	}
	return strings.Join(s, ", ")
}

func isTTY() bool { return term.IsTerminal(os.Stdout.Fd()) }
