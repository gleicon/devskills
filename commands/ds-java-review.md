Review Java code with Tiger Style constraints and Java idioms.

Applies to: Java 25+ (LTS). Backend services, APIs, CLIs.

## Arguments

Scan the invocation for the `--no-tiger` and `--fix` flags. Treat every other argument as review scope (files or directories); if no scope is given, review the changed files on the current branch.

- `--no-tiger` present → skip the Tiger Style section; run Java Idioms, Performance, Security, and Testing only.
- `--no-tiger` absent → run all sections (default).
- `--fix` → after reporting, apply only the violations whose fix is **mechanical and unambiguous** (a rename to the idiom, a missing error check the review is certain about). Anything that changes logic or rests on an unverified assumption — especially security and correctness findings — **stays report-only**. After applying, re-run any build/test/lint check already in the loop and revert any fix that breaks it — or that touched more than the intended mechanical edit. End with a summary of what was applied and what was left.

Example: `/ds-java-review --no-tiger src/main/ src/test/` reviews both trees without Tiger Style.

## Review Checklist

Use the checklist as a lens, not a scorecard: reason about the actual change, report real violations anchored to `file:line`, and flag issues even when they aren't listed. Don't manufacture findings to fill a category. Report only violations — no praise, no summary.

### Tiger Style

Skip this section entirely if `--no-tiger` was passed. Otherwise it is mandatory.
- [ ] Non-trivial methods validate preconditions and key invariants (`Objects.requireNonNull`, guard clauses, `assert` for internal invariants) — don't demand checks in trivial getters or thin wrappers
- [ ] All loops over external input have explicit bounds; no unbounded iteration
- [ ] No recursion without provable termination
- [ ] All exceptions handled or propagated deliberately — no empty `catch` blocks, no swallowed errors
- [ ] No post-initialization allocation in paths actually identified as hot — don't flag ordinary allocation
- [ ] Methods under 70 lines
- [ ] Variable names include units/qualifiers where applicable

### Java Idioms
- [ ] `record` for immutable data carriers instead of hand-written boilerplate classes
- [ ] `sealed` interfaces/classes with exhaustive `switch` pattern matching for closed hierarchies; `switch` expressions over statement chains
- [ ] Immutability favored: `final` fields, `List.of`/`Map.of`/unmodifiable collections; internal mutable collections not exposed
- [ ] `Optional` used only as a return type — not as a field or parameter; no `.get()` without a presence check
- [ ] No `null` returned for collections/arrays/strings — return an empty value; references treated as non-null by default and validated at boundaries (`Objects.requireNonNull`)
- [ ] Resources closed with try-with-resources, not manual `finally` on `AutoCloseable`
- [ ] Catch the most specific exception; no `catch (Exception)`/`Throwable` catch-all unless re-raising; cause preserved (`new X("ctx", e)`)
- [ ] Streams used for transformation without side effects in `map`/`filter` — and a plain loop preferred where a stream would be contorted
- [ ] `var` only where the right-hand side makes the type obvious
- [ ] Concurrent fan-out uses `StructuredTaskScope` over manual `Future` juggling, and `ScopedValue` over `ThreadLocal` for per-request context, where the project is on 25 (structured concurrency is preview — `--enable-preview`)
- [ ] No `static` mutable shared state; constants are `static final`

### Performance
- [ ] No allocation or boxing in hot paths that profiling has identified (autoboxing in tight loops, `String` concatenation in loops → `StringBuilder`)
- [ ] Collections pre-sized with a known capacity (`new ArrayList<>(n)`, `new HashMap<>(n)`)
- [ ] Database queries not issued inside loops; N+1 patterns absent
- [ ] Streams not materialized to intermediate lists unnecessarily; large data streamed rather than fully buffered

### Security
- [ ] No string-built SQL (concatenation/`String.format`) — use `PreparedStatement` parameters or the ORM
- [ ] No command injection via `Runtime.exec`/`ProcessBuilder` with unsanitized input
- [ ] No unsafe deserialization of untrusted data (Java native serialization, unconfigured XML/YAML parsers); XML parsers disable external entities (XXE)
- [ ] Filesystem paths from user input validated and confined to an expected root — no path traversal
- [ ] No hardcoded credentials or secrets; read from env/secret store
- [ ] All external input validated and bounded at the boundary; network/HTTP calls set timeouts

### Testing
- [ ] Public surface and error paths have meaningful coverage — flag notable gaps, not every untested accessor
- [ ] Error paths tested with `assertThrows`, not just the happy path
- [ ] Input variants parametrized (`@ParameterizedTest`) rather than copy-pasted
- [ ] No real network/filesystem in unit tests — fakes/mocks and `@TempDir`; no `Thread.sleep` to coordinate — await latches/futures with a timeout

## Output Format

```
<file>:<line>: <severity>: <problem>. <fix>.
```

Severity levels: `critical` (correctness/security), `major` (reliability/performance), `minor` (idiom/style).

Skip formatting nits unless they affect correctness or readability significantly.
