package bench

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeScenario builds a scenario dir under a temp root with the given
// expectations.yaml content and optional fixture subdirs.
func writeScenario(t *testing.T, name, yaml string, subdirs ...string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), name)
	for _, sub := range subdirs {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if yaml != "" {
		if err := os.WriteFile(filepath.Join(dir, "expectations.yaml"), []byte(yaml), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

const validReport = `task: "Review the diff"
tier: planted-defect
style: report
expectations:
  - file: main.go
    keywords: [dead code, unused]
`

func TestLoadScenarioValid(t *testing.T) {
	dir := writeScenario(t, "finds-dead-code", validReport, "base", "change")
	s, err := LoadScenario(dir)
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "finds-dead-code" || s.Task != "Review the diff" || s.Tier != TierPlantedDefect || s.Style != StyleReport {
		t.Errorf("loaded scenario = %+v", s)
	}
	if len(s.Expectations) != 1 || s.Expectations[0].File != "main.go" || len(s.Expectations[0].Keywords) != 2 {
		t.Errorf("expectations = %+v", s.Expectations)
	}
}

func TestLoadScenarioApplyStyle(t *testing.T) {
	y := `task: "Fix the code"
tier: planted-defect
style: apply
expectations:
  - file: main.go
    anchors: ["x := compute()"]
`
	if _, err := LoadScenario(writeScenario(t, "s", y, "base", "change")); err != nil {
		t.Fatal(err)
	}
}

func TestLoadScenarioStructuralAndSmoke(t *testing.T) {
	structural := `task: "Produce the doc"
tier: structural
elements: [Summary, Findings]
`
	if _, err := LoadScenario(writeScenario(t, "s", structural, "base", "change")); err != nil {
		t.Fatal(err)
	}
	smoke := `task: "Run it"
tier: smoke
`
	if _, err := LoadScenario(writeScenario(t, "s", smoke, "base", "change")); err != nil {
		t.Fatal(err)
	}
}

func TestLoadScenarioRejectsMalformed(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		subdirs []string
		wantErr string
	}{
		{"missing file", "", []string{"base", "change"}, "expectations.yaml"},
		{"bad yaml", "task: [unclosed", []string{"base", "change"}, "expectations.yaml"},
		{"unknown field", "task: t\ntier: smoke\nbogus: x\n", []string{"base", "change"}, "bogus"},
		{"missing task", "tier: smoke\n", []string{"base", "change"}, "task is required"},
		{"bad tier", "task: t\ntier: nope\n", []string{"base", "change"}, `tier = "nope"`},
		{"missing base dir", validReport, []string{"change"}, "missing base/"},
		{"missing change dir", validReport, []string{"base"}, "missing change/"},
		{"planted-defect without style", "task: t\ntier: planted-defect\nexpectations:\n  - {file: f, keywords: [k]}\n", []string{"base", "change"}, `style = ""`},
		{"planted-defect without expectations", "task: t\ntier: planted-defect\nstyle: report\n", []string{"base", "change"}, "requires expectations"},
		{"expectation without file", "task: t\ntier: planted-defect\nstyle: report\nexpectations:\n  - {keywords: [k]}\n", []string{"base", "change"}, "file is required"},
		{"report without keywords", "task: t\ntier: planted-defect\nstyle: report\nexpectations:\n  - {file: f}\n", []string{"base", "change"}, "requires keywords"},
		{"apply without anchors", "task: t\ntier: planted-defect\nstyle: apply\nexpectations:\n  - {file: f, keywords: [k]}\n", []string{"base", "change"}, "requires anchors"},
		{"structural without elements", "task: t\ntier: structural\n", []string{"base", "change"}, "requires elements"},
		{"structural with expectations", "task: t\ntier: structural\nelements: [E]\nexpectations:\n  - {file: f, keywords: [k]}\n", []string{"base", "change"}, "elements only"},
		{"smoke with elements", "task: t\ntier: smoke\nelements: [E]\n", []string{"base", "change"}, "no style, expectations, or elements"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadScenario(writeScenario(t, "s", tt.yaml, tt.subdirs...))
			if err == nil {
				t.Fatal("want error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err, tt.wantErr)
			}
			if !strings.Contains(err.Error(), "scenario s") {
				t.Errorf("error = %q, want it to name the scenario", err)
			}
		})
	}
}

func TestLoadScenariosSortedByName(t *testing.T) {
	evals := t.TempDir()
	for _, name := range []string{"zeta", "alpha"} {
		dir := filepath.Join(evals, "ds-deslop", name)
		for _, sub := range []string{"base", "change"} {
			if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(filepath.Join(dir, "expectations.yaml"), []byte(validReport), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	scenarios, err := LoadScenarios(evals, "ds-deslop")
	if err != nil {
		t.Fatal(err)
	}
	if len(scenarios) != 2 || scenarios[0].Name != "alpha" || scenarios[1].Name != "zeta" {
		t.Errorf("scenarios = %v, %v", scenarios[0].Name, scenarios[1].Name)
	}
}

func TestLoadScenariosMissingSkill(t *testing.T) {
	if _, err := LoadScenarios(t.TempDir(), "ds-nope"); err == nil || !strings.Contains(err.Error(), "ds-nope") {
		t.Errorf("want error naming the skill, got %v", err)
	}
}
