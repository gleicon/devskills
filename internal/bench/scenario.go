// Package bench loads benchmark scenarios from evals/ and materializes
// their fixtures into sandbox git repositories.
package bench

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"
)

// Check tiers (FR-10).
const (
	TierPlantedDefect = "planted-defect"
	TierStructural    = "structural"
	TierSmoke         = "smoke"
)

// Skill output styles for planted-defect checking.
const (
	StyleReport = "report"
	StyleApply  = "apply"
)

// Scenario is one benchmark case: a fixture repo (base/ + change/) plus the
// task prompt and expectations from expectations.yaml.
type Scenario struct {
	Name string // scenario directory name
	Dir  string // path to the scenario directory

	Task  string `yaml:"task"`
	Tier  string `yaml:"tier"`
	Style string `yaml:"style"`
	// Expectations drive planted-defect checking; Elements drive structural.
	Expectations []Expectation `yaml:"expectations"`
	Elements     []string      `yaml:"elements"`
}

// Expectation is one planted defect. Report-style skills hit by mentioning
// File plus any keyword; apply-style skills hit by changing the lines the
// anchors identify in the fixture.
type Expectation struct {
	File     string   `yaml:"file"`
	Keywords []string `yaml:"keywords"`
	Anchors  []string `yaml:"anchors"`
}

// ExpectedHits is the per-run hit denominator: planted defects or required
// elements; smoke has none.
func (s *Scenario) ExpectedHits() int {
	if s.Tier == TierStructural {
		return len(s.Elements)
	}
	return len(s.Expectations)
}

// LoadScenarios loads every scenario for a skill under evalsDir, sorted by
// name. It fails when the skill has no scenario directory.
func LoadScenarios(evalsDir, skill string) ([]*Scenario, error) {
	skillDir := filepath.Join(evalsDir, skill)
	entries, err := os.ReadDir(skillDir)
	if err != nil {
		return nil, fmt.Errorf("no scenarios for skill %q: %w", skill, err)
	}
	var scenarios []*Scenario
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		s, err := LoadScenario(filepath.Join(skillDir, e.Name()))
		if err != nil {
			return nil, err
		}
		scenarios = append(scenarios, s)
	}
	if len(scenarios) == 0 {
		return nil, fmt.Errorf("no scenarios for skill %q in %s", skill, skillDir)
	}
	sort.Slice(scenarios, func(i, j int) bool { return scenarios[i].Name < scenarios[j].Name })
	return scenarios, nil
}

// LoadScenario loads and validates a single scenario directory.
func LoadScenario(dir string) (*Scenario, error) {
	name := filepath.Base(dir)
	b, err := os.ReadFile(filepath.Join(dir, "expectations.yaml"))
	if err != nil {
		return nil, fmt.Errorf("scenario %s: %w", name, err)
	}
	s := &Scenario{Name: name, Dir: dir}
	if err := yaml.UnmarshalWithOptions(b, s, yaml.Strict()); err != nil {
		return nil, fmt.Errorf("scenario %s: expectations.yaml: %w", name, err)
	}
	if err := s.validate(); err != nil {
		return nil, fmt.Errorf("scenario %s: %w", name, err)
	}
	return s, nil
}

func (s *Scenario) validate() error {
	var errs []error
	if s.Task == "" {
		errs = append(errs, errors.New("task is required"))
	}
	for _, sub := range []string{"base", "change"} {
		if fi, err := os.Stat(filepath.Join(s.Dir, sub)); err != nil || !fi.IsDir() {
			errs = append(errs, fmt.Errorf("missing %s/ directory", sub))
		}
	}
	switch s.Tier {
	case TierPlantedDefect:
		errs = append(errs, s.validatePlantedDefect()...)
	case TierStructural:
		if len(s.Elements) == 0 {
			errs = append(errs, errors.New("structural tier requires elements"))
		}
		if s.Style != "" || len(s.Expectations) > 0 {
			errs = append(errs, errors.New("structural tier takes elements only, not style or expectations"))
		}
	case TierSmoke:
		if s.Style != "" || len(s.Expectations) > 0 || len(s.Elements) > 0 {
			errs = append(errs, errors.New("smoke tier takes no style, expectations, or elements"))
		}
	default:
		errs = append(errs, fmt.Errorf("tier = %q, want %s, %s, or %s", s.Tier, TierPlantedDefect, TierStructural, TierSmoke))
	}
	return errors.Join(errs...)
}

func (s *Scenario) validatePlantedDefect() []error {
	var errs []error
	if s.Style != StyleReport && s.Style != StyleApply {
		errs = append(errs, fmt.Errorf("style = %q, want %s or %s", s.Style, StyleReport, StyleApply))
	}
	if len(s.Elements) > 0 {
		errs = append(errs, errors.New("planted-defect tier takes expectations, not elements"))
	}
	if len(s.Expectations) == 0 {
		errs = append(errs, errors.New("planted-defect tier requires expectations"))
		return errs
	}
	for i, e := range s.Expectations {
		if e.File == "" {
			errs = append(errs, fmt.Errorf("expectations[%d]: file is required", i))
		}
		switch s.Style {
		case StyleReport:
			if len(e.Keywords) == 0 {
				errs = append(errs, fmt.Errorf("expectations[%d]: report style requires keywords", i))
			}
		case StyleApply:
			if len(e.Anchors) == 0 {
				errs = append(errs, fmt.Errorf("expectations[%d]: apply style requires anchors", i))
			}
		}
	}
	if s.Style == StyleReport {
		errs = append(errs, s.keywordCollisions()...)
	}
	return errs
}

// keywordCollisions rejects keywords shared across expectations. Report hits
// match the whole output with a case-insensitive Contains, so one expectation's
// keyword appearing inside another's cross-hits: a single model sentence would
// score two planted defects.
func (s *Scenario) keywordCollisions() []error {
	var errs []error
	for i, a := range s.Expectations {
		for _, ka := range a.Keywords {
			for j, b := range s.Expectations[i+1:] {
				for _, kb := range b.Keywords {
					la, lb := strings.ToLower(ka), strings.ToLower(kb)
					if strings.Contains(la, lb) || strings.Contains(lb, la) {
						errs = append(errs, fmt.Errorf("expectations[%d] keyword %q collides with expectations[%d] keyword %q: keyword sets must stay disjoint across expectations", i, ka, i+1+j, kb))
					}
				}
			}
		}
	}
	return errs
}
