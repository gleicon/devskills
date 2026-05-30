# Grill Me — Recipes & Extras

`/grill-me` is the most versatile command in the set, so it gets its own page. The base command is tiny — *interview me about this plan, one decision at a time, until we share the same understanding* — but what you point it at changes everything. This doc is a menu of those uses.

For the command's args and behavior see [commands.md](commands.md). For the canonical pre-PR loop see [recipes.md](recipes.md#the-draft-pr-grill-loop). To *generate* the options before grilling, reach for `/explore` first — it surfaces the candidate approaches and open questions that `/grill-me` then walks you through.

## Why it works

The idea is Fred Brooks' (*The Design of Design*): every feature has a **tree of decisions** ahead of it, and the expensive failures come from branches you never noticed you were choosing. `/grill-me` walks that tree one branch at a time — structured rubber-ducking that forces the implicit decisions into the open *before* they cost you a rewrite. It works from a fully-formed spec or from "a vague couple of sentences."

## How to drive a session

The mechanics are the same whatever you point it at:

- **One question at a time**, each with a *recommended answer*. If you agree, just say "yes" and move on — that's what makes a 30–45 minute session cover so much ground.
- It **explores the codebase instead of asking** whenever a question is answerable from the code. Don't answer things it can verify itself.
- **`--record`** appends each resolved decision (question, answer, one-line rationale) to `DECISIONS.md`. Use it — the artifact is half the value.
- It **ends** when no decision branch is left unresolved, with a summary of the resolved plan.

> Tip: the output is an input. Feed the summary or `DECISIONS.md` into `/spec`, into a PR description, or into your issue tracker — don't let a 45-minute conversation evaporate.

## The recipes

### 1. Requirements discovery from a vague idea

Earliest use, before a spec exists. Start from one or two sentences and let the grilling surface the requirements you haven't thought of yet.

```
/grill-me --record  I want a CLI that watches a directory and uploads new files to S3
```

Then turn the resolved decisions into the contract:

```
/spec
```

### 2. Stress-test a spec or design — before any PR exists

When you *have* a design but no code yet. This is the cheapest possible time to find out the design is wrong: no diff to throw away, no reviewer's attention spent.

```
/grill-me --record  the design in SPEC.md: <or paste the design>
```

Point it at the spec, the architecture sketch, or just your stated plan. It will probe the edge cases, the alternatives you skipped, and the invariants you're quietly assuming.

### 3. Review the approach in a draft PR

When the code already exists and you want the *approach* challenged before a human looks at it. This loop — open a draft PR, grill it, incorporate, mark ready — is documented canonically in [recipes.md](recipes.md#the-draft-pr-grill-loop). Use recipe 2 instead when there's no code yet.

### 4. Refactor or migration planning

Before a risky change, grill the *plan for the change* rather than the change itself.

```
/grill-me --record  plan: migrate the auth layer from sessions to JWTs without downtime
```

Forces the rollout order, the backwards-compat window, and the failure/rollback paths into the open. Pair with [`/zoom-out`](commands.md) first if you don't already hold the area in your head.

### 5. Architecture decision → capture an ADR

When the grilling lands on a decision that is **hard to reverse, surprising without context, and the result of a real trade-off**, capture it as an ADR (Architecture Decision Record) so the *why* survives. Ask the agent to write a short `docs/adr/NNNN-<slug>.md` with: context, the decision, the alternatives considered, and the consequences. Offer ADRs *sparingly* — if a decision isn't all three of the above, a line in `DECISIONS.md` is enough.

### 6. Domain & terminology grilling

When the problem is that everyone says "account" and means three different things. Grill the *language* of the domain: every time a term is overloaded or conflicts with how the code uses it, have it called out and a precise canonical term proposed. Optionally capture the resolved terms in a `CONTEXT.md` glossary (kept free of implementation detail — it's a glossary, not a spec).

```
/grill-me  sharpen the domain language for the ordering module — flag every overloaded term against the code
```

> This recipe is inspired by Matt Pocock's `grill-with-docs` skill (see Origins). We ship it as a recipe rather than a separate command — same engine, pointed at terminology.

### 7. Non-coding decisions

Nothing about `/grill-me` is code-specific. It's just as good on a product roadmap, a course outline, a hiring rubric, or any decision with a branching tree. Point it at the plan and let it interrogate.

## Tips & anti-patterns

- **Don't rubber-stamp the recommended answers.** The speed comes from agreeing fast — but the *value* comes from the branches where you disagree. Slow down there.
- **Timebox it.** A full tree can run long; stop when the branches that matter for the next step are resolved, not when every conceivable branch is.
- **Record, then use.** A session with no `DECISIONS.md` and nothing fed into a spec/PR was entertainment, not work.
- **Let it read the code.** Answering questions it could verify itself wastes the session and risks you stating something the code contradicts.
- **Sequence it:** `/zoom-out` (understand) → `/grill-me` (decide) → `/spec` or build. Grilling in code you don't understand yet produces shallow questions.

## Origins

- Fred Brooks, *The Design of Design* — the decision-tree framing.
- [mattpocock/skills — `grill-me`](https://github.com/mattpocock/skills/blob/main/skills/productivity/grill-me/SKILL.md) and [`grill-with-docs`](https://github.com/mattpocock/skills/blob/main/skills/engineering/grill-with-docs/SKILL.md) — the upstream skills `/grill-me` is adapted from (credited in the README References).
- [My "Grill Me" Skill Went Viral](https://www.aihero.dev/my-grill-me-skill-has-gone-viral) — Matt Pocock on how he uses it.
