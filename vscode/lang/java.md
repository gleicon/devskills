### Java

- `record` for immutable data; `sealed` types with exhaustive `switch` pattern matching; `var` for obvious locals.
- No `null` returns for collections/strings (return empty); `Optional` only as a return type; references non-null by default.
- Never swallow exceptions; preserve the cause (`new AppException("ctx", e)`). Catch the most specific type.
- Close `AutoCloseable` with try-with-resources. Virtual threads for I/O-bound fan-out; `StructuredTaskScope` for scoped fan-out (preview in 25), `ScopedValue` over `ThreadLocal`. Every blocking call has a timeout.
- JUnit 5 with `@ParameterizedTest`; `assertThrows` for error paths; no `Thread.sleep` in tests.
