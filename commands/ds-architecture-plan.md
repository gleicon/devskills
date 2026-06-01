Analyze an existing codebase's architecture and produce a sequenced refactoring plan — assess module boundaries, dependency structure, and layering, then lay out ordered, risk-tagged steps. Language-agnostic. Reports a plan; changes nothing.

When invoked, audit the code against one question: **is the architecture itself sound — and if not, what's the highest-leverage way to fix it, and in what order?** Not file/function simplification (that's `/ds-code-quality-review`, which works *within* the architecture), not a map (that's `/ds-zoom-out`, which renders no judgment). This operates at the **module / dependency / boundary altitude** and is willing to propose big, sequenced structural change. Every finding is anchored to a concrete symptom in *this* codebase — generic best-practice with no local evidence is banned. Proposed refactors preserve observable behavior. **Do not edit any files.** Map first with `/ds-zoom-out`; turn the roadmap into tasks with `/ds-project-plan`; add characterization tests at a seam with `/ds-test-mode` or `/ds-tdd-mode` before a risky move. (To design a *new* architecture rather than review an existing one, use `/ds-blueprint`.)

## Arguments

- Positional args are scope (directories, packages, the whole repo). No scope → the whole project (or the changed area, if you say so).
- `--max-level=<1|2|3>` → clamp: suppress steps above that tier. `--max-level=1` = safe, in-place wins only.
- `--no-tiger` → skip the Tiger Style / simplicity section.

## Levels (risk / blast-radius)

- **L1 — In-place, low-risk**: move/rename, extract a module within existing boundaries, break a cycle with an interface, relocate misplaced logic. Individually shippable, behavior-preserving.
- **L2 — Restructure within the current style**: introduce a missing layer/boundary, consolidate duplicated subsystems, invert a dependency direction, split a god-package.
- **L3 — Architecture-style change**: monolith→modular, ports-and-adapters, service split, persistence-boundary change. Justify why L1/L2 won't suffice; give a migration path.

## What to assess

Build the picture from evidence — the dependency/import graph, and (when available) which files change together in git history. Don't pattern-match a style onto the code.

**1. Module boundaries & cohesion.** God packages/files, grab-bag `utils`/`common`/`helpers` modules, low cohesion, unclear ownership of a concept.

**2. Dependency structure.** Import cycles, dependency-direction violations (a lower layer importing an upper one), hub modules everything depends on, unstable things depended upon by stable ones.

**3. Layering & separation of concerns.** Business logic in the wrong layer (in HTTP handlers, in the DB layer, in views); transport/persistence/domain bleeding into each other; missing seams.

**4. Coupling & change amplification.** Modules that always change together (shotgun surgery), a change that ripples across unrelated areas, hidden temporal coupling.

**5. Duplicated subsystems.** Parallel implementations of one concept, divergent copies, multiple sources of truth.

**6. System-level abstraction leaks.** Implementation details leaking through a public API; a boundary that doesn't actually encapsulate.

**7. Architectural-style fit (L3).** Does the style (or its absence) match the system's actual needs and scale? Only with strong evidence — this is where cargo-culting hides.

## Output

A sequenced refactoring roadmap.

First, a 3–5 line **assessment**: the current architecture (what style, what's healthy, what's not) and the 2–3 highest-leverage problems.

Then **ordered steps**, ranked by leverage (impact ÷ risk-effort) within ordering constraints. Respect `--max-level`. Each step:

- The move, in one line, with its **level tag** (L1/L2/L3).
- **The symptom it fixes** — anchored to evidence: `file:line`, an import-cycle path, or the set of files that co-change. No symptom → not a step.
- **Why now / what it unblocks** — the ordering rationale.
- **Blast radius & risk.**
- **Safety**: whether characterization tests are needed at the seam *before* the move (point to `/ds-test-mode` or `/ds-tdd-mode`); confirm observable behavior is preserved.

Rules:

- **No recommendation without a concrete symptom in this codebase.** Generic best-practice with no local evidence is banned.
- **Simplicity first** — prefer the change that removes structure over the one that adds it; resist speculative generality.
- Behavior-preserving: refactors keep observable behavior; recommend characterization tests before risky moves.
- `Safety > Performance > Developer Experience`.
- Each step independently shippable where possible.
- Change nothing. The output is the plan.

## Tiger Style (skipped with `--no-tiger`)

- Simplicity first; the simplest structure that meets the requirements. No speculative generality.
- Explicit boundaries; separate control plane from data plane.
- Design for the actual scale, not an imagined one.
- `Safety > Performance > Developer Experience`.
