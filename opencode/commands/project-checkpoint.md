Persist current state into `.project/PLAN.md` so the session can be cleared or ended safely.

When invoked, update `.project/PLAN.md` so a fresh session can pick up exactly where this one left off. Run it before `/clear` or at the end of a session.

## Arguments

- `--handoff` — also write a full handoff to `.project/handoff.md` (richer than the `## Now` block: context, what was tried, gotchas). Use it when the next session needs more than the plan state — handing off to another person, or a long pause.

## Process

1. Create `.project/` if needed.
2. Update `## Roadmap` task statuses in `.project/PLAN.md` to match reality (`[ ]` / `[~]` / `[x]`).
3. Rewrite the `## Now` section with:
   - **State** — where things stand, in 2–4 lines.
   - **Next** — the single next action.
   - **Open questions** — anything unresolved that blocks progress.
4. If `--handoff`: write `.project/handoff.md` (current goal, what's done, what remains, key decisions, gotchas). Reference existing artifacts (`PLAN.md`, `DECISIONS.md`, commits) by path rather than duplicating them.

## Rules

- `## Now` is short and current — overwrite it, do not append.
- Do not copy the roadmap into `## Now`; state is "where on the roadmap are we", not a duplicate of it.

## Output

Display the updated `## Now` block (and the handoff path, if written).
