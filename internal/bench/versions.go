package bench

import (
	"bytes"
	"fmt"
	"io/fs"
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
// "old" from the main branch via git (never the embedded catalog, FR-3.a) and
// "new" from the working tree. Each version carries the skill's whole
// directory — SKILL.md plus any companions — so the sandbox install matches a
// real one. When the skill is absent on the main branch the slice holds only
// "new" — baseline mode (FR-3.c).
func LoadVersions(root, skill string) ([]SkillVersion, error) {
	files, err := workingTreeFiles(root, skill)
	if err != nil {
		return nil, err
	}
	newSHA, err := gitStdout(root, "hash-object", filepath.Join(root, "skills", skill, "SKILL.md"))
	if err != nil {
		return nil, err
	}
	newV := SkillVersion{Name: skill, Label: LabelNew, SHA: trimSHA(newSHA), Files: files}

	branch, err := mainBranch(root)
	if err != nil {
		return nil, err
	}
	spec := branch + ":skills/" + skill + "/SKILL.md"
	if _, err := gitStdout(root, "cat-file", "-e", spec); err != nil {
		return []SkillVersion{newV}, nil
	}
	old, err := branchFiles(root, branch, skill)
	if err != nil {
		return nil, err
	}
	oldSHA, err := gitStdout(root, "rev-parse", spec)
	if err != nil {
		return nil, err
	}
	return []SkillVersion{{Name: skill, Label: LabelOld, SHA: trimSHA(oldSHA), Files: old}, newV}, nil
}

// workingTreeFiles reads the skill's directory from the working tree, keyed by
// slash-separated path relative to skills/<skill>/.
func workingTreeFiles(root, skill string) (map[string][]byte, error) {
	dir := filepath.Join(root, "skills", skill)
	if _, err := os.Stat(filepath.Join(dir, "SKILL.md")); err != nil {
		return nil, fmt.Errorf("skill %q not found in working tree: %w", skill, err)
	}
	files := map[string][]byte{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)] = b
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("skill %q: %w", skill, err)
	}
	return files, nil
}

// branchFiles reads the skill's directory as committed on branch, keyed like
// workingTreeFiles.
func branchFiles(root, branch, skill string) (map[string][]byte, error) {
	prefix := "skills/" + skill + "/"
	list, err := gitStdout(root, "ls-tree", "-r", "--name-only", branch, "--", "skills/"+skill)
	if err != nil {
		return nil, err
	}
	files := map[string][]byte{}
	for line := range strings.Lines(string(list)) {
		p := strings.TrimSpace(line)
		if p == "" {
			continue
		}
		b, err := gitStdout(root, "show", branch+":"+p)
		if err != nil {
			return nil, err
		}
		files[strings.TrimPrefix(p, prefix)] = b
	}
	return files, nil
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
