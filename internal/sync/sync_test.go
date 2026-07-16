package sync

import (
	"io/fs"
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
	mkfile(t, filepath.Join(legacyDir, "ds-modes.md"), "old")
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
		filepath.Join(legacyDir, "ds-modes.md"),
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

	before := snapshot(t, skillsDir, legacyDir)
	if _, err := New(fakeCatalog()).Plan(Target{SkillsDir: skillsDir, LegacyDir: legacyDir}); err != nil {
		t.Fatal(err)
	}
	if after := snapshot(t, skillsDir, legacyDir); !slices.Equal(before, after) {
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
	p, err := e.Plan(Target{SkillsDir: skillsDir})
	if err != nil {
		t.Fatal(err)
	}
	if err := e.Apply(p); err != nil {
		t.Fatal(err)
	}
	assertAbsent(t, filepath.Join(skillsDir, "ds-a", "stale.md"))
	assertFile(t, filepath.Join(skillsDir, "ds-a", "SKILL.md"), "a")
}

func TestApplyCodexSidecar(t *testing.T) {
	skillsDir := t.TempDir()
	e := New(fakeCatalog())
	p, err := e.Plan(Target{SkillsDir: skillsDir, Codex: true})
	if err != nil {
		t.Fatal(err)
	}
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
	p, err := e.Plan(Target{SkillsDir: skillsDir})
	if err != nil {
		t.Fatal(err)
	}
	if err := e.Apply(p); err != nil {
		t.Fatal(err)
	}
	assertAbsent(t, filepath.Join(skillsDir, "ds-a", "agents"))
}

func TestUninstallPlanRemovesFootprintNotUserFiles(t *testing.T) {
	skillsDir := t.TempDir()
	legacyDir := t.TempDir()
	// devskills' footprint: a current-catalog skill, a retired skill, a legacy command.
	mkfile(t, filepath.Join(skillsDir, "ds-a", "SKILL.md"), "a")
	mkfile(t, filepath.Join(skillsDir, "ds-typeset", "SKILL.md"), "old")
	mkfile(t, filepath.Join(legacyDir, "ds-workflow.md"), "old")
	// Not ours — must be left alone.
	mkfile(t, filepath.Join(skillsDir, "ds-mine", "SKILL.md"), "mine")
	mkfile(t, filepath.Join(legacyDir, "ds-custom.md"), "mine")

	p, err := New(fakeCatalog()).UninstallPlan(Target{SkillsDir: skillsDir, LegacyDir: legacyDir})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Writes) != 0 {
		t.Errorf("Writes = %v, want none", p.Writes)
	}
	got := removePaths(p)
	want := []string{
		filepath.Join(skillsDir, "ds-a"),
		filepath.Join(skillsDir, "ds-typeset"),
		filepath.Join(legacyDir, "ds-workflow.md"),
	}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("Removes = %v, want %v", got, want)
	}
}

// TestUninstallPlanIsReadOnly extends the --dry-run guarantee to uninstall:
// planning removals must not delete anything.
func TestUninstallPlanIsReadOnly(t *testing.T) {
	skillsDir := t.TempDir()
	legacyDir := t.TempDir()
	mkfile(t, filepath.Join(skillsDir, "ds-a", "SKILL.md"), "a")
	mkfile(t, filepath.Join(legacyDir, "ds-workflow.md"), "old")

	before := snapshot(t, skillsDir, legacyDir)
	if _, err := New(fakeCatalog()).UninstallPlan(Target{SkillsDir: skillsDir, LegacyDir: legacyDir}); err != nil {
		t.Fatal(err)
	}
	if after := snapshot(t, skillsDir, legacyDir); !slices.Equal(before, after) {
		t.Errorf("UninstallPlan mutated the target: before %v, after %v", before, after)
	}
}

func TestApplyUninstallRemovesFootprintLeavesUserFiles(t *testing.T) {
	skillsDir := t.TempDir()
	mkfile(t, filepath.Join(skillsDir, "ds-a", "SKILL.md"), "a")
	mkfile(t, filepath.Join(skillsDir, "ds-typeset", "SKILL.md"), "old")
	mkfile(t, filepath.Join(skillsDir, "ds-mine", "SKILL.md"), "mine")

	e := New(fakeCatalog())
	p, err := e.UninstallPlan(Target{SkillsDir: skillsDir})
	if err != nil {
		t.Fatal(err)
	}
	if err := e.Apply(p); err != nil {
		t.Fatal(err)
	}
	assertAbsent(t, filepath.Join(skillsDir, "ds-a"))
	assertAbsent(t, filepath.Join(skillsDir, "ds-typeset"))
	assertFile(t, filepath.Join(skillsDir, "ds-mine", "SKILL.md"), "mine")
}

// TestApplyUninstallDoesNotCreateSkillsDir guards the write-only MkdirAll: a
// never-installed harness (no skills dir) must stay a no-op, not conjure the dir.
func TestApplyUninstallDoesNotCreateSkillsDir(t *testing.T) {
	skillsDir := filepath.Join(t.TempDir(), "skills")
	e := New(fakeCatalog())
	p, err := e.UninstallPlan(Target{SkillsDir: skillsDir})
	if err != nil {
		t.Fatal(err)
	}
	if err := e.Apply(p); err != nil {
		t.Fatal(err)
	}
	assertAbsent(t, skillsDir)
}

// Guards the rename mistake — adding the new skill name but leaving the old one
// in retiredSkills while it's still a live dir. Reads the real skills/ tree; go
// test's CWD is the package dir, so it's two up (as catalog_test does).
func TestRetiredSkillsNotInCatalog(t *testing.T) {
	entries, err := fs.ReadDir(os.DirFS("../.."), "skills")
	if err != nil {
		t.Fatal(err)
	}
	live := make(map[string]bool, len(entries))
	for _, d := range entries {
		live[d.Name()] = true
	}
	for _, name := range retiredSkills {
		if live[name] {
			t.Errorf("%s is in retiredSkills but still a live skill in skills/", name)
		}
	}
}

func TestPlanProtectsRetiredNameBackInCatalog(t *testing.T) {
	// A maintainer returns ds-typeset to the catalog but leaves it in retiredSkills.
	// Plan must write it and never schedule the path it just wrote for removal.
	catalog := fstest.MapFS{
		"skills/ds-typeset/SKILL.md": {Data: []byte("back")},
		"skills/ds-a/SKILL.md":       {Data: []byte("a")},
	}
	skillsDir := t.TempDir()
	mkfile(t, filepath.Join(skillsDir, "ds-typeset", "SKILL.md"), "present")

	p, err := New(catalog).Plan(Target{SkillsDir: skillsDir})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(p.Writes, "ds-typeset") {
		t.Errorf("Writes = %v, want it to include ds-typeset", p.Writes)
	}
	for _, rm := range p.Removes {
		if rm.Path == filepath.Join(skillsDir, "ds-typeset") {
			t.Error("ds-typeset is in the catalog; it must not be pruned")
		}
	}
}

func TestPlanPrunesDanglingSymlinkRetiredSkill(t *testing.T) {
	skillsDir := t.TempDir()
	link := filepath.Join(skillsDir, "ds-typeset")
	if err := os.Symlink(filepath.Join(skillsDir, "gone"), link); err != nil {
		t.Skipf("symlinks unsupported here: %v", err)
	}
	p, err := New(fakeCatalog()).Plan(Target{SkillsDir: skillsDir})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, rm := range p.Removes {
		if rm.Path == link {
			found = true
		}
	}
	if !found {
		t.Error("a dangling symlink at a retired-skill path should be pruned")
	}
}

func snapshot(t *testing.T, roots ...string) []string {
	t.Helper()
	var out []string
	for _, root := range roots {
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
