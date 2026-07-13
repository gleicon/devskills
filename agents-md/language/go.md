## Language Profile — Go

Target: Go 1.24+. Backend services, CLIs, APIs, systems tooling.

Apply these conventions to all Go code in this session.

### Toolchain

Test with `go test -race ./...`. Lint with `golangci-lint run`. Benchmark with `go test -bench=. -benchmem`.

### Project Layout

`cmd/<name>/` for entrypoints, `internal/` for private packages, `pkg/` only for a real library, `api/` for protobuf/OpenAPI. Name packages by what they provide — avoid `util/`, `common/`, `helpers/`.

### Standard Library & Idioms

- Reach for `slices`, `maps`, and `cmp` over hand-rolled helpers; use the `min`/`max`/`clear` builtins. Range over integers with `for i := range n`.
- Expose sequences as `iter.Seq`/`iter.Seq2` (range-over-func) and prefer `slices.Sorted`/`slices.Collect`/`maps.Keys` over manual accumulation. *(Go 1.23+)*
- `json:",omitzero"` over `omitempty` when the intent is "omit the zero value" — it handles zero `time.Time` and zero structs correctly.

### Error Handling

- Wrap with context: `fmt.Errorf("operation: %w", err)`. Never discard an error with `_`.
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

Always set timeouts: `http.Server` gets `ReadTimeout`/`WriteTimeout`/`IdleTimeout`; `http.Client` gets `Timeout`. Use `net/http/httptest` for handler tests. When joining request-controlled input into filesystem paths, open a root with `os.OpenRoot` and use the `*os.Root` methods — they refuse traversal and symlink escapes, unlike manual `filepath.Clean`/prefix checks.

### Testing

- Table-driven tests with subtests (`t.Run`); `t.Helper()` in helpers.
- No real network or filesystem in unit tests — use interfaces and fakes.
- Integration tests behind `//go:build integration`.
- Benchmarks use `for b.Loop() { ... }`, not `for i := 0; i < b.N; i++` — setup outside the loop runs once and results stay live without `runtime.KeepAlive`.
- Test time-dependent or concurrent code with `testing/synctest` and its fake clock instead of real `time.Sleep`. *(Go 1.25+; `synctest.Test`)*

### Tiger Style

- Non-trivial functions assert their preconditions; validate length/nil on user-supplied slices and maps.
- No `panic` in library code (acceptable in `main` for configuration errors only).
