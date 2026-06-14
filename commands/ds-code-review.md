Review code for single source of truth violations — duplicate implementations, constant drift, parallel-agent conflicts, and unjustified dependencies.

When invoked, diff the current branch against the main branch and report findings. Reports only by default; `--fix` applies the mechanical, unambiguous removals (dead wrappers, redundant helpers, duplicate literals with a clear canonical home). Structural consolidations — merging two implementations, deleting a subsystem — stay reported.

One goal: **maintain a single source of truth.** Prefer existing code over new code. Prefer deletion over addition.

## Arguments

Treat any argument as scope (files or directories). With no scope, review the branch diff against the main branch.

`--fix` → after reporting, apply only the mechanical and behavior-preserving findings. Re-run any build/test/lint check already in the loop; revert any fix that breaks it. Summarize applied vs. left.

## Review Checks

### Duplication

Flag when functionality already exists elsewhere: second HTTP client, second retry, second cache, second validation layer, second serializer, second configuration system, second logging wrapper, second feature flag mechanism.

`duplicate: existing implementation already covers this. Reuse <location>.`

---

### Constant Drift

Flag duplicated literals, configuration values, limits, timeouts, paths, feature flags, environment variables, regexes, or protocol values.

`constant: value duplicated. Reuse <location>.`

---

### Parallel Agent Conflicts

Assume multiple agents may have modified the codebase. Flag: same concept implemented differently in multiple places, competing helpers, multiple utility files solving the same problem, naming divergence for identical concepts.

`conflict: competing implementation exists at <location>. Consolidate.`

---

### Pattern Violations

Prefer project conventions over generic best practices. Flag: new patterns where the project already has one, new architecture style in a different-style codebase, introducing repositories/services/factories when not already used, introducing frameworks for already-solved problems.

`pattern: existing project pattern is <location>. Follow it.`

---

### Idiomatic Usage

Flag: hand-written functionality already provided by the language or stdlib, non-idiomatic loops, unnecessary wrappers, unnecessary interfaces, unnecessary generics, unnecessary inheritance, unnecessary dependency usage.

`idiom: replace with standard language construct.`

---

### Dependency Audit

Every dependency must justify itself. Flag: dependency used for one trivial function, dependency replaceable by stdlib, dependency replaceable by platform feature.

`dependency: remove. Use <replacement>.`

## Output

One finding per line, anchored to `file:line` where possible. End with one verdict:

- `risk: low` — no significant violations
- `risk: medium` — minor duplications or drifted constants; addressable without restructuring
- `risk: high` — duplicate implementations, competing helpers, or multiple sources of truth introduced

If no findings: `Looks consistent. Ship it.`
