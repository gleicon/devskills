## Language Profile — Python

Target: Python 3.11+. Backend services, APIs, CLIs, data pipelines, automation.

Apply these conventions to all Python code in this session.

### Toolchain

Manage the project with `uv` (or `pip` + a virtualenv if the repo already uses it) — never install into the system interpreter. Lint and format with `ruff` (`ruff check` + `ruff format`). Type-check with `mypy --strict` (or `pyright`). Test with `pytest`. Pin dependencies in `pyproject.toml`; commit the lockfile.

### Project Layout

`src/<package>/` layout — importable code under `src/`, not the repo root. `tests/` mirrors the package tree. Entrypoints via `[project.scripts]` in `pyproject.toml`, not loose top-level scripts. One responsibility per module; avoid `utils.py`/`helpers.py` grab-bags.

### Typing

- Annotate every function signature and module-level constant. Code is written to pass `mypy --strict` — no implicit `Any`.
- Modern syntax: `list[str]`, `dict[str, int]`, `X | None` over `Optional[X]`. `from __future__ import annotations` where it helps.
- `@dataclass(frozen=True, slots=True)` for value objects; `Protocol` for structural interfaces at the point of use; `Enum` over magic strings.
- No `# type: ignore` without a trailing reason comment.

### Error Handling

- Catch specific exceptions, never bare `except:` or `except Exception` without re-raising. Preserve the chain with `raise NewError(...) from err`.
- Define a package-level exception hierarchy (`class AppError(Exception)`); don't signal failure with sentinel return values or `None` where an exception is clearer.
- `try` blocks wrap only the line that can fail. Release resources with context managers (`with`), never manual `try/finally` for files, locks, or connections.

### Concurrency

- `asyncio` for I/O-bound concurrency; no blocking calls (`time.sleep`, `requests`, sync DB drivers) inside `async def` — use the async client or `asyncio.to_thread`. Every `await` on external I/O has a timeout.
- CPU-bound work goes to `ProcessPoolExecutor`, not threads (the GIL serializes them).
- Never mutate shared state across tasks without an `asyncio.Lock`/`threading.Lock`.

### Testing

- `pytest` with plain `assert`. Parametrize variants with `@pytest.mark.parametrize`; share setup through fixtures, not class hierarchies.
- No real network or filesystem in unit tests — use fakes and `tmp_path`. `pytest.raises` for error paths; `freezegun`/injected clocks instead of real time.
- `pytest-asyncio` for async tests. Test behavior through the public API, not private functions.

### Tiger Style

- Non-trivial functions validate their preconditions (`assert` for invariants, raised exceptions for caller errors and bad input). Don't assert in thin wrappers or trivial accessors.
- All loops over external input have explicit bounds; no unbounded recursion without provable termination.
- No mutable default arguments (`def f(x=[])`) — use `None` and create inside.
- Keep functions under 70 lines; refactor past that without being asked.
