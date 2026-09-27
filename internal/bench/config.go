package bench

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/goccy/go-yaml"

	"github.com/gleicon/devskills/internal/harness"
)

// Config is the checked-in bench configuration (evals/bench.yaml): the model
// each harness is pinned to, so runs are reproducible without flags, and the
// list prices that turn token counts into cost where an assistant reports
// none.
type Config struct {
	Models map[string]string `yaml:"models"`
	Prices map[string]Price  `yaml:"prices"`
}

// Price is a model's API list price in USD per 1M tokens, standard tier.
type Price struct {
	Checked    string  `yaml:"checked"` // YYYY-MM-DD the rates were read; they change without notice
	Input      float64 `yaml:"input"`
	CacheRead  float64 `yaml:"cache_read"`
	CacheWrite float64 `yaml:"cache_write"`
	Output     float64 `yaml:"output"`
}

// Cost prices a run's tokens.
func (p Price) Cost(u Usage) float64 {
	return (float64(u.Input)*p.Input + float64(u.CacheRead)*p.CacheRead +
		float64(u.CacheWrite)*p.CacheWrite + float64(u.Output)*p.Output) / 1e6
}

func (p Price) String() string {
	return fmt.Sprintf("list price checked %s: $%.2f input, $%.2f cache read, $%.2f cache write, $%.2f output per 1M tokens",
		p.Checked, p.Input, p.CacheRead, p.CacheWrite, p.Output)
}

// LoadConfig reads and validates the bench config at path.
func LoadConfig(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("bench config: %w", err)
	}
	var c Config
	if err := yaml.UnmarshalWithOptions(b, &c, yaml.Strict()); err != nil {
		return Config{}, fmt.Errorf("bench config %s: %w", path, err)
	}
	if len(c.Models) == 0 {
		return Config{}, fmt.Errorf("bench config %s: models is required", path)
	}
	var errs []error
	for id, model := range c.Models {
		if !harness.ID(id).Valid() {
			errs = append(errs, fmt.Errorf("unknown harness %q", id))
		}
		if model == "" {
			errs = append(errs, fmt.Errorf("harness %q has an empty model pin", id))
		}
	}
	for model, p := range c.Prices {
		if _, err := time.Parse(time.DateOnly, p.Checked); err != nil {
			errs = append(errs, fmt.Errorf("price for %q: checked must be a YYYY-MM-DD date, got %q", model, p.Checked))
		}
		if min(p.Input, p.CacheRead, p.CacheWrite, p.Output) < 0 {
			errs = append(errs, fmt.Errorf("price for %q has a negative rate", model))
		}
	}
	if err := errors.Join(errs...); err != nil {
		return Config{}, fmt.Errorf("bench config %s: %w", path, err)
	}
	return c, nil
}

// Model returns the pinned model for a harness.
func (c Config) Model(id harness.ID) (string, error) {
	m, ok := c.Models[string(id)]
	if !ok {
		return "", fmt.Errorf("no model pin for harness %q in bench config", id)
	}
	return m, nil
}

// Price returns a model's list price, if the config has one.
func (c Config) Price(model string) (Price, bool) {
	p, ok := c.Prices[model]
	return p, ok
}
