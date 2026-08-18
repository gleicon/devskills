package cli

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_CONFIG_SYSTEM="+os.DevNull,
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// benchRoot builds a git repo root with one committed skill (then diverged in
// the working tree, so old and new versions differ), a bench config, and the
// named scenarios.
func benchRoot(t *testing.T, scenarios ...string) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, "skills/ds-x/SKILL.md", "OLDSKILL\n")
	writeFile(t, root, "evals/bench.yaml", "models:\n  claude: pinned-model\n  codex: codex-pin\n  opencode: oc-pin\n")
	for _, s := range scenarios {
		dir := "evals/ds-x/" + s + "/"
		writeFile(t, root, dir+"expectations.yaml", `task: "Do the thing"
tier: planted-defect
style: report
expectations:
  - file: main.go
    keywords: [slop]
`)
		writeFile(t, root, dir+"base/main.go", "package main // v1\n")
		writeFile(t, root, dir+"change/main.go", "package main // v2\n")
	}
	gitRun(t, root, "init", "-q", "-b", "main")
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "commit", "-q", "-m", "base")
	writeFile(t, root, "skills/ds-x/SKILL.md", "NEWSKILL\n")
	return root
}

func fakeHarnessCLI(t *testing.T, name, script string) {
	t.Helper()
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func fakeClaudeCLI(t *testing.T, script string) {
	t.Helper()
	fakeHarnessCLI(t, "claude", script)
}

func TestRunBenchOldVsNew(t *testing.T) {
	root := benchRoot(t, "alpha")
	// The fake harness prints the installed skill, proving each version's
	// content reached its sandbox.
	fakeClaudeCLI(t, `cat .claude/skills/ds-x/SKILL.md`)
	var out strings.Builder
	if err := runBench(context.Background(), &out, io.Discard, root, benchOptions{Skill: "ds-x", Runs: 2}); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{
		"old run 1/2", "old run 2/2", "new run 1/2", "new run 2/2",
		"OLDSKILL", "NEWSKILL",
		"model pinned-model", "-- stdout --", "-- stderr --", "-- diff --",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q:\n%s", want, got)
		}
	}
	if strings.Count(got, "== ds-x/alpha") != 4 {
		t.Errorf("run headers = %d, want 2 versions x 2 runs", strings.Count(got, "== ds-x/alpha"))
	}
	if strings.Contains(got, "baseline mode") {
		t.Error("baseline mode announced for a skill present on main")
	}
}

func TestRunBenchBaselineMode(t *testing.T) {
	root := benchRoot(t, "alpha")
	// ds-fresh exists only in the working tree.
	writeFile(t, root, "skills/ds-fresh/SKILL.md", "FRESH\n")
	writeFile(t, root, "evals/ds-fresh/s1/expectations.yaml", "task: t\ntier: smoke\n")
	writeFile(t, root, "evals/ds-fresh/s1/base/main.go", "package main // v1\n")
	writeFile(t, root, "evals/ds-fresh/s1/change/main.go", "package main // v2\n")
	fakeClaudeCLI(t, `echo ok`)
	var out strings.Builder
	if err := runBench(context.Background(), &out, io.Discard, root, benchOptions{Skill: "ds-fresh", Runs: 1}); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "baseline mode") {
		t.Errorf("output = %q, want baseline mode announced", got)
	}
	if strings.Contains(got, "old run") || strings.Count(got, "== ds-fresh/s1") != 1 {
		t.Errorf("output = %q, want a single new-version run", got)
	}
}

func TestRunBenchScoresPlantedDefectRuns(t *testing.T) {
	root := benchRoot(t, "alpha")
	fakeClaudeCLI(t, `echo "main.go: slop found"`)
	var out strings.Builder
	if err := runBench(context.Background(), &out, io.Discard, root, benchOptions{Skill: "ds-x", Runs: 1}); err != nil {
		t.Fatal(err)
	}
	if strings.Count(out.String(), "hits: 1/1, extra findings: 0") != 2 {
		t.Errorf("output = %q, want a score line per version run", out.String())
	}
}

func TestRunBenchScenarioFilter(t *testing.T) {
	root := benchRoot(t, "alpha", "beta")
	fakeClaudeCLI(t, `echo ok`)
	var out strings.Builder
	if err := runBench(context.Background(), &out, io.Discard, root, benchOptions{Skill: "ds-x", Scenario: "beta", Runs: 1}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "ds-x/alpha") || !strings.Contains(out.String(), "ds-x/beta") {
		t.Errorf("output = %q, want only beta", out.String())
	}
}

func TestRunBenchModelOverride(t *testing.T) {
	root := benchRoot(t, "alpha")
	fakeClaudeCLI(t, `echo ok`)
	var out strings.Builder
	if err := runBench(context.Background(), &out, io.Discard, root, benchOptions{Skill: "ds-x", Model: "override-model", Runs: 1}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "model override-model") {
		t.Errorf("output = %q, want the override model", out.String())
	}
}

func TestRunBenchAllRunsFailed(t *testing.T) {
	root := benchRoot(t, "alpha")
	fakeClaudeCLI(t, `exit 1`)
	var out strings.Builder
	err := runBench(context.Background(), &out, io.Discard, root, benchOptions{Skill: "ds-x", Runs: 1})
	if err == nil || !strings.Contains(err.Error(), "all 2 runs failed") {
		t.Errorf("error = %v, want all-failed over both versions", err)
	}
	if !strings.Contains(out.String(), "run failed:") {
		t.Errorf("output = %q, want failures reported inline", out.String())
	}
}

func TestRunBenchPartialFailureExitsZero(t *testing.T) {
	root := benchRoot(t, "alpha")
	// Fail only the run against the old skill version.
	fakeClaudeCLI(t, `grep -q OLDSKILL .claude/skills/ds-x/SKILL.md && exit 1
echo ok`)
	var out strings.Builder
	if err := runBench(context.Background(), &out, io.Discard, root, benchOptions{Skill: "ds-x", Runs: 1}); err != nil {
		t.Fatalf("partial failure must not fail the command: %v", err)
	}
	if !strings.Contains(out.String(), "run failed:") {
		t.Errorf("output = %q, want the old-version failure reported", out.String())
	}
}

func TestRunBenchHarnessFanOut(t *testing.T) {
	root := benchRoot(t, "alpha")
	fakeClaudeCLI(t, `echo ok`)
	fakeHarnessCLI(t, "codex", `echo ok`)
	var out strings.Builder
	opts := benchOptions{Skill: "ds-x", Runs: 1, Harness: "claude,codex", Format: "pr-md"}
	if err := runBench(context.Background(), &out, io.Discard, root, opts); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{
		"## Claude Code — model `pinned-model`",
		"## OpenAI Codex — model `codex-pin`",
		"--harness claude,codex --runs 1",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("report missing %q:\n%s", want, got)
		}
	}
	// Two harnesses: --model can't carry both pins, so repro omits it.
	if strings.Contains(got, "--runs 1 --model") {
		t.Error("multi-harness repro must not pick one harness's model")
	}
}

func TestRunBenchMissingHarnessCLIRecorded(t *testing.T) {
	root := benchRoot(t, "alpha")
	if _, err := exec.LookPath("opencode"); err == nil {
		t.Skip("opencode installed on this machine; the missing-CLI path can't be exercised")
	}
	fakeClaudeCLI(t, `echo ok`)
	// opencode is not on PATH: its runs must fail loudly, not vanish.
	var out strings.Builder
	opts := benchOptions{Skill: "ds-x", Runs: 1, Harness: "claude,opencode"}
	if err := runBench(context.Background(), &out, io.Discard, root, opts); err != nil {
		t.Fatalf("claude runs succeeded, command must exit zero: %v", err)
	}
	if !strings.Contains(out.String(), "run failed:") {
		t.Errorf("output = %q, want the missing-CLI failure recorded", out.String())
	}
}

func TestRunBenchUnknownHarness(t *testing.T) {
	err := runBench(context.Background(), &strings.Builder{}, io.Discard, t.TempDir(), benchOptions{Skill: "ds-x", Runs: 1, Harness: "gemini"})
	if err == nil || !strings.Contains(err.Error(), "gemini") {
		t.Errorf("error = %v, want unknown-harness", err)
	}
}

func TestRunBenchPrMdFormat(t *testing.T) {
	root := benchRoot(t, "alpha")
	fakeClaudeCLI(t, `echo "main.go: slop found"`)
	var out, errOut strings.Builder
	if err := runBench(context.Background(), &out, &errOut, root, benchOptions{Skill: "ds-x", Runs: 1, Format: "pr-md"}); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{
		"# Bench report: ds-x",
		"| run | old | new |",
		"Claude Code — model `pinned-model`",
		"Reproduce: `devskills bench ds-x --harness claude --runs 1 --model pinned-model --format pr-md`",
		"<details>",
		"1/1 hits",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("report missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "== ds-x/") || strings.Contains(got, "-- stdout --") {
		t.Error("pr-md to stdout must not interleave streaming output")
	}
	// pr-md progress is diagnostics: run headers and scores stream to stderr.
	if !strings.Contains(errOut.String(), "== ds-x/alpha") || !strings.Contains(errOut.String(), "hits: 1/1") {
		t.Errorf("stderr = %q, want run headers and score lines streamed", errOut.String())
	}
	if !strings.Contains(got, "(main branch), new `") {
		t.Error("report missing version SHAs")
	}
}

func TestRunBenchPrMdOut(t *testing.T) {
	root := benchRoot(t, "alpha")
	fakeClaudeCLI(t, `echo ok`)
	outPath := filepath.Join(t.TempDir(), "report.md")
	var out, errOut strings.Builder
	if err := runBench(context.Background(), &out, &errOut, root, benchOptions{Skill: "ds-x", Runs: 1, Format: "pr-md", Out: outPath}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "# Bench report: ds-x") {
		t.Errorf("report file = %q", b)
	}
	// With --out, the file gets the report, stdout the written-to note, and
	// progress streams to stderr like every pr-md run.
	if !strings.Contains(out.String(), "report written to") || strings.Contains(out.String(), "== ds-x/alpha") {
		t.Errorf("stdout = %q, want only the written-to note", out.String())
	}
	if !strings.Contains(errOut.String(), "== ds-x/alpha") {
		t.Errorf("stderr = %q, want progress lines", errOut.String())
	}
	if strings.Contains(out.String(), "# Bench report") {
		t.Error("report must not also go to stdout when --out is set")
	}
}

func TestRunBenchRejectsUnknownFormat(t *testing.T) {
	err := runBench(context.Background(), &strings.Builder{}, io.Discard, t.TempDir(), benchOptions{Skill: "ds-x", Runs: 1, Format: "html"})
	if err == nil || !strings.Contains(err.Error(), "--format") {
		t.Errorf("error = %v, want format validation", err)
	}
}

func TestRunBenchUnknownSkill(t *testing.T) {
	root := benchRoot(t, "alpha")
	err := runBench(context.Background(), &strings.Builder{}, io.Discard, root, benchOptions{Skill: "ds-nope", Runs: 1})
	if err == nil || !strings.Contains(err.Error(), "ds-nope") {
		t.Errorf("error = %v, want it to name the missing skill", err)
	}
}

func TestRunBenchRejectsBadRuns(t *testing.T) {
	err := runBench(context.Background(), &strings.Builder{}, io.Discard, t.TempDir(), benchOptions{Skill: "ds-x", Runs: 0})
	if err == nil || !strings.Contains(err.Error(), "--runs") {
		t.Errorf("error = %v, want runs validation", err)
	}
}

func TestRunBenchRejectsNegativeTimeout(t *testing.T) {
	err := runBench(context.Background(), &strings.Builder{}, io.Discard, t.TempDir(), benchOptions{Skill: "ds-x", Runs: 1, Timeout: -time.Second})
	if err == nil || !strings.Contains(err.Error(), "--timeout") {
		t.Errorf("error = %v, want timeout validation", err)
	}
}

func TestRunBenchTimeoutFlagBoundsRuns(t *testing.T) {
	root := benchRoot(t, "alpha")
	fakeClaudeCLI(t, `sleep 5`)
	var out strings.Builder
	err := runBench(context.Background(), &out, io.Discard, root, benchOptions{Skill: "ds-x", Runs: 1, Timeout: 100 * time.Millisecond})
	if err == nil || !strings.Contains(err.Error(), "all 2 runs failed") {
		t.Errorf("error = %v, want every run timed out", err)
	}
	if !strings.Contains(out.String(), "timed out after 100ms") {
		t.Errorf("output = %q, want the flag's timeout reported", out.String())
	}
}

func TestRunBenchInterruptedContextAborts(t *testing.T) {
	root := benchRoot(t, "alpha")
	fakeClaudeCLI(t, `echo ok`)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out strings.Builder
	err := runBench(ctx, &out, io.Discard, root, benchOptions{Skill: "ds-x", Runs: 3})
	if err == nil || !strings.Contains(err.Error(), "interrupted") {
		t.Errorf("error = %v, want the bench aborted on cancellation", err)
	}
	// One aborted run at most — never the full grind of fake failures.
	if n := strings.Count(out.String(), "== ds-x/alpha"); n > 1 {
		t.Errorf("run headers = %d, want the bench to stop at the first canceled run", n)
	}
}
