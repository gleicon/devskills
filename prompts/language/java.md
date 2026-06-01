## Language Profile — Java

Target: Java 25+ (LTS). Backend services, APIs, CLIs, systems tooling.

Apply these conventions to all Java code in this session.

### Toolchain

Build with Maven or Gradle — whichever the repo already uses; don't introduce a second build tool. Test with JUnit 5 (`mvn test` / `gradle test`). Format with the repo's configured formatter, wired into the build (`google-java-format` via the Spotless plugin is a reasonable default when a project has none). Enable compiler lint (`javac -Xlint:all`) and treat warnings as errors where the project allows, plus a static-analysis tool the project standardizes on (Error Prone, SpotBugs, …).

### Project Layout

Standard `src/main/java` / `src/test/java` source roots with the package tree mirrored under each. One public top-level type per file. Package by feature, not by layer (`orders/` over `services/` + `dtos/`); avoid `util`/`common`/`helpers` grab-bag packages. Keep packages cohesive and depend inward.

### Language & Idioms

- `record` for immutable data carriers; `sealed` interfaces/classes with exhaustive `switch` pattern matching for closed type hierarchies. Prefer `switch` expressions over statement chains.
- `var` for local variables where the type is obvious from the right-hand side; spell out the type when it aids readability.
- Flexible constructor bodies: validate and normalize arguments *before* `super(...)`/`this(...)` instead of deferring to a factory wrapper. *(Java 25)*
- Favor immutability: `final` fields, unmodifiable collections (`List.of`, `Map.of`, `Collectors.toUnmodifiableList`). Don't expose internal mutable collections.
- Streams for transformation pipelines, but keep them readable — a plain loop beats a contorted stream. No side effects inside `map`/`filter`.
- `Optional` as a return type for "maybe absent"; never as a field or method parameter, and never call `.get()` without a presence check.
- Don't return `null` for collections, arrays, or strings — return an empty value and reserve `Optional` for a scalar that may be absent. Treat references as non-null by default, validating at boundaries with `Objects.requireNonNull`; mark the deliberate exceptions `@Nullable` if the project has adopted JSpecify.

### Error Handling

- Prefer unchecked exceptions for application errors — they compose with streams and lambdas and don't leak implementation across architectural seams. Reserve checked exceptions for the rare genuinely recoverable, caller-actionable condition, and don't let them cross layer boundaries. (Don't blanket-wrap in `RuntimeException` to dodge the compiler either — model real error types.)
- No empty `catch` blocks. Preserve the cause: `throw new AppException("context", e)`. Don't log-and-rethrow the same exception at every layer.
- Catch the most specific exception type; never `catch (Exception e)` (or `Throwable`) as a catch-all unless re-raising.

### Resources & Concurrency

- Always close resources with try-with-resources; never rely on manual `finally` for `AutoCloseable`.
- Prefer the `java.util.concurrent` high-level constructs (`ExecutorService`, `CompletableFuture`, concurrent collections) over raw `Thread` and `synchronized`/`wait`/`notify`.
- Use virtual threads (`Executors.newVirtualThreadPerTaskExecutor()`) for I/O-bound fan-out; don't pool virtual threads. Every blocking call has a timeout.
- For subtasks that must succeed or fail together, prefer structured concurrency (`StructuredTaskScope.open(...)` with a `Joiner`) over hand-rolled `Future` juggling — it scopes cancellation and error propagation to the call. *(Preview in 25 — requires `--enable-preview`; widely used in practice as the virtual-thread fan-out idiom.)*
- Share immutable per-request context with `ScopedValue` rather than `ThreadLocal` — immutable, structured, and virtual-thread-friendly. *(Java 25)*
- Guard shared mutable state explicitly; prefer immutable objects and `java.util.concurrent.atomic` over locks where possible.

### Testing

- JUnit 5 with its built-in assertions (`assertEquals`, `assertThrows`); `@ParameterizedTest` for input variants instead of copy-pasted cases. A fluent assertion library (AssertJ) is fine when the project already uses it.
- No real network or filesystem in unit tests — prefer hand-written fakes and JUnit's `@TempDir`; reach for a mocking library (Mockito) only where a fake is impractical. Assert on behavior through the public API.
- Test error paths with `assertThrows`, not just the happy path. No `Thread.sleep` to coordinate tests — await on latches/futures with a bounded timeout.

### Tiger Style

- Non-trivial methods validate their preconditions (`Objects.requireNonNull`, guard clauses, or `assert` for internal invariants); don't assert in trivial getters or thin wrappers.
- All loops over external input have explicit bounds; no unbounded recursion without provable termination.
