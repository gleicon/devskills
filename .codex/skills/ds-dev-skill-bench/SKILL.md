---
name: ds-dev-skill-bench
description: "Bench the skills changed on this branch with devskills bench, then improve them from the results or produce PR evidence."
disable-model-invocation: true
---

When invoked, run `devskills bench` against the skills under development and act on the results in one of two modes. This skill drives the bench; it never changes how the bench scores. Read `docs/bench.md` before the first run if you haven't this session.

## Arguments

`/ds-dev-skill-bench [skill…] [improve|report] [bench flags]`

- **Skills** — explicit names bench exactly those, skipping auto-detection.
- **Mode** — `improve` or `report`. Free text may resolve it when unambiguous ("…for the PR" → report). Otherwise ask at pre-flight; never guess — a wrongly guessed mode wastes a paid run.
- **Assistants** — bare assistant names (`claude`, `opencode`, `codex`) or free text ("all assistants") resolve to `--harness`; registry assistants only.
- **Bench flags** — `--runs`, `--harness`, `--scenario`, `--model` pass through verbatim. Don't add flags the user didn't give; the bench owns its defaults.

## Targets

Without explicit skills, targets are the union of:

- skills changed in `git diff $(git merge-base main HEAD)..HEAD -- skills/`
- skills with uncommitted working-tree changes under `skills/`

Skill names come from the `skills/<name>/` path segment. Evals-only changes (scenario touched, skill untouched) are mentioned but never auto-benched — with SKILL.md unchanged, old and new are identical and the comparison is a no-op; the user can still name the skill explicitly to validate a scenario. No targets and no arguments: say so and stop.

## Pre-flight

Before the first bench invocation, echo the resolved targets with each one's scenario count and total run count, plus the assistant set the runs will use — the bench default when none was chosen, named explicitly so it never rides silently. Ask for the mode if it is still unresolved. One confirmation for the whole batch, not one per skill; the user can amend the assistant set at this gate. Zero bench runs happen before this gate.

## Running the bench

Always the working tree, never an installed binary, always the pr-md report:

```bash
make bench SKILL=<skill> ARGS="--format pr-md --out <scratch>/<skill>.md [passthrough flags]"
```

(equivalently `go run . bench <skill> …`). The report's hit tables plus embedded transcripts are the diagnosis input and the PR artifact — no second format is ever needed. Bench targets run sequentially.

## Improve mode

1. Bench each target; read the report.
2. Diagnose each miss from the transcripts: what did the run actually say or do, and what in the skill's prose allowed it?
3. Propose edits as a diff over `skills/<skill>/SKILL.md` with a per-miss rationale. Present it and **wait for approval — apply nothing before a yes.**
4. After an approved apply, offer a re-bench. Never run it, or iterate further, unprompted.

**Hard rule — the improver never grades its own homework:** propose edits only to `skills/<skill>/SKILL.md` of the targeted skills. Never edit anything under `evals/` — not expectations, not fixtures, not keyword lists. When a miss looks like a miscalibrated check (legitimate phrasing outside a narrow keyword list), report that with the transcript evidence and leave the eval edit to the user.

**Zero-scenario targets:** when a targeted skill has no scenarios under `evals/`, you may draft one (`base/` + `change/` + `expectations.yaml`, per `docs/bench.md`) — only before any bench run for that skill, and presented for approval like every other edit. Once at least one scenario exists, the hard rule covers it with no exception.

## Report mode

Default: write the report to a scratch path, display it, and print the path with a reminder to include it in the PR body. Create nothing in the repo tree. On explicit request, write `evals/reports/<skill>.md` instead for the user to commit.

Then adopt this standing rule for the rest of the session: **if a PR is opened, its body includes a `## Benchmark evidence` section** — embedding the full report (scratch variant) or referencing the committed file (committed variant). This satisfies CI's `# Bench report:` presence check.

## Rules

- Never draw a verdict from a report — no "pass", "improved", or cross-assistant comparison, in output, PR sections, or proposed prose. Numbers and transcripts speak; the author narrates, the reviewer judges.
- This skill never commits or pushes; version control follows the session's git posture.
