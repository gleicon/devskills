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

func TestChooseConfig(t *testing.T) {
	avail := []string{"ds-git-mode", "ds-step-mode"}
	tests := []struct {
		name       string
		o          configOpts
		want       []string
		wantPrompt bool
		wantErr    bool
	}{
		{name: "modes flag wins over tty", o: configOpts{changed: true, modeCSV: "ds-git-mode", tty: true}, want: []string{"ds-git-mode"}},
		{name: "dedupes", o: configOpts{changed: true, modeCSV: "ds-git-mode, ds-git-mode"}, want: []string{"ds-git-mode"}},
		{name: "empty clears", o: configOpts{changed: true, modeCSV: ""}},
		{name: "uninstalled mode errors", o: configOpts{changed: true, modeCSV: "ds-tdd-mode"}, wantErr: true},
		{name: "tty prompts", o: configOpts{tty: true}, wantPrompt: true},
		{name: "piped without modes errors", o: configOpts{}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, prompt, err := chooseConfig(tt.o, avail)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if prompt != tt.wantPrompt {
				t.Errorf("prompt = %v, want %v", prompt, tt.wantPrompt)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("modes = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCatalogModes(t *testing.T) {
	catalog := fstest.MapFS{
		"skills/ds-git-mode/SKILL.md":   {Data: []byte("x")},
		"skills/ds-step-mode/SKILL.md":  {Data: []byte("x")},
		"skills/ds-go-review/SKILL.md":  {Data: []byte("x")},
		"skills/ds-quality-gate/README": {Data: []byte("x")},
	}
	got, err := catalogModes(catalog)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"ds-git-mode", "ds-step-mode"}; !slices.Equal(got, want) {
		t.Errorf("modes = %v, want %v", got, want)
	}
}

func TestConfiguredModes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.md")
	if got, err := configuredModes(path); err != nil || got != nil {
		t.Fatalf("absent file: modes = %v, err = %v; want nil, nil", got, err)
	}
	// A hand-added mode outside our block still counts, so the prompt doesn't
	// silently drop it.
	body := "# Project config\n\n- ds-tdd-mode\n\n<!-- BEGIN devskills:modes -->\n## Modes\n\n- ds-git-mode\n- ds-git-mode\n<!-- END devskills:modes -->\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := configuredModes(path)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"ds-tdd-mode", "ds-git-mode"}; !slices.Equal(got, want) {
		t.Errorf("modes = %v, want %v", got, want)
	}
}

func TestRunConfigWritesBlockAndPreservesUserContent(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".project", "config.md")

	if err := runConfig(io.Discard, path, []string{"ds-git-mode", "ds-step-mode"}, false); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, path)
	for _, want := range []string{"<!-- BEGIN devskills:modes -->", "## Modes", "- ds-git-mode", "- ds-step-mode"} {
		if !strings.Contains(got, want) {
			t.Errorf("config.md missing %q:\n%s", want, got)
		}
	}

	// Anything the user adds outside the block survives a re-run, which is what
	// lets config.md grow beyond modes.
	if err := os.WriteFile(path, []byte(got+"\n## My own section\n\nkeep me\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runConfig(io.Discard, path, []string{"ds-step-mode"}, false); err != nil {
		t.Fatal(err)
	}
	got = readFile(t, path)
	if !strings.Contains(got, "keep me") {
		t.Errorf("user content dropped:\n%s", got)
	}
	if strings.Contains(got, "- ds-git-mode") {
		t.Errorf("deselected mode still listed:\n%s", got)
	}
}

func TestRunConfigDryRunWritesNothing(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".project", "config.md")
	if err := runConfig(io.Discard, path, []string{"ds-git-mode"}, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("dry run created %s (err = %v)", path, err)
	}
}
