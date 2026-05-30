Suggest candidate approaches to a problem — research and lay out options, but never decide.

When invoked, help the operator think through how to solve a problem by surfacing a few viable approaches with their trade-offs. You suggest; you do not decide and you do not implement. The output is a comparison the operator can act on — and an input to `/grill-me`, which is where the actual decision gets made.

## Arguments

- A problem or question to explore. If none is given, ask once, then proceed.
- `--web` — opt-in: do bounded web research (a few high-signal sources, each cited). Off by default; the command works from the project context and your own knowledge.

## Process

1. Establish the problem. If the user gave none, ask once.
2. Read the context so options respect reality:
   - `.project/PROJECT.md`, `.project/PLAN.md`, `.project/DECISIONS.md` if present — honor recorded decisions and constraints, don't re-litigate them.
   - the relevant code.
3. Gather information:
   - With `--web`: research, bounded — a handful of sources, cite each, no open-ended crawl.
   - Without `--web`: work from the context and your knowledge. **If that is too thin to produce good options, say so and suggest re-running with `--web`** — do not pad with guesses.
4. Lay out **2–4 candidate approaches**. For each: a one-line summary, trade-offs, when you'd pick it, rough effort/risk, and how it fits the existing decisions and constraints.
5. List the **open questions** the choice hinges on — these are what `/grill-me` will walk through.
6. Write the artifact to `.project/EXPLORE.md` if `.project/` exists, else a fresh `mktemp -d`; return the path.

## Rules

- Suggest, never decide. If you have a lean, state it and mark it explicitly as the operator's call.
- Do not edit code, `PLAN.md`, `PROJECT.md`, or `DECISIONS.md`. This command only reads and writes its own artifact.
- Build options on top of recorded decisions — do not reopen what `DECISIONS.md` already settled.
- Stay bounded: 2–4 options, not an exhaustive survey. Cite any web sources.
- `.project/EXPLORE.md` is a scratchpad — overwrite it. It is not part of the durable plan/state.

## Output

The candidate approaches and open questions inline, plus the artifact path. End by pointing the operator at `/grill-me` to decide between the options.
