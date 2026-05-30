Restore working context from `.project/PLAN.md` (and a fresh handoff, if any).

When invoked, read the project's persisted state and report where to pick up — the counterpart to `/project-checkpoint`. Safe to run at the start of any session.

## Process

1. If `.project/PLAN.md` does not exist, say so and suggest `/project-map` then `/project-plan`. Stop.
2. Read `.project/PLAN.md` — focus on `## Now` (state, next, open questions) and the `## Roadmap` status.
3. Read `.project/PROJECT.md` if present, for the repo map and constraints.
4. If `.project/handoff.md` exists, check whether it is still current — by **file modification time, not git** (the workflow must work when `.project/` is git-ignored or the repo has no git):
   - If `handoff.md` is newer than `.project/PLAN.md`, load it — it's the freshest context.
   - If it is older than `PLAN.md` (a checkpoint happened after it), treat it as **stale**: mention it exists and its date, but do not rely on it.
   - If the repo uses git, you may *optionally* also flag the handoff as stale when commits have landed since it was written — but never require git; the file-time comparison is the source of truth.
5. Summarize: current state, the next action, open questions, and anything stale worth noting.

## Rules

- Read-only. Do not modify `.project/` files.
- Trust `## Now` over a stale `handoff.md`.

## Output

A short orientation — where we are, the next step, and open questions — enough to start working immediately.
