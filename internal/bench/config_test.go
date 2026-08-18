package bench

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gleicon/devskills/internal/harness"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "bench.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadConfig(t *testing.T) {
	c, err := LoadConfig(writeConfig(t, "models:\n  claude: claude-sonnet-5\n"))
	if err != nil {
		t.Fatal(err)
	}
	m, err := c.Model(harness.Claude)
	if err != nil || m != "claude-sonnet-5" {
		t.Errorf("Model(claude) = %q, %v", m, err)
	}
	if _, err := c.Model(harness.Codex); err == nil || !strings.Contains(err.Error(), "codex") {
		t.Errorf("Model(codex) = %v, want missing-pin error", err)
	}
}

func TestLoadConfigRejectsMalformed(t *testing.T) {
	tests := []struct {
		name, yaml, wantErr string
	}{
		{"empty models", "models: {}\n", "models is required"},
		{"unknown harness", "models:\n  gemini: g-1\n", `unknown harness "gemini"`},
		{"empty pin", "models:\n  claude: \"\"\n", "empty model pin"},
		{"unknown field", "models:\n  claude: m\nbogus: x\n", "bogus"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadConfig(writeConfig(t, tt.yaml))
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
	if _, err := LoadConfig(filepath.Join(t.TempDir(), "absent.yaml")); err == nil {
		t.Error("want error for missing config file")
	}
}

func TestCommittedBenchConfig(t *testing.T) {
	if _, err := LoadConfig("../../evals/bench.yaml"); err != nil {
		t.Fatal(err)
	}
}
