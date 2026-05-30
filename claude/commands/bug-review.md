Run a strict correctness review of code changes — hunt real bugs, not style or structure. Reports a findings list; changes nothing.

When invoked, audit the code in scope against one question: **will this misbehave at runtime?** Not "is it clean" (that's `/code-quality-review`), not "is it idiomatic" (that's the language reviews) — *is it correct.* Find the defects that would actually fire: wrong logic, mishandled edges, ignored failures, races, leaks. Every finding must name the condition that triggers it; a bug nobody can reach is noise.

Like `/code-quality-review` and `/doc-quality-review`, this produces a prioritized list. **Do not edit any files.** When a finding is confirmed, `/debug` root-causes the fix and `/verify-this` proves it.

## Arguments

- Treat positional args as scope (files, directories, globs). With no scope, review the code changed on the current branch.
- Freeform scope ("the parser", "the whole diff") is interpreted reasonably.

## What to hunt

Read the actual execution path — don't pattern-match. Run the tests or a quick repro when it's cheap; a bug you can trigger beats one you can argue about.

**1. Logic errors.** Inverted or wrong conditionals, off-by-one, wrong operator (`<` vs `<=`, `&&` vs `||`), wrong comparison, mishandled edge values (empty, zero, negative, max), incorrect arithmetic or unit/sign mistakes.

**2. Null / absent values.** Dereferences of something that can actually be null/nil/None/undefined on some path; unchecked optionals; missing-key access. Only where the value *can* be absent — a guard on something never null is the opposite problem.

**3. Error and failure paths.** Swallowed or ignored errors; the unhappy path left half-done (state mutated, then a later step fails with no rollback → inconsistent state); resources not released on the error path — files, locks, connections, transactions, goroutines/tasks.

**4. Boundary and data handling.** Integer overflow/truncation, precision loss, unsafe type coercion, encoding/charset assumptions, unbounded input, parsing that trusts its shape.

**5. Concurrency.** Data races on shared state, check-then-act (TOCTOU), deadlock and lock-ordering, missing synchronization, `await`/async mistakes, assumptions about ordering or atomicity that don't hold.

**6. Control flow & lifecycle.** Missing `break`/`return`, unintended fallthrough, an early return that skips cleanup, use-after-close/free, double-free/close, uninitialized use, stale or never-invalidated cache.

**7. Contract misuse.** Wrong argument order, ignored return values that signal failure, off-contract use of a library or API, mismatched units between caller and callee.

## Output

A prioritized findings list, ordered by severity (likelihood × impact):

1. Critical — data loss/corruption, crash on a reachable path, or a security-relevant defect
2. Likely-wrong — produces incorrect results under normal use
3. Edge-case — fires only on specific, less-common inputs

For each finding:

- Anchor to `file:line`.
- State the bug in one line, **name the exact condition that triggers it**, then the fix.
- Give your confidence. If it rests on an assumption you couldn't verify, say which.

Rules:

- Real defects only. No "could theoretically be null" without a path that reaches it — that's how false positives bury the real ones.
- A short, high-confidence list beats a long speculative one.
- Change nothing. The output is the list.
