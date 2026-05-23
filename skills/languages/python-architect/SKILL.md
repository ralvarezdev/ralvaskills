---
name: python-architect
version: 1.0.0
description: Python 3.14 enterprise standards — modern typing (PEP 649), immutable dataclasses, Protocol-based DI, asyncio discipline, pytest 9, psycopg + .sql files via importlib.resources. Use when writing, reviewing, or scaffolding Python code.
---

# Python Architecture Standards

Targets **Python 3.14**. See [STACK.md](STACK.md) for pinned dependency versions.

## 1. Typing & Domain Safety

- **Modern syntax:** Built-in generics (`list[str]`, `dict[K, V]`, `X | None`). Never the legacy `typing.List` / `typing.Optional`.
- **Deferred annotations (PEP 649, 3.14):** Annotations are no longer eagerly evaluated — forward references no longer need quotes (`def f(arg: NotYetDefined)` works). Inspect via `annotationlib.get_annotations()`, not `__annotations__` directly.
- **Domain types:** `typing.NewType` to separate distinct concepts (`UserId` vs `OrderId`).
- **Constraints:** `Literal` for string flags; `Enum` (with `__str__`) for states.
- **Structured payloads:** `TypedDict` over `dict[str, Any]` for known-shape mappings.
- **Subclassing safety:** `typing.override` decorator (3.12+) on every overriding method — mypy flags broken overrides.

## 2. Data Structures & Memory

- **Immutability:** Default to `@dataclass(slots=True, frozen=True)` for DTOs and value objects.
- **Pydantic vs dataclass boundary:** Pydantic only at application boundaries (API request/response, DB row parsing, config). Standard dataclasses for core domain logic — keeps the domain free of validation-framework coupling.
- **Memory:** `__slots__` (explicit or via `dataclass(slots=True)`) on high-volume instances.

## 3. Interfaces & DI

- **Protocols:** `typing.Protocol` (structural typing) over deep `abc.ABC` inheritance. Define protocols where consumed.
- **DI:** Pass dependencies into `__init__`. Never instantiate external clients inside a class.
- **State:** No globals. `contextvars` only when request-scoped state is unavoidable.

## 4. Concurrency & Resources

- **Asyncio discipline:** Never block the event loop. Offload sync I/O or CPU work via `asyncio.to_thread()`.
- **Task groups:** `asyncio.TaskGroup` for concurrent coroutines — handles cancellation and exception aggregation properly. Avoid bare `asyncio.gather`.
- **Multiple interpreters (PEP 734, 3.14):** Use `concurrent.interpreters` for CPU-bound parallelism — true multi-core without `multiprocessing`'s overhead, no GIL contention.
- **Introspection:** Debug live async apps with `python -m asyncio ps <PID>` / `pstree <PID>` (3.14).
- **Free-threaded builds (PEP 703):** Be aware of the no-GIL variant. Design hot paths to avoid shared mutable state regardless of GIL presence.
- **Resources:** Wrap I/O in `with` / `async with`. Use `contextlib` for compositions.

## 5. Packages & Imports

- **Imports:** Three groups separated by blank lines — stdlib, third-party, local. Prefer absolute imports.
- **`__init__.py`:** Minimal. Use `__all__ = [...]` to declare the public API explicitly.
- **Bundled resources:** Use `importlib.resources.files(__package__).joinpath("...").read_text()` for embedded files (SQL, templates). Survives wheel and zipapp packaging — never use `__file__`-relative paths for shipped assets.

## 6. Errors & Testing

- **Exceptions:** A base custom exception per module. Always chain (`raise NewError(...) from err`). Never bare `except:`.
- **Bracketless except (PEP 758, 3.14):** `except TimeoutError, ConnectionRefusedError:` is now valid without parens when no `as` clause.
- **Finally hazards (PEP 765, 3.14):** `return` / `break` / `continue` inside `finally` now emits SyntaxWarning — refactor it out.
- **Iterables:** `map(strict=True)` (3.14) when consuming parallel iterables, matching `zip(strict=True)`.
- **Testing:** `pytest 9` with `conftest.py` fixtures. Never the legacy `unittest` module. `pytest-asyncio` for async tests.

## 7. Documentation

- **Docstrings:** Google style (Args, Returns, Raises).
- **DRY:** Don't repeat type info already in hints.
- **Focus:** Explain *why* (domain rules, edge cases), not *what*.

## 8. Stdlib defaults

Prefer stdlib when it covers the use case.

- `pathlib.Path` for all paths — never `os.path` strings. New in 3.14: `Path.copy()`, `Path.move()`, `Path.copy_into()`, `Path.move_into()` for recursive operations.
- `compression.zstd` (3.14) over `gzip` / `bz2` for new payloads — `gzip` / `bz2` / `lzma` / `zlib` are now re-exported under `compression.*`.
- `importlib.resources` for shipped files (see §5).
- `contextlib` for resource lifecycle composition.
- `dataclasses` for data containers (see §2).

## 9. Database access — SQL files + `importlib.resources`

**Recommended pattern, not mandatory.** Mirrors the Go `sqlx + //go:embed` philosophy: raw SQL in `.sql` files, loaded once at module import, executed via `psycopg 3`. No ORM by default — keeps queries auditable in git and gives editors full SQL syntax highlighting and linting.

```python
from importlib.resources import files
import psycopg
from psycopg.rows import class_row

GET_USER_BY_ID = files(__package__).joinpath("queries/get_user_by_id.sql").read_text()

class UserRepo:
    def __init__(self, conn: psycopg.AsyncConnection) -> None:
        self._conn = conn

    async def get_by_id(self, user_id: int) -> User | None:
        async with self._conn.cursor(row_factory=class_row(User)) as cur:
            await cur.execute(GET_USER_BY_ID, (user_id,))
            return await cur.fetchone()
```

Layout:

```
src/myapp/userrepo/
├── __init__.py
├── repo.py
└── queries/
    ├── get_user_by_id.sql
    ├── insert_user.sql
    └── list_users.sql
```

- **Driver:** `psycopg 3` — sync + async, server-side cursors, COPY, prepared statements.
- **Migrations:** `alembic` — versioned, works with raw SQL (no SQLAlchemy ORM required).
- **Dynamic queries:** Compose `.sql` fragments in Python; never concatenate user input — bind parameters.
- **When an ORM is genuinely needed:** SQLAlchemy 2.x (Core or ORM). Record the decision in an ADR.

## 10. Tooling

- **Environment + packaging:** `uv` — replaces `pip`, `pip-tools`, `virtualenv`, `pyenv`. Single binary, fast.
- **Lint + format:** `ruff` — replaces `black`, `isort`, `flake8`, `pyupgrade`. One config, one tool. Drop-in template: [`assets/ruff.toml`](assets/ruff.toml) — copy to your project root as `ruff.toml` (or fold into `pyproject.toml` under `[tool.ruff]`) and set `known-first-party` to your package name.
- **Type checking:** `mypy --strict` as the baseline. No `Any`-by-default escape hatches.
- **Test:** `pytest 9` + `pytest-asyncio` for async paths.

## Canonical libraries

See [STACK.md](STACK.md) for the full pinned list — pydantic, pydantic-settings, fastapi, uvicorn, httpx, pytest, pytest-asyncio, mypy, ruff, uv, typer, psycopg, alembic.
