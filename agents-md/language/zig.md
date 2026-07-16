## Language Profile — Zig

Target: Zig 0.16 (current stable). Systems programming, CLIs, embedded, performance-critical code.

Zig is pre-1.0 and breaks across releases — code is written against one toolchain version, not a floor. Pin `minimum_zig_version` in `build.zig.zon` and move it deliberately on upgrade. Tiger Style originates in this world (TigerBeetle), so the constraints below are native to Zig, not bolted on.

Apply these conventions to all Zig code in this session.

### Toolchain

Build with `zig build` (declare targets and steps in `build.zig`, dependencies in `build.zig.zon`). Test with `zig build test` / `zig test`. Format with `zig fmt` — the canonical formatter, no config and no debate — and keep it in CI. Build and ship in `Debug`/`ReleaseSafe` wherever safety matters; reserve `ReleaseFast`/`ReleaseSmall` for code you have measured and where dropping the safety checks is a justified, deliberate trade.

### Memory & Allocators

- No hidden allocations. Every function that allocates takes an `std.mem.Allocator` parameter explicitly — never reach for a global or default allocator.
- Pair each allocation with `defer`/`errdefer` for release in the same scope. Use an arena (`std.heap.ArenaAllocator`) for batch or request-scoped lifetimes you can free in one shot.
- Tests run against `std.testing.allocator` — it fails the test on a leak. In debug builds `std.heap.DebugAllocator` (the renamed `GeneralPurposeAllocator`) catches leaks and use-after-free.
- Prefer slices (`[]T`, bounds-checked) over many-item pointers; pass `[]const T` for read-only views.

### Error Handling

- Errors are values: return error unions (`!T`), define explicit error sets, propagate with `try`, and clean up with `errdefer`. No hidden control flow.
- Don't paper over an error with `catch unreachable` unless the impossibility is genuinely provable — and then assert it. Use `catch |err|` to handle, `orelse` for optionals.
- No `unreachable` or `.?` on values derived from input — only where an invariant guarantees presence.

### Language & Idioms

- `comptime` for generics and compile-time computation instead of macros or runtime reflection — keep logic explicit and traceable.
- Model closed alternatives as a tagged union (`union(enum)`) and `switch` over it exhaustively — omit the `else` branch when you want the compiler to flag unhandled variants as the type grows.
- Optionals (`?T`) for "maybe absent"; unwrap with `orelse` or `if (x) |v|`, never a blind `.?` on an uncertain value.
- No hidden control flow: no operator overloading, no exceptions, no lossy implicit conversions. Make narrowing casts visible (`@intCast`, `@truncate`).
- Be explicit about integer overflow: `+`/`-`/`*` are checked in safe builds — use the wrapping (`+%`) or saturating (`+|`) operators only when wrap/saturate is the actual intent.
- `defer`/`errdefer` for all cleanup, declared right next to the acquisition.

### Testing

- `test "..."` blocks colocated with the code; run with `zig build test`. Use `std.testing.allocator` so leaks fail the test.
- `std.testing.expect`, `expectEqual`, and `expectError` for error paths — exercise the error set, not just the happy path.
- No real network or filesystem in unit tests — inject dependencies and use in-memory fakes; `std.testing.tmpDir` only when a real filesystem is unavoidable.

### Tiger Style

Zig is Tiger Style's native context — apply it fully.
- Assert preconditions, postconditions, and invariants with `std.debug.assert`; assert the things that "can't happen," and assert on both sides of a boundary.
- All loops over external input have explicit upper bounds; no unbounded recursion. Prefer fixed-size buffers with named capacity constants over dynamic growth in hot or embedded paths.
- Named constants for every limit and size — no magic numbers.
