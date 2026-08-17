package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// benchRoot builds a repo root with one skill, a bench config, and the named
// scenarios, each a minimal valid planted-defect fixture.
func benchRoot(t *testing.T, scenarios ...string) string {
	t.Helper()
	root := t.TempDir()
	write := func(rel, content string) {
		t.Helper()
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("skills/ds-x/SKILL.md", "SKILLBODY\n")
	write("evals/bench.yaml", "models:\n  claude: pinned-model\n")
	for _, s := range scenarios {
		dir := "evals/ds-x/" + s + "/"
		write(dir+"expectations.yaml", `task: "Do the thing"
tier: planted-defect
style: report
expectations:
  - file: main.go
    keywords: [slop]
`)
		write(dir+"base/main.go", "package main // v1\n")
		write(dir+"change/main.go", "package main // v2\n")
	}
	return root
}

func fakeClaudeCLI(t *testing.T, script string) {
	t.Helper()
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "claude"), []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestRunBench(t *testing.T) {
	root := benchRoot(t, "alpha", "beta")
	fakeClaudeCLI(t, `echo "found slop in main.go"`)
	var out strings.Builder
	if err := runBench(context.Background(), &out, root, "ds-x", "", ""); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if strings.Count(got, "== ds-x/") != 2 {
		t.Errorf("output = %q, want both scenarios run", got)
	}
	for _, want := range []string{"model pinned-model", "-- stdout --", "found slop in main.go", "-- stderr --", "-- diff --"} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q:\n%s", want, got)
		}
	}
}

func TestRunBenchScenarioFilter(t *testing.T) {
	root := benchRoot(t, "alpha", "beta")
	fakeClaudeCLI(t, `echo ok`)
	var out strings.Builder
	if err := runBench(context.Background(), &out, root, "ds-x", "beta", ""); err != nil {
		t.Fatal(err)
	}
	if strings.Count(out.String(), "== ds-x/") != 1 || !strings.Contains(out.String(), "ds-x/beta") {
		t.Errorf("output = %q, want only beta", out.String())
	}
}

func TestRunBenchModelOverride(t *testing.T) {
	root := benchRoot(t, "alpha")
	fakeClaudeCLI(t, `echo ok`)
	var out strings.Builder
	if err := runBench(context.Background(), &out, root, "ds-x", "", "override-model"); err != nil {
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
	err := runBench(context.Background(), &out, root, "ds-x", "", "")
	if err == nil || !strings.Contains(err.Error(), "all 1 runs failed") {
		t.Errorf("error = %v, want all-failed", err)
	}
	if !strings.Contains(out.String(), "run failed:") {
		t.Errorf("output = %q, want the failure reported inline", out.String())
	}
}

func TestRunBenchUnknownSkill(t *testing.T) {
	root := benchRoot(t, "alpha")
	err := runBench(context.Background(), &strings.Builder{}, root, "ds-nope", "", "")
	if err == nil || !strings.Contains(err.Error(), "ds-nope") {
		t.Errorf("error = %v, want it to name the missing skill", err)
	}
}
