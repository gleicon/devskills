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

## Context Discipline

- Locate before reading: find the relevant place, then read scoped regions — not whole files "to be safe".
- If `PROJECT.md` exists, prefer its map over re-deriving structure. When the map and the code disagree, the code wins.
- Sufficiency beats thrift: when unsure, read more. A wrong answer costs far more than the extra context.

## What to Avoid

- Magic numbers without named constants.
- Global mutable state.
- Deep nesting — flatten with early returns.
- Feature flags for one-time changes — just change the code.
- Backwards-compatibility shims for internal code.
