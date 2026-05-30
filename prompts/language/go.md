## Language Profile — Go

Target: Go 1.22+. Backend services, CLIs, APIs, systems tooling.

Apply these conventions to all Go code in this session.

### Toolchain

Test with `go test -race ./...`. Lint with `golangci-lint run`. Benchmark with `go test -bench=. -benchmem`.

### Project Layout

`cmd/<name>/` for entrypoints, `internal/` for private packages, `pkg/` only for a real library, `api/` for protobuf/OpenAPI. Name packages by what they provide — avoid `util/`, `common/`, `helpers/`.

### Error Handling

- Wrap with context: `fmt.Errorf("operation: %w", err)`. Never discard with `_` — every error is handled or propagated.
- Sentinel errors (`var ErrNotFound = errors.New(...)`) checked with `errors.Is`, not `==`.

### Concurrency

- Every goroutine has an explicit exit condition. `context.Context` for cancellation, always the first parameter.
- `errgroup` for parallel work with error collection; `sync.WaitGroup` for fan-out, channels for coordination.
- Protect shared state with a named `sync.Mutex` field (conventionally first) — not embedded; embedding leaks `Lock`/`Unlock` into the struct's public API.

```go
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { return doWork(ctx) })
if err := g.Wait(); err != nil { ... }
```

### HTTP Services

Always set timeouts: `http.Server` gets `ReadTimeout`/`WriteTimeout`/`IdleTimeout`; `http.Client` gets `Timeout`. Use `net/http/httptest` for handler tests.

### Testing

- Table-driven tests with subtests (`t.Run`); `t.Helper()` in helpers.
- No real network or filesystem in unit tests — use interfaces and fakes.
- Integration tests behind `//go:build integration`.

### Tiger Style

- Non-trivial functions assert their preconditions; validate length/nil on user-supplied slices and maps.
- No `panic` in library code (acceptable in `main` for configuration errors only).
- Refactor functions over 70 lines without being asked.
