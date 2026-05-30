Run a strict review of test quality — whether the code that matters is well tested, and whether the tests are good ones. Reports a findings list; changes nothing.

When invoked, audit the tests in scope against one governing principle: **test what matters, and test it well.** Coverage is not the goal — a percentage says nothing about whether a test would catch a real failure, or whether it pins the code in place. Hunt two failures with equal energy: critical code that is *under-tested* (the bug waiting to ship) and tests that are *bad* (present and green, but worthless or actively harmful — they lock the design and tax every change).

Like `/code-quality-review` and `/doc-quality-review`, this produces a prioritized list. It's the after-the-fact audit; to write tests this way from the start, use `/test` (alongside normal work) or `/tdd` (test-first). **Do not edit any files.**

## Arguments

- Treat positional args as scope (files, directories, globs). With no scope, review the tests covering code changed on the current branch.
- Freeform scope ("the auth module", "the whole suite") is interpreted reasonably.

## What to check

Identify the code worth testing first, then judge the tests against it. Run the suite when it's safe and fast — a test that can't run isn't protecting anything.

**1. Is the critical code tested at all? (highest priority.)** Find the code that would hurt most if it broke — core domain logic, money/auth/permissions, data integrity, parsing, state machines, anything with real branching — and check it has tests that would actually fail if it regressed. A gap here matters far more than a high coverage number elsewhere. Ignore the trivia: getters, glue, framework behavior, the obvious.

**2. Edge cases and failure modes.** For the risky logic, are the bugs-that-actually-happen covered — boundaries (empty, zero, one, max, off-by-one), invalid and malicious input, error paths, partial failure, concurrency and ordering? Happy-path-only tests on branching logic are a gap.

**3. Test quality — does it test behavior?** Good tests exercise real behavior through the public interface and survive a behavior-preserving refactor. Flag tests that:
- mock internal collaborators, or assert on call counts / call order
- reach through a side channel (raw DB query, private method) instead of the interface
- assert nothing meaningful — a real break would still pass them
- are non-deterministic: depend on wall-clock, network, or test order

**4. Tests that lock the design (worse than no test).** A test coupled to implementation breaks on every refactor that changes nothing observable. It doesn't prevent failures — it taxes change and trains people to ignore red. Mock-the-world, snapshot-everything, and internals-asserting tests belong here. Recommend rewriting against the interface, or deleting.

**5. Bloat and noise.** Tests of trivial code, duplicated assertions, tests that re-verify the framework or the language itself. They add maintenance and run time without buying confidence.

**Not coverage.** Do not recommend tests to hit a number. A well-tested core with untested trivia is the correct state, not a finding. If you see coverage-chasing — tests written only to color a line green — call it out as the anti-pattern it is.

## Output

A prioritized findings list, in this order:

1. Critical/core code with no real test (the failure waiting to ship)
2. Missing edge cases and failure modes on risky logic
3. Tests that lock the design (implementation-coupled — rewrite or delete)
4. Weak tests that wouldn't catch a real break
5. Bloat — tests to cut

For each finding:

- Anchor to `file:line`.
- State the problem in one line, then the suggested fix — and when the fix is "delete it", say so plainly.
- For a coverage gap, name the specific behavior or edge case that's unprotected, not "needs more tests".

Rules:

- A short high-conviction list beats a long pedantic one. Don't pad it with low-value nits.
- Judge by risk, not count. The question is always "would this catch a real failure, and does it survive a refactor?"
- Change nothing. The output is the list.
