Remove AI-generated code slop introduced on the current branch and align it with the surrounding codebase.

When invoked, diff the current branch against the main branch and remove the slop the branch introduced. Preserve behavior.

The test for slop is simple: **code an experienced engineer working in this language and this codebase wouldn't have written.** Judge against the surrounding code and the language's idioms — not a fixed checklist.

## Arguments

Treat any argument as scope (files or directories). With no scope, review the branch's diff against the main branch.

## What to look for

- **Narrating comments** — comments that restate what the code does, or break the file's existing comment density and style.
- **Defensive overkill** — error handling or guards abnormal for a trusted path. The form is language-specific: a swallowed `try/catch`, an `if err != nil` that can't fire, an `unwrap_or` masking a real invariant, a null check on a value that's never null.
- **Type escape hatches used only to dodge the checker** — judged by the language's norms: `any` / `as` / `@ts-ignore` in TypeScript or `# type: ignore` in Python are usually slop; `any` / `interface{}` in Go is often idiomatic. Flag the ones that exist only to silence a type error, not the idiomatic uses.
- **Needless nesting** — deep conditionals that read better as early returns or guard clauses (or the language's equivalent: `?` in Rust, `match`, and so on).
- **Speculative scaffolding** — unused helpers, premature abstractions, config knobs with one possible value, placeholder TODOs.
- **Anything else inconsistent** with the file and surrounding code.

## Guardrails

- Apply the principle, not the syntax — what's slop in one language is idiomatic in another.
- Keep behavior unchanged unless fixing a clear bug.
- Prefer minimal, focused edits over broad rewrites. Match each file's existing conventions; do not impose a new style.

## Output

Apply the edits, then give a concise 1–3 sentence summary of what was removed or simplified.
