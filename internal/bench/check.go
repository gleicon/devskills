package bench

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
)

// CheckResult scores one run against a scenario's expectations.
type CheckResult struct {
	// Hits is parallel to Scenario.Expectations: whether each planted defect
	// was found.
	Hits []bool
	// Extras are findings matching no expectation — counted and listed but
	// never scored (FR-11). Report style: output lines mentioning a fixture
	// file that no expectation accounts for. Apply style: changed files no
	// expectation names.
	Extras []string
}

// HitCount is the number of planted defects the run found.
func (c CheckResult) HitCount() int {
	n := 0
	for _, h := range c.Hits {
		if h {
			n++
		}
	}
	return n
}

// Check scores a run deterministically per the scenario's tier (FR-10). Only
// planted-defect is supported so far; structural and smoke land separately.
func Check(s *Scenario, res Result) (CheckResult, error) {
	if s.Tier != TierPlantedDefect {
		return CheckResult{}, fmt.Errorf("scenario %s: checking for tier %q is not implemented yet", s.Name, s.Tier)
	}
	switch s.Style {
	case StyleReport:
		return checkReport(s, res.Stdout)
	case StyleApply:
		return checkApply(s, res.Diff), nil
	}
	return CheckResult{}, fmt.Errorf("scenario %s: unknown style %q", s.Name, s.Style)
}

// checkReport hits an expectation when the output mentions its file and at
// least one keyword (FR-10.a). File paths match exactly; keywords are
// case-insensitive.
func checkReport(s *Scenario, stdout string) (CheckResult, error) {
	lower := strings.ToLower(stdout)
	c := CheckResult{Hits: make([]bool, len(s.Expectations))}
	for i, e := range s.Expectations {
		c.Hits[i] = strings.Contains(stdout, e.File) && anyKeyword(lower, e.Keywords)
	}

	files, err := fixtureFiles(s)
	if err != nil {
		return CheckResult{}, err
	}
	for line := range strings.Lines(stdout) {
		line = strings.TrimRight(line, "\n")
		lowerLine := strings.ToLower(line)
		for _, f := range files {
			if !strings.Contains(line, f) {
				continue
			}
			expected := false
			for _, e := range s.Expectations {
				if e.File == f && anyKeyword(lowerLine, e.Keywords) {
					expected = true
					break
				}
			}
			if !expected {
				c.Extras = append(c.Extras, line)
			}
			break // one verdict per line, on its first mentioned file
		}
	}
	return c, nil
}

// checkApply hits an expectation when the post-run diff removes or rewrites a
// line containing one of its anchors in the expected file.
func checkApply(s *Scenario, diff string) CheckResult {
	sections := diffSections(diff)
	c := CheckResult{Hits: make([]bool, len(s.Expectations))}
	for i, e := range s.Expectations {
		for _, removed := range sections[e.File] {
			if anyAnchor(removed, e.Anchors) {
				c.Hits[i] = true
				break
			}
		}
	}
	expected := make([]string, 0, len(s.Expectations))
	for _, e := range s.Expectations {
		expected = append(expected, e.File)
	}
	for file := range sections {
		if !slices.Contains(expected, file) {
			c.Extras = append(c.Extras, file)
		}
	}
	slices.Sort(c.Extras)
	return c
}

// diffSections parses a unified git diff into file → removed lines (content
// of "-" lines, prefix stripped).
func diffSections(diff string) map[string][]string {
	sections := map[string][]string{}
	var current string
	for line := range strings.Lines(diff) {
		line = strings.TrimRight(line, "\n")
		if rest, ok := strings.CutPrefix(line, "diff --git a/"); ok {
			// "a/<path> b/<path>": the b-path is the file's post-run name.
			if _, b, ok := strings.Cut(rest, " b/"); ok {
				current = b
				sections[current] = nil
			}
			continue
		}
		if current == "" || strings.HasPrefix(line, "---") {
			continue
		}
		if removed, ok := strings.CutPrefix(line, "-"); ok {
			sections[current] = append(sections[current], removed)
		}
	}
	return sections
}

// fixtureFiles lists the scenario's fixture file paths (slash-separated,
// relative to the repo root) across base/ and change/.
func fixtureFiles(s *Scenario) ([]string, error) {
	var files []string
	for _, sub := range []string{"base", "change"} {
		root := filepath.Join(s.Dir, sub)
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			if f := filepath.ToSlash(rel); !slices.Contains(files, f) {
				files = append(files, f)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("scenario %s: %w", s.Name, err)
		}
	}
	return files, nil
}

func anyKeyword(lowerText string, keywords []string) bool {
	for _, k := range keywords {
		if strings.Contains(lowerText, strings.ToLower(k)) {
			return true
		}
	}
	return false
}

func anyAnchor(line string, anchors []string) bool {
	for _, a := range anchors {
		if strings.Contains(line, a) {
			return true
		}
	}
	return false
}
