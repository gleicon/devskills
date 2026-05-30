## Language Profile — Rust

Target: Rust stable. Systems programming, performance-critical services, experimental large projects.

Apply these conventions to all Rust code in this session.

### Toolchain

- Build: `cargo build`
- Test: `cargo test`
- Lint: `cargo clippy -- -D warnings`
- Format: `cargo fmt`
- Audit: `cargo audit`
- Benchmark: `cargo bench` with Criterion
- Check: `cargo check` (faster than build for type checking)

### Cargo.toml Conventions

```toml
[profile.release]
lto = "thin"
codegen-units = 1
panic = "abort"   # for binaries; libraries keep "unwind"

[profile.dev]
debug = true
```

Minimize dependencies. Each crate is a liability. Prefer `std` over a crate when the implementation is straightforward.

### Error Handling

```rust
// Use thiserror for library errors
#[derive(Debug, thiserror::Error)]
pub enum AppError {
    #[error("not found: {0}")]
    NotFound(String),
    #[error("io: {0}")]
    Io(#[from] std::io::Error),
}

// anyhow for application code (binaries)
use anyhow::{Context, Result};

fn load_config(path: &Path) -> Result<Config> {
    let data = fs::read_to_string(path)
        .with_context(|| format!("reading config from {}", path.display()))?;
    Ok(toml::from_str(&data)?)
}
```

Never `.unwrap()` in library code or production paths. Acceptable in tests and `main` with documented reason.

### Memory and Ownership

- Prefer references over cloning. Clone only when ownership is genuinely needed.
- Use `Arc<T>` for shared ownership across threads. `Rc<T>` for single-threaded.
- `Mutex<T>` wraps data, not functions. Lock scopes are minimal.
- Avoid `unsafe` unless writing FFI or performance-critical code that cannot be expressed safely. Document every `unsafe` block with the invariant it relies on.

### Async (when applicable)

- Runtime: Tokio.
- `async fn` for I/O-bound work. CPU-bound work goes in `tokio::task::spawn_blocking`.
- Cancellation safety: document whether async functions are cancellation-safe.
- No blocking calls in async context (`std::thread::sleep`, `fs::read` in async fn).

```rust
#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // ...
}
```

### Testing

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_name_describes_scenario() {
        // arrange
        // act
        // assert
    }

    #[tokio::test]
    async fn async_test() { ... }
}
```

- Test modules in the same file as the code they test (unit tests).
- Integration tests in `tests/` directory.
- Use `assert_eq!` with the expected value first.

### Naming

- Types: `PascalCase`
- Functions/methods/variables: `snake_case`
- Constants: `SCREAMING_SNAKE_CASE`
- Lifetimes: short lowercase (`'a`, `'src`, `'buf`)
- Modules: `snake_case`

### Performance

- Allocate upfront where possible. Use `Vec::with_capacity` when size is known.
- Profile with `perf`, `flamegraph`, or `cargo-flamegraph` before optimizing.
- Prefer stack allocation for small, fixed-size data.
- Zero-copy parsing with `&str` and `&[u8]` slices.
- `#[inline]` only after profiling shows the call site is hot.

### Tiger Style Integration

- Assertions via `assert!` and `debug_assert!` for invariants.
- `debug_assert!` for checks that are too expensive for release builds.
- Every public function documents its preconditions in doc comments.
- Panics only on programmer errors (violated invariants), never on bad input.
- Functions over 70 lines are refactored.
