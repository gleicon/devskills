### Go

- Wrap errors: `fmt.Errorf("context: %w", err)`. Context is first parameter.
- Every goroutine has explicit exit condition.
- Table-driven tests; benchmarks use `for b.Loop()`; no `time.Sleep` (use `testing/synctest`).
- Prefer `slices`/`maps`/`cmp` and `min`/`max`/`clear`; `os.OpenRoot` for user-controlled paths.
