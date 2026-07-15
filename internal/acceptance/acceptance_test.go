//go:build integration

// End-to-end acceptance for the built binary — the durable form of the one-off
// acceptance.sh harness. It builds the real binary once, then drives
// install/init/doctor/version against a throwaway $HOME and asserts the
// sandbox acceptance criteria from .project/SPEC.md (AC-4/5/6/7/13/17/18).
//
// Opt-in, never part of `go test ./...`:  go test -tags integration ./internal/acceptance/
//
// Safety: every invocation runs with HOME and the three harness config-dir env
// vars forced inside a t.TempDir() sandbox, so no real ~/.claude, ~/.codex, or
// ~/.config/opencode is ever touched. doctor --fix is only run with --dry-run.
package acceptance

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const (
	skillCount = 47
	rootPkg    = "github.com/gleicon/devskills"
)

var dsBin string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "ds-acceptance")
	if err != nil {
		panic(err)
	}
	dsBin = filepath.Join(tmp, "devskills")
	// Build the root binary by import path — works regardless of this package's
	// location in the module tree.
	build := exec.Command("go", "build",
		"-ldflags", "-X "+rootPkg+"/internal/cli.version=v0.0.0-acceptance",
		"-o", dsBin, rootPkg)
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	if err := build.Run(); err != nil {
		_ = os.RemoveAll(tmp)
		panic("build failed: " + err.Error())
	}
	code := m.Run()
	_ = os.RemoveAll(tmp)
	os.Exit(code)
}

// TestInstall covers the global install path across all three harnesses.
func TestInstall(t *testing.T) {
	sb := t.TempDir()
	cSkills := filepath.Join(sb, ".claude", "skills")
	oSkills := filepath.Join(sb, ".config", "opencode", "skills")
	xSkills := filepath.Join(sb, ".codex", "skills")
	cCmds := filepath.Join(sb, ".claude", "commands")

	// seed a bash-era layout: a ds- legacy command, a non-ds legacy command
	// (user-owned, must be backed up before purge), and a retired skill dir.
	writeFile(t, filepath.Join(cCmds, "ds-code-review.md"), "legacy\n")
	writeFile(t, filepath.Join(cCmds, "test.md"), "user command\n")
	writeFile(t, filepath.Join(cSkills, "ds-typeset", "SKILL.md"), "stale\n")

	// AC-5: --dry-run writes nothing.
	before := fingerprint(t, sb)
	mustRun(t, sb, sb, "install", "--all", "--dry-run")
	if fingerprint(t, sb) != before {
		t.Error("AC-5: install --dry-run mutated the sandbox")
	}

	mustRun(t, sb, sb, "install", "--all")

	// AC-4: each harness skills dir holds the full catalog.
	for name, dir := range map[string]string{"claude": cSkills, "opencode": oSkills, "codex": xSkills} {
		if n := countDirs(t, dir); n != skillCount {
			t.Errorf("AC-4: %s skills dir has %d skills, want %d", name, n, skillCount)
		}
	}

	// AC-4: purge + backup semantics.
	assertAbsent(t, "AC-4 ds- legacy purged", filepath.Join(cCmds, "ds-code-review.md"))
	assertAbsent(t, "AC-4 ds- purge leaves no .bak", filepath.Join(cCmds, "ds-code-review.md.bak"))
	assertAbsent(t, "AC-4 non-ds legacy purged", filepath.Join(cCmds, "test.md"))
	assertPresent(t, "AC-4 non-ds legacy backed up", filepath.Join(cCmds, "test.md.bak"))
	assertAbsent(t, "AC-4 retired skill pruned", filepath.Join(cSkills, "ds-typeset"))

	// AC-6: Codex sidecar on every installed skill.
	if n := countSidecars(t, xSkills); n != skillCount {
		t.Errorf("AC-6: %d Codex sidecars, want %d", n, skillCount)
	}

	// AC-7: a second install is byte-for-byte identical (idempotent).
	before2 := fingerprint(t, sb)
	mustRun(t, sb, sb, "install", "--all")
	if fingerprint(t, sb) != before2 {
		t.Error("AC-7: re-install was not idempotent")
	}
}

// TestInit covers project scaffolding: block writing, user-content safety,
// idempotent re-run, and clean uninstall (AC-13).
func TestInit(t *testing.T) {
	sb := t.TempDir()
	proj := filepath.Join(sb, "project")
	mustMkdir(t, proj)
	// git-init so the binary's projectRoot() resolves to proj deterministically.
	if out, err := exec.Command("git", "-C", proj, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	agents := filepath.Join(proj, "AGENTS.md")
	claude := filepath.Join(proj, "CLAUDE.md")
	writeFile(t, agents, "# My Project\n\nhand-written notes.\n")

	mustRun(t, proj, sb, "init", "--lang", "go,typescript")
	a := readFile(t, agents)
	for _, want := range []string{"devskills:base", "devskills:language:go", "devskills:language:typescript", "hand-written notes"} {
		if !strings.Contains(a, want) {
			t.Errorf("AC-13: AGENTS.md missing %q", want)
		}
	}
	if !strings.Contains(readFile(t, claude), "@AGENTS.md") {
		t.Error("AC-13: CLAUDE.md missing @AGENTS.md import")
	}

	// re-run is a no-op.
	mustRun(t, proj, sb, "init", "--lang", "go,typescript")
	if readFile(t, agents) != a {
		t.Error("AC-13: re-running init was not a no-op")
	}

	// uninstall removes only devskills blocks, keeps user content.
	mustRun(t, proj, sb, "init", "--uninstall")
	a = readFile(t, agents)
	if strings.Contains(a, "devskills:") {
		t.Error("AC-13: uninstall left devskills blocks")
	}
	if !strings.Contains(a, "hand-written notes") {
		t.Error("AC-13: uninstall dropped user content")
	}
}

// TestDoctor covers the external-tool report and the fix-dry-run path, asserting
// nothing is ever really installed (AC-17).
func TestDoctor(t *testing.T) {
	sb := t.TempDir()
	out := mustRun(t, sb, sb, "doctor") // check-only never errors
	for _, tok := range []string{"osv-scanner", "ast-grep", "tldt", "ds-osv", "ds-security-review", "ds-tldt"} {
		if !strings.Contains(out, tok) {
			t.Errorf("AC-17: doctor output missing %q", tok)
		}
	}
	dry := mustRun(t, sb, sb, "doctor", "--fix", "--dry-run")
	for _, line := range strings.Split(dry, "\n") {
		if strings.HasPrefix(line, "install ") { // a real install line, not "would "
			t.Errorf("AC-17: --dry-run performed a real install: %q", line)
		}
	}
}

// TestVersion asserts the ldflags version stamp reaches the CLI (AC-18).
func TestVersion(t *testing.T) {
	sb := t.TempDir()
	out := mustRun(t, sb, sb, "version")
	if !strings.Contains(out, "v0.0.0-acceptance") {
		t.Errorf("AC-18: version output missing ldflags stamp; got:\n%s", out)
	}
}

// --- helpers --------------------------------------------------------------

// mustRun executes the built binary in the sandbox environment and fails the
// test on a non-zero exit. dir is the working directory (matters for init's
// project detection); sandbox is the fake $HOME.
func mustRun(t *testing.T, dir, sandbox string, args ...string) string {
	t.Helper()
	cmd := exec.Command(dsBin, args...)
	cmd.Dir = dir
	cmd.Env = sandboxEnv(sandbox)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("devskills %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

// sandboxEnv forces HOME and every harness config-dir var inside sandbox,
// dropping any inherited copy so the child's first-match lookup can't see a real
// config dir.
func sandboxEnv(sandbox string) []string {
	override := map[string]string{
		"HOME":              sandbox,
		"CLAUDE_CONFIG_DIR": filepath.Join(sandbox, ".claude"),
		"CODEX_HOME":        filepath.Join(sandbox, ".codex"),
		"XDG_CONFIG_HOME":   filepath.Join(sandbox, ".config"),
	}
	var env []string
	for _, kv := range os.Environ() {
		i := strings.IndexByte(kv, '=')
		if i >= 0 {
			if _, ok := override[kv[:i]]; ok {
				continue
			}
		}
		env = append(env, kv)
	}
	for k, v := range override {
		env = append(env, k+"="+v)
	}
	return env
}

// fingerprint hashes every file's content + relative path under root into one
// order-stable digest — used to prove --dry-run and re-install change nothing.
func fingerprint(t *testing.T, root string) string {
	t.Helper()
	var lines []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(b)
		lines = append(lines, hex.EncodeToString(sum[:])+"  "+rel)
		return nil
	})
	if err != nil {
		t.Fatalf("fingerprint %s: %v", root, err)
	}
	sort.Strings(lines)
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return hex.EncodeToString(sum[:])
}

func countDirs(t *testing.T, dir string) int {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}
	n := 0
	for _, e := range entries {
		if e.IsDir() {
			n++
		}
	}
	return n
}

// countSidecars counts agents/openai.yaml files under root and checks each pins
// allow_implicit_invocation:false.
func countSidecars(t *testing.T, root string) int {
	t.Helper()
	n := 0
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Base(p) != "openai.yaml" || filepath.Base(filepath.Dir(p)) != "agents" {
			return nil
		}
		n++
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if !strings.Contains(string(b), "allow_implicit_invocation: false") {
			t.Errorf("AC-6: sidecar %s missing allow_implicit_invocation:false", p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return n
}

func assertPresent(t *testing.T, desc, path string) {
	t.Helper()
	if _, err := os.Lstat(path); err != nil {
		t.Errorf("%s: expected %s to exist", desc, path)
	}
}

func assertAbsent(t *testing.T, desc, path string) {
	t.Helper()
	if _, err := os.Lstat(path); err == nil {
		t.Errorf("%s: expected %s to be gone", desc, path)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	mustMkdir(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func mustMkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
}
