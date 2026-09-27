package bench

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing/fstest"
	"time"

	"github.com/gleicon/devskills/internal/harness"
	devsync "github.com/gleicon/devskills/internal/sync"
)

// DefaultTimeout bounds one harness invocation.
const DefaultTimeout = 5 * time.Minute

// SkillVersion is one version of a skill under bench: the working tree ("new")
// or the main-branch content ("old").
type SkillVersion struct {
	Name  string
	Label string // LabelOld or LabelNew
	SHA   string // git blob SHA of the version's SKILL.md, for reproducible reports (NFR-3)
	// Files is the skill's whole directory — SKILL.md plus any companions —
	// keyed by slash-separated path relative to skills/<Name>/.
	Files map[string][]byte
}

// Result captures one harness invocation. Err records a failed or timed-out
// invocation (FR-9: loud, never skipped); Stdout/Stderr are kept even then.
type Result struct {
	Stdout string
	Stderr string
	Diff   string // post-run git diff of the sandbox, harness dirs excluded
	Err    error
	Usage  *Usage // nil when the assistant reported none
}

// Runner invokes one harness with a pinned model.
type Runner struct {
	Harness harness.ID
	Model   string
	Timeout time.Duration // 0 defers to the scenario's timeout, then DefaultTimeout
	Price   *Price        // list price for assistants that report tokens but no cost; nil leaves cost unknown
}

// Run benches one skill version against one scenario: materialize the fixture
// and the blocks paired with the version into a fresh sandbox, install the
// version project-locally alongside extras (the scenario's Skills), invoke the
// harness headlessly in the sandbox, capture output and the post-run diff. The
// returned error is infrastructural (sandbox, git); harness failures land in
// Result.Err.
func (r Runner) Run(ctx context.Context, s *Scenario, skill SkillVersion, extras []SkillVersion, blocks []Block) (Result, error) {
	argv, err := headlessArgs(r.Harness, s.Task, r.Model, skill.Name)
	if err != nil {
		return Result{}, err
	}
	sandbox, err := os.MkdirTemp("", "devskills-bench-*")
	if err != nil {
		return Result{}, err
	}
	defer os.RemoveAll(sandbox)
	if err := Materialize(s, sandbox, blocks); err != nil {
		return Result{}, err
	}
	// The diff base is the materialized tip, pinned by SHA: a run that
	// commits (git-mode skills) moves HEAD, and diffing against HEAD would
	// hide the run's own changes.
	baseSHA, err := gitOutput(sandbox, "rev-parse", "HEAD")
	if err != nil {
		return Result{}, err
	}
	baseSHA = strings.TrimSpace(baseSHA)
	if err := excludeHarnessDirs(sandbox); err != nil {
		return Result{}, err
	}
	if err := installSkills(sandbox, r.Harness, append([]SkillVersion{skill}, extras...)); err != nil {
		return Result{}, err
	}

	timeout := cmp.Or(r.Timeout, s.Timeout, DefaultTimeout)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = sandbox
	if r.Harness == harness.OpenCode {
		// opencode has no --safe-mode; an empty OPENCODE_CONFIG_DIR drops the
		// operator's global config, rules, agents, and commands, and
		// OPENCODE_DISABLE_CLAUDE_CODE stops the compat fallback from reading
		// ~/.claude/CLAUDE.md. Auth lives in the data dir, so it is untouched.
		configDir, err := os.MkdirTemp("", "devskills-bench-opencode-*")
		if err != nil {
			return Result{}, err
		}
		defer os.RemoveAll(configDir)
		// cmd.Environ, not os.Environ: Go only points PWD at cmd.Dir when Env
		// is nil, and opencode resolves its project from PWD — an inherited
		// one runs it in the caller's repo instead of the sandbox.
		cmd.Env = append(cmd.Environ(),
			"OPENCODE_CONFIG_DIR="+configDir,
			"OPENCODE_DISABLE_CLAUDE_CODE=1",
		)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	runErr := cmd.Run()
	// Only a failed run is a timeout: a clean exit microseconds before the
	// deadline must not be rewritten into a failure.
	if runErr != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
		runErr = fmt.Errorf("timed out after %s", timeout)
	}

	diff, err := postRunDiff(sandbox, baseSHA)
	if err != nil {
		return Result{}, err
	}
	res := Result{Stdout: stdout.String(), Stderr: stderr.String(), Diff: diff, Err: runErr}
	switch r.Harness {
	case harness.Claude:
		res.applyParsed(parseClaude(res.Stdout))
	case harness.Codex:
		res.applyParsed(parseCodex(res.Stdout, r.Price))
	case harness.OpenCode:
		res.applyParsed(parseOpenCode(res.Stdout))
	}
	return res, nil
}

// applyParsed swaps the raw JSON stdout for the final text, so the checker
// scores what the model said, and records usage. Output that yields no usage
// keeps the raw stdout for the transcript; a parse error fails an otherwise
// clean run.
func (res *Result) applyParsed(text string, usage *Usage, err error) {
	if usage != nil {
		res.Stdout, res.Usage = text, usage
	}
	if res.Err == nil {
		res.Err = err
	}
}

// headlessArgs builds the non-interactive invocation for a harness.
func headlessArgs(id harness.ID, task, model, skill string) ([]string, error) {
	switch id {
	case harness.Claude:
		// Claude Code does not surface project-local skills to a headless run,
		// so the task naming `/skill` mid-sentence loads nothing and the score
		// would only reflect the bare task text. Point the model at the file
		// bench installed — the same thing Codex does for itself by grepping.
		prompt := fmt.Sprintf("Read .claude/skills/%s/SKILL.md and follow it as your instructions for this task: %s", skill, task)
		// --setting-sources project drops the operator's user-level CLAUDE.md,
		// skills, plugins, hooks and output styles, which move scores independent
		// of the skill; --safe-mode would also drop the sandbox's own CLAUDE.md.
		// --dangerously-skip-permissions runs approvals-off: only the cwd is the
		// throwaway sandbox — the process is unconfined, so scenario tasks are
		// trusted input (see the trust model in docs/bench.md).
		return []string{"claude", "-p", prompt, "--model", model, "--output-format", "json", "--setting-sources", "project", "--dangerously-skip-permissions"}, nil
	case harness.Codex:
		// exec is codex's non-interactive mode; workspace-write confines
		// model-run commands to the sandbox repo.
		return []string{"codex", "exec", "--json", "--model", model, "--sandbox", "workspace-write", task}, nil
	case harness.OpenCode:
		// --pure drops the operator's external plugins; --auto approves the
		// permission prompts a headless run can't answer — approvals-off like
		// Claude, unconfined beyond the sandbox cwd, so scenario tasks are
		// trusted input.
		return []string{"opencode", "run", task, "--model", model, "--format", "json", "--pure", "--auto"}, nil
	}
	return nil, fmt.Errorf("harness %q is not supported by bench", id)
}

// installSkills writes each skill — its full directory, companions included —
// under the harness's project-local skills path in the sandbox, reusing the
// sync engine with an in-memory catalog so install semantics (layout, Codex
// sidecar) stay single-sourced.
func installSkills(sandbox string, id harness.ID, skills []SkillVersion) error {
	dir, err := harness.Resolver{ProjectRoot: sandbox}.SkillsDir(id, harness.Local)
	if err != nil {
		return err
	}
	catalog := fstest.MapFS{}
	for _, skill := range skills {
		for rel, data := range skill.Files {
			catalog["skills/"+skill.Name+"/"+rel] = &fstest.MapFile{Data: data}
		}
	}
	eng := devsync.New(catalog)
	plan, err := eng.Plan(devsync.Target{Name: id.Name(), SkillsDir: dir, Codex: id == harness.Codex})
	if err != nil {
		return err
	}
	return eng.Apply(plan)
}

// excludeHarnessDirs keeps the project-local harness state (installed skill,
// session files) out of the post-run diff via .git/info/exclude, which unlike
// a .gitignore leaves the work tree itself untouched.
func excludeHarnessDirs(sandbox string) error {
	info := filepath.Join(sandbox, ".git", "info")
	if err := os.MkdirAll(info, 0o755); err != nil {
		return err
	}
	var lines strings.Builder
	for _, id := range harness.All() {
		lines.WriteString(harness.LocalDir(id))
		lines.WriteString("/\n")
	}
	return os.WriteFile(filepath.Join(info, "exclude"), []byte(lines.String()), 0o644)
}

// postRunDiff stages everything and diffs against the materialized tip, so
// files the harness added are captured alongside edits — even when the run
// committed and moved HEAD. The sandbox is disposable; mutating its index is
// fine.
func postRunDiff(sandbox, baseSHA string) (string, error) {
	if err := git(sandbox, "add", "-A"); err != nil {
		return "", err
	}
	return gitOutput(sandbox, "diff", "--cached", baseSHA)
}
