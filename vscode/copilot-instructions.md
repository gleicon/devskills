# GitHub Copilot Instructions

These instructions apply to all GitHub Copilot interactions in this workspace.

## Engineering Philosophy

Apply Tiger Style principles (https://tigerstyle.dev/) to all generated code:

- Safety first, performance second, developer experience third.
- Minimum 2 assertions per function (validate arguments on entry, return values before exit).
- No recursion without provable termination.
- All loops have explicit upper bounds.
- Every error is handled explicitly. No silent discard.
- Functions under 70 lines.
- Variable names include units and qualifiers: `timeout_ms`, `size_bytes`, `is_valid`.
- Zero external dependencies unless strictly necessary.

## Code Generation Rules

- Write only what is asked. No speculative abstractions.
- No comments that restate what the code does.
- No placeholder TODO comments without a linked issue.
- Explicit over implicit. Pass parameters directly; do not rely on ambient state.
- Validate all inputs at system boundaries (user input, external APIs, file reads).

## Language-Specific Notes

### Go

- Wrap errors: `fmt.Errorf("context: %w", err)`.
- Context is first parameter.
- Every goroutine has explicit exit condition.
- Table-driven tests.

### TypeScript

- `strict: true`. No `any`.
- Discriminated unions for state.
- Runtime validation at boundaries.
- Named exports. No barrel files inside feature folders.

### Rust

- `thiserror` for library errors, `anyhow` for binaries.
- No `.unwrap()` in production paths.
- `cargo clippy -- -D warnings` must pass.

### JavaScript

- `const` by default.
- ESM (`import`/`export`) everywhere.
- Async errors always handled.

### Python

- Type-annotated; passes `mypy --strict`. Modern syntax (`list[str]`, `X | None`).
- Catch specific exceptions, never bare `except:`. Chain with `raise ... from err`.
- No mutable default arguments. Resources via `with`.
- `pytest` with plain `assert`; `ruff` for lint/format, `uv` for deps.

## What to Avoid

- Magic numbers without named constants.
- Global mutable state.
- Deep nesting — flatten with early returns.
- Feature flags for one-time changes — just change the code.
- Backwards-compatibility shims for internal code.
