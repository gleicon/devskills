# Benchmarking skills

`devskills bench` produces PR-ready before/after evidence for a skill change. It runs the **old** version of the skill (from the main branch) and the **new** version (your working tree) against committed scenarios, scores each run deterministically — no LLM judging anywhere — and emits a markdown report you paste into the PR.

```bash
devskills bench ds-deslop                      # stream raw runs, Claude Code only
devskills bench ds-deslop --format pr-md       # PR-ready markdown report
devskills bench ds-deslop --scenario narrated-greeting --runs 1
devskills bench ds-deslop --harness claude,codex,opencode
```

In the devskills repo itself, bench the working tree — never a stale installed binary — with `make bench SKILL=ds-deslop ARGS="--format pr-md"`, or equivalently `go run . bench ds-deslop --format pr-md`.

## How a run works

For each scenario, bench materializes the fixture into a temp git repo (base tree committed on `main`, change tree committed on a `change` branch), installs the skill version under test project-locally, and invokes the assistant headlessly with the scenario's task prompt. The model is pinned per assistant in `evals/bench.yaml`; `--model` overrides. Stdout, stderr, and the post-run diff are captured and scored.

- Old version comes from the main branch via git; new from the working tree. Each carries the skill's whole directory — SKILL.md plus any companion files — so the sandbox install matches a real one. A skill absent on main runs **baseline mode**: new version only.
- `--runs` (default 3) repeats each version per scenario; skills are nondeterministic, one run proves little.
- A missing CLI or timed-out run is reported loudly, never skipped. The command exits non-zero only when every run failed.
- Reports never compare scores across assistants — per-assistant tables only. Interpretation belongs to the PR author and reviewer; the report carries no verdict.
- Claude and OpenCode runs are isolated from the operator's global config (Claude's `--safe-mode`; an empty `OPENCODE_CONFIG_DIR` for OpenCode). Codex offers no equivalent off-switch, so its runs inherit `~/.codex/config.toml` and the global `AGENTS.md` — read Codex numbers with that in mind.

## Scenario anatomy

Scenarios live at `evals/<skill>/<scenario>/`:

```
evals/ds-deslop/narrated-greeting/
├── base/               # fixture tree, committed on the default branch
├── change/             # overlay, committed on the work branch
└── expectations.yaml   # task prompt, check tier, expectations
```

`base/` is the repo as it stood; `change/` overlays it as the branch under review. Both directories are required and must contain at least one file. `evals/` is never embedded in the binary.

`expectations.yaml` declares the task and one of three check tiers:

### `planted-defect` — the skill must find (or fix) what you planted

The change tree plants known defects; the checker scores how many each run catches. Declare the skill's output style:

**`style: report`** — the skill lists findings. A hit requires the output to mention the expectation's file *and* at least one keyword (case-insensitive):

```yaml
task: "Run /ds-bug-review on this branch."
tier: planted-defect
style: report
expectations:
  - file: greet.go
    keywords: [nil map, uninitialized]
```

**`style: apply`** — the skill edits the tree. A hit requires the post-run diff to remove or rewrite a line containing one of the expectation's anchors in the named file:

```yaml
task: "Run /ds-deslop to remove the AI slop this branch introduced."
tier: planted-defect
style: apply
expectations:
  - file: greet.go
    anchors:
      - "// First we get the greeting for the name."
```

Findings matching no expectation are counted and listed as **extra findings**, never scored.

### `structural` — the produced artifact must contain required elements

For skills that produce a document or other artifact. Each element is a literal string that must appear in the run's stdout or on a line the post-run diff added:

```yaml
task: "Run /ds-project-checkpoint to record where this project stands."
tier: structural
elements:
  - "# now"
  - "# next"
  - "# settled"
  - "# hazards"
```

### `smoke` — the invocation must work at all

The weakest tier, for skills with genuinely no checkable output: the run passes when the assistant exits zero and prints non-blank output — even a refusal counts. Reach for it last; a skill that mandates any literal in its output (a confirmation line, a heading) supports the structural tier instead:

```yaml
task: "Run /ds-example to do something with no checkable output."
tier: smoke
```

## Writing good keyword lists

Keywords absorb phrasing variance — the same defect described three ways should still hit.

- **Any-of semantics**: one match suffices. List the distinct *names* for the defect, not sentence fragments: `[nil map, uninitialized map, assignment to entry]`.
- **Short and specific.** A keyword like `bug` or `issue` matches everything and proves nothing; `dead guard` or `narrating comment` matches only the planted defect.
- **Undercounts are fixable**: when a legitimate finding misses because the model phrased it unexpectedly, widen the list — that's authoring, not gaming.
- Apply-style **anchors** are exact substrings of planted lines. Anchor on the distinctive part of the line (a comment, a condition), never on line numbers — fixtures shift.

Every committed scenario is exercised by `go test ./internal/bench` (`TestCommittedScenarios`), which verifies each anchor actually exists in the fixture's work-branch file — expectations cannot drift from the fixtures they point at.

## Evidence in PRs

Skills with at least one scenario are **covered**. Two checks enforce coverage — neither runs a benchmark or spends a token:

- A catalog test fails when a newly added skill has no scenario under `evals/` (existing skills are grandfathered; the list never grows).
- CI fails a PR touching a covered skill unless it carries a bench report — `# Bench report:` in the PR body or in a committed markdown file.

Generate the evidence with:

```bash
devskills bench <skill> --format pr-md
```

and paste the output into the PR's evidence section. The report includes the exact reproduction command, model IDs, and skill-version SHAs, so a reviewer can re-run it identically.
