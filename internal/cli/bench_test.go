package cli

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	fakeHarnessCLI(t, "claude", claudeJSON(script))
}

// codexJSON is the event stream a fake codex prints: one agent message
// carrying text, then a completed turn.
func codexJSON(text string) string {
	return `echo '{"type":"item.completed","item":{"type":"agent_message","text":"` + text + `"}}'
echo '{"type":"turn.completed","usage":{"input_tokens":1000000,"cached_input_tokens":0,"output_tokens":0}}'`
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
	if strings.Count(out.String(), "score: 1/1 hits, 0 extra") != 2 {
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

// A reproduction that drops the filter would re-run the wrong scenario set
// (NFR-3), so the pr-md Reproduce line must carry --scenario.
func TestRunBenchScenarioFilterInRepro(t *testing.T) {
	root := benchRoot(t, "alpha", "beta")
	fakeClaudeCLI(t, `echo ok`)
	var out strings.Builder
	opts := benchOptions{Skill: "ds-x", Scenario: "beta", Runs: 1, Format: "pr-md"}
	if err := runBench(context.Background(), &out, io.Discard, root, opts); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "bench ds-x --scenario beta --harness claude --runs 1") {
		t.Errorf("report = %q, want the repro to carry --scenario beta", out.String())
	}
}

// Skill and scenario names are joined into paths and git specs; anything
// path-shaped is rejected before touching the filesystem.
func TestRunBenchRejectsPathyNames(t *testing.T) {
	for _, opts := range []benchOptions{
		{Skill: "../../etc", Runs: 1},
		{Skill: "ds-x", Scenario: "a/b", Runs: 1},
	} {
		err := runBench(context.Background(), io.Discard, io.Discard, t.TempDir(), opts)
		if err == nil || !strings.Contains(err.Error(), "bare directory name") {
			t.Errorf("opts %+v: err = %v, want bare-name rejection", opts, err)
		}
	}
}

// claude runs approvals-off and unconfined; the warning must land on stderr
// before any run — and not for Codex, whose sandbox actually confines.
func TestRunBenchWarnsApprovalsOff(t *testing.T) {
	root := benchRoot(t, "alpha")
	fakeClaudeCLI(t, `echo ok`)
	var out, errOut strings.Builder
	if err := runBench(context.Background(), &out, &errOut, root, benchOptions{Skill: "ds-x", Runs: 1}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errOut.String(), "approvals-off") {
		t.Errorf("errOut = %q, want the approvals-off warning", errOut.String())
	}
	if strings.Contains(out.String(), "approvals-off") {
		t.Error("warning must go to stderr, not the product stream")
	}
}

func TestRunBenchNoWarningForCodexOnly(t *testing.T) {
	root := benchRoot(t, "alpha")
	fakeHarnessCLI(t, "codex", codexJSON("ok"))
	var errOut strings.Builder
	if err := runBench(context.Background(), io.Discard, &errOut, root, benchOptions{Skill: "ds-x", Runs: 1, Harness: "codex"}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(errOut.String(), "approvals-off") {
		t.Errorf("errOut = %q, codex runs sandboxed — no warning owed", errOut.String())
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
	fakeHarnessCLI(t, "codex", codexJSON("ok"))
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

func TestRunBenchPricesCodexFromConfig(t *testing.T) {
	tests := []struct {
		name     string
		prices   string
		want     []string
		wantNone []string
	}{
		{
			name:   "priced model",
			prices: "prices:\n  codex-pin:\n    checked: \"2026-09-26\"\n    input: 2\n    output: 12\n",
			want:   []string{"Cost at list price checked 2026-09-26", "| 1 | 0/1 hits, 0 extra · $2.0000 | 0/1 hits, 0 extra · $2.0000 |", "| **median cost** | $2.0000 | $2.0000 |"},
		},
		{
			name:     "unpriced model",
			want:     []string{"cost unknown"},
			wantNone: []string{"Cost at", "median cost"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := benchRoot(t, "alpha")
			writeFile(t, root, "evals/bench.yaml", "models:\n  codex: codex-pin\n"+tt.prices)
			fakeHarnessCLI(t, "codex", codexJSON("ok"))
			var out strings.Builder
			opts := benchOptions{Skill: "ds-x", Runs: 1, Harness: "codex", Format: "pr-md"}
			if err := runBench(context.Background(), &out, io.Discard, root, opts); err != nil {
				t.Fatal(err)
			}
			got := out.String()
			for _, w := range tt.want {
				if !strings.Contains(got, w) {
					t.Errorf("report missing %q:\n%s", w, got)
				}
			}
			for _, w := range tt.wantNone {
				if strings.Contains(got, w) {
					t.Errorf("report has %q:\n%s", w, got)
				}
			}
		})
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
	if !strings.Contains(errOut.String(), "== ds-x/alpha") || !strings.Contains(errOut.String(), "1/1 hits") {
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

func TestRunBenchReproCarriesTimeout(t *testing.T) {
	root := benchRoot(t, "alpha")
	fakeClaudeCLI(t, `echo ok`)
	var out strings.Builder
	if err := runBench(context.Background(), &out, io.Discard, root, benchOptions{Skill: "ds-x", Runs: 1, Format: "pr-md", Timeout: 2 * time.Minute}); err != nil {
		t.Fatal(err)
	}
	// NFR-3: a non-default timeout shapes which runs fail, so the repro
	// command must carry it.
	if !strings.Contains(out.String(), "--timeout 2m0s") {
		t.Errorf("report = %q, want the non-default --timeout in the repro command", out.String())
	}
}

func TestRunBenchInstallsScenarioSkills(t *testing.T) {
	root := benchRoot(t, "alpha")
	writeFile(t, root, "skills/ds-y/SKILL.md", "EXTRASKILL\n")
	writeFile(t, root, "evals/ds-x/alpha/expectations.yaml", `task: "Do the thing"
tier: smoke
skills: [ds-y]
`)
	fakeClaudeCLI(t, `cat .claude/skills/ds-y/SKILL.md`)
	var out strings.Builder
	if err := runBench(context.Background(), &out, io.Discard, root, benchOptions{Skill: "ds-x", Runs: 1}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "EXTRASKILL") {
		t.Errorf("output = %q, want the scenario's declared skill installed in the sandbox", out.String())
	}
}

func TestRunBenchUnknownScenarioSkillFailsBeforeRuns(t *testing.T) {
	root := benchRoot(t, "alpha")
	writeFile(t, root, "evals/ds-x/alpha/expectations.yaml", "task: t\ntier: smoke\nskills: [ds-missing]\n")
	marker := filepath.Join(t.TempDir(), "ran")
	t.Setenv("MARKER", marker)
	fakeClaudeCLI(t, `touch "$MARKER"`)
	err := runBench(context.Background(), &strings.Builder{}, io.Discard, root, benchOptions{Skill: "ds-x", Runs: 1})
	if err == nil || !strings.Contains(err.Error(), "ds-missing") {
		t.Errorf("error = %v, want the missing skill named", err)
	}
	if _, statErr := os.Stat(marker); statErr == nil {
		t.Error("harness ran despite the scenario naming an unknown skill")
	}
}

func TestRunBenchDedupesHarnesses(t *testing.T) {
	root := benchRoot(t, "alpha")
	fakeClaudeCLI(t, `echo ok`)
	var out strings.Builder
	if err := runBench(context.Background(), &out, io.Discard, root, benchOptions{Skill: "ds-x", Runs: 1, Harness: "claude,claude"}); err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(out.String(), "== ds-x/alpha"); n != 2 {
		t.Errorf("run headers = %d, want 2 — a duplicated --harness entry must not double the runs", n)
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

// blocksRoot adds agents-md blocks to a benchRoot: base and the go profile
// committed on main then diverged, and concise new in the working tree.
func blocksRoot(t *testing.T) string {
	t.Helper()
	root := benchRoot(t, "alpha")
	writeFile(t, root, "agents-md/system/agents-base.md", "OLDBASE\n")
	writeFile(t, root, "agents-md/language/go.md", "OLDGO\n")
	gitRun(t, root, "add", "agents-md")
	gitRun(t, root, "commit", "-q", "-m", "blocks")
	writeFile(t, root, "agents-md/system/agents-base.md", "NEWBASE\n")
	writeFile(t, root, "agents-md/language/go.md", "NEWGO\n")
	writeFile(t, root, "agents-md/system/concise.md", "NEWCONCISE\n")
	return root
}

func TestRunBenchInstallsBlocksPerVersion(t *testing.T) {
	root := blocksRoot(t)
	fakeClaudeCLI(t, `cat AGENTS.md`)
	var out strings.Builder
	opts := benchOptions{Skill: "ds-x", Runs: 1, Format: "pr-md", Blocks: "go,concise,base"}
	if err := runBench(context.Background(), &out, io.Discard, root, opts); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	oldRun, newRun, ok := strings.Cut(got, "#### new run 1")
	if !ok {
		t.Fatalf("report has no new-run transcript:\n%s", got)
	}
	if !strings.Contains(oldRun, "OLDBASE") || !strings.Contains(oldRun, "OLDGO") || strings.Contains(oldRun, "NEWCONCISE") {
		t.Errorf("old run should get main's blocks and no branch-only block:\n%s", oldRun)
	}
	iBase, iConcise, iGo := strings.Index(newRun, "NEWBASE"), strings.Index(newRun, "NEWCONCISE"), strings.Index(newRun, "NEWGO")
	if iBase < 0 || iConcise < iBase || iGo < iConcise {
		t.Errorf("new run should get working-tree blocks in init's order (base, layers, languages):\n%s", newRun)
	}
	blocks := regexp.MustCompile("- Blocks, each from its version's tree:\n" +
		"  - `base`: old `[0-9a-f]{40}`, new `[0-9a-f]{40}`\n" +
		"  - `concise`: new `[0-9a-f]{40}` \\(absent on the main branch, so old runs go without\\)\n" +
		"  - `language:go`: old `[0-9a-f]{40}`, new `[0-9a-f]{40}`\n")
	if !blocks.MatchString(got) {
		t.Errorf("report should list each block with its SHAs in init's order:\n%s", got)
	}
	if want := "--blocks base,concise,go --format pr-md"; !strings.Contains(got, want) {
		t.Errorf("report missing %q:\n%s", want, got)
	}
}

func TestRunBenchUnknownBlockFailsBeforeRuns(t *testing.T) {
	root := blocksRoot(t)
	fakeClaudeCLI(t, `echo ran`)
	var out strings.Builder
	err := runBench(context.Background(), &out, io.Discard, root, benchOptions{Skill: "ds-x", Runs: 1, Blocks: "base,cobol"})
	if err == nil || !strings.Contains(err.Error(), `unknown block "cobol"`) || !strings.Contains(err.Error(), "go") {
		t.Errorf("want an unknown-block error listing the available blocks, got %v", err)
	}
	if strings.Contains(out.String(), "ran") {
		t.Error("a bad block name must fail before any run spends tokens")
	}
}
