## Language Profile — Rust

Target: Rust stable. Systems programming, performance-critical services, experimental large projects.

Apply these conventions to all Rust code in this session.

### Toolchain

Lint with `cargo clippy -- -D warnings`. Audit with `cargo audit`. Benchmark with Criterion. Use `cargo check` for fast type-checking.

### Cargo.toml Conventions

```toml
[profile.release]
lto = "thin"
codegen-units = 1
panic = "abort"   # binaries only; libraries keep "unwind"
```

Minimize dependencies — each crate is a liability. Prefer `std` when the implementation is straightforward.

### Error Handling

- Libraries: typed errors with `thiserror`. Binaries: `anyhow` with `.context(...)`.
- Propagate with `?`. Never `.unwrap()` in library or production paths (acceptable in tests and `main` with a documented reason).

```rust
#[derive(Debug, thiserror::Error)]
pub enum AppError {
    #[error("not found: {0}")]
    NotFound(String),
    #[error("io: {0}")]
    Io(#[from] std::io::Error),
}
```

### Memory and Ownership

- References over cloning; clone only when ownership is genuinely needed.
- `Arc<T>` for cross-thread sharing, `Rc<T>` single-threaded. `Mutex<T>` wraps data, not functions; lock scopes minimal, never held across an await point.
- Avoid `unsafe` outside FFI / unavoidable perf code. Document every `unsafe` block with the invariant it relies on.

### Async (when applicable)

Tokio. `async fn` for I/O; CPU-bound work in `tokio::task::spawn_blocking`. No blocking calls (`std::thread::sleep`, `std::fs`) in async context. Document whether each async fn is cancellation-safe.

### Performance

- Allocate upfront (`Vec::with_capacity`) when size is known. Zero-copy parsing with `&str`/`&[u8]`.
- Profile (`flamegraph`, `perf`) before optimizing. `#[inline]` only after the call site shows hot.

### Testing

Unit tests in the same file (`#[cfg(test)] mod tests`); integration tests in `tests/`. `#[tokio::test]` for async. Each error variant and each `unsafe` path gets a test.

### Tiger Style

- Assert invariants with `assert!` / `debug_assert!` (the latter for checks too expensive for release). Document public functions' preconditions.
- Panic only on programmer errors (violated invariants), never on bad input.
