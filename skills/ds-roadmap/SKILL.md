---
name: ds-roadmap
description: "Turn a goal or another command's output into an ordered task roadmap."
disable-model-invocation: true
---

When invoked, sequence work into a `## Roadmap` checklist. You order the work; you do not choose the architecture or make technical decisions — those are the engineer's. The companion to `/ds-spec`: spec defines the WHAT, this orders the work to get there.

## Input

Use whatever the user provides as the source of tasks:
- a goal or feature description,
- a `SPEC.md` (`.project/SPEC.md` if present),
- pasted output from another command (e.g. `/ds-code-quality-review` findings, a list of bugs).

If nothing is provided, ask once for the goal, then proceed.

## Process

1. Resolve the target file: `.project/PLAN.md` if `.project/` exists, else `PLAN.md` in the current directory. Read it first if it already exists. Do not create `.project/` — write to the current directory when it's absent.
2. Break the input into the smallest ordered, independently-shippable tasks.
3. Write them under `## Roadmap` as a status checklist, appending to existing tasks rather than discarding them:
   - `[ ]` todo · `[~]` in progress · `[x]` done
4. Preserve a `## Now` section if present — that section belongs to `/ds-project-checkpoint`, not here.

## Rules

- Order and scope only. No library choices, no patterns, no architecture — record what the engineer decides, don't decide it.
- Tasks are outcomes ("parser rejects malformed input"), not instructions on how to implement them.
- Keep it terse. No estimates, no owners, no ceremony.

## Output

Display the updated `## Roadmap` and the path.
