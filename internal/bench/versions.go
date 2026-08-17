package bench

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Version labels used in run headers and reports.
const (
	LabelOld = "old"
	LabelNew = "new"
)

// LoadVersions resolves the skill versions to bench in the repo at root:
// "old" from the main branch via git show (never the embedded catalog, FR-3.a)
// and "new" from the working tree. When the skill is absent on the main branch
// the slice holds only "new" — baseline mode (FR-3.c).
func LoadVersions(root, skill string) ([]SkillVersion, error) {
	path := filepath.Join(root, "skills", skill, "SKILL.md")
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("skill %q not found in working tree: %w", skill, err)
	}
	newSHA, err := gitStdout(root, "hash-object", path)
	if err != nil {
		return nil, err
	}
	newV := SkillVersion{Name: skill, Label: LabelNew, SHA: trimSHA(newSHA), Content: content}

	branch, err := mainBranch(root)
	if err != nil {
		return nil, err
	}
	spec := branch + ":skills/" + skill + "/SKILL.md"
	if _, err := gitStdout(root, "cat-file", "-e", spec); err != nil {
		return []SkillVersion{newV}, nil
	}
	old, err := gitStdout(root, "show", spec)
	if err != nil {
		return nil, err
	}
	oldSHA, err := gitStdout(root, "rev-parse", spec)
	if err != nil {
		return nil, err
	}
	return []SkillVersion{{Name: skill, Label: LabelOld, SHA: trimSHA(oldSHA), Content: old}, newV}, nil
}

// mainBranch finds the repo's default branch.
func mainBranch(root string) (string, error) {
	for _, b := range []string{"main", "master"} {
		if _, err := gitStdout(root, "rev-parse", "--verify", "--quiet", b); err == nil {
			return b, nil
		}
	}
	return "", fmt.Errorf("no main or master branch found in %s", root)
}

func trimSHA(raw []byte) string { return strings.TrimSpace(string(raw)) }

// gitStdout runs git and returns stdout verbatim — unlike gitOutput it never
// mixes stderr into the result, so file content from git show stays exact.
func gitStdout(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}
