Review Zig code with Tiger Style constraints and Zig idioms.

Applies to: Zig 0.16 (current stable). Systems programming, CLIs, embedded.

Zig is Tiger Style's native context (TigerBeetle): explicit allocators, no hidden control flow, and no hidden allocations map almost one-to-one onto the Tiger Style constraints. Lean into that section harder than for other languages.

## Arguments

Scan the invocation for the `--no-tiger` and `--fix` flags. Treat every other argument as review scope (files or directories); if no scope is given, review the changed files on the current branch.

- `--no-tiger` present → skip the Tiger Style section; run Memory & Allocators, Error Handling, Zig Idioms, Safety, and Testing only.
- `--no-tiger` absent → run all sections (default).
- `--fix` → after reporting, apply only the violations whose fix is **mechanical and unambiguous** (a rename to the idiom, a missing error check the review is certain about). Anything that changes logic or rests on an unverified assumption — especially security and correctness findings — **stays report-only**. After applying, re-run any build/test/lint check already in the loop and revert any fix that breaks it — or that touched more than the intended mechanical edit. End with a summary of what was applied and what was left.

Example: `/ds-zig-review --no-tiger src/` reviews `src/` without Tiger Style.

## Automated Checks (run first if tools are available)

```bash
zig fmt --check .                    # formatting drift
zig build test                       # tests + std.testing.allocator leak detection
zig build -Doptimize=ReleaseSafe     # compile + safety-check validation
```

Run these, report what they surface, then do the manual review below. They are baseline context — anchor findings to the code in scope; don't report pre-existing failures outside the change as if it introduced them.

## Review Checklist

Use the checklist as a lens, not a scorecard: reason about the actual change, report real violations anchored to `file:line`, and flag issues even when they aren't listed. Don't manufacture findings to fill a category. Report only violations — no praise, no summary.

### Tiger Style

Skip this section entirely if `--no-tiger` was passed. Otherwise it is mandatory.
- [ ] Non-trivial functions assert preconditions, postconditions, and key invariants with `std.debug.assert` — assert the "can't happen" cases and on both sides of a boundary; don't demand assertions in trivial accessors
- [ ] All loops over external input have explicit upper bounds; no unbounded recursion without provable termination
- [ ] Named constants for every limit, size, and capacity — no unexplained magic numbers
- [ ] Functions under 70 lines
- [ ] Errors handled or propagated explicitly — never silently dropped

### Memory & Allocators
- [ ] Every allocating function takes an `std.mem.Allocator` explicitly — no hidden global/default allocator
- [ ] Each allocation paired with `defer`/`errdefer` for release in the same scope; no leak on the error path
- [ ] Arena (`ArenaAllocator`) used for batch/request-scoped lifetimes instead of scattered individual frees
- [ ] Slices (`[]T`, bounds-checked) preferred over many-item pointers; `[]const T` for read-only views
- [ ] No use-after-free: freed memory not retained or returned; ownership of returned allocations is clear

### Error Handling
- [ ] Errors returned as error unions (`!T`) with explicit error sets, not encoded in sentinels or out-params
- [ ] Propagated with `try`; cleanup on the error path via `errdefer`
- [ ] No `catch unreachable` unless the impossibility is provable (and asserted); no swallowed errors via empty `catch`
- [ ] No `unreachable` or `.?` on values derived from input — only where an invariant guarantees presence

### Zig Idioms
- [ ] `comptime` for generics/compile-time computation rather than runtime reflection workarounds
- [ ] Closed alternatives modeled as a tagged union (`union(enum)`) with an exhaustive `switch` — no catch-all `else` where the compiler should force handling of new variants
- [ ] Optionals (`?T`) unwrapped with `orelse` / `if (x) |v|`, not a blind `.?`
- [ ] No hidden control flow introduced — no reliance on implicit conversions; narrowing casts made visible (`@intCast`, `@truncate`)
- [ ] Integer overflow intent explicit: `+%`/`+|` only where wrap/saturate is wanted, plain `+`/`-`/`*` otherwise
- [ ] `defer`/`errdefer` declared next to the acquisition they release

### Safety
- [ ] Code intended for production runs in `Debug`/`ReleaseSafe`; any `ReleaseFast`/`ReleaseSmall` choice that drops safety checks is deliberate and justified
- [ ] All external/untrusted input validated and bounded (lengths, indices, sizes) before use
- [ ] No out-of-bounds risk: slice indices and lengths checked; no raw pointer arithmetic without an asserted bound
- [ ] No command/shell execution or filesystem path built from unsanitized input; no hardcoded secrets

### Testing
- [ ] `test "..."` blocks cover the public surface and error paths — flag notable gaps, not every trivial accessor
- [ ] Tests use `std.testing.allocator` so leaks fail the test
- [ ] Error paths exercised with `expectError`, not just the happy path
- [ ] No real network/filesystem in unit tests — injected fakes; `std.testing.tmpDir` only where a real FS is unavoidable

## Output Format

```
<file>:<line>: <severity>: <problem>. <fix>.
```

Severity levels: `critical` (memory safety / correctness), `major` (reliability / leak), `minor` (idiom/style).

Skip formatting nits unless they affect correctness or readability significantly.
