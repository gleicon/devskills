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

// fakeClaude puts a shell script named claude on PATH, keeping the real PATH
// so git stays reachable.
func fakeClaude(t *testing.T, script string) {
	t.Helper()
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "claude"), []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
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

func TestRunnerUnsupportedHarness(t *testing.T) {
	r := Runner{Harness: harness.Codex, Model: "m"}
	if _, err := r.Run(context.Background(), fixtureScenario(t), SkillVersion{Name: "s", Content: []byte("s")}); err == nil {
		t.Fatal("want error for a harness bench does not support yet")
	}
}
