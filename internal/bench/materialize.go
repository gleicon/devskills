package bench

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	if err := copyTree(filepath.Join(s.Dir, "base"), dir); err != nil {
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
	if err := copyTree(filepath.Join(s.Dir, "change"), dir); err != nil {
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
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

// copyTree copies src into dst, overwriting existing files so change/ can
// overlay base/. os.CopyFS is not used because it refuses to overwrite.
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			out.Close()
			return err
		}
		return out.Close()
	})
}
