package bench

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	devsync "github.com/gleicon/devskills/internal/sync"
)

// Branch names of the materialized fixture repo.
const (
	DefaultBranch = "main"
	WorkBranch    = "change"
)

// Materialize turns a scenario into a git repository at dir (an existing empty
// directory): the base/ tree committed on DefaultBranch, then the change/ tree
// overlaid and committed on WorkBranch, which is left checked out. Host and
// system git config are ignored so runs are reproducible across machines.
func Materialize(s *Scenario, dir string) error {
	if err := devsync.CopyTree(os.DirFS(filepath.Join(s.Dir, "base")), ".", dir); err != nil {
		return fmt.Errorf("scenario %s: copy base/: %w", s.Name, err)
	}
	steps := [][]string{
		{"init", "-q", "-b", DefaultBranch},
		{"add", "-A"},
		{"commit", "-q", "-m", "base"},
		{"checkout", "-q", "-b", WorkBranch},
	}
	for _, args := range steps {
		if err := git(dir, args...); err != nil {
			return fmt.Errorf("scenario %s: %w", s.Name, err)
		}
	}
	if err := devsync.CopyTree(os.DirFS(filepath.Join(s.Dir, "change")), ".", dir); err != nil {
		return fmt.Errorf("scenario %s: copy change/: %w", s.Name, err)
	}
	for _, args := range [][]string{{"add", "-A"}, {"commit", "-q", "-m", "change"}} {
		if err := git(dir, args...); err != nil {
			return fmt.Errorf("scenario %s: %w", s.Name, err)
		}
	}
	return nil
}

func git(dir string, args ...string) error {
	_, err := gitOutput(dir, args...)
	return err
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_CONFIG_SYSTEM="+os.DevNull,
		"GIT_AUTHOR_NAME=devskills-bench",
		"GIT_AUTHOR_EMAIL=bench@devskills.local",
		"GIT_COMMITTER_NAME=devskills-bench",
		"GIT_COMMITTER_EMAIL=bench@devskills.local",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}
