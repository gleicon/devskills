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
	fakeCLI(t, "claude", claudeJSON(script))
}

// claudeJSON wraps a fake claude script so its stdout becomes the result
// text of a JSON result, the way --output-format json reports it. The text
// must hold no quotes or backslashes: the wrapper does not escape them.
func claudeJSON(script string) string {
	return "out=$( (\n" + script + "\n) ); rc=$?\n" + `printf '%s' "$out" | awk '
BEGIN { ORS = ""; print "{\"type\":\"result\",\"subtype\":\"success\",\"is_error\":false,\"result\":\"" }
{ if (NR > 1) print "\\n"; print }
END { print "\",\"total_cost_usd\":0.01,\"modelUsage\":{}}\n" }'
exit $rc
`
}

// benchSkill is a minimal one-file skill version for runner tests.
func benchSkill(content string) SkillVersion {
	return SkillVersion{Name: "ds-x", Files: map[string][]byte{"SKILL.md": []byte(content)}}
}

func TestRunnerRun(t *testing.T) {
	argsFile := filepath.Join(t.TempDir(), "args")
	t.Setenv("ARGS_OUT", argsFile)
	fakeClaude(t, `printf '%s\n' "$@" > "$ARGS_OUT"
cat .claude/skills/ds-x/SKILL.md .claude/skills/ds-x/ref.md
printf 'package main // cleaned\n' > main.go
echo "one warning" >&2
`)

	skill := benchSkill("SKILLBODY\n")
	skill.Files["ref.md"] = []byte("COMPANION\n")
	r := Runner{Harness: harness.Claude, Model: "pin-model"}
	res, err := r.Run(context.Background(), fixtureScenario(t), skill, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Err != nil {
		t.Fatalf("Result.Err = %v", res.Err)
	}

	if !strings.Contains(res.Stdout, "SKILLBODY") {
		t.Errorf("stdout = %q, want the project-locally installed skill content", res.Stdout)
	}
	if !strings.Contains(res.Stdout, "COMPANION") {
		t.Errorf("stdout = %q, want the skill's companion file installed beside SKILL.md", res.Stdout)
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
	for _, want := range []string{"-p", "Review the diff", "--model", "pin-model", "--output-format", "json", "--safe-mode", "--dangerously-skip-permissions"} {
		if !strings.Contains(string(args), want) {
			t.Errorf("claude args = %q, missing %q", args, want)
		}
	}
}

func TestRunnerReadsClaudeResult(t *testing.T) {
	fakeCLI(t, "claude", `cat <<'EOF'
{"type":"result","subtype":"success","is_error":false,"result":"found the slop","total_cost_usd":0.25,
 "modelUsage":{"claude-sonnet-5":{"inputTokens":100,"outputTokens":20,"cacheReadInputTokens":3000,"cacheCreationInputTokens":400}}}
EOF`)
	r := Runner{Harness: harness.Claude, Model: "m"}
	res, err := r.Run(context.Background(), fixtureScenario(t), benchSkill("s"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Err != nil {
		t.Fatalf("Result.Err = %v", res.Err)
	}
	if res.Stdout != "found the slop" {
		t.Errorf("stdout = %q, want the result text, not the raw JSON", res.Stdout)
	}
	want := Usage{Input: 100, CacheRead: 3000, CacheWrite: 400, Output: 20, CostUSD: 0.25}
	if res.Usage == nil || *res.Usage != want {
		t.Errorf("usage = %+v, want %+v", res.Usage, want)
	}
}

func TestRunnerFailsOnNonJSONClaudeOutput(t *testing.T) {
	fakeCLI(t, "claude", `echo "plain text"`)
	r := Runner{Harness: harness.Claude, Model: "m"}
	res, err := r.Run(context.Background(), fixtureScenario(t), benchSkill("s"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Err == nil {
		t.Fatal("want Result.Err when claude prints no JSON result")
	}
	if !strings.Contains(res.Stdout, "plain text") || res.Usage != nil {
		t.Errorf("stdout = %q, usage = %v; want the raw output kept and no usage", res.Stdout, res.Usage)
	}
}

func TestRunnerDiffSurvivesHarnessCommit(t *testing.T) {
	// A run that commits its work (git-mode skills) moves HEAD; the post-run
	// diff is pinned to the materialized tip, so the change must still show.
	fakeClaude(t, `printf 'package main // committed\n' > main.go
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_SYSTEM=/dev/null
git add -A
git -c user.name=h -c user.email=h@h commit -q -m done`)
	r := Runner{Harness: harness.Claude, Model: "m"}
	res, err := r.Run(context.Background(), fixtureScenario(t), benchSkill("s"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Err != nil {
		t.Fatalf("Result.Err = %v", res.Err)
	}
	if !strings.Contains(res.Diff, "committed") || !strings.Contains(res.Diff, "main.go") {
		t.Errorf("diff = %q, want the committed change captured", res.Diff)
	}
}

func TestRunnerRecordsFailure(t *testing.T) {
	fakeClaude(t, `echo "boom" >&2; exit 3`)
	r := Runner{Harness: harness.Claude, Model: "m"}
	res, err := r.Run(context.Background(), fixtureScenario(t), benchSkill("s"), nil)
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
	res, err := r.Run(context.Background(), fixtureScenario(t), benchSkill("s"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Err == nil || !strings.Contains(res.Err.Error(), "timed out") {
		t.Errorf("Result.Err = %v, want timeout", res.Err)
	}
}

func TestRunnerScenarioTimeoutPrecedence(t *testing.T) {
	tests := []struct {
		name     string
		runner   time.Duration
		scenario time.Duration
		want     string
	}{
		{"scenario timeout replaces the default", 0, 100 * time.Millisecond, "timed out after 100ms"},
		{"runner timeout beats the scenario", 100 * time.Millisecond, 10 * time.Second, "timed out after 100ms"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClaude(t, `sleep 5`)
			s := fixtureScenario(t)
			s.Timeout = tt.scenario
			r := Runner{Harness: harness.Claude, Model: "m", Timeout: tt.runner}
			res, err := r.Run(context.Background(), s, benchSkill("s"), nil)
			if err != nil {
				t.Fatal(err)
			}
			if res.Err == nil || !strings.Contains(res.Err.Error(), tt.want) {
				t.Errorf("Result.Err = %v, want %q", res.Err, tt.want)
			}
		})
	}
}

func TestRunnerInstallsExtraSkills(t *testing.T) {
	fakeClaude(t, `cat .claude/skills/ds-x/SKILL.md .claude/skills/ds-y/SKILL.md`)
	extra := SkillVersion{Name: "ds-y", Files: map[string][]byte{"SKILL.md": []byte("EXTRABODY\n")}}
	r := Runner{Harness: harness.Claude, Model: "m"}
	res, err := r.Run(context.Background(), fixtureScenario(t), benchSkill("SKILLBODY\n"), []SkillVersion{extra})
	if err != nil {
		t.Fatal(err)
	}
	if res.Err != nil {
		t.Fatalf("Result.Err = %v", res.Err)
	}
	if !strings.Contains(res.Stdout, "SKILLBODY") || !strings.Contains(res.Stdout, "EXTRABODY") {
		t.Errorf("stdout = %q, want both the skill under test and the extra installed", res.Stdout)
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
	res, err := r.Run(context.Background(), fixtureScenario(t), benchSkill("s"), nil)
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
	res, err := r.Run(context.Background(), fixtureScenario(t), benchSkill("S"), nil)
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
[ -d "$OPENCODE_CONFIG_DIR" ] && [ -z "$(ls -A "$OPENCODE_CONFIG_DIR")" ] && [ "$OPENCODE_DISABLE_CLAUDE_CODE" = 1 ] && echo ISOLATED
cat .opencode/skills/ds-x/SKILL.md`)
	r := Runner{Harness: harness.OpenCode, Model: "anthropic/some-model"}
	res, err := r.Run(context.Background(), fixtureScenario(t), benchSkill("OCSKILL"), nil)
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
	for _, want := range []string{"run", "Review the diff", "--model", "anthropic/some-model", "--pure", "--auto"} {
		if !strings.Contains(string(args), want) {
			t.Errorf("opencode args = %q, missing %q", args, want)
		}
	}
	// The run must see an empty OPENCODE_CONFIG_DIR and the compat fallback
	// disabled, so the operator's global config never moves scores.
	if !strings.Contains(res.Stdout, "ISOLATED") {
		t.Errorf("stdout = %q, want the isolation env marker", res.Stdout)
	}
	if !strings.Contains(res.Stdout, "OCSKILL") {
		t.Errorf("stdout = %q, want the installed skill under .opencode/skills", res.Stdout)
	}
}
