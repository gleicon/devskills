package doctor

import (
	"slices"
	"testing"
)

func TestInstallerPicksFirstApplicable(t *testing.T) {
	tool := Tool{
		Name: "x",
		Installers: []Installer{
			{Requires: "brew", GOOS: "darwin", Command: []string{"brew", "install", "x"}},
			{Requires: "go", Command: []string{"go", "install", "x@latest"}},
		},
	}
	has := func(present ...string) func(string) bool {
		return func(b string) bool { return slices.Contains(present, b) }
	}
	tests := []struct {
		name    string
		goos    string
		present []string
		wantCmd []string
		wantOK  bool
	}{
		{"darwin with brew takes brew", "darwin", []string{"brew", "go"}, []string{"brew", "install", "x"}, true},
		{"darwin without brew falls to go", "darwin", []string{"go"}, []string{"go", "install", "x@latest"}, true},
		{"linux skips darwin brew, takes go", "linux", []string{"brew", "go"}, []string{"go", "install", "x@latest"}, true},
		{"nothing applicable", "linux", nil, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in, ok := tool.Installer(tt.goos, has(tt.present...))
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && !slices.Equal(in.Command, tt.wantCmd) {
				t.Errorf("cmd = %v, want %v", in.Command, tt.wantCmd)
			}
		})
	}
}

func TestToolsTableWellFormed(t *testing.T) {
	seen := map[string]bool{}
	for _, tool := range Tools {
		if tool.Name == "" || tool.Skill == "" || tool.DocURL == "" {
			t.Errorf("%+v: missing Name/Skill/DocURL", tool)
		}
		if seen[tool.Name] {
			t.Errorf("duplicate tool %q", tool.Name)
		}
		seen[tool.Name] = true
		if len(tool.Installers) == 0 {
			t.Errorf("%s: no installers", tool.Name)
		}
		for _, in := range tool.Installers {
			if in.Requires == "" || len(in.Command) == 0 {
				t.Errorf("%s: installer missing Requires/Command", tool.Name)
			}
		}
	}
}
