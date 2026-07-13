---
name: ds
description: "The devskills router — names every skill by phase (orient → spec → plan → build → clean → review → verify → ship) and when to reach for it."
disable-model-invocation: true
---

The index to the `ds-*` family. Every skill is user-invoked, so nothing here fires on its own — this names each one and when to reach for it. Read it, pick the skill that fits, invoke that skill. `/ds` itself does no work.

The phase spine is the usual arc of a change; the groups below it (modes, language reviews, project memory, utilities) cut across phases — reach for them whenever they apply. Everything works standalone: no skill requires `.project/` or any other.

## Phase spine

### Orient — *"Where am I? What does this code do?"*
- `/ds-zoom-out` — step up a level and map how the current code fits the bigger picture, before you change it.

To restore a persisted session, see `/ds-project-resume` under Project memory.

### Spec — *"What exactly are we building?"*
- `/ds-spec` — turn a description into a structured spec with acceptance criteria.
- `/ds-explore` — at a fork, research and lay out candidate approaches without deciding.
- `/ds-grill-me` — get interviewed about a plan or design until the understanding is shared and the gaps are exposed.

### Plan — *"In what order — and what structure?"*
- `/ds-blueprint` — design a target architecture for a **new** system: module boundaries, dependency rules, seams, build order.
- `/ds-architecture-plan` — assess an **existing** codebase and produce a sequenced refactoring plan.
- `/ds-roadmap` — turn a goal or another skill's output into an ordered task roadmap.
- `/ds-perf-plan` — analyze code for optimization opportunities and produce a graded, ranked performance plan.

### Build — *"Make it work."*
- `/ds-debug` — find the root cause of a failure with the scientific method: reproduce, isolate, fix, prove.

Turn on the relevant **Modes** (below) for the build — they set the engineering bar as you write.

### Clean — *"Strip the AI slop before anyone looks at it."*
- `/ds-deslop` — remove narrating comments, defensive overkill, and type escape hatches; align the diff with the surrounding code. Run before any review.

### Review — *"Is it correct, safe, idiomatic, maintainable?"*
- `/ds-code-quality-review` — maintainability + single-source-of-truth: is the diff making the codebase worse?
- `/ds-bug-review` — correctness: real bugs, not style.
- `/ds-security-review` — exploitable weaknesses; reach for it when the change touches input, auth, secrets, or I/O.
- `/ds-data-review` — data-correctness, integrity, and migration-safety.
- `/ds-test-quality-review` — is the risky logic actually covered, and are the tests real?
- `/ds-doc-quality-review` — docs accuracy against the code, broken links, staleness.
- `/ds-ui-quality-review` — is the interface soundly engineered, crafted, and accessible?
- `/ds-comment-review` — do the comments earn their place?
- `/ds-notebook-review` — notebook state, output hygiene, reproducibility.
- `/ds-quality-gate` — run the review pipeline as a gate over the whole branch/feature.
- `/ds-osv` — scan dependencies for known vulnerabilities (OSV).

For a specific language, prefer its **Language review** (below) — it folds in idioms and security.

### Verify — *"Does it actually do what I claimed?"*
- `/ds-verify-this` — before/after repro with a hard verdict (VERIFIED / NOT VERIFIED). Give it something measurable; it refuses vague claims.

### Ship — *"Hand it off."*
- `/ds-handoff` — compact the session into a handoff doc so a fresh agent or another person can continue.

To persist `.project/` state before shipping, see `/ds-project-checkpoint` under Project memory. Then `git push` + `gh pr create`.

## Modes — standing postures, compose anytime

Turn one on and it governs the rest of the session; several can be active at once.

- `/ds-tiger-style-mode` — safety + explicitness engineering bar.
- `/ds-senior-mode` — write like a super-senior engineer everywhere: terse, step-gated, no slop (folds in git/test/deslop/step).
- `/ds-git-mode` — commit each working unit as it lands, terse messages, never pushes.
- `/ds-step-mode` — smallest reviewable step, then hand back.
- `/ds-test-mode` — keep the core tested by risk as you build.
- `/ds-tdd-mode` — drive implementation test-first, one vertical slice at a time.
- `/ds-ui-mode` — component/state discipline + design craft for UI work.
- `/ds-data-mode` — rigor for data/analytics work.

## Language reviews — one per language, idioms + security

`/ds-go-review` · `/ds-python-review` · `/ds-rust-review` · `/ds-java-review` · `/ds-ts-review` · `/ds-zig-review`

Each reviews with Tiger Style constraints plus that language's idioms; reach for it over the general reviews when the diff is single-language.

## Project memory — optional `.project/` persistence

Durable context across sessions. The workflow runs fine without any of these.

- `/ds-project-map` — map the repo into `.project/PROJECT.md`.
- `/ds-project-config` — set project preferences (e.g. modes auto-applied on resume) in `.project/config.md`.
- `/ds-project-resume` — restore where you left off and apply configured modes. Start an ongoing project here.
- `/ds-project-checkpoint` — sweep the session and route durable context to its owning file before `/clear` or end of session.
- `/ds-project-compact` — housekeeping over the persisted `.project/` state.

## Local knowledge & utilities

- `/ds-recall` — inject prior local context from recall into this session.
- `/ds-recall-capture` — store this session's outcome into recall's knowledge base.
- `/ds-recall-setup` — initialize recall and its session integration.
- `/ds-tldt` — extractive summary of a long doc before it enters context.
- `/ds-typeset` — turn a Markdown file into a self-contained, themed HTML document.
