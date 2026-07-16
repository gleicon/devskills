---
name: ds-tiger-style-mode
description: "Activate Tiger Style enforcement for this session."
disable-model-invocation: true
---

Tiger Style is TigerBeetle's engineering philosophy. Source: https://tigerstyle.dev/
Full spec: https://github.com/tigerbeetle/tigerbeetle/blob/main/docs/TIGER_STYLE.md

Apply these constraints to all code you write or review for the remainder of this session.

## Priority Order

Safety > Performance > Developer Experience. Never trade safety for convenience. Never trade correctness for speed.

## Safety Rules

**Assertions**
- Minimum 2 assertions per function: validate arguments on entry, validate return values and invariants before return.
- Assert at the boundary of every external call (OS, disk, network).
- Prefer pair assertions: enforce each property from at least two distinct code paths.
- Assertions are not error handling. They document invariants that MUST hold. If an assertion fails, it is a bug in your code, not the user's input.

**Control Flow**
- No recursion unless termination is formally proven. Use iterative equivalents.
- Every loop has an explicit iteration bound. Assert the bound is not exceeded.
- No goto. No setjmp/longjmp equivalents.
- Minimize branches. Flatten nesting. Early returns are acceptable.

**Memory**
- No dynamic allocation after initialization. Allocate everything upfront, statically.
- No unbounded data structures. Every queue, buffer, and collection has a fixed maximum capacity.
- Assert capacity before every append or push.

**Error Handling**
- Every error must be handled explicitly. No silent discard.
- Do not use exceptions for control flow.
- Propagate errors explicitly through return values, not side channels.

**Dependencies**
- Zero external dependencies unless strictly necessary. Justify each dependency.
- Treat each dependency as a liability: it adds surface area, upgrade cost, and failure modes.

## Performance Rules

**Design-time performance**
- Before writing a non-trivial function, sketch the performance budget: network, disk, memory, compute.
- Batch operations to amortize per-call costs.
- Separate control plane (low-frequency, high-latency-tolerant) from data plane (high-frequency, latency-critical).
- Avoid allocations in hot paths. Prefer pre-allocated buffers.

**Measurement**
- Do not optimize without a measurement. State the baseline before the change.
- Profile at the system level, not the function level.

## Code Structure Rules

**Functions**
- Maximum 70 lines per function. If longer, extract logical units.
- One purpose per function. If you cannot name it with a verb-noun pair, split it.
- Explicit is better than implicit. Pass parameters explicitly; do not rely on global or ambient state.

**Naming**
- snake_case for all identifiers.
- Include units in variable names: `timeout_ms`, `size_bytes`, `count_max`.
- Include qualifiers: `is_valid`, `has_error`, `was_committed`.
- No abbreviations except for loop indices and well-established domain terms.
- Big-endian naming: general concept before specific qualifier (`timestamp_created_at` not `created_at_timestamp`).

**Scope**
- Minimize variable scope. Declare variables as close to use as possible.
- No global mutable state in hot paths.

**Comments**
- No comments that restate what the code does. Code must be readable without them.
- Comments target humans and explain WHY, not WHAT — a hidden constraint, non-obvious invariant, or workaround. One line by default, only where the reason isn't clear from the code; never restate code or cite plan/ticket IDs. A comment past a few lines is rare and signals "this matters" — keep that signal meaningful.
- Assertions often replace comments: if you want to write "x must be positive here", assert it instead.

## Technical Debt

- Zero tolerance. Fix problems during design and implementation.
- If a fix cannot be done now, file a tracked issue with a deadline. Do not leave TODO comments without a ticket.
- Shortcuts taken now cost exponentially more in production.

## Reporting

When flagging a violation during review, name the rule category, the `file:line`, what the code does, and the fix — enough for the author to act.

Confirm activation with "Tiger Style active." Activating a mode only turns on this posture; it is not approval to begin work — continue with whatever the user already asked for, or wait for their next instruction.
