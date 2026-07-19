---
name: ds-explore
description: "Suggest candidate approaches to a problem — research and lay out options, but never decide."
disable-model-invocation: true
---

When invoked, help the operator think through how to solve a problem by surfacing a few viable approaches with their trade-offs. You suggest; you do not decide and you do not implement. The output is a comparison the operator can act on — and an input to `/ds-grill-me`, which is where the actual decision gets made. (When the decision is specifically a *new system's architecture*, `/ds-blueprint` is the decisive counterpart that commits to one structure.)

## Arguments

- A problem or question to explore. If none is given, ask once, then proceed.
- `--web` — opt-in: do bounded web research (a few high-signal sources, each cited). Off by default; the command works from the project context and your own knowledge.

## Process

1. Establish the problem. If the user gave none, ask once.
2. Read the context so options respect reality:
   - `.project/map.md` and `.project/state.md` if present — honor what `# settled` and `# hazards` record, don't re-litigate them.
   - the relevant code.
3. Gather information:
   - With `--web`: research, bounded — a handful of sources, cite each, no open-ended crawl.
   - Without `--web`: work from the context and your knowledge. **If that is too thin to produce good options, say so and suggest re-running with `--web`** — do not pad with guesses.
4. Lay out **2–4 candidate approaches**. For each: a one-line summary, trade-offs, when you'd pick it, rough effort/risk, and how it fits the existing decisions and constraints.
5. List the **open questions** the choice hinges on — these are what `/ds-grill-me` will walk through.
6. Write the artifact to `EXPLORE.md` in the current directory; return the path.

## Rules

- Suggest, never decide. If you have a lean, state it and mark it explicitly as the operator's call.
- Do not edit code or anything under `.project/`. This command only reads, and writes its own artifact.
- Build options on top of what is already settled — do not reopen a call `state.md` records.
- Stay bounded: 2–4 options, not an exhaustive survey. Cite any web sources.
- `EXPLORE.md` is a scratchpad — overwrite it. It never goes in `.project/`, which holds session state, not work products.

## Output

The candidate approaches and open questions inline, plus the artifact path. End by pointing the operator at `/ds-grill-me` to decide between the options.
