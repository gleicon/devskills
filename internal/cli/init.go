package cli

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/gleicon/devskills/internal/harness"
	"github.com/gleicon/devskills/internal/scaffold"
)

// retiredBlocks are managed-block ids devskills once wrote and no longer does;
// init sweeps them every run, the way the sync ledger sweeps retired skills. The
// bare "language" block predates the per-language language:<lang> ids.
var retiredBlocks = []string{"language"}

// layers are the optional AGENTS.md blocks init offers beside base. Adding one
// here is the whole change: flag, flagsChanged, prompt, and write all derive
// from this table.
var layers = []struct {
	id    string // flag name and managed-block id
	asset string // file under agents-md/
	help  string // flag usage line
	title string // prompt confirm title
}{
	{"concise", "system/concise.md", "add the terse-response directive", "Add the terse-response directive (concise)?"},
	{"interaction", "system/interaction.md", "add the question/handback interaction rules", "Add question/handback interaction rules (interaction)?"},
	{"plain-language", "system/plain-language.md", "add the plain-language writing rules", "Add plain-language writing rules (plain-language)?"},
	{"phases", "system/phase-hints.md", "add phase-aware suggestions", "Add phase-aware suggestions (phases)?"},
	{"spec-discipline", "system/spec-discipline.md", "add the SPEC/GRILL amendment-discipline rules", "Add SPEC/GRILL amendment-discipline rules (spec-discipline)?"},
}

type initSelection struct {
	langs  []string
	layers []string // selected layer ids
}

type initOpts struct {
	langCSV      string
	layers       []string // layer ids whose flags were set
	yes          bool
	tty          bool
	flagsChanged bool // --lang or any layer flag explicitly set
}

func newInitCmd(catalog fs.FS) *cobra.Command {
	var (
		langCSV                string
		yes, dryRun, uninstall bool
	)
	layerOn := make([]bool, len(layers))
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold AGENTS.md (and a CLAUDE.md import) for the current project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			r, err := harness.NewResolver(nil)
			if err != nil {
				return err
			}
			if uninstall {
				return runInitUninstall(cmd.OutOrStdout(), r.ProjectRoot, dryRun)
			}
			avail, err := availableLanguages(catalog)
			if err != nil {
				return err
			}
			o := initOpts{langCSV: langCSV, yes: yes, tty: isTTY(), flagsChanged: cmd.Flags().Changed("lang")}
			for i, l := range layers {
				if layerOn[i] {
					o.layers = append(o.layers, l.id)
				}
				o.flagsChanged = o.flagsChanged || cmd.Flags().Changed(l.id)
			}
			sel, needPrompt, err := chooseInit(o, avail)
			if err != nil {
				return err
			}
			if needPrompt {
				sel, err = promptInit(avail)
				if err != nil {
					return err
				}
			}
			return runInit(cmd.OutOrStdout(), catalog, r.ProjectRoot, sel, dryRun)
		},
	}
	f := cmd.Flags()
	f.StringVar(&langCSV, "lang", "", "comma-separated language profiles (e.g. go,typescript)")
	for i, l := range layers {
		f.BoolVar(&layerOn[i], l.id, false, l.help)
	}
	f.BoolVarP(&yes, "yes", "y", false, "accept defaults (base only) without prompting")
	f.BoolVar(&dryRun, "dry-run", false, "print what would change without writing")
	f.BoolVar(&uninstall, "uninstall", false, "remove devskills' blocks from AGENTS.md/CLAUDE.md")
	return cmd
}

// chooseInit resolves the block selection with no I/O. Explicit flags win and skip
// the prompt; a bare TTY with no flags prompts; anything else (piped, --yes) takes
// the base-only default.
func chooseInit(o initOpts, avail []string) (initSelection, bool, error) {
	if o.flagsChanged {
		langs, err := parseLangCSV(o.langCSV, avail)
		if err != nil {
			return initSelection{}, false, err
		}
		return initSelection{langs: langs, layers: o.layers}, false, nil
	}
	if o.tty && !o.yes {
		return initSelection{}, true, nil
	}
	return initSelection{}, false, nil
}

// parseLangCSV validates, trims, and dedupes the requested profiles. Empty is
// valid (base only) — unlike --harness, a language is optional.
func parseLangCSV(csv string, avail []string) ([]string, error) {
	var langs []string
	for tok := range strings.SplitSeq(csv, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		if !slices.Contains(avail, tok) {
			return nil, fmt.Errorf("unknown language profile %q (available: %s)", tok, strings.Join(avail, ", "))
		}
		if !slices.Contains(langs, tok) {
			langs = append(langs, tok)
		}
	}
	return langs, nil
}

func promptInit(avail []string) (initSelection, error) {
	langOpts := make([]huh.Option[string], len(avail))
	for i, l := range avail {
		langOpts[i] = huh.NewOption(l, l)
	}
	var langs []string
	on := make([]bool, len(layers))
	fields := []huh.Field{
		huh.NewMultiSelect[string]().
			Title("Language profiles to include").
			Height(len(langOpts) + 1).
			Options(langOpts...).
			Value(&langs),
	}
	for i, l := range layers {
		fields = append(fields, huh.NewConfirm().Title(l.title).Value(&on[i]))
	}
	form := huh.NewForm(huh.NewGroup(fields...)).WithTheme(huh.ThemeFunc(formTheme))
	if err := form.Run(); err != nil {
		return initSelection{}, err
	}
	sel := initSelection{langs: langs}
	for i, l := range layers {
		if on[i] {
			sel.layers = append(sel.layers, l.id)
		}
	}
	return sel, nil
}

// runInit writes AGENTS.md's managed blocks (base + selected layers) and points
// CLAUDE.md at it. The retired bare "language" block is swept first so a bash-era
// project migrates to the per-language ids in place.
func runInit(out io.Writer, catalog fs.FS, root string, sel initSelection, dryRun bool) error {
	agentsPath := filepath.Join(root, "AGENTS.md")
	claudePath := filepath.Join(root, "CLAUDE.md")
	e := scaffold.New(dryRun, func(s string) { lipgloss.Fprintf(out, "  %s\n", s) })

	if err := e.RemoveBlocks(agentsPath, retiredBlocks...); err != nil {
		return err
	}

	add := func(id, rel string) error {
		body, err := readAsset(catalog, rel)
		if err != nil {
			return err
		}
		return e.Upsert(agentsPath, id, body)
	}
	if err := add("base", "system/agents-base.md"); err != nil {
		return err
	}
	// Iterate the table, not sel.layers, so block order stays canonical.
	for _, l := range layers {
		if slices.Contains(sel.layers, l.id) {
			if err := add(l.id, l.asset); err != nil {
				return err
			}
		}
	}
	for _, lang := range sel.langs {
		body, err := readAsset(catalog, "language/"+lang+".md")
		if err != nil {
			return err
		}
		note := fmt.Sprintf("<!-- profile: %s — managed by devskills; edits between these markers are overwritten -->", lang)
		if err := e.Upsert(agentsPath, "language:"+lang, note+"\n"+body); err != nil {
			return err
		}
	}
	return e.EnsureClaudeImport(claudePath)
}

// runInitUninstall backs devskills out of the project: every managed block in
// AGENTS.md and CLAUDE.md (each file removed only if nothing else remained), plus
// the pre-AGENTS.md profile file some old setups left behind.
func runInitUninstall(out io.Writer, root string, dryRun bool) error {
	e := scaffold.New(dryRun, func(s string) { lipgloss.Fprintf(out, "  %s\n", s) })
	for _, f := range []string{filepath.Join(root, "AGENTS.md"), filepath.Join(root, "CLAUDE.md")} {
		ids, err := e.BlockIDs(f)
		if err != nil {
			return err
		}
		if err := e.RemoveBlocks(f, ids...); err != nil {
			return err
		}
	}
	return removeLegacyProfile(out, root, dryRun)
}

// removeLegacyProfile sweeps the profile file older setups recorded in
// .devskills/language; the directory follows if it is then empty.
func removeLegacyProfile(out io.Writer, root string, dryRun bool) error {
	file := filepath.Join(root, ".devskills", "language")
	_, err := os.Stat(file)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if dryRun {
		lipgloss.Fprintln(out, "  [dry] would remove legacy .devskills/language")
		return nil
	}
	if err := os.Remove(file); err != nil {
		return err
	}
	os.Remove(filepath.Join(root, ".devskills")) // best effort: only succeeds if empty
	lipgloss.Fprintln(out, "  removed legacy .devskills/language")
	return nil
}

func readAsset(catalog fs.FS, rel string) (string, error) {
	b, err := fs.ReadFile(catalog, path.Join("agents-md", rel))
	if err != nil {
		return "", fmt.Errorf("read embedded %s: %w", rel, err)
	}
	return string(b), nil
}

// availableLanguages lists the language profiles the binary embeds, so the flag
// validation and prompt stay in sync with the shipped content.
func availableLanguages(catalog fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(catalog, "agents-md/language")
	if err != nil {
		return nil, fmt.Errorf("read language profiles: %w", err)
	}
	var langs []string
	for _, d := range entries {
		if name, ok := strings.CutSuffix(d.Name(), ".md"); ok {
			langs = append(langs, name)
		}
	}
	return langs, nil
}
