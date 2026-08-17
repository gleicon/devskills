package bench

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// skillRepo builds a git repo with skills/ds-x/SKILL.md committed on the given
// default branch, then diverges the working tree copy.
func skillRepo(t *testing.T, branch string) string {
	t.Helper()
	root := t.TempDir()
	path := filepath.Join(root, "skills", "ds-x", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("OLD\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"init", "-q", "-b", branch},
		{"add", "-A"},
		{"commit", "-q", "-m", "base"},
	} {
		if err := git(root, args...); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(path, []byte("NEW\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestLoadVersions(t *testing.T) {
	for _, branch := range []string{"main", "master"} {
		t.Run(branch, func(t *testing.T) {
			vs, err := LoadVersions(skillRepo(t, branch), "ds-x")
			if err != nil {
				t.Fatal(err)
			}
			if len(vs) != 2 {
				t.Fatalf("got %d versions, want old and new", len(vs))
			}
			if vs[0].Label != LabelOld || string(vs[0].Content) != "OLD\n" {
				t.Errorf("old = %q %q", vs[0].Label, vs[0].Content)
			}
			if vs[1].Label != LabelNew || string(vs[1].Content) != "NEW\n" {
				t.Errorf("new = %q %q", vs[1].Label, vs[1].Content)
			}
		})
	}
}

func TestLoadVersionsBaseline(t *testing.T) {
	root := skillRepo(t, "main")
	// A second skill exists only in the working tree — absent on main.
	path := filepath.Join(root, "skills", "ds-fresh", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("FRESH\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	vs, err := LoadVersions(root, "ds-fresh")
	if err != nil {
		t.Fatal(err)
	}
	if len(vs) != 1 || vs[0].Label != LabelNew || string(vs[0].Content) != "FRESH\n" {
		t.Errorf("versions = %+v, want baseline new only", vs)
	}
}

func TestLoadVersionsMissingSkill(t *testing.T) {
	_, err := LoadVersions(skillRepo(t, "main"), "ds-nope")
	if err == nil || !strings.Contains(err.Error(), "ds-nope") {
		t.Errorf("error = %v, want it to name the skill", err)
	}
}

func TestLoadVersionsNoDefaultBranch(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "skills", "ds-x", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("X\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := git(root, "init", "-q", "-b", "trunk"); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadVersions(root, "ds-x"); err == nil || !strings.Contains(err.Error(), "no main or master") {
		t.Errorf("error = %v, want no-default-branch", err)
	}
}
