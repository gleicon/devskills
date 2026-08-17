package bench

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// reportScenario returns a loaded report-style scenario with two expectations
// and fixture files main.go, greet.go, other.go.
func reportScenario(t *testing.T) *Scenario {
	t.Helper()
	dir := writeScenario(t, "s", `task: t
tier: planted-defect
style: report
expectations:
  - file: greet.go
    keywords: [narrating comment, restates]
  - file: main.go
    keywords: [dead guard]
`, "base", "change")
	for _, f := range []string{"base/main.go", "base/greet.go", "base/other.go", "change/main.go", "change/greet.go", "change/other.go"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("package main\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	s, err := LoadScenario(dir)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestCheckReportHitsAndMisses(t *testing.T) {
	s := reportScenario(t)
	tests := []struct {
		name   string
		stdout string
		want   []bool
	}{
		{"both hit", "greet.go has a Narrating Comment.\nmain.go: dead guard removed.\n", []bool{true, true}},
		{"keyword case-insensitive", "greet.go: RESTATES the code\n", []bool{true, false}},
		{"file without keyword misses", "greet.go looks fine\n", []bool{false, false}},
		{"keyword without file misses", "there is a narrating comment somewhere\n", []bool{false, false}},
		{"empty output", "", []bool{false, false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := Check(s, Result{Stdout: tt.stdout})
			if err != nil {
				t.Fatal(err)
			}
			for i := range tt.want {
				if c.Hits[i] != tt.want[i] {
					t.Errorf("Hits[%d] = %v, want %v", i, c.Hits[i], tt.want[i])
				}
			}
		})
	}
}

func TestCheckReportExtras(t *testing.T) {
	s := reportScenario(t)
	stdout := "greet.go has a narrating comment\n" + // expected finding, not extra
		"other.go: unused helper here\n" + // fixture file, no expectation -> extra
		"greet.go: also has needless nesting\n" + // expected file, no expected keyword -> extra
		"unrelated chatter with no file mention\n"
	c, err := Check(s, Result{Stdout: stdout})
	if err != nil {
		t.Fatal(err)
	}
	if c.HitCount() != 1 {
		t.Errorf("HitCount = %d, want 1", c.HitCount())
	}
	if len(c.Extras) != 2 {
		t.Fatalf("Extras = %q, want 2", c.Extras)
	}
	if !strings.Contains(c.Extras[0], "other.go") || !strings.Contains(c.Extras[1], "needless nesting") {
		t.Errorf("Extras = %q", c.Extras)
	}
}

// applyScenario mirrors the committed narrated-greeting shape: anchors on
// planted lines in greet.go and main.go.
func applyScenario(t *testing.T) *Scenario {
	t.Helper()
	dir := writeScenario(t, "s", `task: t
tier: planted-defect
style: apply
expectations:
  - file: greet.go
    anchors: ["// First we get the greeting"]
  - file: main.go
    anchors: ['if g != "" {']
`, "base", "change")
	for _, f := range []string{"base/x.go", "change/x.go"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("package main\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	s, err := LoadScenario(dir)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

const applyDiff = `diff --git a/greet.go b/greet.go
index 111..222 100644
--- a/greet.go
+++ b/greet.go
@@ -1,5 +1,3 @@
 func shout(name string) string {
-	// First we get the greeting for the name.
 	g := greeting(name)
diff --git a/notes.md b/notes.md
new file mode 100644
--- /dev/null
+++ b/notes.md
@@ -0,0 +1 @@
+harness scratch file
`

func TestCheckApply(t *testing.T) {
	s := applyScenario(t)
	c, err := Check(s, Result{Diff: applyDiff})
	if err != nil {
		t.Fatal(err)
	}
	if !c.Hits[0] {
		t.Error("removed anchor line in greet.go must hit")
	}
	if c.Hits[1] {
		t.Error("main.go untouched, must miss")
	}
	if len(c.Extras) != 1 || c.Extras[0] != "notes.md" {
		t.Errorf("Extras = %q, want the unexpected changed file", c.Extras)
	}
}

func TestCheckApplyRewrittenLineHits(t *testing.T) {
	s := applyScenario(t)
	diff := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,3 +1,3 @@
-	if g != "" {
+	if g == "" { // inverted guard, still a change to the planted line
`
	c, err := Check(s, Result{Diff: diff})
	if err != nil {
		t.Fatal(err)
	}
	if !c.Hits[1] {
		t.Error("rewritten anchor line must hit")
	}
	if c.Hits[0] || len(c.Extras) != 0 {
		t.Errorf("Hits[0] = %v, Extras = %q", c.Hits[0], c.Extras)
	}
}

func TestCheckApplyEmptyDiffMissesAll(t *testing.T) {
	c, err := Check(applyScenario(t), Result{})
	if err != nil {
		t.Fatal(err)
	}
	if c.HitCount() != 0 || len(c.Extras) != 0 {
		t.Errorf("empty diff: HitCount = %d, Extras = %q", c.HitCount(), c.Extras)
	}
}

func TestCheckUnsupportedTier(t *testing.T) {
	dir := writeScenario(t, "s", "task: t\ntier: smoke\n", "base", "change")
	s, err := LoadScenario(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Check(s, Result{}); err == nil || !strings.Contains(err.Error(), "smoke") {
		t.Errorf("error = %v, want unsupported-tier", err)
	}
}
