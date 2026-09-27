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
		{"price without date", "models:\n  codex: m\nprices:\n  m:\n    input: 1\n", "checked must be a YYYY-MM-DD date"},
		{"price with bad date", "models:\n  codex: m\nprices:\n  m:\n    checked: \"last week\"\n", "checked must be a YYYY-MM-DD date"},
		{"negative rate", "models:\n  codex: m\nprices:\n  m:\n    checked: \"2026-09-26\"\n    output: -1\n", "negative rate"},
		{"unknown price field", "models:\n  codex: m\nprices:\n  m:\n    checked: \"2026-09-26\"\n    cached: 1\n", "cached"},
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

func TestLoadConfigPrices(t *testing.T) {
	c, err := LoadConfig(writeConfig(t, "models:\n  codex: m\nprices:\n  m:\n    checked: \"2026-09-26\"\n    input: 2\n    cache_read: 0.2\n    cache_write: 2.5\n    output: 12\n"))
	if err != nil {
		t.Fatal(err)
	}
	p, ok := c.Price("m")
	want := Price{Checked: "2026-09-26", Input: 2, CacheRead: 0.2, CacheWrite: 2.5, Output: 12}
	if !ok || p != want {
		t.Errorf("Price(m) = %+v, %v; want %+v", p, ok, want)
	}
	if _, ok := c.Price("other"); ok {
		t.Error("Price(other) found, want no entry")
	}
}

func TestCommittedBenchConfig(t *testing.T) {
	if _, err := LoadConfig("../../evals/bench.yaml"); err != nil {
		t.Fatal(err)
	}
}
