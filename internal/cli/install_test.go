package cli

import (
	"path/filepath"
	"slices"
	"testing"

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
