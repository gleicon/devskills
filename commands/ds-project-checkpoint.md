Sweep the session for durable context and route each piece to its owning file, so the session can be cleared or ended safely.

When invoked, make sure nothing durable is lost before `/clear` or the end of a session — the counterpart to `/ds-project-resume`. Sweep the conversation for context a fresh agent will need, write each piece to its right home, update the plan state, and confirm it's safe to clear. Run it after *any* session — this guards the decisions and facts that otherwise live only in the conversation.

## Arguments

- `--handoff` — also write a full handoff to `.project/handoff.md` (richer than the `## Now` block: context, what was tried, gotchas). Use it when the next session needs more than the plan state — handing off to another person, or a long pause.

## Process

1. Create `.project/` if needed.
2. **Sweep the session** for anything durable that isn't already persisted somewhere — resolved decisions, new structural/map facts, scope/requirement changes, current state, and soft context (things to keep watching).
3. **Route each item to its owning home:**
   - **Resolved decision** → append to `.project/DECISIONS.md` (create it on first write): one entry per decision — the question, the chosen answer, a one-line rationale; plain Markdown, leave existing entries intact.
   - **New stable structural/map fact** (a new top-level dir, dependency, build/test command) → append to the right section of `.project/PROJECT.md`. Never rewrite or refresh existing sections, never touch human-authored prose or constraints. If `PROJECT.md` is absent, or the map looks broadly stale (more than a one-line fact), don't edit it — flag it ("run `/ds-project-map`").
   - **Scope/requirement change** → record it as a decision in `DECISIONS.md`, then flag `SPEC.md` as stale ("no longer reflects decision X — re-run `/ds-spec`"). Never edit `SPEC.md` — it is a structured artifact owned by `/ds-spec`.
   - **State / next / open questions / soft context** → the `## Now` section of `.project/PLAN.md` (step 5).
   - **Roadmap progress** → `## Roadmap` statuses in `PLAN.md` (step 5).
4. **Dedupe, then get approval for reader-affecting writes.** Compare against what each file already holds and add only what's missing — don't duplicate the map, an existing decision, or anything already recorded. For each `DECISIONS.md` entry and each `PROJECT.md` additive line, show the exact write and get explicit approval before writing it — one item at a time; approval on one item is never approval of another (honor a batch approval if the user asks). The `PLAN.md` updates in step 5 are checkpoint's own state and need no approval.
5. **Update `PLAN.md`** (automatic):
   - `## Roadmap` task statuses to match reality (`[ ]` / `[~]` / `[x]`).
   - Rewrite `## Now`: **State** (where things stand, 2–4 lines), **Next** (the single next action), **Open questions** (anything unresolved that blocks progress), and a short **Watch** line for non-blocking things to keep an eye on, if any.
6. **No-op fast path:** if the sweep found nothing durable beyond state, say so — just update `## Now`/roadmap and report "nothing durable to route."
7. **If `--handoff`:** also write `.project/handoff.md` (current goal, what's done, what remains, key decisions, gotchas). Reference existing artifacts (`PLAN.md`, `DECISIONS.md`, commits) by path rather than duplicating them.
8. **Confirm safe to clear** — state what is now persisted and where.

## Rules

- **Route to the owning file; don't pile everything into one.** A resolved decision goes to `DECISIONS.md`, a map fact to `PROJECT.md`, state to `PLAN.md` `## Now` — never summarize all of it into one file.
- **Don't invent new files.** Soft context lives in `## Now`, not a new notes file. `EXPLORE.md` stays a scratchpad — durable findings route to `DECISIONS.md`/`PROJECT.md`, never back into it.
- **Format-faithful, additive writes.** Match each target file's format and leave existing entries intact — an append must not corrupt `DECISIONS.md` or duplicate an entry.
- **Approval per item for reader-affecting writes** (`DECISIONS.md`, `PROJECT.md`). `PLAN.md` is checkpoint's own — overwrite it without prompting.
- **Defer owned files to their owners.** Only `/ds-project-map` refreshes `PROJECT.md`; only `/ds-spec` regenerates `SPEC.md`. Checkpoint appends the occasional `PROJECT.md` fact and otherwise flags drift.
- `## Now` is short and current — overwrite it, do not append. Don't copy the roadmap into `## Now`; state is "where on the roadmap are we", not a duplicate of it.

## Output

The updated `## Now` block, a list of what was routed and where (decisions appended, `PROJECT.md` facts added, files flagged stale), and a one-line "safe to clear" confirmation — plus the handoff path, if written.
