package cli

import (
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/gleicon/devskills/internal/doctor"
)

// doctorEnv injects system access so doctor is testable without real detection or
// installs.
type doctorEnv struct {
	lookPath func(string) (string, error)
	goos     string
	run      func(name string, args ...string) error
}

func newDoctorCmd() *cobra.Command {
	var fix, dryRun bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check (with --fix, install) the external tools some skills need",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			env := doctorEnv{
				lookPath: exec.LookPath,
				goos:     runtime.GOOS,
				run: func(name string, args ...string) error {
					c := exec.CommandContext(cmd.Context(), name, args...)
					c.Stdout, c.Stderr = cmd.OutOrStdout(), cmd.ErrOrStderr()
					return c.Run()
				},
			}
			return runDoctor(cmd.OutOrStdout(), env, fix, dryRun)
		},
	}
	f := cmd.Flags()
	f.BoolVar(&fix, "fix", false, "install the missing tools")
	f.BoolVar(&dryRun, "dry-run", false, "with --fix, print the install commands without running them")
	return cmd
}

// runDoctor reports each external tool's status and, with fix, installs the
// missing ones. Check-only never fails; --fix returns an error only if an install
// command failed (a tool with no applicable installer is reported, not an error).
func runDoctor(out io.Writer, env doctorEnv, fix, dryRun bool) error {
	hasBinary := func(b string) bool { _, err := env.lookPath(b); return err == nil }
	var failed []string

	for _, t := range doctor.Tools {
		if path, err := env.lookPath(t.Name); err == nil {
			lipgloss.Fprintf(out, "ok       %-12s (/%s) → %s\n", t.Name, t.Skill, path)
			continue
		}
		in, ok := t.Installer(env.goos, hasBinary)
		if !ok {
			lipgloss.Fprintf(out, "missing  %-12s (/%s) → no installer for this platform; see %s\n", t.Name, t.Skill, t.DocURL)
			continue
		}
		if !fix {
			lipgloss.Fprintf(out, "missing  %-12s (/%s) → %s   [devskills doctor --fix]\n", t.Name, t.Skill, strings.Join(in.Command, " "))
			continue
		}
		if dryRun {
			lipgloss.Fprintf(out, "would    %-12s → %s\n", t.Name, strings.Join(in.Command, " "))
			continue
		}
		lipgloss.Fprintf(out, "install  %-12s → %s\n", t.Name, strings.Join(in.Command, " "))
		if err := env.run(in.Command[0], in.Command[1:]...); err != nil {
			lipgloss.Fprintf(out, "failed   %-12s → %v (see %s)\n", t.Name, err, t.DocURL)
			failed = append(failed, t.Name)
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("install failed: %s", strings.Join(failed, ", "))
	}
	return nil
}
