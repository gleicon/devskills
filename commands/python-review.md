Review Python code with Tiger Style constraints and Python idioms.

Applies to: Python 3.11+. Backend services, APIs, CLIs, data pipelines.

## Arguments

Scan the invocation for the `--no-tiger` flag. Treat every other argument as review scope (files or directories); if no scope is given, review the changed files on the current branch.

- `--no-tiger` present → skip the Tiger Style section; run Python Idioms, Typing, Performance, Security, and Testing only.
- `--no-tiger` absent → run all sections (default).

Example: `/python-review --no-tiger pkg1/ pkg2/` reviews `pkg1/` and `pkg2/` without Tiger Style.

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

### Typing
- [ ] Every public signature annotated; passes `mypy --strict` (no implicit `Any`)
- [ ] Modern syntax: `list[str]`, `X | None` over `Optional[X]`
- [ ] `@dataclass`/`Protocol`/`Enum` used instead of loose dicts and magic strings where they fit
- [ ] No `# type: ignore` without a trailing reason comment

### Performance
- [ ] No blocking calls (`time.sleep`, `requests`, sync DB drivers) inside `async def` — use async clients or `asyncio.to_thread`
- [ ] Every external `await` / network / DB call has a timeout
- [ ] CPU-bound work uses `ProcessPoolExecutor`, not threads (the GIL serializes threads)
- [ ] Database queries not issued inside loops; N+1 patterns absent
- [ ] Generators/streaming for large data instead of building full lists in memory

### Security
- [ ] No string-built SQL (f-string/`%`/`+`) — use parameterized queries / the ORM
- [ ] No `subprocess` with `shell=True` on user input; no `eval`/`exec`/`pickle` on untrusted data
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
