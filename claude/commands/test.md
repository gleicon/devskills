Activate pragmatic testing mode for this session.

When active, make sure the things that matter get well tested as you build — without test-first ceremony. This is the pragmatic sibling of `/tdd`: `/tdd` drives the design by writing the test first; this rides alongside normal work and ensures the risk is covered before you call something done. To audit tests that already exist, use `/test-quality-review`.

## What earns a test

Test by risk, not by rule. Spend tests where a failure would actually hurt:

- Core domain logic, money/auth/permissions, data integrity, parsing, state machines — anything with real branching.
- The edge cases bugs hide in: boundaries (empty, zero, one, max, off-by-one), invalid input, error paths, partial failure, concurrency and ordering.
- Every bug you fix: a test that reproduces it first, so it can't come back.

Skip the trivia — getters, glue code, framework behavior, the obvious. A test there is maintenance with no payoff.

## What a good test looks like

- Exercises real behavior through the public interface — not internals.
- Survives a behavior-preserving refactor unchanged. If it breaks when nothing observable changed, it was testing the wrong thing.
- One logical assertion; a name that states the behavior; deterministic (no wall-clock, network, or test-order dependence).

## Don't lock the design

A test coupled to implementation is a liability, not an asset — it taxes every change and trains the team to ignore red. As you write tests, avoid:

- Mocking internal collaborators; asserting on call counts or call order.
- Testing private methods, or reaching through a side channel instead of the interface.
- Snapshotting everything as a substitute for asserting what actually matters.

Prefer fewer, sharper tests that would each catch a real failure over many that chase coverage. Coverage is a side effect of testing the right things, never the goal.

Confirm activation with "Pragmatic testing mode active." then proceed.
