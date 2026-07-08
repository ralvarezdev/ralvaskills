# Stack Versions

| Dependency | Pinned version | Purpose |
|---|---|---|
| python | 3.14 | Language |
| pydantic | 2.13 | Boundary validation (API in/out, config) |
| pydantic-settings | 2.14 | Env + file layered config via Pydantic |
| fastapi | 0.136 | Async web framework |
| uvicorn | 0.47 | ASGI server |
| httpx | 0.28 | Async HTTP client |
| pytest | 9.0 | Test runner — replaces `unittest` entirely |
| pytest-asyncio | 1.3 | Async test support |
| mypy | 2.1 | Static type checking — `--strict` baseline |
| ruff | 0.15 | Lint + format — replaces black / isort / flake8 / pyupgrade |
| uv | 0.11 | Environment + packaging — replaces pip / virtualenv / pyenv |
| typer | 0.25 | CLI framework |
| psycopg | 3.3 | Postgres driver — sync + async, used with `importlib.resources` for `.sql` files |
| alembic | 1.18 | Database migrations — works with raw SQL (no SQLAlchemy ORM required) |

## Notes

- **No ORM by default.** Recommended data-access pattern is `psycopg 3` + raw SQL in `.sql` files loaded via `importlib.resources`. SQLAlchemy 2.x (Core or ORM) is acceptable when multi-DB abstraction or complex query composition justifies it — record the choice in an ADR.
- **No `unittest`.** Standardize on `pytest 9`. `unittest`-style assertions are an anti-pattern in new code.
- **Free-threaded builds (PEP 703).** Design hot paths to avoid shared mutable state regardless of GIL presence — code should run correctly on both standard and free-threaded interpreters.
- **No task queue pinned.** Neither `arq` nor `celery` is canonical right now — add when a real use case appears.

_Last reviewed: 2026-07-08_
_Skill version at last review: 1.3.0_
