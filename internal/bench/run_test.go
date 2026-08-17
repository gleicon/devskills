package bench

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gleicon/devskills/internal/harness"
)

// fakeCLI puts a shell script with the given name on PATH, keeping the real
// PATH so git stays reachable.
func fakeCLI(t *testing.T, name, script string) {
	t.Helper()
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func fakeClaude(t *testing.T, script string) {
	t.Helper()
	fakeCLI(t, "claude", script)
}

func TestRunnerRun(t *testing.T) {
	argsFile := filepath.Join(t.TempDir(), "args")
	t.Setenv("ARGS_OUT", argsFile)
	fakeClaude(t, `printf '%s\n' "$@" > "$ARGS_OUT"
cat .claude/skills/ds-x/SKILL.md
printf 'package main // cleaned\n' > main.go
echo "one warning" >&2
`)

	skill := SkillVersion{Name: "ds-x", Content: []byte("SKILLBODY\n")}
	r := Runner{Harness: harness.Claude, Model: "pin-model"}
	res, err := r.Run(context.Background(), fixtureScenario(t), skill)
	if err != nil {
		t.Fatal(err)
	}
	if res.Err != nil {
		t.Fatalf("Result.Err = %v", res.Err)
	}

	if !strings.Contains(res.Stdout, "SKILLBODY") {
		t.Errorf("stdout = %q, want the project-locally installed skill content", res.Stdout)
	}
	if !strings.Contains(res.Stderr, "one warning") {
		t.Errorf("stderr = %q", res.Stderr)
	}
	if !strings.Contains(res.Diff, "cleaned") || !strings.Contains(res.Diff, "main.go") {
		t.Errorf("diff = %q, want the harness's edit to main.go", res.Diff)
	}
	if strings.Contains(res.Diff, ".claude") || strings.Contains(res.Diff, "SKILLBODY") {
		t.Errorf("diff leaks harness dir contents: %q", res.Diff)
	}

	args, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"-p", "Review the diff", "--model", "pin-model", "--dangerously-skip-permissions"} {
		if !strings.Contains(string(args), want) {
			t.Errorf("claude args = %q, missing %q", args, want)
		}
	}
}

func TestRunnerRecordsFailure(t *testing.T) {
	fakeClaude(t, `echo "boom" >&2; exit 3`)
	r := Runner{Harness: harness.Claude, Model: "m"}
	res, err := r.Run(context.Background(), fixtureScenario(t), SkillVersion{Name: "ds-x", Content: []byte("s")})
	if err != nil {
		t.Fatal(err)
	}
	if res.Err == nil {
		t.Fatal("want Result.Err for a non-zero harness exit")
	}
	if !strings.Contains(res.Stderr, "boom") {
		t.Errorf("stderr = %q, want it kept on failure", res.Stderr)
	}
}

func TestRunnerRecordsTimeout(t *testing.T) {
	fakeClaude(t, `sleep 5`)
	r := Runner{Harness: harness.Claude, Model: "m", Timeout: 100 * time.Millisecond}
	res, err := r.Run(context.Background(), fixtureScenario(t), SkillVersion{Name: "ds-x", Content: []byte("s")})
	if err != nil {
		t.Fatal(err)
	}
	if res.Err == nil || !strings.Contains(res.Err.Error(), "timed out") {
		t.Errorf("Result.Err = %v, want timeout", res.Err)
	}
}

func TestRunnerMissingCLI(t *testing.T) {
	// PATH with git only, no claude: materialization works, the invoke fails.
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not found")
	}
	bin := t.TempDir()
	if err := os.Symlink(gitPath, filepath.Join(bin, "git")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	r := Runner{Harness: harness.Claude, Model: "m"}
	res, err := r.Run(context.Background(), fixtureScenario(t), SkillVersion{Name: "ds-x", Content: []byte("s")})
	if err != nil {
		t.Fatal(err)
	}
	if res.Err == nil {
		t.Fatal("want Result.Err when the harness CLI is missing")
	}
}

func TestRunnerCodexInvocation(t *testing.T) {
	argsFile := filepath.Join(t.TempDir(), "args")
	t.Setenv("ARGS_OUT", argsFile)
	fakeCLI(t, "codex", `printf '%s\n' "$@" > "$ARGS_OUT"
cat .codex/skills/ds-x/agents/openai.yaml`)
	r := Runner{Harness: harness.Codex, Model: "codex-model"}
	res, err := r.Run(context.Background(), fixtureScenario(t), SkillVersion{Name: "ds-x", Content: []byte("S")})
	if err != nil {
		t.Fatal(err)
	}
	if res.Err != nil {
		t.Fatalf("Result.Err = %v", res.Err)
	}
	args, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"exec", "--model", "codex-model", "--sandbox", "workspace-write", "Review the diff"} {
		if !strings.Contains(string(args), want) {
			t.Errorf("codex args = %q, missing %q", args, want)
		}
	}
	// The sync engine's Codex sidecar must be emitted in the sandbox install.
	if !strings.Contains(res.Stdout, "allow_implicit_invocation: false") {
		t.Errorf("stdout = %q, want the codex sidecar policy", res.Stdout)
	}
}

func TestRunnerOpenCodeInvocation(t *testing.T) {
	argsFile := filepath.Join(t.TempDir(), "args")
	t.Setenv("ARGS_OUT", argsFile)
	fakeCLI(t, "opencode", `printf '%s\n' "$@" > "$ARGS_OUT"
cat .opencode/skills/ds-x/SKILL.md`)
	r := Runner{Harness: harness.OpenCode, Model: "anthropic/some-model"}
	res, err := r.Run(context.Background(), fixtureScenario(t), SkillVersion{Name: "ds-x", Content: []byte("OCSKILL")})
	if err != nil {
		t.Fatal(err)
	}
	if res.Err != nil {
		t.Fatalf("Result.Err = %v", res.Err)
	}
	args, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"run", "Review the diff", "--model", "anthropic/some-model"} {
		if !strings.Contains(string(args), want) {
			t.Errorf("opencode args = %q, missing %q", args, want)
		}
	}
	if !strings.Contains(res.Stdout, "OCSKILL") {
		t.Errorf("stdout = %q, want the installed skill under .opencode/skills", res.Stdout)
	}
}
