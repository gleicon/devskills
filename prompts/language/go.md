## Language Profile — Go

Target: Go 1.22+. Backend services, CLIs, APIs, systems tooling.

Apply these conventions to all Go code in this session.

### Toolchain

- Build: `go build ./...`
- Test: `go test -race ./...`
- Lint: `golangci-lint run`
- Format: `gofmt -s` or `goimports`
- Vet: `go vet ./...`
- Benchmark: `go test -bench=. -benchmem ./...`

### Project Layout

Follow standard Go project structure:
```
cmd/<name>/main.go      # entrypoints
internal/               # private packages
pkg/                    # exported packages (only if library)
api/                    # protobuf or OpenAPI definitions
```

Avoid: `util/`, `common/`, `helpers/` — name packages by what they provide, not what they are.

### Error Handling

```go
// Always wrap with context
if err != nil {
    return fmt.Errorf("operation name: %w", err)
}

// Sentinel errors for callers to check
var ErrNotFound = errors.New("not found")

// Check with errors.Is, not ==
if errors.Is(err, ErrNotFound) { ... }
```

Never discard errors with `_`. Every error is handled or propagated.

### Concurrency

- Goroutines always have an explicit exit condition.
- Use `context.Context` for cancellation — first parameter always.
- `sync.WaitGroup` for fan-out, channels for coordination.
- Protect shared state with `sync.Mutex`, embedded as first field in owning struct.
- Never share mutable state between goroutines without synchronization.
- Use `errgroup` for parallel work with error collection.

```go
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { return doWork(ctx) })
if err := g.Wait(); err != nil { ... }
```

### HTTP Services

```go
// Always set timeouts
client := &http.Client{
    Timeout: 30 * time.Second,
}

// Server
srv := &http.Server{
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}
```

Use `net/http/httptest` for handler tests.

### Testing

- Table-driven tests using `[]struct{ name, input, expected }`.
- `t.Helper()` in helper functions.
- Subtests with `t.Run(tc.name, ...)` for parallel execution.
- No real network or filesystem in unit tests — use interfaces and fakes.
- Integration tests in `_test.go` files with build tag `//go:build integration`.

### Naming

- Exported: `PascalCase`. Unexported: `camelCase`.
- Error types: `ErrSomething` (sentinel) or `SomethingError` (struct).
- Interfaces: noun or noun phrase. One-method interfaces: method name + `-er` (`Reader`, `Closer`).
- Receiver name: short, consistent, never `self` or `this`.
- Context variable always named `ctx`.

### Tiger Style Integration

- Every exported function has at minimum one guard clause asserting pre-conditions.
- All slice/map operations on user-supplied data have length/nil checks.
- No `panic` in library code. `panic` acceptable in `main` for configuration errors only.
- Functions over 70 lines are refactored without being asked.
