package cli

import (
	"errors"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"testing/fstest"

	"github.com/gleicon/devskills/internal/harness"
	dsync "github.com/gleicon/devskills/internal/sync"
)

var universe = []harness.ID{harness.Claude, harness.OpenCode, harness.Codex}

func TestChooseHarnesses(t *testing.T) {
	detected := []harness.ID{harness.Claude}
	tests := []struct {
		name       string
		o          installOpts
		detected   []harness.ID
		wantIDs    []harness.ID
		wantPrompt bool
		wantErr    bool
	}{
		{name: "all wins", o: installOpts{all: true}, wantIDs: universe},
		{name: "csv explicit", o: installOpts{harnessCSV: "claude,codex"}, wantIDs: []harness.ID{harness.Claude, harness.Codex}},
		{name: "csv unknown errors", o: installOpts{harnessCSV: "claude,bogus"}, wantErr: true},
		{name: "tty no selection prompts", o: installOpts{tty: true}, wantPrompt: true},
		{name: "global non-tty uses detected", o: installOpts{}, detected: detected, wantIDs: detected},
		{name: "global --yes uses detected", o: installOpts{yes: true, tty: true}, detected: detected, wantIDs: detected},
		{name: "global non-tty none detected", o: installOpts{}, wantIDs: nil},
		{name: "local non-tty errors", o: installOpts{local: true}, wantErr: true},
		{name: "local --yes errors", o: installOpts{local: true, yes: true, tty: true}, wantErr: true},
		{name: "local explicit works", o: installOpts{local: true, all: true}, wantIDs: universe},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids, prompt, err := chooseHarnesses(tt.o, universe, tt.detected)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if prompt != tt.wantPrompt {
				t.Errorf("prompt = %v, want %v", prompt, tt.wantPrompt)
			}
			if !slices.Equal(ids, tt.wantIDs) {
				t.Errorf("ids = %v, want %v", ids, tt.wantIDs)
			}
		})
	}
}

func TestParseHarnessCSVDedupAndTrim(t *testing.T) {
	ids, err := parseHarnessCSV(" claude , codex ,claude", universe)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(ids, []harness.ID{harness.Claude, harness.Codex}) {
		t.Errorf("ids = %v, want [claude codex]", ids)
	}
	if _, err := parseHarnessCSV(" , ", universe); err == nil {
		t.Error("want error for empty selection")
	}
}

func TestBuildTarget(t *testing.T) {
	// Home/ProjectRoot use OS-native separators via FromSlash so the wants match
	// on any platform.
	r := harness.Resolver{Home: fromSlash("/home/u"), ProjectRoot: fromSlash("/repo")}
	tests := []struct {
		name  string
		id    harness.ID
		scope harness.Scope
		want  dsync.Target
	}{
		{
			name: "codex global carries legacy dir and sidecar",
			id:   harness.Codex, scope: harness.Global,
			want: dsync.Target{Name: "OpenAI Codex", SkillsDir: fromSlash("/home/u/.codex/skills"), LegacyDir: fromSlash("/home/u/.codex/prompts"), Codex: true},
		},
		{
			name: "claude local drops legacy dir and sidecar",
			id:   harness.Claude, scope: harness.Local,
			want: dsync.Target{Name: "Claude Code", SkillsDir: fromSlash("/repo/.claude/skills")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildTarget(r, tt.id, tt.scope)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("target = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func fromSlash(p string) string { return filepath.FromSlash(p) }

// treeSnapshot maps every path under root to its content ("dir" for
// directories), so byte-identity of a tree is one maps.Equal away.
func treeSnapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	snap := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			snap[p] = "dir"
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		snap[p] = string(b)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return snap
}

// runInstall is the wiring around the sync engine: the install/uninstall branch
// and the dry-run gate. Uninstall is the binary's destructive path, so the
// wiring is pinned here — the engine's own semantics live in internal/sync.
// The Claude override keeps the resolver off the real machine env entirely.
func TestRunInstallUninstallWiring(t *testing.T) {
	home := t.TempDir()
	cfg := filepath.Join(home, "claude-cfg")
	r := harness.Resolver{Home: home, Overrides: map[harness.ID]string{harness.Claude: cfg}}
	catalog := fstest.MapFS{"skills/ds-x/SKILL.md": &fstest.MapFile{Data: []byte("x")}}
	ids := []harness.ID{harness.Claude}

	if err := runInstall(io.Discard, catalog, r, harness.Global, ids, false, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(cfg, "skills", "ds-x", "SKILL.md")); err != nil {
		t.Fatalf("install did not write the skill: %v", err)
	}
	legacy := filepath.Join(cfg, "commands", "ds-blueprint.md")
	if err := os.MkdirAll(filepath.Dir(legacy), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacy, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	before := treeSnapshot(t, home)
	if err := runInstall(io.Discard, catalog, r, harness.Global, ids, true, true); err != nil {
		t.Fatal(err)
	}
	if !maps.Equal(before, treeSnapshot(t, home)) {
		t.Error("uninstall --dry-run changed the tree")
	}

	if err := runInstall(io.Discard, catalog, r, harness.Global, ids, false, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(cfg, "skills", "ds-x")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("uninstall left the skill dir: %v", err)
	}
	if _, err := os.Stat(legacy); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("uninstall left the legacy command file: %v", err)
	}
}
