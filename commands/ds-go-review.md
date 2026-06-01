Review Go code with Tiger Style constraints and Go idioms.

Applies to: Go 1.24+. Backend services, CLIs, APIs.

## Arguments

Scan the invocation for the `--no-tiger` and `--fix` flags. Treat every other argument as review scope (files or directories); if no scope is given, review the changed files on the current branch.

- `--no-tiger` present → skip the Tiger Style section; run Go Idioms, Performance, Security, and Testing only.
- `--no-tiger` absent → run all sections (default).
- `--fix` → after reporting, apply only the violations whose fix is **mechanical and unambiguous** (a rename to the idiom, a missing error check the review is certain about). Anything that changes logic or rests on an unverified assumption — especially security and correctness findings — **stays report-only**. After applying, re-run any build/test/lint check already in the loop and revert any fix that breaks it — or that touched more than the intended mechanical edit. End with a summary of what was applied and what was left.

Example: `/ds-go-review --no-tiger dir1/ dir2/` reviews `dir1/` and `dir2/` without Tiger Style.

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
- [ ] Prefer stdlib `slices`/`maps`/`cmp` and the `min`/`max`/`clear` builtins over hand-rolled equivalents; `for i := range n` for integer counts
- [ ] Sequences exposed as `iter.Seq`/`iter.Seq2` where iteration is the public API (Go 1.23+)
- [ ] `json:",omitzero"` (not `omitempty`) when the intent is to omit zero values — notably zero `time.Time`
- [ ] Tool dependencies declared in `go.mod` (`go get -tool`, run via `go tool`), not a legacy `tools.go` blank-import shim (Go 1.24+)
- [ ] Legacy idioms a modernizer would rewrite are flagged — pre-`slices`/`maps` helpers, `interface{}` over `any`, manual `b.N` loops (run `golangci-lint`'s `modernize`, or `go fix` on Go 1.26+)

### Performance
- [ ] No allocation in request hot path (profile with `go test -benchmem` if critical)
- [ ] Slice pre-allocated with `make([]T, 0, knownCap)` where capacity is known
- [ ] `sync.Pool` used for frequently allocated/freed objects in hot paths
- [ ] Database queries not issued inside loops
- [ ] N+1 query patterns absent

### Security
- [ ] No `fmt.Sprintf` constructing SQL queries — use parameterized queries
- [ ] No `exec.Command` with unsanitized user input
- [ ] Filesystem paths built from user input go through `os.OpenRoot`/`*os.Root`, not manual `filepath.Clean`/prefix checks — blocks traversal and symlink escape (Go 1.24+)
- [ ] No hardcoded credentials or secrets
- [ ] HTTP handlers validate and bound all user-controlled inputs
- [ ] `http.Server` sets ReadTimeout/WriteTimeout/IdleTimeout; `http.Client` sets `Timeout` or uses context deadlines

### Testing
- [ ] Public surface and error paths have meaningful coverage — flag notable gaps, not every untested accessor
- [ ] Error paths tested, not just happy path
- [ ] Benchmarks for performance-critical functions, written with `for b.Loop()` rather than a manual `for i := 0; i < b.N` loop (Go 1.24)
- [ ] No `time.Sleep` in tests — use channels or sync primitives, or `testing/synctest`'s fake clock for time-dependent concurrency (Go 1.25+)

## Output Format

```
<file>:<line>: <severity>: <problem>. <fix>.
```

Severity levels: `critical` (correctness/security), `major` (reliability/performance), `minor` (idiom/style).

Skip formatting nits unless they affect correctness or readability significantly.
