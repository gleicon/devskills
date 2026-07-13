---
name: ds-tdd-mode
description: "Drive implementation with tests, one vertical slice at a time."
disable-model-invocation: true
---

When invoked, build the feature test-first. Each test verifies observable behavior through a public interface — never implementation details.

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
- Break during refactoring when behavior has not changed

A test that breaks on a behavior-preserving refactor is coupled to implementation. Rewrite it against the interface.

## Design for testability

- **Inject dependencies** — accept them as parameters; do not construct them inside the unit.
- **Prefer pure functions** — return results rather than mutating shared state.
- **Keep the API surface small** — fewer methods and parameters mean fewer test scenarios. Deep modules (small interface, substantial hidden implementation) test better than shallow ones.

## Output

For each slice: the test, then the implementation. Report which behaviors remain untested.
