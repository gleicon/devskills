package bench

import (
	"errors"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"

	"github.com/gleicon/devskills/internal/harness"
)

// Config is the checked-in bench configuration (evals/bench.yaml): the model
// each harness is pinned to, so runs are reproducible without flags.
type Config struct {
	Models map[string]string `yaml:"models"`
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
