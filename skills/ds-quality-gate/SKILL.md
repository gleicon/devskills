---
name: ds-quality-gate
description: "Run the review pipeline against the current branch or feature as a finite quality gate — a sequenced set of strict review passes."
disable-model-invocation: true
---

When invoked, run the quality gate: six strict review passes in sequence, **bookended by `/ds-deslop`** — run first to clean the incoming diff, and again last to strip any slop the between-pass fixes introduced. Each pass surfaces findings, you accept or reject them, the agent implements the accepted ones, then the next pass runs. The gate is **finite** — it runs once through the passes and finishes, not a standing mode. Say "stop" or "skip the rest" at any point to end it early.

Scope: the changed files on the current branch (same as `/ds-code-quality-review` with no argument). Narrow it by passing a path — `/ds-quality-gate src/handlers/` limits every pass to that path.

## The pipeline (in order)

```
1. /ds-deslop              — strip narrating comments, defensive overkill, type escape hatches (clean the incoming diff)
2. /ds-code-quality-review — single source of truth + structure: delete duplicates, competing implementations, and dead abstractions, then is the diff making the codebase worse? (run early, so later passes don't audit code that should have been deleted)
3. /ds-test-quality-review — is the risky logic covered with real, non-trivial tests?
4. /ds-security-review     — exploitability: input, auth, secrets, I/O, injection
5. /ds-bug-review          — correctness: real bugs, not style
6. /ds-data-review         — data correctness: schema, queries, transactions, migrations. Only when the change touches one of those; otherwise skip with a note
7. /ds-doc-quality-review  — is the public API, config, and non-obvious behavior documented?
8. /ds-deslop              — final cleanup: re-run to strip any slop the between-pass fixes introduced
```

Each pass answers a **different question**. They do not overlap. The order matters: strip noise *first* so the structural and correctness passes see signal, not slop — and strip it *again last*, because this gate implements accepted fixes between passes, and those fixes are themselves freshly-generated code that can carry slop.

## Per-pass loop

After each pass:
1. Show findings for that pass only.
2. Ask: "Accept all / reject all / list numbers to skip?"
3. Implement accepted findings immediately before moving to the next pass.
4. Confirm what changed, then proceed.

If a pass finds nothing: say so in one line and move on. Do not pad.

## Invocation

```
/ds-quality-gate              # run the full pipeline on branch changes
/ds-quality-gate src/api/     # scope every pass to a path
```

Say "stop" or "skip the rest" at any point to end the gate early. After the final pass, report a one-paragraph summary: what was fixed, what was skipped, and whether the branch is ready for `/ds-verify-this`.

## Rules

- Never skip a pass silently. If a pass is irrelevant (e.g., `/ds-security-review` on a pure-data-model change with no I/O, or `/ds-data-review` on a change that touches no schema, queries, transactions, or migrations), say so and move on.
- Implement accepted findings before the next pass — don't batch them all at the end.
- `/ds-deslop` **bookends** the pipeline: first so reviewers (and the later passes) see clean code, and last so slop introduced by the gate's own between-pass fixes doesn't survive. The final deslop is the closing step — nothing runs after it.
- If the user says "skip" or "next" mid-pass, honor it and proceed.
