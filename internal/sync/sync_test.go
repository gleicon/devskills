package sync

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"testing/fstest"
)

// fakeCatalog is a two-skill catalog; ds-a carries a companion file to prove
// whole-directory copies.
func fakeCatalog() fstest.MapFS {
	return fstest.MapFS{
		"skills/ds-a/SKILL.md": {Data: []byte("a")},
		"skills/ds-a/extra.md": {Data: []byte("companion")},
		"skills/ds-b/SKILL.md": {Data: []byte("b")},
	}
}

func mkfile(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func removePaths(p Plan) []string {
	out := make([]string, len(p.Removes))
	for i, r := range p.Removes {
		out[i] = r.Path
	}
	return out
}

func TestPlanWriteSet(t *testing.T) {
	e := New(fakeCatalog())
	p, err := e.Plan(Target{SkillsDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(p.Writes, []string{"ds-a", "ds-b"}) {
		t.Errorf("Writes = %v, want [ds-a ds-b]", p.Writes)
	}
	if len(p.Removes) != 0 {
		t.Errorf("Removes = %v, want none", p.Removes)
	}
}

func TestPlanPrunesPresentRetiredAndLegacy(t *testing.T) {
	skillsDir := t.TempDir()
	legacyDir := t.TempDir()
	// Present: a retired skill dir, and two legacy command files.
	mkfile(t, filepath.Join(skillsDir, "ds-typeset", "SKILL.md"), "old")
	mkfile(t, filepath.Join(legacyDir, "ds-workflow.md"), "old")
	mkfile(t, filepath.Join(legacyDir, "frontend.md"), "old")
	// Not in any ledger — must be left alone.
	mkfile(t, filepath.Join(legacyDir, "ds-custom.md"), "mine")

	p, err := New(fakeCatalog()).Plan(Target{SkillsDir: skillsDir, LegacyDir: legacyDir})
	if err != nil {
		t.Fatal(err)
	}
	got := removePaths(p)
	want := []string{
		filepath.Join(skillsDir, "ds-typeset"),
		filepath.Join(legacyDir, "ds-workflow.md"),
		filepath.Join(legacyDir, "frontend.md"),
	}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("Removes = %v, want %v", got, want)
	}
	for _, r := range p.Removes {
		if filepath.Base(r.Path) == "ds-custom.md" {
			t.Error("planned removal of a non-ledger file")
		}
	}
}

func TestPlanSkipsLegacyWithoutLegacyDir(t *testing.T) {
	skillsDir := t.TempDir()
	mkfile(t, filepath.Join(skillsDir, "ds-typeset", "SKILL.md"), "old")
	p, err := New(fakeCatalog()).Plan(Target{SkillsDir: skillsDir}) // no LegacyDir
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range p.Removes {
		if r.Kind == LegacyCommand {
			t.Errorf("legacy removal planned without a LegacyDir: %s", r.Path)
		}
	}
}

// TestPlanIsReadOnly is the engine-level --dry-run guarantee: planning must not
// change the filesystem, so the retired dir and legacy files survive an unapplied
// plan and nothing new is written.
func TestPlanIsReadOnly(t *testing.T) {
	skillsDir := t.TempDir()
	legacyDir := t.TempDir()
	mkfile(t, filepath.Join(skillsDir, "ds-typeset", "SKILL.md"), "old")
	mkfile(t, filepath.Join(legacyDir, "ds-workflow.md"), "old")

	before := snapshot(t, skillsDir)
	if _, err := New(fakeCatalog()).Plan(Target{SkillsDir: skillsDir, LegacyDir: legacyDir}); err != nil {
		t.Fatal(err)
	}
	if after := snapshot(t, skillsDir); !slices.Equal(before, after) {
		t.Errorf("Plan mutated the target: before %v, after %v", before, after)
	}
}

func TestApplyWritesAndPrunes(t *testing.T) {
	skillsDir := t.TempDir()
	// A user file that is not devskills' — must survive.
	mkfile(t, filepath.Join(skillsDir, "ds-mine", "SKILL.md"), "mine")
	// A retired skill — must be pruned.
	mkfile(t, filepath.Join(skillsDir, "ds-typeset", "SKILL.md"), "old")

	e := New(fakeCatalog())
	p, err := e.Plan(Target{SkillsDir: skillsDir})
	if err != nil {
		t.Fatal(err)
	}
	if err := e.Apply(p); err != nil {
		t.Fatal(err)
	}

	assertFile(t, filepath.Join(skillsDir, "ds-a", "SKILL.md"), "a")
	assertFile(t, filepath.Join(skillsDir, "ds-a", "extra.md"), "companion")
	assertFile(t, filepath.Join(skillsDir, "ds-b", "SKILL.md"), "b")
	assertAbsent(t, filepath.Join(skillsDir, "ds-typeset"))
	assertFile(t, filepath.Join(skillsDir, "ds-mine", "SKILL.md"), "mine")
}

func TestApplyIdempotentDropsStaleCompanion(t *testing.T) {
	skillsDir := t.TempDir()
	// A stale companion left by a previous install must not linger after re-sync.
	mkfile(t, filepath.Join(skillsDir, "ds-a", "stale.md"), "stale")

	e := New(fakeCatalog())
	p, _ := e.Plan(Target{SkillsDir: skillsDir})
	if err := e.Apply(p); err != nil {
		t.Fatal(err)
	}
	assertAbsent(t, filepath.Join(skillsDir, "ds-a", "stale.md"))
	assertFile(t, filepath.Join(skillsDir, "ds-a", "SKILL.md"), "a")
}

func TestApplyCodexSidecar(t *testing.T) {
	skillsDir := t.TempDir()
	e := New(fakeCatalog())
	p, _ := e.Plan(Target{SkillsDir: skillsDir, Codex: true})
	if err := e.Apply(p); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"ds-a", "ds-b"} {
		assertFile(t, filepath.Join(skillsDir, name, "agents", "openai.yaml"),
			"policy:\n  allow_implicit_invocation: false\n")
	}
}

func TestApplyWithoutCodexHasNoSidecar(t *testing.T) {
	skillsDir := t.TempDir()
	e := New(fakeCatalog())
	p, _ := e.Plan(Target{SkillsDir: skillsDir})
	if err := e.Apply(p); err != nil {
		t.Fatal(err)
	}
	assertAbsent(t, filepath.Join(skillsDir, "ds-a", "agents"))
}

func snapshot(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(root, func(p string, _ os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		out = append(out, p)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(out)
	return out
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(b) != want {
		t.Errorf("%s = %q, want %q", path, b, want)
	}
}

func assertAbsent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("%s should be absent, stat err = %v", path, err)
	}
}
