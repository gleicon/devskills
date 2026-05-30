Review Go code with Tiger Style constraints and Go idioms.

Applies to: Go 1.22+. Backend services, CLIs, APIs.

## Arguments

Scan the invocation for the `--no-tiger` flag. Treat every other argument as review scope (files or directories); if no scope is given, review the changed files on the current branch.

- `--no-tiger` present → skip the Tiger Style section; run Go Idioms, Performance, Security, and Testing only.
- `--no-tiger` absent → run all sections (default).

Example: `/go-review --no-tiger dir1/ dir2/` reviews `dir1/` and `dir2/` without Tiger Style.

## Review Checklist

Use the checklist as a lens, not a scorecard: reason about the actual change, report real violations anchored to `file:line`, and flag issues even when they aren't listed. Don't manufacture findings to fill a category. Report only violations — no praise, no summary.

### Tiger Style

Skip this section entirely if `--no-tiger` was passed. Otherwise it is mandatory.
- [ ] Non-trivial functions assert their preconditions and key invariants — don't demand assertions in thin wrappers or trivial accessors
- [ ] All loops have explicit bounds; no unbounded iteration over external input
- [ ] No recursion without provable termination
- [ ] All errors explicitly handled — no `_` discard of error returns
- [ ] No post-initialization dynamic allocation in paths actually identified as hot — don't flag ordinary allocation
- [ ] Functions under 70 lines
- [ ] Variable names include units/qualifiers where applicable

### Go Idioms
- [ ] Errors are values. Check them. Wrap with context using `fmt.Errorf("op: %w", err)`
- [ ] Interfaces defined at point of use (consumer package), not in producer package
- [ ] No exported interface with a single method that could be a function
- [ ] Goroutines have explicit lifecycle: context cancellation, WaitGroup, or channel close
- [ ] No goroutine leak: every goroutine has a clear exit condition
- [ ] Context is first parameter when present: `func F(ctx context.Context, ...)`
- [ ] Prefer table-driven tests
- [ ] No `init()` functions that have side effects
- [ ] Struct fields that are sync primitives (Mutex, WaitGroup) are not copied — pointer receivers used

### Performance
- [ ] No allocation in request hot path (profile with `go test -benchmem` if critical)
- [ ] Slice pre-allocated with `make([]T, 0, knownCap)` where capacity is known
- [ ] `sync.Pool` used for frequently allocated/freed objects in hot paths
- [ ] Database queries not issued inside loops
- [ ] N+1 query patterns absent

### Security
- [ ] No `fmt.Sprintf` constructing SQL queries — use parameterized queries
- [ ] No `exec.Command` with unsanitized user input
- [ ] No hardcoded credentials or secrets
- [ ] HTTP handlers validate and bound all user-controlled inputs
- [ ] `http.Server` sets ReadTimeout/WriteTimeout/IdleTimeout; `http.Client` sets `Timeout` or uses context deadlines

### Testing
- [ ] Public surface and error paths have meaningful coverage — flag notable gaps, not every untested accessor
- [ ] Error paths tested, not just happy path
- [ ] Benchmarks for performance-critical functions
- [ ] No `time.Sleep` in tests — use channels or sync primitives

## Output Format

```
<file>:<line>: <severity>: <problem>. <fix>.
```

Severity levels: `critical` (correctness/security), `major` (reliability/performance), `minor` (idiom/style).

Skip formatting nits unless they affect correctness or readability significantly.
