Analyze code for optimization opportunities and produce a graded performance plan — a ranked set of moves, each tagged by the architectural cost it incurs (L1/L2/L3). Language-agnostic. Reports a plan; changes nothing.

When invoked, analyze the code in scope against one question: **where is this doing more work than it needs to, and what would each speedup cost?** Not "is it clean" (that's `/ds-code-quality-review`), not frontend render performance (that's `/ds-ui-quality-review`), not the idiom-level checklist the language reviews already run — this is the cross-cutting hotspot hunt that names the *price* of each optimization so a speedup that damages the architecture is a deliberate choice, not a silent one. Every finding carries a cost model; a "slow" claim with no model behind it is noise. **Do not edit any files.** When a finding is worth pursuing, `/ds-verify-this` proves the win with a same-machine baseline/treatment.

## Arguments

- Treat positional args as scope (files, directories, globs). With no scope, review the code changed on the current branch.
- `--max-level=<1|2|3>` (or freeform "no architecture changes", "free wins only") → clamp: suppress findings above that tier. Default: report all tiers, tagged.
- `--no-tiger` → skip the Tiger Style section; run the hotspot hunt and cost models only.

## Levels

Every finding is tagged with the architectural cost of *applying* it:

- **L1 — Free win.** Behavior- *and* structure-preserving: better algorithm/data structure within the same boundaries, remove redundant work, hoist invariants, fix N+1, cut hot-path allocations, batch IO. Safe to apply as-is.
- **L2 — Localized restructuring.** Preserves public contracts and layer boundaries, but changes a module's internals: caching/memoization, a data-structure change that ripples within the module, precompute, reshape a loop nest.
- **L3 — Architectural / boundary-breaking.** Trades clean architecture for speed: denormalize, inline across layers, collapse abstractions, hand-rolled allocators, SIMD, API-reshaping batching. **State the architecture/clarity cost explicitly** — per Tiger Style this trade is never made blindly.

## What to hunt

Read the actual execution path and estimate cost — don't pattern-match "loops are slow." Run a benchmark or profile when it's cheap; measured beats argued.

**1. Algorithmic complexity.** Quadratic-or-worse work on input that grows, nested loops over the same collection, repeated recomputation of a value that's invariant in the loop, accidental O(n²) from `contains`/`indexOf` inside a loop.

**2. Data structures.** Linear scan where a map/set/index fits, wrong container for the access pattern, sorted-vs-hashed mismatch, missing DB index behind a frequent query.

**3. Redundant work.** Re-fetching, re-serializing, re-parsing, or recomputing across calls; work inside a loop that could be hoisted; eager computation of something rarely used.

**4. Allocation & memory.** Allocation in a path *actually identified as hot*, avoidable copies, unbounded growth/retention. Do not flag ordinary allocation — only where a cost model shows it matters.

**5. I/O & data access.** N+1 queries, chatty round-trips that could batch, synchronous IO on a latency-critical path, missing caching where the data is safely cacheable.

**6. Concurrency & parallelism.** Serial work that's embarrassingly parallel, lock contention, false sharing, oversized critical sections. Propose carefully — concurrency trades correctness risk for speed.

**7. Architecture-level (L3).** Where a clean boundary forces repeated cost: denormalization, crossing-layer batching, collapsing an abstraction that's on a hot path.

## Output

A plan: the candidate moves ranked by **impact ÷ cost** (free, high-impact wins first), grouped by level. Respect `--max-level` if given.

For each finding:

- Anchor to `file:line`.
- State the hotspot in one line.
- **Give the cost model** — *why* it's slow: Big-O, allocation/IO/query/syscall counts, or a measured profile. No cost model → don't report it.
- The optimization, and its **level tag**. For L2/L3, name the architecture/clarity cost it incurs.
- **Evidence label**: `measured` (profile/benchmark), `reasoned` (sound cost model, unmeasured), or `speculative` (plausible, unproven) — plus your confidence.
- The `/ds-verify-this` claim that would prove the win (e.g. "`tool process big.json` runs ≥2× faster on this branch vs parent").

Rules:

- Real, measurable wins over theoretical ones. A short, high-confidence list beats a long speculative one.
- Never propose an optimization that trades away correctness or safety (`Safety > Performance`). If a speedup adds a correctness risk, say so and treat it as a cost.
- Don't optimize without a measurement — state the baseline. Mark unmeasured findings `reasoned`/`speculative` honestly.
- Change nothing. The output is the list.

## Tiger Style (skipped with `--no-tiger`)

- Sketch the performance budget — network, disk, memory, compute — before proposing a change.
- Separate control plane (low-frequency) from data plane (high-frequency, latency-critical); spend the effort on the data plane.
- Prefer pre-allocated buffers; no dynamic allocation in hot paths after initialization.
- `Safety > Performance > Developer Experience`. Never trade correctness for speed.
