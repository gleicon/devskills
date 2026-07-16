---
name: ds-tdd-mode
description: "Drive implementation with tests, one vertical slice at a time."
disable-model-invocation: true
---

When active, build the feature test-first. Each test verifies observable behavior through a public interface — never implementation details.

## Anti-pattern: horizontal slicing

Do not write all tests before any implementation. Tests written against imagined behavior verify guessed data shapes, not real user-facing outcomes. Slice vertically instead: one test, one implementation, repeat.

## Workflow

1. **Plan** — Confirm the interface changes needed. List the behaviors to test, highest-value first. Design for testability (see below). Get user approval before writing code.
2. **Tracer bullet** — Prove the path end to end: one test, then the minimal implementation that passes it.
3. **Incremental loop** — Write one test, implement the minimal code to pass it. Do not anticipate future tests.
4. **Refactor** — Only once tests pass. Remove duplication and apply design principles while keeping every test green.

## Good tests vs bad tests

Good tests:
- Exercise real code paths through the public API
- Describe WHAT the system does, not HOW
- Make one logical assertion
- Survive internal refactors unchanged

Bad tests — red flags:
- Mock internal collaborators
- Assert on call counts or call order
- Test private methods
- Verify through a side channel (e.g. a raw DB query) instead of the interface
- Recompute the expected value the way the code does (`assert add(a, b) == a + b`) — the test can never disagree with the code. Draw expected values from an independent source: a known-good literal, a worked example, or the spec.
- Break during refactoring when behavior has not changed

A test that breaks on a behavior-preserving refactor is coupled to implementation. Rewrite it against the interface.

## Design for testability

- **Inject dependencies** — accept them as parameters; do not construct them inside the unit.
- **Prefer pure functions** — return results rather than mutating shared state.
- **Keep the API surface small** — fewer methods and parameters mean fewer test scenarios. Deep modules (small interface, substantial hidden implementation) test better than shallow ones.
- **Mock only at system boundaries** — external APIs, the database, time, randomness, the filesystem. Never mock your own classes or internal collaborators; prefer a real or in-memory implementation (mocking internals couples the test to the design).

## Output

For each slice: the test, then the implementation. Report which behaviors remain untested.

Confirm activation with "TDD mode active." Activating a mode only turns on this posture; it is not approval to begin work — continue with whatever the user already asked for, or wait for their next instruction.
