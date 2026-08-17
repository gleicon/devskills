package catalog

import (
	"os"
	"path/filepath"
	"testing"
)

// grandfathered lists the skills that existed before bench coverage became
// mandatory (FR-15). Never add to it: a new skill ships with at least one
// scenario under evals/<skill>/.
var grandfathered = map[string]bool{
	"ds":                     true,
	"ds-architecture-plan":   true,
	"ds-blueprint":           true,
	"ds-bug-review":          true,
	"ds-code-quality-review": true,
	"ds-comment-review":      true,
	"ds-data-mode":           true,
	"ds-data-review":         true,
	"ds-debug":               true,
	"ds-doc-quality-review":  true,
	"ds-explore":             true,
	"ds-go-review":           true,
	"ds-grill-me":            true,
	"ds-handoff":             true,
	"ds-java-review":         true,
	"ds-notebook-review":     true,
	"ds-onboarding":          true,
	"ds-osv":                 true,
	"ds-perf-plan":           true,
	"ds-project-map":         true,
	"ds-project-resume":      true,
	"ds-python-review":       true,
	"ds-quality-gate":        true,
	"ds-recall":              true,
	"ds-recall-capture":      true,
	"ds-recall-setup":        true,
	"ds-retro":               true,
	"ds-roadmap":             true,
	"ds-rust-review":         true,
	"ds-security-review":     true,
	"ds-semgrep":             true,
	"ds-spec":                true,
	"ds-step-mode":           true,
	"ds-tdd-mode":            true,
	"ds-test-mode":           true,
	"ds-test-quality-review": true,
	"ds-tiger-style-mode":    true,
	"ds-tldt":                true,
	"ds-ts-review":           true,
	"ds-ui-mode":             true,
	"ds-ui-quality-review":   true,
	"ds-verify-this":         true,
	"ds-zig-review":          true,
	"ds-zoom-out":            true,
}

// TestNewSkillsHaveScenarios fails, naming the skill, when a skill outside
// the grandfather list has no scenario under evals/ (FR-15, AC-11).
func TestNewSkillsHaveScenarios(t *testing.T) {
	skills, err := os.ReadDir("../../skills")
	if err != nil {
		t.Fatal(err)
	}
	for _, sk := range skills {
		if !sk.IsDir() || grandfathered[sk.Name()] {
			continue
		}
		if !hasScenario(sk.Name()) {
			t.Errorf("skill %s has no scenario under evals/%s/ — add one; never extend the grandfather list", sk.Name(), sk.Name())
		}
	}
}

func hasScenario(skill string) bool {
	entries, err := os.ReadDir(filepath.Join("../../evals", skill))
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			return true
		}
	}
	return false
}
