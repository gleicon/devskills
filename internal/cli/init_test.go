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

// fakeAssets is a minimal embedded tree: the three system profiles plus two
// languages, enough to drive init without the real content.
func fakeAssets() fstest.MapFS {
	return fstest.MapFS{
		"agents-md/system/agents-base.md":  {Data: []byte("BASE PRINCIPLES\n")},
		"agents-md/system/concise.md":      {Data: []byte("BE TERSE\n")},
		"agents-md/system/phase-hints.md":  {Data: []byte("PHASES\n")},
		"agents-md/language/go.md":         {Data: []byte("GO PROFILE\n")},
		"agents-md/language/typescript.md": {Data: []byte("TS PROFILE\n")},
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
			if !slices.Equal(sel.langs, tt.wantSel.langs) || sel.concise != tt.wantSel.concise || sel.phases != tt.wantSel.phases {
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
	sel := initSelection{langs: []string{"go"}, concise: true}
	if err := runInit(io.Discard, fakeAssets(), root, sel, false); err != nil {
		t.Fatal(err)
	}
	agents := readFile(t, filepath.Join(root, "AGENTS.md"))
	for _, want := range []string{
		"<!-- BEGIN devskills:base -->", "BASE PRINCIPLES",
		"<!-- BEGIN devskills:concise -->", "BE TERSE",
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
