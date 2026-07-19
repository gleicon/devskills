---
name: ds-blueprint
description: "Design a target architecture for a new system — module boundaries, dependency rules, seams, and build order — from its requirements. Reports a blueprint; changes nothing."
disable-model-invocation: true
---

When invoked, take the requirements (a `SPEC.md`, a chosen approach from `/ds-explore`, or a freeform description) and commit to a concrete architecture. This is the **decisive** counterpart to `/ds-explore`: where explore surveys options and abstains, this **recommends one** structure and notes key alternatives briefly. It describes the structural *how*, not the behavioral *what* (that's `/ds-spec`). To critique or refactor an architecture that *already exists*, use `/ds-architecture-plan` instead. **Do not write code or scaffold files** — the output is the plan. Hand it to `/ds-roadmap` to turn the build order into tasks.

## Arguments

- Positional args / freeform: the requirements source (a path like `SPEC.md`, a description, or a chosen approach). With none, ask for the requirements or read `SPEC.md` if present.
- `--no-tiger` → skip the Tiger Style / simplicity section.

## Principles

- **Simplicity first (YAGNI).** Default to the simplest architecture that meets the stated requirements. Every layer, boundary, queue, cache, or service split must trace to a requirement — not to "best practice" or imagined future scale.
- **Deep modules.** Maximize behavior hidden behind a small interface. For each module, ask: can I cut methods, simplify parameters, push more complexity inside? A wide interface over a thin implementation — a shallow pass-through — is a design smell, not a layer.
- **Testable by construction.** Accept dependencies rather than constructing them inside a module; return results rather than mutating shared state — so seams fall out of the design instead of being retrofitted.
- **No structure without a requirement that demands it.** Generic patterns (hexagonal, DDD, microservices, event sourcing) are proposed only when a stated need forces them, and the need is named.
- **Decisive.** Recommend one architecture. Note the strongest alternative in a line or two and why you didn't pick it.
- **Design for the actual scale**, not an imagined one.

## Process

1. Read the requirements; extract the forces that actually shape structure — data and its lifecycle, the integrations, the consistency/latency/scale needs, the team's constraints.
2. Choose the simplest decomposition that satisfies those forces. Name the modules/boundaries and what each owns. For the load-bearing modules, sketch 2–3 genuinely different interfaces (minimize surface / maximize flexibility / optimize the common caller) before committing — your first cut is rarely the deepest.
3. Fix the dependency direction rules (what may depend on what; keep it acyclic).
4. Identify the seams — where to test, where the system can change later without a rewrite.
5. Order the build: a walking skeleton first, then the increments. State what's deliberately deferred.

## Output

A target-architecture blueprint:

- **Shape** — the chosen architecture in 3–5 lines, and the tier it's pitched at (walking skeleton → full target) with why.
- **Modules / boundaries** — each with its single responsibility and what it owns, plus its **contract**: key invariants, ordering/error modes, and any config or performance characteristic a caller must know — not just the responsibility.
- **Dependency rules** — the allowed direction(s); confirm it's acyclic.
- **Seams** — where behavior is tested and where the design can flex later. Classify each dependency: in-process (no seam), local-substitutable (a test stand-in), remote-but-owned (define a port + real & in-memory adapters), or true-external (inject a port, mock it). One adapter means a hypothetical seam, two means a real one — don't add a port unless something actually varies.
- **Build order** — the walking skeleton, then increments, each independently shippable.
- **Deferred** — what's intentionally left out, and what requirement would justify adding it.
- **Alternative considered** — the strongest other option and why it lost.

Rules:

- No structural element without a requirement behind it. Name the requirement.
- Before committing to a structure, check `# settled` in `.project/state.md` (if present) so you don't re-decide something already made; if the blueprint would override one, say so explicitly.
- Prefer the simpler structure; when unsure, defer and say what would justify escalating.
- `Safety > Performance > Developer Experience`.
- Change nothing. The output is the blueprint.

## Tiger Style (skipped with `--no-tiger`)

- Simplicity first; the simplest structure that meets the requirements. No speculative generality.
- Explicit boundaries; separate control plane from data plane.
- Design for the actual scale, not an imagined one.
- `Safety > Performance > Developer Experience`.
