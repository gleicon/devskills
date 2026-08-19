package bench

import (
	"os"
	"strings"
	"testing"
)

// TestCommittedScenarios exercises every scenario committed under evals/
// end to end: load, materialize, and check that each expectation's anchors
// exist in the work-branch version of its file — so expectations can never
// drift from the fixtures they point at.
func TestCommittedScenarios(t *testing.T) {
	skills, err := os.ReadDir("../../evals")
	if err != nil {
		t.Fatal(err)
	}
	for _, skill := range skills {
		if !skill.IsDir() {
			continue
		}
		scenarios, err := LoadScenarios("../../evals", skill.Name())
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range scenarios {
			t.Run(skill.Name()+"/"+s.Name, func(t *testing.T) {
				repo := t.TempDir()
				if err := Materialize(s, repo); err != nil {
					t.Fatal(err)
				}
				for i, e := range s.Expectations {
					content := gitOut(t, repo, "show", WorkBranch+":"+e.File)
					for _, anchor := range e.Anchors {
						if !strings.Contains(content, anchor) {
							t.Errorf("expectations[%d]: anchor %q not found in %s on work branch", i, anchor, e.File)
						}
					}
				}
			})
		}
	}
}

// TestEvalsNotEmbedded guards that the go:embed directive in main.go never
// picks up evals/ — scenarios ship in the repo, not the binary (docs/bench.md).
func TestEvalsNotEmbedded(t *testing.T) {
	b, err := os.ReadFile("../../main.go")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for line := range strings.Lines(string(b)) {
		if !strings.HasPrefix(strings.TrimSpace(line), "//go:embed") {
			continue
		}
		found = true
		if strings.Contains(line, "evals") {
			t.Errorf("main.go embeds evals/: %s", strings.TrimSpace(line))
		}
	}
	if !found {
		t.Fatal("no //go:embed directive found in main.go")
	}
}
