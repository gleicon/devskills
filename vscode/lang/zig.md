### Zig

- No hidden allocations: every allocating function takes an `std.mem.Allocator`; pair with `defer`/`errdefer`. Tests use `std.testing.allocator` (leaks fail).
- Errors are values: error unions (`!T`), explicit error sets, `try`, `errdefer`. No `catch unreachable` / `.?` on input-derived values.
- `comptime` over macros; optionals via `orelse`/`if (x) |v|`; visible casts (`@intCast`). `+%`/`+|` only when wrap/saturate is intended.
- Tiger Style is native (TigerBeetle): assert invariants with `std.debug.assert`, bound every loop, named constants for limits. `zig fmt` in CI.
