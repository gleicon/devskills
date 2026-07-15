package harness

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func fakeEnv(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestSkillsDirDefaults(t *testing.T) {
	r := Resolver{Home: "/home/u", ProjectRoot: "/repo", getenv: fakeEnv(nil)}
	tests := []struct {
		id    ID
		scope Scope
		want  string
	}{
		{Claude, Global, "/home/u/.claude/skills"},
		{Codex, Global, "/home/u/.codex/skills"},
		{OpenCode, Global, "/home/u/.config/opencode/skills"},
		{Claude, Local, "/repo/.claude/skills"},
		{OpenCode, Local, "/repo/.opencode/skills"},
		{Codex, Local, "/repo/.codex/skills"},
	}
	for _, tt := range tests {
		t.Run(string(tt.id)+"-"+scopeName(tt.scope), func(t *testing.T) {
			got, err := r.SkillsDir(tt.id, tt.scope)
			if err != nil {
				t.Fatalf("SkillsDir: %v", err)
			}
			if got != filepath.FromSlash(tt.want) {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSkillsDirUnknownHarness(t *testing.T) {
	r := Resolver{Home: "/home/u", getenv: fakeEnv(nil)}
	if _, err := r.SkillsDir(ID("nope"), Global); err == nil {
		t.Error("want error for unknown harness, got nil")
	}
}

func TestSkillsDirPrecedence(t *testing.T) {
	tests := []struct {
		name      string
		id        ID
		scope     Scope
		env       map[string]string
		overrides map[ID]string
		want      string
	}{
		{
			name: "env overrides default",
			id:   Claude, scope: Global,
			env:  map[string]string{"CLAUDE_CONFIG_DIR": "/env/claude"},
			want: "/env/claude/skills",
		},
		{
			name: "flag overrides env",
			id:   Claude, scope: Global,
			env:       map[string]string{"CLAUDE_CONFIG_DIR": "/env/claude"},
			overrides: map[ID]string{Claude: "/flag/claude"},
			want:      "/flag/claude/skills",
		},
		{
			name: "opencode XDG env appends opencode",
			id:   OpenCode, scope: Global,
			env:  map[string]string{"XDG_CONFIG_HOME": "/xdg"},
			want: "/xdg/opencode/skills",
		},
		{
			name: "opencode flag replaces whole config dir",
			id:   OpenCode, scope: Global,
			env:       map[string]string{"XDG_CONFIG_HOME": "/xdg"},
			overrides: map[ID]string{OpenCode: "/flag/oc"},
			want:      "/flag/oc/skills",
		},
		{
			name: "local ignores env and overrides",
			id:   Claude, scope: Local,
			env:       map[string]string{"CLAUDE_CONFIG_DIR": "/env/claude"},
			overrides: map[ID]string{Claude: "/flag/claude"},
			want:      "/repo/.claude/skills",
		},
		{
			name: "tilde expands against home",
			id:   Codex, scope: Global,
			overrides: map[ID]string{Codex: "~/xdg-codex"},
			want:      "/home/u/xdg-codex/skills",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Resolver{Home: "/home/u", ProjectRoot: "/repo", Overrides: tt.overrides, getenv: fakeEnv(tt.env)}
			got, err := r.SkillsDir(tt.id, tt.scope)
			if err != nil {
				t.Fatalf("SkillsDir: %v", err)
			}
			if got != filepath.FromSlash(tt.want) {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLegacyCommandDir(t *testing.T) {
	r := Resolver{Home: "/home/u", getenv: fakeEnv(nil)}
	tests := []struct {
		id   ID
		want string
		ok   bool
	}{
		{Claude, "/home/u/.claude/commands", true},
		{Codex, "/home/u/.codex/prompts", true},
		{OpenCode, "/home/u/.opencode/commands", true},
	}
	for _, tt := range tests {
		t.Run(string(tt.id), func(t *testing.T) {
			got, ok := r.LegacyCommandDir(tt.id)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if got != filepath.FromSlash(tt.want) {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetected(t *testing.T) {
	found := func(string) (string, error) { return "/usr/bin/x", nil }
	missing := func(string) (string, error) { return "", errors.New("not found") }

	t.Run("binary on PATH", func(t *testing.T) {
		r := Resolver{Home: t.TempDir(), getenv: fakeEnv(nil), lookPath: found}
		if !r.Detected(Claude) {
			t.Error("want detected via PATH")
		}
	})

	t.Run("config dir exists", func(t *testing.T) {
		home := t.TempDir()
		if err := os.MkdirAll(filepath.Join(home, ".codex"), 0o755); err != nil {
			t.Fatal(err)
		}
		r := Resolver{Home: home, getenv: fakeEnv(nil), lookPath: missing}
		if !r.Detected(Codex) {
			t.Error("want detected via config dir")
		}
	})

	t.Run("opencode legacy dir counts", func(t *testing.T) {
		home := t.TempDir()
		if err := os.MkdirAll(filepath.Join(home, ".opencode"), 0o755); err != nil {
			t.Fatal(err)
		}
		r := Resolver{Home: home, getenv: fakeEnv(nil), lookPath: missing}
		if !r.Detected(OpenCode) {
			t.Error("want detected via legacy ~/.opencode")
		}
	})

	t.Run("absent", func(t *testing.T) {
		r := Resolver{Home: t.TempDir(), getenv: fakeEnv(nil), lookPath: missing}
		if r.Detected(Claude) {
			t.Error("want not detected")
		}
	})
}

func scopeName(s Scope) string {
	if s == Local {
		return "local"
	}
	return "global"
}
