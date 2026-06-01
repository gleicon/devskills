Activate (or deactivate) the quality gate pipeline for the current branch or feature.

When active, the quality gate runs a seven-pass review pipeline in sequence. Each pass surfaces findings, you accept or reject them, the agent implements the accepted ones, then the next pass runs. This is a mode — it stays on until you say "stop quality gate" or "/ds-quality-gate-mode off".

Scope: the changed files on the current branch (same as `/ds-code-quality-review` with no argument). Scope can be narrowed: `/ds-quality-gate-mode src/handlers/` limits every pass to that path.

## The pipeline (in order)

```
1. /ds-deslop              — strip narrating comments, defensive overkill, type escape hatches
2. /ds-test-quality-review — is the risky logic covered with real, non-trivial tests?
3. /ds-security-review     — exploitability: input, auth, secrets, I/O, injection
4. /ds-bug-review          — correctness: real bugs, not style
5. /ds-data-review         — data correctness: schema, queries, transactions, migrations. Only when the change touches one of those; otherwise skip with a note
6. /ds-code-quality-review — structure: is the diff making the codebase worse?
7. /ds-doc-quality-review  — is the public API, config, and non-obvious behavior documented?
```

Each pass answers a **different question**. They do not overlap. The order matters: strip noise first so structural and correctness passes see signal, not slop.

## Loop behavior (when active)

After each pass:
1. Show findings for that pass only.
2. Ask: "Accept all / reject all / list numbers to skip?"
3. Implement accepted findings immediately before moving to the next pass.
4. Confirm what changed, then proceed.

If a pass finds nothing: say so in one line and move on. Do not pad.

## Activation

```
/ds-quality-gate-mode              # run the full pipeline on branch changes
/ds-quality-gate-mode src/api/     # scope every pass to a path
/ds-quality-gate-mode off          # deactivate; resume normal behavior
```

After the final pass, report a one-paragraph summary: what was fixed, what was skipped, and whether the branch is ready for `/ds-verify-this`.

## Rules

- Never skip a pass silently. If a pass is irrelevant (e.g., `/ds-security-review` on a pure-data-model change with no I/O, or `/ds-data-review` on a change that touches no schema, queries, transactions, or migrations), say so and move on.
- Implement accepted findings before the next pass — don't batch them all at the end.
- `/ds-deslop` runs first, always. Reviewers (including the subsequent passes) should see clean code.
- If the user says "skip" or "next" mid-pass, honor it and proceed.
