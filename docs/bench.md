# Benchmarking skills

`devskills bench` produces PR-ready before/after evidence for a skill change. It runs the **old** version of the skill (from the main branch) and the **new** version (your working tree) against committed scenarios, scores each run deterministically — no LLM judging anywhere — and emits a markdown report you paste into the PR.

```bash
devskills bench ds-deslop                      # stream raw runs, Claude Code only
devskills bench ds-deslop --format pr-md       # PR-ready markdown report
devskills bench ds-deslop --format pr-md --out report.md
devskills bench ds-deslop --scenario narrated-greeting --runs 1
devskills bench ds-deslop --harness claude,codex,opencode
devskills bench ds-deslop --blocks base,concise  # with agents-md blocks installed
```

In pr-md mode the report is the product: it goes to stdout (or the `--out` file), and every progress line — run headers, scores, failures — streams to stderr.

In the devskills repo itself, bench the working tree — never a stale installed binary — with `make bench SKILL=ds-deslop ARGS="--format pr-md"`, or equivalently `go run . bench ds-deslop --format pr-md`.

## How a run works

For each scenario, bench materializes the fixture into a temp git repo (base tree committed on `main`, change tree committed on a `change` branch), installs the skill version under test project-locally, and invokes the assistant headlessly with the scenario's task prompt. The model is pinned per assistant in `evals/bench.yaml`; `--model` overrides. Stdout, stderr, and the post-run diff are captured and scored.

- Old version comes from the main branch via git; new from the working tree. Each carries the skill's whole directory — SKILL.md plus any companion files — so the sandbox install matches a real one. A skill absent on main runs **baseline mode**: new version only.
- `--blocks` installs `agents-md` blocks into the sandbox's `AGENTS.md`, with a `CLAUDE.md` that imports it — the files `devskills init` writes. It takes `base`, init's layer ids (`concise`, `interaction`, `plain-language`, `phases`, `spec-discipline`) and language names (`go`, `shell`). Blocks follow the skill's split: old runs get the main branch's copies, new runs the working tree's, and a block absent on main is left out of old runs. The report lists each block's old and new SHA. They land in the fixture's base commit, so they sit on both branches and never show up in the change a skill reviews. Without `--blocks`, no block reaches a run. To measure a block change, bench a skill whose behavior the block should shift — the skill is identical on both sides, so the delta is the block's.
- `--runs` (default 3) repeats each version per scenario; skills are nondeterministic, one run proves little. `--timeout` bounds each run: the flag when given, else the scenario's `timeout:`, else 5m.
- Every run also records token usage and cost, from each assistant's JSON output (Claude `--output-format json`, Codex `exec --json`, OpenCode `run --format json`). The report adds each run's cost and two rows per scenario: **cost / success** — the spend of every run, failures included, divided by the runs that hit every expectation (smoke: produced output) — and **median cost**.
- Cost is a list price, never what a subscription bills. Claude computes its own. Codex reports tokens only, so bench prices them from the `prices:` table in `evals/bench.yaml`, and the report states the rates and the date they were checked. A model with no entry there — a `--model` override, say — reports tokens with the cost marked unknown. OpenCode prices runs itself but reports 0 for any model it cannot price, including every model behind a ChatGPT sign-in, so bench reads 0 as unknown, not free. A timed-out run prints no usage, so its spend is missing from both rows. OpenCode's stream leaves out subagent steps, so a skill that spawns subagents is undercounted there.
- A missing CLI or timed-out run is reported loudly, never skipped. The command exits non-zero only when every run failed.
- Reports never compare scores across assistants — per-assistant tables only. Interpretation belongs to the PR author and reviewer; the report carries no verdict.
- In devskills 0.8.1 and earlier, OpenCode runs did not run in the sandbox: they inherited the caller's `PWD`, which OpenCode trusts over its working directory, so they worked on whatever repo `devskills bench` was started from. Discard OpenCode bench results from those versions.
- Claude and OpenCode runs are isolated from the operator's global config (Claude's `--setting-sources project`, which still reads the sandbox's own `CLAUDE.md`, plus auto-memory off and `--strict-mcp-config` for your MCP servers and claude.ai connectors; an empty `OPENCODE_CONFIG_DIR` for OpenCode). Codex offers no equivalent off-switch, so its runs inherit `~/.codex/config.toml` and the global `AGENTS.md` — read Codex numbers with that in mind.
- Runs are approvals-off: the sandbox is a throwaway *working directory*, not an OS boundary — only Codex (`--sandbox workspace-write`) is genuinely confined. A bench run executes the branch's committed tasks and fixtures with your local CLIs, so review `evals/` changes like code before benching an untrusted branch.

## Scenario anatomy

Scenarios live at `evals/<skill>/<scenario>/`:

```
evals/ds-deslop/narrated-greeting/
├── base/               # fixture tree, committed on the default branch
├── change/             # overlay, committed on the work branch
└── expectations.yaml   # task prompt, check tier, expectations
```

`base/` is the repo as it stood; `change/` overlays it as the branch under review. Both directories are required and must contain at least one file. `evals/` is never embedded in the binary.

`expectations.yaml` declares the task and one of three check tiers. Two optional fields cover a task that delegates:

- `skills: [ds-bug-review]` installs the named skills from the working tree beside the skill under test. The sandbox holds nothing else, so a task naming a second `/ds-*` skill loads nothing without it.
- `timeout: 20m` replaces the 5m default for this scenario; `--timeout` still overrides it.

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
- **Keyword sets stay disjoint across expectations.** Hits match the whole output, so a keyword shared between two expectations cross-hits — one model sentence would score two planted defects. The loader rejects collisions, including one expectation's keyword containing another's (case-insensitive). Plant distinct defect *types*; two same-type defects in different files are indistinguishable under whole-output matching.
- **Short and specific.** A keyword like `bug` or `issue` matches everything and proves nothing; `dead guard` or `narrating comment` matches only the planted defect.
- **Undercounts are fixable**: when a legitimate finding misses because the model phrased it unexpectedly, widen the list — that's authoring, not gaming.
- Apply-style **anchors** are exact substrings of planted lines. Anchor on the distinctive part of the line (a comment, a condition), never on line numbers — fixtures shift.

Every committed scenario is exercised by `go test ./internal/bench` (`TestCommittedScenarios`), which verifies each anchor actually exists in the fixture's work-branch file — expectations cannot drift from the fixtures they point at.

## Evidence in PRs

Skills with at least one scenario are **covered**. Two checks enforce coverage — neither runs a benchmark or spends a token:

- A catalog test fails when a newly added skill has no scenario under `evals/` (existing skills are grandfathered; the list never grows).
- CI fails a PR touching a covered skill or `agents-md/` unless it carries a bench report — `# Bench report:` in the PR body or in a committed markdown file. For a block change, bench with `--blocks`.

Generate the evidence with:

```bash
devskills bench <skill> --format pr-md
```

and paste the output into the PR's evidence section. The report includes the exact reproduction command, model IDs, and skill-version SHAs, so a reviewer can re-run it identically.
