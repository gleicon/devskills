---
name: ds-arena
description: "Fan one task out to several candidates on different models, cross-judge them against a rubric, pick a base, graft the winners' parts into it, and record the synthesis in ARENA.md. Given a review skill as the task, the candidates are findings lists and the merge is a verified, bucketed union."
disable-model-invocation: true
---

When invoked, run the same task N times in parallel, then synthesize — never average. One attempt at a non-trivial artifact locks in the first shape it finds; several independent attempts expose the shape that was not obvious, and different models catch different real bugs. The deliverable is one artifact plus a synthesis note; nothing is applied to the user's tree beyond that.

## Arguments

- A task: the artifact to produce, in the user's words.
- A review skill (`/ds-arena /ds-bug-review`, with any scope the review takes): each candidate runs that review; the merge is a union of findings, not a graft.

## Candidates

- One candidate per distinct model the assistant can delegate to, in one message, each read-only unless the task requires writing. When only one model is reachable, run N independent candidates on it and say so in the note — the diversity is then independence, not model.
- Every candidate gets the same prompt: the task, the same grounding, and an instruction to return the artifact plus a short rationale naming the alternatives it rejected. Candidates never see the rubric.
- Each candidate writes to its own path — a git worktree when the task edits the tree, otherwise a directory under `mktemp -d`. Candidates never share a working tree.
- A candidate that returns nothing is a dropout: proceed with N-1 and record it.

## Task mode

1. **Frame.** State the artifact. Derive a rubric of 3 to 6 gradeable criteria from what success means for this task. Assign paths.
2. **Fan out.** Spawn all candidates in one message. Wait for all of them before judging.
3. **Cross-judge.** One read-only judge, on a model other than your own where possible, sees the rubric and the candidates by label, scores each criterion, and recommends a base with a reason.
4. **Pick.** Read every candidate end to end. Score criterion by criterion, not on feel. Compare with the judge: agreement confirms the base; disagreement means a biased reader or an ambiguous rubric — read both rationales before deciding. Prefer the candidate a maintainer can extend without breaking its invariants; between ties, the smaller surface.
5. **Graft.** Walk each losing candidate once for what is worth porting — usually one or two things, not most of it. Fold each graft in by hand so the result stays coherent under one mental model. When candidates converge on one shape, ship the consensus and record the convergence; when they wildly diverge, the framing was under-specified — reframe and re-run rather than average.
6. **Verify.** The synthesized artifact meets the same bar as any other output: run what can be run. A problem the arena missed means the framing was wrong or a graft was missed; go back, do not paper over.

## Review mode

1. **State the intent** of the change in one paragraph, from the user's words, the commits and the code. Ask if unsure.
2. **Fan out.** A subagent cannot invoke a skill — every `ds-*` skill is user-invoked. Read the review skill's `SKILL.md` and any companion file it names, and paste that text verbatim into each candidate's prompt with the scope, as the instructions to follow. Every candidate gets the identical text.
3. **Merge.** Deduplicate findings that describe one defect in different words; record which models raised each. Consensus across two or more models is the strongest signal; a lone finding is kept and read, never dropped for being alone. Where one model contradicts another, keep both and say so.
4. **Verify** each surviving finding against the code before bucketing it — the union is of confirmed findings, not of claims.
5. **Judge as lead reviewer**, not as an aggregator. Buckets: **act on** (would block a real merge), **consider** (legitimate, cost unclear), **noted** (valid, not actionable now), **dismissed** (wrong or missing context, with one line why).

## ARENA.md

The synthesis note is a work product in the current directory, overwritten on each run, never under `.project/`. Headings, in order:

- Task mode: `## Rubric`, `## Candidates` (model per label, dropouts), `## Base` (which, why, the judge's verdict), `## Grafts` (each with its source candidate), `## Rejected` (what and why), `## Verification`.
- Review mode: `## Intent`, `## Reviewers` (model per label, finding count), `## Act on`, `## Consider`, `## Noted`, `## Dismissed` (each finding with the models that raised it), `## Agreement map` (where models agreed, where they diverged, and what the pattern says).

## Output

The artifact, the path to `ARENA.md`, and the note's headline: the base and grafts, or the act-on list. In review mode nothing is fixed; hand the act-on findings to the user.
