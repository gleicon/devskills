Turn a goal or another command's output into an ordered task list in `.project/PLAN.md`.

When invoked, sequence work into `.project/PLAN.md` under `## Roadmap`. You order the work; you do not choose the architecture or make technical decisions — those are the engineer's.

## Input

Use whatever the user provides as the source of tasks:
- a goal or feature description,
- a `.project/SPEC.md` (or `SPEC.md`),
- pasted output from another command (e.g. `/code-quality-review` findings, a list of bugs).

If nothing is provided, ask once for the goal, then proceed.

## Process

1. Create `.project/` if needed. Read `.project/PLAN.md` if it exists.
2. Break the input into the smallest ordered, independently-shippable tasks.
3. Write them under `## Roadmap` as a status checklist, appending to existing tasks rather than discarding them:
   - `[ ]` todo · `[~]` in progress · `[x]` done
4. Preserve the `## Now` section if present — that section belongs to `/project-checkpoint`, not here.

## Rules

- Order and scope only. No library choices, no patterns, no architecture — record what the engineer decides, don't decide it.
- Tasks are outcomes ("parser rejects malformed input"), not instructions on how to implement them.
- Keep it terse. No estimates, no owners, no ceremony.

## Output

Display the updated `## Roadmap` and the path.
