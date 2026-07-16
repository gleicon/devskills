// Package catalog validates the skill catalog embedded in the binary.
package catalog

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"
)

// Validate checks every skill directory under skills/ in fsys against the
// user-invoked contract: a SKILL.md whose frontmatter name matches its
// directory, carries a description, and sets disable-model-invocation: true.
// It reports all violations at once, not just the first.
func Validate(fsys fs.FS) error {
	dirs, err := fs.ReadDir(fsys, "skills")
	if err != nil {
		return fmt.Errorf("read skills/: %w", err)
	}
	if len(dirs) == 0 {
		return errors.New("no skills found")
	}
	var errs []error
	for _, d := range dirs {
		name := d.Name()
		if !d.IsDir() {
			errs = append(errs, fmt.Errorf("%s: not a directory", name))
			continue
		}
		errs = append(errs, validateSkill(fsys, name))
	}
	return errors.Join(errs...)
}

func validateSkill(fsys fs.FS, name string) error {
	b, err := fs.ReadFile(fsys, "skills/"+name+"/SKILL.md")
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	fm, ok := frontmatter(b)
	if !ok {
		return fmt.Errorf("%s: missing --- frontmatter block", name)
	}
	if fm["name"] != name {
		return fmt.Errorf("%s: frontmatter name = %q, want %q", name, fm["name"], name)
	}
	if fm["description"] == "" {
		return fmt.Errorf("%s: missing description", name)
	}
	if fm["disable-model-invocation"] != "true" {
		return fmt.Errorf("%s: missing disable-model-invocation: true", name)
	}
	return nil
}

// frontmatter parses the leading --- fenced key: value block into a flat map,
// trimming surrounding quotes from values. Line-based on purpose: the contract
// is a handful of scalar keys, not arbitrary YAML.
func frontmatter(b []byte) (map[string]string, bool) {
	// Normalize CRLF so a file checked out with Windows line endings still parses.
	s := strings.ReplaceAll(string(b), "\r\n", "\n")
	rest, ok := strings.CutPrefix(s, "---\n")
	if !ok {
		return nil, false
	}
	block, _, ok := strings.Cut(rest, "\n---")
	if !ok {
		return nil, false
	}
	m := map[string]string{}
	for line := range strings.SplitSeq(block, "\n") {
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		m[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(val), `"`)
	}
	return m, true
}
