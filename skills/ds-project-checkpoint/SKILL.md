---
name: ds-project-checkpoint
description: "Sweep the session for durable context and route each piece to its owning file, so the session can be cleared or ended safely."
disable-model-invocation: true
---

When invoked, make sure nothing durable is lost before `/clear` or the end of a session — the counterpart to `/ds-project-resume`. Sweep the conversation for context a fresh agent will need, write each piece to its right home, update the plan state, and confirm it's safe to clear. Run it after *any* session — this guards the decisions and facts that otherwise live only in the conversation.

## Arguments

- `--handoff` — also write a full handoff to `.project/handoff.md` (richer than the `## Now` block: context, what was tried, gotchas). Use it when the next session needs more than the plan state — handing off to another person, or a long pause.

## Process

1. Create `.project/` if needed.
2. **Sweep the session** for anything durable that isn't already persisted somewhere — resolved decisions, new structural/map facts, scope/requirement changes, current state, code-keyed constraints someone will trip over later, and soft context (things to keep watching).
3. **Route each item to its owning home:**
   - **Resolved decision** → append to `.project/DECISIONS.md` (create it on first write): one entry per decision — the question, the chosen answer, a one-line rationale; plain Markdown, leave existing entries intact. If it reverses a decision already recorded, mark the original too (see Rules).
   - **New stable structural/map fact** (a new top-level dir, dependency, build/test command) → append to the right section of `.project/PROJECT.md`. Never rewrite or refresh existing sections, never touch human-authored prose or constraints. If `PROJECT.md` is absent, or the map looks broadly stale (more than a one-line fact), don't edit it — flag it ("run `/ds-project-map`").
   - **Code-keyed constraint** — a hazard whoever edits will walk into ("the iOS probe is load-bearing, don't simplify it", "never let a Tailwind class touch `${`") → append a one-line imperative row to `## Landmines` in `.project/PROJECT.md`, keyed by **scope**: a file path, a glob, or `repo-wide`. Create the section on first write with its contract as the header line — *read the row before editing that scope; never recite it*. If a decision already carries the rationale, point at it instead of repeating it.
   - **Scope/requirement change** → record it as a decision in `DECISIONS.md`, then flag `SPEC.md` as stale in `## Watch` ("no longer reflects decision X — re-run `/ds-spec`"). Never edit `SPEC.md` — it is a structured artifact owned by `/ds-spec`.
   - **State / next / open questions** → the `## Now` section of `.project/PLAN.md`; **soft context** → the top-level `## Watch` section (step 5).
   - **Roadmap progress** → `## Roadmap` statuses in `PLAN.md` (step 5).
4. **Dedupe, then ask only where you are inferring.** Compare against what each file already holds and add only what's missing — don't duplicate the map, an existing decision, or anything already recorded. Then write what you are transcribing, and ask about what you concluded:
   - **Write it** — a decision the user stated outright, a structural fact, a landmine they described, a `## Watch` entry, and everything in `PLAN.md`.
   - **Ask** — a decision *inferred* from discussion (the session weighed three options and you decided one won), a supersession call (it retires a decision that is still governing), and any write that would alter existing prose rather than append to it.

   Show the whole batch at once and let the user strike what doesn't belong; never walk them through it item by item. Most sessions have nothing in the ask column — write, report, done.
5. **Update `PLAN.md`** (automatic):
   - `## Roadmap` task statuses to match reality (`[ ]` / `[~]` / `[x]`).
   - Rewrite `## Now`: **State** (where things stand — 2 lines, and no feature inventories; what shipped is `git log`'s job, not State's), **Next** (the single next action), and **Open questions** *only when something unresolved actually blocks progress* — when nothing does, leave the heading out altogether rather than writing "none" beneath it.
   - `## Watch` — a **top-level section beside `## Now`, not inside it**. Append a one-line entry for anything unresolved that needs no decision now (a doc flagged stale, a deferred upgrade, an advisory); delete entries that are resolved. Never rewrite the section wholesale: a Watch entry outlives many sessions, and `## Now` is rewritten every one — that is the whole reason it lives outside.
   - Increment `<!-- checkpoints-since-verify: N -->` at the foot of the file (start it at 1 if absent). It is a comment so it stays out of the human sections; `/ds-project-verify` resets it, and `/ds-project-resume` uses it to know when drift is worth a look.
   - **First run on a project whose `## Watch` sits inside `## Now`:** split it once. Code-keyed constraints go to `## Landmines` in `PROJECT.md`, everything else to the promoted `## Watch`. Show the split as one batch and let the user strike items — re-filing their own words is transcription, but judging which bullet is a landmine is not. Ask once per project, never again.
6. **No-op fast path:** if the sweep found nothing durable beyond state, say so — just update `## Now`/roadmap and report "nothing durable to route."
7. **If `--handoff`:** also write `.project/handoff.md` (current goal, what's done, what remains, key decisions, gotchas). Reference existing artifacts (`PLAN.md`, `DECISIONS.md`, commits) by path rather than duplicating them. **Redact secrets/PII — the handoff is plaintext a fresh session will read.**
8. **Confirm safe to clear** — state what is now persisted and where.

## Rules

- **Route to the owning file; don't pile everything into one.** A resolved decision goes to `DECISIONS.md`, a map fact to `PROJECT.md`, state to `PLAN.md` `## Now` — never summarize all of it into one file.
- **Don't invent new files.** Soft context lives in `## Watch`, not a new notes file. `EXPLORE.md` stays a scratchpad — durable findings route to `DECISIONS.md`/`PROJECT.md`, never back into it.
- **Format-faithful, additive writes.** Match each target file's format and leave existing entries intact — an append must not corrupt `DECISIONS.md` or duplicate an entry.
- **Record nothing that changes without a decision.** A merge SHA, a released version, "X is device-verified" — settled events, safe to write down. A branch position, a test/file/line count, "tree is clean" — these rot on the very next commit, including one this same session makes, and every later read then reports your own file as stale. Anything countable or derivable is read from the repo when it's needed, never frozen into prose.
- **Reversing a decision marks the original, never edits it** — the one exception to leaving entries intact. Append the new decision, then add a one-line pointer on the superseded one. An unmarked reversed decision is indistinguishable from a live one, and everything downstream reads it as still governing.
- **Ask about inference, not about files.** What gates a write is whether you concluded it or the user said it — never which file it lands in. An approval prompt that is always answered "yes, all" is not a control, it is a delay that teaches rubber-stamping, and it lets a genuinely wrong entry through in the same breath as nine obvious ones.
- **Sort by when it is needed, not by what it is about.** Something that needs a decision at session start — a stale doc, a deferred upgrade — goes to `## Watch`. Something that only matters while editing particular code goes to `## Landmines`, whose header says to apply it and not read it aloud. A single section holding both is what makes a resume recite constraints back at the person who wrote them.
- **Defer owned files to their owners.** Only `/ds-project-map` refreshes `PROJECT.md`; only `/ds-spec` regenerates `SPEC.md`. Checkpoint appends the occasional `PROJECT.md` fact or landmine row and otherwise flags drift.
- `## Now` is short and current — overwrite it, do not append. Don't copy the roadmap into `## Now`; state is "where on the roadmap are we", not a duplicate of it. A slot with nothing to say is dropped, not filled with "none" — an empty heading invites the next writer to find something to put under it.

## Output

The updated `## Now` block, a list of what was routed and where (decisions appended, `PROJECT.md` facts and landmine rows added, `## Watch` entries added or cleared, files flagged stale), and a one-line "safe to clear" confirmation — plus the handoff path, if written.
