package cli

import (
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

// fakeAssets is a minimal embedded tree: the six system profiles plus two
// languages, enough to drive init without the real content.
func fakeAssets() fstest.MapFS {
	return fstest.MapFS{
		"agents-md/system/agents-base.md":     {Data: []byte("BASE PRINCIPLES\n")},
		"agents-md/system/concise.md":         {Data: []byte("BE TERSE\n")},
		"agents-md/system/interaction.md":     {Data: []byte("ONE QUESTION\n")},
		"agents-md/system/plain-language.md":  {Data: []byte("PLAIN WORDS\n")},
		"agents-md/system/phase-hints.md":     {Data: []byte("PHASES\n")},
		"agents-md/system/spec-discipline.md": {Data: []byte("AMEND INLINE\n")},
		"agents-md/language/go.md":            {Data: []byte("GO PROFILE\n")},
		"agents-md/language/typescript.md":    {Data: []byte("TS PROFILE\n")},
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestChooseInit(t *testing.T) {
	avail := []string{"go", "typescript"}
	tests := []struct {
		name       string
		o          initOpts
		wantSel    initSelection
		wantPrompt bool
		wantErr    bool
	}{
		{name: "lang flag wins", o: initOpts{flagsChanged: true, langCSV: "go", tty: true}, wantSel: initSelection{langs: []string{"go"}}},
		{name: "concise flag skips prompt", o: initOpts{flagsChanged: true, concise: true, tty: true}, wantSel: initSelection{concise: true}},
		{name: "interaction flag skips prompt", o: initOpts{flagsChanged: true, interaction: true, tty: true}, wantSel: initSelection{interaction: true}},
		{name: "plain-language flag skips prompt", o: initOpts{flagsChanged: true, plain: true, tty: true}, wantSel: initSelection{plain: true}},
		{name: "spec-discipline flag skips prompt", o: initOpts{flagsChanged: true, discipline: true, tty: true}, wantSel: initSelection{discipline: true}},
		{name: "unknown lang errors", o: initOpts{flagsChanged: true, langCSV: "cobol"}, wantErr: true},
		{name: "tty no flags prompts", o: initOpts{tty: true}, wantPrompt: true},
		{name: "yes takes base-only default", o: initOpts{tty: true, yes: true}, wantSel: initSelection{}},
		{name: "non-tty base only", o: initOpts{}, wantSel: initSelection{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sel, prompt, err := chooseInit(tt.o, avail)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if prompt != tt.wantPrompt {
				t.Errorf("prompt = %v, want %v", prompt, tt.wantPrompt)
			}
			if !slices.Equal(sel.langs, tt.wantSel.langs) || sel.concise != tt.wantSel.concise || sel.interaction != tt.wantSel.interaction || sel.plain != tt.wantSel.plain || sel.phases != tt.wantSel.phases || sel.discipline != tt.wantSel.discipline {
				t.Errorf("sel = %+v, want %+v", sel, tt.wantSel)
			}
		})
	}
}

func TestParseLangCSV(t *testing.T) {
	avail := []string{"go", "typescript"}
	got, err := parseLangCSV(" go , typescript ,go", avail)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, []string{"go", "typescript"}) {
		t.Errorf("got %v, want [go typescript]", got)
	}
	if got, err := parseLangCSV("", avail); err != nil || got != nil {
		t.Errorf("empty = (%v, %v), want (nil, nil)", got, err)
	}
	if _, err := parseLangCSV("cobol", avail); err == nil {
		t.Error("want error for unknown language")
	}
}

func TestAvailableLanguages(t *testing.T) {
	got, err := availableLanguages(fakeAssets())
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, []string{"go", "typescript"}) {
		t.Errorf("languages = %v, want [go typescript]", got)
	}
}

func TestRunInitWritesBlocks(t *testing.T) {
	root := t.TempDir()
	sel := initSelection{langs: []string{"go"}, concise: true, interaction: true, plain: true, discipline: true}
	if err := runInit(io.Discard, fakeAssets(), root, sel, false); err != nil {
		t.Fatal(err)
	}
	agents := readFile(t, filepath.Join(root, "AGENTS.md"))
	for _, want := range []string{
		"<!-- BEGIN devskills:base -->", "BASE PRINCIPLES",
		"<!-- BEGIN devskills:concise -->", "BE TERSE",
		"<!-- BEGIN devskills:interaction -->", "ONE QUESTION",
		"<!-- BEGIN devskills:plain-language -->", "PLAIN WORDS",
		"<!-- BEGIN devskills:spec-discipline -->", "AMEND INLINE",
		"<!-- BEGIN devskills:language:go -->", "GO PROFILE",
		"profile: go — managed by devskills",
	} {
		if !strings.Contains(agents, want) {
			t.Errorf("AGENTS.md missing %q\n%s", want, agents)
		}
	}
	if strings.Contains(agents, "devskills:phases") {
		t.Error("phases block present but not selected")
	}
	claude := readFile(t, filepath.Join(root, "CLAUDE.md"))
	if !strings.Contains(claude, "<!-- BEGIN devskills:import -->") || !strings.Contains(claude, "@AGENTS.md") {
		t.Errorf("CLAUDE.md missing import: %q", claude)
	}
}

func TestRunInitMigratesBareLanguageBlock(t *testing.T) {
	root := t.TempDir()
	// A bash-era project: a bare "language" block instead of language:<lang>.
	seed := "<!-- BEGIN devskills:base -->\nold base\n<!-- END devskills:base -->\n\n" +
		"<!-- BEGIN devskills:language -->\nold lang\n<!-- END devskills:language -->\n"
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runInit(io.Discard, fakeAssets(), root, initSelection{langs: []string{"go"}}, false); err != nil {
		t.Fatal(err)
	}
	agents := readFile(t, filepath.Join(root, "AGENTS.md"))
	if strings.Contains(agents, "<!-- BEGIN devskills:language -->") {
		t.Error("retired bare language block should be migrated away")
	}
	if !strings.Contains(agents, "<!-- BEGIN devskills:language:go -->") {
		t.Error("language:go block missing after migration")
	}
}

func TestRunInitDryRunWritesNothing(t *testing.T) {
	root := t.TempDir()
	sel := initSelection{langs: []string{"go"}, concise: true}
	if err := runInit(io.Discard, fakeAssets(), root, sel, true); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"AGENTS.md", "CLAUDE.md"} {
		if _, err := os.Stat(filepath.Join(root, name)); !os.IsNotExist(err) {
			t.Errorf("dry run wrote %s", name)
		}
	}
}

func TestRunInitUninstallRemovesOurFiles(t *testing.T) {
	root := t.TempDir()
	// Scaffold, then back out: both files held only devskills content.
	if err := runInit(io.Discard, fakeAssets(), root, initSelection{langs: []string{"go"}, concise: true}, false); err != nil {
		t.Fatal(err)
	}
	if err := runInitUninstall(io.Discard, root, false); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"AGENTS.md", "CLAUDE.md"} {
		if _, err := os.Stat(filepath.Join(root, name)); !os.IsNotExist(err) {
			t.Errorf("%s should be removed (held only devskills content)", name)
		}
	}
}

func TestRunInitUninstallKeepsUserContent(t *testing.T) {
	root := t.TempDir()
	agents := filepath.Join(root, "AGENTS.md")
	seed := "# My project rules\n\n<!-- BEGIN devskills:base -->\nours\n<!-- END devskills:base -->\n\n# Keep this\n"
	if err := os.WriteFile(agents, []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runInitUninstall(io.Discard, root, false); err != nil {
		t.Fatal(err)
	}
	want := "# My project rules\n\n# Keep this\n"
	if got := readFile(t, agents); got != want {
		t.Errorf("content = %q, want %q", got, want)
	}
}

func TestRunInitUninstallDryRun(t *testing.T) {
	root := t.TempDir()
	if err := runInit(io.Discard, fakeAssets(), root, initSelection{langs: []string{"go"}}, false); err != nil {
		t.Fatal(err)
	}
	before := readFile(t, filepath.Join(root, "AGENTS.md"))
	if err := runInitUninstall(io.Discard, root, true); err != nil {
		t.Fatal(err)
	}
	if after := readFile(t, filepath.Join(root, "AGENTS.md")); after != before {
		t.Error("dry-run uninstall mutated AGENTS.md")
	}
}

func TestRunInitUninstallSweepsLegacyProfile(t *testing.T) {
	root := t.TempDir()
	legacyDir := filepath.Join(root, ".devskills")
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacyDir, "language"), []byte("go"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runInitUninstall(io.Discard, root, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(legacyDir); !os.IsNotExist(err) {
		t.Error("empty .devskills dir should be swept after removing its only file")
	}
}
