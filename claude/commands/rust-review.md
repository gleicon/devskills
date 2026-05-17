Review Rust code for safety metrics, error handling, and idiomatic Rust.

Applies to: Rust stable. Systems programming, CLI tools, services.

## Automated Checks (run first if tools are available)

```bash
# Safety surface metrics
cargo geiger 2>/dev/null          # count unsafe blocks across crate graph
cargo geiger --output-format json  # machine-readable

# Lints — must pass clean
cargo clippy -- -D warnings

# Security audit
cargo audit

# Unused dependencies
cargo +nightly udeps --all-targets 2>/dev/null

# Warnings in clean build
cargo clean && cargo build 2>&1 | grep -E "^warning"

# Count unsafe/panic/unwrap manually
grep -rn "unsafe " src/
grep -rn "\.unwrap()" src/
grep -rn "\.expect(" src/
grep -rn "panic!(" src/
grep -rn "todo!()" src/
grep -rn "unimplemented!()" src/
```

Run these commands, report counts, then do manual review below.

## Safety Metrics

Report these numbers at the top of the review:

```
unsafe blocks:      <N>  (cargo geiger or manual grep)
unwrap() calls:     <N>
expect() calls:     <N>
panic!() calls:     <N>
todo!()/unimpl!():  <N>
clippy warnings:    <N>
audit findings:     <N>
```

Each non-zero number requires justification in the review findings.

## Review Checklist

### Unsafe Code
- [ ] Every `unsafe` block has a `// SAFETY:` comment explaining the invariant relied upon
- [ ] Unsafe is confined to the smallest possible scope
- [ ] Each unsafe block has a corresponding test exercising the unsafe path
- [ ] No raw pointer arithmetic without bounds assertions

### Panic and Unwrap
- [ ] No `.unwrap()` in library code (crates without `main.rs`)
- [ ] Each `.expect("msg")` has a descriptive message that identifies WHAT failed and WHY it should not
- [ ] No `panic!()` on recoverable errors (bad input, missing file, network failure)
- [ ] `panic!()` acceptable only for: violated invariants, programmer errors, OOM in systems context

### Error Handling
- [ ] Library errors use `thiserror` with typed enum variants
- [ ] Binary errors use `anyhow` with `.context("operation description")`
- [ ] `?` operator used consistently — no manual `match Err(e) => return Err(e.into())`
- [ ] All `Result`-returning functions have their error handled at call sites — no implicit `.ok()`  discard on meaningful operations

### Ownership and Borrowing
- [ ] No unnecessary `.clone()` — check if a reference suffices
- [ ] No `Arc<Mutex<T>>` where single-threaded `RefCell<T>` or owned data works
- [ ] `Mutex` lock scope is minimal — no lock held across await points
- [ ] No `Box<dyn Error>` in library public API — use typed errors

### Async (if applicable)
- [ ] No blocking calls inside `async fn` (no `std::thread::sleep`, `std::fs::read`)
- [ ] CPU-bound work in `tokio::task::spawn_blocking`
- [ ] Futures are `Send` when used with multi-thread Tokio runtime
- [ ] `select!` arms handle cancellation correctly

### Idiomatic Rust
- [ ] Iterator chains preferred over manual loops where clarity is not lost
- [ ] `if let` / `while let` over `.unwrap()` on Options
- [ ] `matches!()` macro for pattern matching in boolean context
- [ ] Struct fields ordered: frequently accessed first (cache line consideration for hot structs)
- [ ] `#[derive]` used where applicable (Debug, Clone, PartialEq)
- [ ] No custom `Display` that duplicates `Debug` — both should exist independently

### Performance
- [ ] No allocation inside loops where pre-allocation is possible
- [ ] `String::with_capacity(n)` / `Vec::with_capacity(n)` where size is known
- [ ] No repeated `HashMap::get` + `HashMap::insert` — use `entry()` API
- [ ] Large values passed by reference, not moved through multiple stack frames unnecessarily

### Testing
- [ ] Each `unsafe` block has a dedicated test
- [ ] Each error variant has a test triggering it
- [ ] `#[should_panic]` tests for intentional panics
- [ ] Fuzzing targets for parsing/deserialization code (`cargo fuzz` or `bolero`)

### Dependencies (Cargo.toml)
- [ ] No duplicate functionality across deps (two HTTP clients, two JSON libs)
- [ ] `features` flags used to minimize compiled surface area
- [ ] Dev-only deps in `[dev-dependencies]`, not `[dependencies]`
- [ ] `cargo audit` clean (no known CVEs)

## Output Format

Header (always emit):
```
unsafe: <N> | unwrap: <N> | expect: <N> | panic: <N> | clippy: <N>
```

Findings:
```
<file>:<line>: <severity>: <problem>. <fix>.
```

Severity: `critical` (soundness/CVE), `major` (reliability/correctness), `minor` (idiom/style).

Skip findings with no actionable fix. Do not report clippy items already caught by `cargo clippy -- -D warnings`.
