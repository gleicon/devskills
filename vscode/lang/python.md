### Python

- Type-annotated; passes `mypy --strict`. Modern syntax (`list[str]`, `X | None`).
- Catch specific exceptions, never bare `except:`. Chain with `raise ... from err`.
- No mutable default arguments. Resources via `with`.
- `pytest` with plain `assert`; `ruff` for lint/format, `uv` for deps.
