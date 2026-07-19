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
2. Break the input into the smallest **vertically-sliced** tasks — each cuts a narrow but complete path through all layers (schema → logic → API → UI → test), shippable and demoable on its own. A single-layer horizontal slice ("all the schema", "all the endpoints") is the anti-pattern — it can't be shipped or verified alone.
3. Write them under `## Roadmap` as a status checklist, appending to existing tasks rather than discarding them:
   - `[ ]` todo · `[~]` in progress · `[x]` done
4. Preserve any `## Now` and top-level `## Watch` sections if present — both belong to `/ds-project-checkpoint`, not here. You own `## Roadmap` and nothing else in the file.
5. Before finalizing, show the breakdown and ask whether the granularity is right — merge or split on the user's steer, don't just emit and move on.

## Rules

- Order and scope only. No library choices, no patterns, no architecture — record what the engineer decides, don't decide it.
- Tasks are outcomes ("parser rejects malformed input"), not instructions on how to implement them.
- Sequence enabling cleanup **first**: if a prefactor makes the real change easy, make it its own early task — make the change easy, then make the easy change.
- **Wide mechanical refactors** (rename a shared symbol/column that breaks every call site at once) don't fit one vertical slice. Sequence **expand → migrate → contract**: add the new form beside the old, migrate call sites in blast-radius-sized batches (each shippable, CI green), then delete the old last, only after every migration batch has landed.
- Keep it terse. No estimates, no owners, no ceremony.

## Output

Display the updated `## Roadmap` and the path.
