package catalog

import (
	"bytes"
	"os"
	"testing"
)

// TestDevSkillCopiesInSync guards the repo-local /ds-dev-skill-bench skill,
// which is committed twice: .claude/skills/ (read by Claude Code, and by
// OpenCode through its Claude-compat paths) and .codex/skills/ (read by
// Codex, with the invoke-only sidecar). The .claude copy is the source;
// after editing it, copy it over the .codex one.
func TestDevSkillCopiesInSync(t *testing.T) {
	source, err := os.ReadFile("../../.claude/skills/ds-dev-skill-bench/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	codex, err := os.ReadFile("../../.codex/skills/ds-dev-skill-bench/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(source, codex) {
		t.Error("the .codex copy of ds-dev-skill-bench differs from the .claude source — copy .claude/skills/ds-dev-skill-bench/SKILL.md over it")
	}

	sidecar, err := os.ReadFile("../../.codex/skills/ds-dev-skill-bench/agents/openai.yaml")
	if err != nil {
		t.Fatal(err)
	}
	const policy = "policy:\n  allow_implicit_invocation: false\n"
	if string(sidecar) != policy {
		t.Errorf("codex sidecar = %q, want %q", sidecar, policy)
	}
}
