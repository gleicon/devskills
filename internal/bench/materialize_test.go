package bench

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// fixtureScenario builds a loadable scenario whose base/ has main.go and
// keep.go, and whose change/ overwrites main.go and adds extra.go.
func fixtureScenario(t *testing.T) *Scenario {
	t.Helper()
	dir := writeScenario(t, "overlay", validReport, "base", "change")
	files := map[string]string{
		"base/main.go":    "package main // v1\n",
		"base/keep.go":    "package main // untouched\n",
		"change/main.go":  "package main // v2\n",
		"change/extra.go": "package main // added\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	s, err := LoadScenario(dir)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func gitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestMaterialize(t *testing.T) {
	s := fixtureScenario(t)
	repo := t.TempDir()
	if err := Materialize(s, repo, nil); err != nil {
		t.Fatal(err)
	}

	if got := gitOut(t, repo, "branch", "--show-current"); got != WorkBranch {
		t.Errorf("checked-out branch = %q, want %q", got, WorkBranch)
	}
	if got := gitOut(t, repo, "log", "--format=%s", DefaultBranch); got != "base" {
		t.Errorf("%s log = %q, want single base commit", DefaultBranch, got)
	}
	if got := gitOut(t, repo, "log", "--format=%s", WorkBranch); got != "change\nbase" {
		t.Errorf("%s log = %q, want change on top of base", WorkBranch, got)
	}
	if got := gitOut(t, repo, "status", "--porcelain"); got != "" {
		t.Errorf("working tree dirty after materialize: %q", got)
	}

	// Overlay semantics: change/ overwrites and adds, base-only files survive.
	if got := gitOut(t, repo, "show", WorkBranch+":main.go"); !strings.Contains(got, "v2") {
		t.Errorf("main.go on work branch = %q, want overwritten v2", got)
	}
	if got := gitOut(t, repo, "show", DefaultBranch+":main.go"); !strings.Contains(got, "v1") {
		t.Errorf("main.go on default branch = %q, want original v1", got)
	}
	for _, f := range []string{"keep.go", "extra.go"} {
		if got := gitOut(t, repo, "show", WorkBranch+":"+f); got == "" {
			t.Errorf("%s missing on work branch", f)
		}
	}
	if out, err := exec.Command("git", "-C", repo, "show", DefaultBranch+":extra.go").CombinedOutput(); err == nil {
		t.Errorf("extra.go exists on default branch: %s", out)
	}
}

func TestMaterializeNestedDirs(t *testing.T) {
	s := fixtureScenario(t)
	nested := filepath.Join(s.Dir, "change", "internal", "deep")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "d.go"), []byte("package deep\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	repo := t.TempDir()
	if err := Materialize(s, repo, nil); err != nil {
		t.Fatal(err)
	}
	if got := gitOut(t, repo, "show", WorkBranch+":internal/deep/d.go"); got == "" {
		t.Error("nested change file missing on work branch")
	}
}

func TestMaterializeEmptyChangeFails(t *testing.T) {
	dir := writeScenario(t, "noop", validReport, "base", "change")
	if err := os.WriteFile(filepath.Join(dir, "base", "main.go"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := LoadScenario(dir)
	if err != nil {
		t.Fatal(err)
	}
	err = Materialize(s, t.TempDir(), nil)
	if err == nil || !strings.Contains(err.Error(), "noop") {
		t.Errorf("want loud failure naming the scenario for an empty change/, got %v", err)
	}
}

func TestMaterializeBlocks(t *testing.T) {
	s := fixtureScenario(t)
	if err := os.WriteFile(filepath.Join(s.Dir, "base", "AGENTS.md"), []byte("Fixture rules.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	repo := t.TempDir()
	if err := Materialize(s, repo, []Block{{ID: "base", Body: "BASEBLOCK\n"}}); err != nil {
		t.Fatal(err)
	}

	for _, branch := range []string{DefaultBranch, WorkBranch} {
		agents := gitOut(t, repo, "show", branch+":AGENTS.md")
		for _, want := range []string{"Fixture rules.", "<!-- BEGIN devskills:base -->", "BASEBLOCK"} {
			if !strings.Contains(agents, want) {
				t.Errorf("AGENTS.md on %s = %q, missing %q", branch, agents, want)
			}
		}
		if got := gitOut(t, repo, "show", branch+":CLAUDE.md"); !strings.Contains(got, "@AGENTS.md") {
			t.Errorf("CLAUDE.md on %s = %q, want the AGENTS.md import", branch, got)
		}
	}
	if got := gitOut(t, repo, "diff", "--name-only", DefaultBranch+"..."+WorkBranch); strings.Contains(got, ".md") {
		t.Errorf("branch diff = %q, want blocks outside the change a skill reviews", got)
	}
	if got := gitOut(t, repo, "ls-files"); strings.Contains(got, ".bak") {
		t.Errorf("tracked files = %q, want no engine backup committed into the fixture", got)
	}
	if got := gitOut(t, repo, "status", "--porcelain"); got != "" {
		t.Errorf("working tree dirty after materialize: %q", got)
	}
}
