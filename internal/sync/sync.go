// Package sync brings a harness's on-disk skills into line with the embedded
// catalog: write the current skills, prune retired skill dirs and legacy command
// files. Planning is read-only, so --dry-run is simply "plan but don't apply".
package sync

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Target is one resolved install destination for a (harness, scope).
type Target struct {
	Name      string // harness display name, for plan messages
	SkillsDir string // where skill directories are written and retired ones pruned
	LegacyDir string // legacy command/prompt dir to purge; empty to skip (local scope)
	Codex     bool   // emit the per-skill Codex sidecar on write
}

// RemoveKind distinguishes what a prune step deletes.
type RemoveKind int

const (
	RetiredSkill RemoveKind = iota
	LegacyCommand
)

func (k RemoveKind) String() string {
	if k == LegacyCommand {
		return "legacy command"
	}
	return "retired skill"
}

// Remove is a single prune step: a present retired skill dir or legacy command
// file to delete.
type Remove struct {
	Path string
	Kind RemoveKind
}

// Plan is the set of changes to sync one Target: every current skill is
// (over)written, and each retired/legacy name that is actually present is
// removed.
type Plan struct {
	Target  Target
	Writes  []string // skill names, in catalog order
	Removes []Remove
}

// Engine plans and applies syncs against the embedded catalog. The fs.FS is
// rooted where skills live under "skills/" (the same layout catalog.Validate
// expects), so main can pass the embed directly.
type Engine struct {
	catalog fs.FS
}

func New(catalog fs.FS) Engine { return Engine{catalog: catalog} }

// Plan computes the changes for t without touching the filesystem beyond
// read-only existence checks — the guarantee behind --dry-run.
func (e Engine) Plan(t Target) (Plan, error) {
	names, err := skillNames(e.catalog)
	if err != nil {
		return Plan{}, err
	}
	p := Plan{Target: t, Writes: names}
	for _, name := range retiredSkills {
		if pth := filepath.Join(t.SkillsDir, name); exists(pth) {
			p.Removes = append(p.Removes, Remove{Path: pth, Kind: RetiredSkill})
		}
	}
	if t.LegacyDir != "" {
		for _, name := range legacyCommandFiles {
			if pth := filepath.Join(t.LegacyDir, name); exists(pth) {
				p.Removes = append(p.Removes, Remove{Path: pth, Kind: LegacyCommand})
			}
		}
	}
	return p, nil
}

// Apply performs the plan's writes and removals. Writes overwrite blindly —
// skills are files devskills owns, so no backups (see DECISIONS.md).
func (e Engine) Apply(p Plan) error {
	if err := os.MkdirAll(p.Target.SkillsDir, 0o755); err != nil {
		return fmt.Errorf("create skills dir: %w", err)
	}
	for _, name := range p.Writes {
		dest := filepath.Join(p.Target.SkillsDir, name)
		if err := e.writeSkill(name, dest, p.Target.Codex); err != nil {
			return err
		}
	}
	for _, rm := range p.Removes {
		if err := os.RemoveAll(rm.Path); err != nil {
			return fmt.Errorf("remove %s %s: %w", rm.Kind, rm.Path, err)
		}
	}
	return nil
}

// writeSkill replaces dest with a fresh copy of the embedded skill. The rm
// before copy keeps re-runs idempotent and drops companion files removed
// upstream, instead of leaving them stranded.
func (e Engine) writeSkill(name, dest string, codex bool) error {
	if err := os.RemoveAll(dest); err != nil {
		return fmt.Errorf("clean %s: %w", dest, err)
	}
	if err := copyTree(e.catalog, path.Join("skills", name), dest); err != nil {
		return fmt.Errorf("write skill %s: %w", name, err)
	}
	if codex {
		return emitCodexSidecar(dest)
	}
	return nil
}

// emitCodexSidecar writes Codex's per-skill policy file. allow_implicit_
// invocation: false keeps the skill user-invoked only — never auto-fired —
// which is devskills' whole identity. Generated at install time, never in the
// repo.
func emitCodexSidecar(skillDir string) error {
	dir := filepath.Join(skillDir, "agents")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("codex sidecar dir: %w", err)
	}
	const policy = "policy:\n  allow_implicit_invocation: false\n"
	if err := os.WriteFile(filepath.Join(dir, "openai.yaml"), []byte(policy), 0o644); err != nil {
		return fmt.Errorf("codex sidecar: %w", err)
	}
	return nil
}

func copyTree(src fs.FS, root, dest string) error {
	return fs.WalkDir(src, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(p, root), "/")
		target := filepath.Join(dest, filepath.FromSlash(rel))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := fs.ReadFile(src, p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
}

func skillNames(catalog fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(catalog, "skills")
	if err != nil {
		return nil, fmt.Errorf("read catalog: %w", err)
	}
	var names []string
	for _, d := range entries {
		if d.IsDir() {
			names = append(names, d.Name())
		}
	}
	if len(names) == 0 {
		return nil, errors.New("catalog has no skills")
	}
	return names, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
