### Python

- Type-annotated; passes `mypy --strict`. Modern syntax (`list[str]`, `X | None`); PEP 695 generics (`class C[T]`, `type` alias), `@override`, `TypeIs` (3.13+).
- Catch specific exceptions, never bare `except:`. Chain with `raise ... from err`. No control flow in `finally`.
- No mutable default arguments. Resources via `with`. `datetime.now(UTC)` over `utcnow()`. CPU-bound → `ProcessPoolExecutor` (GIL serializes threads unless the separate `python3.14t` build, 3.14+).
- `pytest` with plain `assert`; `ruff` (`UP` rules, `py313`) for lint/format, `uv` for deps.
