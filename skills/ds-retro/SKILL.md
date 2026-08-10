---
name: ds-retro
description: "Post-release retrospective — compare what SPEC.md/GRILL.md decided against what the release range actually shipped."
disable-model-invocation: true
---

When invoked, run a post-release retrospective over a release range: compare what was decided (`SPEC.md`, `GRILL.md`) against what shipped, and distill rules for the next cycle. The loop-closer to `/ds-spec` and `/ds-grill-me` — they record decisions going in; this judges them coming out.

## Arguments

- A release range, resolved in order:
  - `vPREV..vNEXT` — used verbatim; any valid git range works, branches and SHAs included (`v0.4.0..HEAD` for a retro before tagging).
  - a single ref — the range is previous-version-tag..ref.
  - nothing — the last two version tags (`git tag --sort=-v:refname`). No tags and no argument: ask for an explicit range, never guess.
- `--record [path]` — also write the retro to `RETRO.md` in the current directory, or to `path`. Without the flag, the retro lives in the session only.

## Sources

Read whatever exists; open the report with a source inventory naming what was found and what analysis its absence disables ("GRILL.md ✓ · SPEC.md ✗ — spec-break analysis skipped").

- **Floor:** at least one of `SPEC.md` / `GRILL.md`. Neither → stop and point at `/ds-spec` and `/ds-grill-me`; with no recorded decisions there is nothing to compare, and the output would be a changelog, not a retro.
- **Pinned reads:** read `SPEC.md`/`GRILL.md` at the range end — `git show vNEXT:SPEC.md` — never the working tree; the next cycle may already be amending them. Mid-implementation amendments are the in-file dated `Amended YYYY-MM-DD:` lines whose dates fall inside the range (tag dates come from `git log -1 --format=%ci <tag>`).
- **Enrichment:** `.project/roadmap.md` and `.project/state.md` (`# settled` / `# hazards`) from the working tree — git-ignored, so the working copy is the only copy. Flag them in the report as "current lens, not range-pinned."

## Discipline audit

The amendment discipline (scaffolded by `devskills init --spec-discipline`) makes the decision files self-describing; git's job is to verify the files, not to be the source:

1. `git log --oneline vPREV..vNEXT -- SPEC.md GRILL.md`.
2. Check each commit is a dedicated `docs(spec):` / `docs(grill):` commit and the commit count squares with the amendment lines found in the files.
3. A mismatch — spec edits riding feature commits, changes with no `Amended:` line — is itself a retro finding: report the drift. Only then read `git log -p` on the offending commits to recover what changed.

## Report

Compare decided vs happened across:

- **Assumptions measurement killed** — assumptions the release's reality disproved.
- **Decisions that held** — calls that survived contact, and why.
- **Spec breaks** — shipped behavior that contradicts an unamended decision.
- **Accepted-consequence accuracy** — did the costs knowingly accepted show up as predicted?
- **Unvalidated-at-scale leftovers** — decisions still untested by real use, carried into the next cycle.

Plus the source inventory and any discipline-drift findings.

## RETRO.md (with --record)

- File shape: a `# Retrospectives` header with one intro line; one section per retro, newest first, inserted directly under the header.
- Section shape: `## vPREV..vNEXT — YYYY-MM-DD`, distilled rule bullets first, then `### Evidence` with condensed support. Rules-first is what lets `/ds-grill-me` and `/ds-spec` read the top of a section and stop.
- Idempotent per range: re-running for a range already present replaces that range's own section; other sections are never touched — git history is RETRO.md's audit trail.

## Companion habit

`/ds-grill-me` and `/ds-spec` read `RETRO.md` at invocation (current directory, then repo root) and surface its rules as priors — cited when they bear on the question at hand, overridable by the user, never silently binding.
