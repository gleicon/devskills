Review Python code with Tiger Style constraints and Python idioms.

Applies to: Python 3.13+. Backend services, APIs, CLIs, data pipelines.

## Arguments

Scan the invocation for the `--no-tiger` and `--fix` flags. Treat every other argument as review scope (files or directories); if no scope is given, review the changed files on the current branch.

- `--no-tiger` present → skip the Tiger Style section; run Python Idioms, Typing, Performance, Security, and Testing only.
- `--no-tiger` absent → run all sections (default).
- `--fix` → after reporting, apply only the violations whose fix is **mechanical and unambiguous** (a rename to the idiom, a missing error check the review is certain about). Anything that changes logic or rests on an unverified assumption — especially security and correctness findings — **stays report-only**. After applying, re-run any build/test/lint check already in the loop and revert any fix that breaks it — or that touched more than the intended mechanical edit. End with a summary of what was applied and what was left.

Example: `/ds-python-review --no-tiger pkg1/ pkg2/` reviews `pkg1/` and `pkg2/` without Tiger Style.

## Review Checklist

Use the checklist as a lens, not a scorecard: reason about the actual change, report real violations anchored to `file:line`, and flag issues even when they aren't listed. Don't manufacture findings to fill a category. Report only violations — no praise, no summary.

### Tiger Style

Skip this section entirely if `--no-tiger` was passed. Otherwise it is mandatory.
- [ ] Non-trivial functions assert their preconditions and key invariants — don't demand assertions in thin wrappers or trivial accessors
- [ ] All loops over external input have explicit bounds; no unbounded iteration
- [ ] No recursion without provable termination
- [ ] All exceptions handled or propagated deliberately — no bare `except:` and no silently swallowed errors
- [ ] No post-initialization allocation in paths actually identified as hot — don't flag ordinary allocation
- [ ] Functions under 70 lines
- [ ] Variable names include units/qualifiers where applicable

### Python Idioms
- [ ] No mutable default arguments (`def f(x=[])`) — use `None` and create inside
- [ ] Resources acquired with context managers (`with`), not manual `try/finally` or unclosed handles
- [ ] Catch specific exceptions, never bare `except:`; chain with `raise ... from err`
- [ ] No signaling failure via `None`/sentinel where an exception or typed result is clearer
- [ ] Iterate with comprehensions / generators over manual index loops; lazy generators for large streams
- [ ] No `from module import *`; no logic in `__init__.py`; entrypoints guarded by `if __name__ == "__main__"`
- [ ] `pathlib` over `os.path`; f-strings over `%`/`.format()`
- [ ] No `return`/`break`/`continue` inside a `finally` block — it silently swallows exceptions (SyntaxWarning on 3.14+, PEP 765)
- [ ] Timezone-aware `datetime.now(UTC)` over the deprecated naive `datetime.utcnow()` / `utcfromtimestamp()`
- [ ] No imports of stdlib modules removed in 3.13 (PEP 594) — `crypt`→`bcrypt`/`argon2-cffi`, `pipes`→`subprocess`+`shlex.quote`, `cgi`/`cgitb`→`urllib.parse`/`email`; also `telnetlib`/`nntplib`/`imghdr`/`uu`/`lib2to3`

### Typing
- [ ] Every public signature annotated; passes `mypy --strict` (no implicit `Any`)
- [ ] Modern syntax: `list[str]`, `X | None` over `Optional[X]`
- [ ] PEP 695 type parameters (`class C[T]`, `def f[T]`, `type` alias statement) over explicit `TypeVar`/`TypeAlias` on new generic code
- [ ] `@override` on methods overriding a base; `typing.TypeIs` over `TypeGuard`; `ReadOnly` for immutable `TypedDict` items (3.13+)
- [ ] Forward refs unquoted and `from __future__ import annotations` dropped where 3.14 deferred evaluation makes them redundant — flag only on 3.14+, keep on 3.13
- [ ] `@dataclass`/`Protocol`/`Enum` used instead of loose dicts and magic strings where they fit
- [ ] No `# type: ignore` without a trailing reason comment

### Performance
- [ ] No blocking calls (`time.sleep`, `requests`, sync DB drivers) inside `async def` — use async clients or `asyncio.to_thread`
- [ ] Every external `await` / network / DB call has a timeout
- [ ] Concurrent tasks managed with `asyncio.TaskGroup` (scoped lifetime, sibling cancellation, `ExceptionGroup`) over bare `asyncio.gather`
- [ ] CPU-bound work uses `ProcessPoolExecutor` (or `InterpreterPoolExecutor`, 3.14+) — threads only parallelize CPU on the separate free-threaded build (`python3.14t`); the stock interpreter's GIL serializes them
- [ ] Database queries not issued inside loops; N+1 patterns absent
- [ ] Generators/streaming for large data instead of building full lists in memory

### Security
- [ ] No string-built SQL (f-string/`%`/`+`) — use parameterized queries / the ORM. (Emerging: a `Template`-aware library can use t-strings, PEP 750, 3.14+, to offer a safe-interpolation API — but a `t'...'` literal sanitizes nothing on its own)
- [ ] No `subprocess` with `shell=True` on user input; no `eval`/`exec`/`pickle` on untrusted data
- [ ] `tarfile` extraction passes `filter='data'` (or stricter) — bare `extractall` is a path-traversal/overwrite hazard and errors by default on 3.14+
- [ ] `yaml.safe_load`, not `yaml.load`; no untrusted deserialization
- [ ] No hardcoded credentials or secrets; read from env/secret store
- [ ] All external input validated and bounded at the boundary (e.g. `pydantic`); requests set timeouts

### Testing
- [ ] Public surface and error paths have meaningful coverage — flag notable gaps, not every untested accessor
- [ ] Error paths tested with `pytest.raises`, not just the happy path
- [ ] Variants parametrized (`@pytest.mark.parametrize`) rather than copy-pasted
- [ ] No real network/filesystem in unit tests — fakes and `tmp_path`; no real `sleep` or wall-clock dependence

## Output Format

```
<file>:<line>: <severity>: <problem>. <fix>.
```

Severity levels: `critical` (correctness/security), `major` (reliability/performance), `minor` (idiom/style).

Skip formatting nits unless they affect correctness or readability significantly.
