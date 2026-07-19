package cli

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/gleicon/devskills/internal/harness"
	"github.com/gleicon/devskills/internal/scaffold"
)

// config.md is the one file in .project/ that no skill may write: it holds the
// user's standing instructions to the assistant (the modes /ds-project-resume
// applies at session start), so the assistant editing it could disarm its own
// leash. The CLI owns it, and manages only its own marker block — anything else
// the user puts in the file is left alone.

type configOpts struct {
	modeCSV string
	tty     bool
	changed bool // --modes explicitly set
}

func newConfigCmd(catalog fs.FS) *cobra.Command {
	var (
		modeCSV string
		dryRun  bool
	)
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Set the modes applied at session start (.project/config.md)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			r, err := harness.NewResolver(nil)
			if err != nil {
				return err
			}
			avail, err := catalogModes(catalog)
			if err != nil {
				return err
			}
			path := filepath.Join(r.ProjectRoot, ".project", "config.md")
			current, err := configuredModes(path)
			if err != nil {
				return err
			}
			o := configOpts{modeCSV: modeCSV, tty: isTTY(), changed: cmd.Flags().Changed("modes")}
			modes, needPrompt, err := chooseConfig(o, avail)
			if err != nil {
				return err
			}
			if needPrompt {
				if modes, err = promptConfig(avail, current); err != nil {
					return err
				}
			}
			return runConfig(cmd.OutOrStdout(), path, modes, dryRun)
		},
	}
	f := cmd.Flags()
	f.StringVar(&modeCSV, "modes", "", "comma-separated mode skills (e.g. ds-git-mode,ds-step-mode)")
	f.BoolVar(&dryRun, "dry-run", false, "print what would change without writing")
	return cmd
}

// chooseConfig resolves the selection with no I/O. --modes wins and skips the
// prompt; a TTY without it prompts. Piped with no --modes is an error rather
// than a silent default: there is no sensible default set of modes.
func chooseConfig(o configOpts, avail []string) ([]string, bool, error) {
	if o.changed {
		modes, err := parseModeCSV(o.modeCSV, avail)
		return modes, false, err
	}
	if o.tty {
		return nil, true, nil
	}
	return nil, false, errors.New("no terminal to prompt on: pass --modes")
}

// parseModeCSV validates, trims, and dedupes the requested modes. Empty is valid
// — it clears the list, which is how you turn every mode off.
func parseModeCSV(csv string, avail []string) ([]string, error) {
	var modes []string
	for tok := range strings.SplitSeq(csv, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		if !slices.Contains(avail, tok) {
			return nil, fmt.Errorf("unknown or uninstalled mode %q (available: %s)", tok, strings.Join(avail, ", "))
		}
		if !slices.Contains(modes, tok) {
			modes = append(modes, tok)
		}
	}
	return modes, nil
}

func promptConfig(avail, current []string) ([]string, error) {
	opts := make([]huh.Option[string], len(avail))
	for i, m := range avail {
		opts[i] = huh.NewOption(m, m).Selected(slices.Contains(current, m))
	}
	modes := slices.Clone(current)
	form := huh.NewForm(huh.NewGroup(
		huh.NewMultiSelect[string]().
			Title("Modes to apply at session start").
			Height(len(opts)+1).
			Options(opts...).
			Value(&modes),
	)).WithTheme(huh.ThemeFunc(formTheme))
	if err := form.Run(); err != nil {
		return nil, err
	}
	return modes, nil
}

func runConfig(out io.Writer, path string, modes []string, dryRun bool) error {
	if !dryRun {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
	}
	e := scaffold.New(dryRun, func(s string) { lipgloss.Fprintf(out, "  %s\n", s) })
	if err := e.Upsert(path, "modes", modesBlock(modes)); err != nil {
		return err
	}
	if len(modes) == 0 {
		lipgloss.Fprintln(out, "  no modes configured; resume will apply none.")
		return nil
	}
	lipgloss.Fprintf(out, "  %s\n", strings.Join(modes, ", "))
	return nil
}

func modesBlock(modes []string) string {
	var b strings.Builder
	b.WriteString("## Modes\n\n")
	for _, m := range modes {
		fmt.Fprintf(&b, "- %s\n", m)
	}
	return b.String()
}

var modeBullet = regexp.MustCompile(`(?m)^-[ \t]+(ds-[a-z0-9-]+-mode)[ \t]*$`)

// configuredModes reads the modes a project already lists. It scans the whole
// file rather than just our block, so a hand-added mode is pre-selected rather
// than appearing unset. One this binary doesn't ship has no option to select and
// won't survive in the block — only outside it, which Upsert leaves alone.
func configuredModes(path string) ([]string, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var modes []string
	for _, m := range modeBullet.FindAllStringSubmatch(string(b), -1) {
		if !slices.Contains(modes, m[1]) {
			modes = append(modes, m[1])
		}
	}
	return modes, nil
}

// catalogModes lists the mode skills this binary ships. Reading the embedded
// catalog rather than the install dirs keeps the offer correct regardless of
// which assistant reads config.md, which config dir the shell points at, and
// what an older install left behind on disk.
func catalogModes(catalog fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(catalog, "skills")
	if err != nil {
		return nil, fmt.Errorf("read skill catalog: %w", err)
	}
	var modes []string
	for _, d := range entries {
		if d.IsDir() && strings.HasSuffix(d.Name(), "-mode") {
			modes = append(modes, d.Name())
		}
	}
	return modes, nil
}
