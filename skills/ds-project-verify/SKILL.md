---
name: ds-project-verify
description: "Reconcile the `.project/` memory files against the code — correct what is mechanically checkable, flag what needs a human."
disable-model-invocation: true
---

When invoked, read the `.project/` files against the actual source and find the claims that stopped being true. Nothing else in the family does this: `/ds-project-checkpoint` sweeps the *session*, `/ds-project-compact` sorts entries by whether they still govern, and neither one opens the code. A statement that was accurate when written and was quietly falsified by a later change has no other way of being caught.

Drift is silent by construction, so it accumulates in projects that are being maintained *well* — every session checkpointed, every write approved. Diligence at recording is not a check on accuracy.

## Process

1. If `.project/` does not exist, say so and stop.
2. **Delegate the sweep to a subagent.** Cross-checking every claim in `SPEC.md`, `DECISIONS.md`, and `PROJECT.md` against source is a large read whose intermediate output nobody needs. Send it out, take back a findings list.
3. **Check each class separately** — they fail differently:
   - **`## Landmines` rows in `PROJECT.md`** — two checks, kept apart. *Scope resolves*: does the path or glob still match anything? Mechanical. *Constraint still holds*: does the code still behave the way the row assumes? That needs reading it, so it is judgment. A row whose file exists can be entirely dead — the fragile thing was rewritten last quarter and the warning outlived it.
   - **`DECISIONS.md`** — entries a later decision reversed that carry no supersession marker. Unmarked, they read as current.
   - **`SPEC.md`** — FR/NFR/AC the code contradicts.
   - **`PROJECT.md` prose** — Overview and Constraints claims the code contradicts.
4. **Correct only what is knowable without the user.** Show the batch once and let them strike items:
   - add a supersession marker to an unmarked reversed decision — additive, and the reversal is already recorded;
   - remove a `## Landmines` row whose scope resolves to nothing.
5. **Flag everything else** into `PLAN.md`'s `## Watch`, one line each, naming the file and what the code says instead.
6. **Reset the counter** — set `<!-- checkpoints-since-verify: 0 -->` at the foot of `PLAN.md`.
7. **Report by what you actually checked** (see Rules).

## Rules

- **Never edit `SPEC.md`.** If FR-3 says X and the code does Y, you cannot tell from here whether the spec drifted or the **code regressed** — the fixes are opposite. Rewriting the spec to match the code would launder a regression into a requirement, and the requirement is the thing that would have caught it. Flag it; a human decides which side is wrong.
- **Never rewrite `PROJECT.md` prose** for the same reason. Overview and Constraints are human-authored claims about intent, and code disagreeing with intent is exactly the case a person needs to rule on.
- **Say which check you ran.** "12 landmine scopes resolve; none checked for whether they still hold" is honest. "12 landmines verified" is the false clean bill of health this command exists to stop — a report stronger than its check is worse than no report, because it ends the search.
- **Never delete a landmine on a guess.** A warning removed in error is one nobody gets a second time. Scope-resolves-to-nothing is the only mechanical basis for removal; anything softer gets flagged.
- Findings are per-claim and anchored — name the file, quote the claim, say what the source shows.

## Output

Findings grouped by file, each with the claim and what the code shows; the batch of corrections applied; the flags written to `## Watch`; and an explicit line stating which checks ran and which did not — mechanical scope resolution and substantive truth are reported separately, never merged into one number.
