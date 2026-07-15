// Package harness resolves per-assistant install targets and detects which
// assistants are present. It knows, for each supported AI coding assistant and
// scope (global vs project-local), where devskills' skills belong.
package harness

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ID identifies a supported AI coding assistant.
type ID string

const (
	Claude      ID = "claude"
	OpenCode    ID = "opencode"
	Codex       ID = "codex"
	Antigravity ID = "antigravity"
)

// Scope is where skills install: shared across the user's projects, or into one
// repo.
type Scope int

const (
	Global Scope = iota
	Local
)

// ErrPluginManaged reports that a (harness, scope) pair has no skills directory
// to sync because the assistant owns its own store. It applies only to
// Antigravity + Global, which installs via `agy plugin install`. Callers match
// it with errors.Is and take the plugin path instead of a file copy.
var ErrPluginManaged = errors.New("harness: target is plugin-managed, no skills dir")

type info struct {
	name   string // human-readable
	binary string // executable name for PATH detection
}

var registry = map[ID]info{
	Claude:      {name: "Claude Code", binary: "claude"},
	OpenCode:    {name: "OpenCode", binary: "opencode"},
	Codex:       {name: "OpenAI Codex", binary: "codex"},
	Antigravity: {name: "Antigravity", binary: "agy"},
}

// All returns every supported harness in install order.
func All() []ID { return []ID{Claude, OpenCode, Codex, Antigravity} }

// Valid reports whether id names a supported harness.
func (id ID) Valid() bool { _, ok := registry[id]; return ok }

// Name is the harness's human-readable name.
func (id ID) Name() string { return registry[id].name }

// Resolver turns a (harness, scope) into a concrete skills directory. Build one
// with NewResolver; the exported fields are overridable for tests and callers
// that already know the environment.
type Resolver struct {
	Home        string        // user home dir; base for default global config dirs
	ProjectRoot string        // repo root; base for --local targets
	Overrides   map[ID]string // per-harness config-dir flag overrides (global only)
	getenv      func(string) string
	lookPath    func(string) (string, error)
}

// NewResolver reads the real environment: home dir, repo root (git top-level,
// else cwd), os.Getenv, and exec.LookPath. overrides holds --<harness>-dir flag
// values (nil if none).
func NewResolver(overrides map[ID]string) (Resolver, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Resolver{}, fmt.Errorf("resolve home dir: %w", err)
	}
	return Resolver{
		Home:        home,
		ProjectRoot: projectRoot(),
		Overrides:   overrides,
		getenv:      os.Getenv,
		lookPath:    exec.LookPath,
	}, nil
}

// SkillsDir resolves the directory devskills' skills install into for id at
// scope. Global applies precedence flag > env > default; local uses the repo's
// per-harness dir and ignores overrides. Returns ErrPluginManaged for the
// Antigravity global target, which has no directory.
func (r Resolver) SkillsDir(id ID, scope Scope) (string, error) {
	if !id.Valid() {
		return "", fmt.Errorf("unknown harness %q", id)
	}
	if scope == Local {
		return filepath.Join(r.ProjectRoot, localSkillsDir(id)), nil
	}
	if id == Antigravity {
		return "", ErrPluginManaged
	}
	return filepath.Join(r.configDir(id), "skills"), nil
}

// Detected reports whether a harness is present, for pre-selecting global
// install targets: its binary is on PATH, or its config dir already exists.
// Local scope has no detection signal, so this is a global-only notion.
func (r Resolver) Detected(id ID) bool {
	h, ok := registry[id]
	if !ok {
		return false
	}
	if _, err := r.find(h.binary); err == nil {
		return true
	}
	if id == Antigravity {
		return false // no config dir; only agy on PATH signals presence
	}
	if _, err := os.Stat(r.configDir(id)); err == nil {
		return true
	}
	// OpenCode's legacy ~/.opencode still counts as installed.
	if id == OpenCode {
		if _, err := os.Stat(filepath.Join(r.Home, ".opencode")); err == nil {
			return true
		}
	}
	return false
}

// LegacyCommandDir returns the harness's pre-skills command/prompt directory —
// the global-only sweep target for the legacy-command purge. The bool is false
// for a harness that never had one (Antigravity).
func (r Resolver) LegacyCommandDir(id ID) (string, bool) {
	switch id {
	case Claude:
		return filepath.Join(r.configDir(Claude), "commands"), true
	case Codex:
		return filepath.Join(r.configDir(Codex), "prompts"), true
	case OpenCode:
		// OpenCode's legacy commands lived under ~/.opencode, not the XDG config dir.
		return filepath.Join(r.Home, ".opencode", "commands"), true
	}
	return "", false
}

// configDir resolves the global config dir (the parent of skills/) with
// precedence override > env > default. Not valid for Antigravity — SkillsDir
// short-circuits it before reaching here.
func (r Resolver) configDir(id ID) string {
	if o := r.Overrides[id]; o != "" {
		return r.expand(o)
	}
	switch id {
	case Claude:
		if v := r.env("CLAUDE_CONFIG_DIR"); v != "" {
			return r.expand(v)
		}
		return filepath.Join(r.Home, ".claude")
	case Codex:
		if v := r.env("CODEX_HOME"); v != "" {
			return r.expand(v)
		}
		return filepath.Join(r.Home, ".codex")
	case OpenCode:
		// XDG_CONFIG_HOME is the parent; opencode/ sits under it.
		if v := r.env("XDG_CONFIG_HOME"); v != "" {
			return filepath.Join(r.expand(v), "opencode")
		}
		return filepath.Join(r.Home, ".config", "opencode")
	}
	return ""
}

func localSkillsDir(id ID) string {
	switch id {
	case Claude:
		return filepath.Join(".claude", "skills")
	case OpenCode:
		return filepath.Join(".opencode", "skills")
	case Codex:
		return filepath.Join(".codex", "skills")
	case Antigravity:
		return filepath.Join(".agents", "skills")
	}
	return ""
}

// expand resolves a leading ~ against Home, since env vars and flag values are
// not shell-expanded.
func (r Resolver) expand(p string) string {
	if p == "~" {
		return r.Home
	}
	if rest, ok := strings.CutPrefix(p, "~/"); ok {
		return filepath.Join(r.Home, rest)
	}
	return p
}

func (r Resolver) env(key string) string {
	if r.getenv != nil {
		return r.getenv(key)
	}
	return os.Getenv(key)
}

func (r Resolver) find(bin string) (string, error) {
	if r.lookPath != nil {
		return r.lookPath(bin)
	}
	return exec.LookPath(bin)
}

// projectRoot returns the git worktree root, falling back to the current dir.
func projectRoot() string {
	if out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output(); err == nil {
		if root := strings.TrimSpace(string(out)); root != "" {
			return root
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return "."
}
