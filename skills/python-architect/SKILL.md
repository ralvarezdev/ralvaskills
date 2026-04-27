---
name: python-architect
description: Defines strict enterprise standards for Python typing, concurrency, testing, and domain architecture.
---

# Python Architecture Standards

## 1. Typing & Domain Safety
* **Modern Syntax:** Use built-in generics (`list[str]`, `X | None`). Do not use the legacy `typing` module for built-ins.
* **Domain Types:** Use `typing.NewType` to separate distinct domain concepts (e.g., `UserId` vs `OrderId`).
* **Constraints:** Use `typing.Literal` for string flags and `Enum` (with `__str__` defined) for states.

## 2. Data Structures & Memory
* **Immutability:** Default to `@dataclass(slots=True, frozen=True)` for DTOs.
* **Pydantic:** Restrict Pydantic strictly to application boundaries (APIs, DB parsing). Use standard dataclasses for core domain logic.
* **Optimization:** Explicitly define `__slots__` on high-volume standard classes to reduce memory footprint.

## 3. Interfaces & DI
* **Protocols:** Prefer `typing.Protocol` (duck typing) over deep `abc.ABC` inheritance trees. Define interfaces where they are consumed.
* **Dependency Injection:** Pass dependencies into `__init__`. Never instantiate external clients inside the class.
* **State:** Ban global variables. Use `contextvars` if request-scoped state is absolutely required.

## 4. Concurrency & Resources
* **AsyncIO:** Never block the event loop. Offload sync I/O or CPU tasks via `asyncio.to_thread()`.
* **Tasks:** Use `asyncio.TaskGroup()` (Python 3.11+) to manage concurrent coroutines safely.
* **Context Managers:** Wrap all I/O connections in `with` blocks or use `contextlib`.

## 5. Packages & Imports
* **Imports:** Group strictly by Standard Library, Third-Party, then Local, separated by blank lines. Prefer absolute imports.
* **`__init__.py`:** Keep minimal. Only use `__all__ = [...]` to explicitly define the public API.

## 6. Errors & Testing
* **Exceptions:** Create a base custom exception for each module. Always chain errors (`raise NewError from original_error`). Never use bare `except:`.
* **Testing:** Standardize on `pytest` and `conftest.py`. Do not use the legacy `unittest` module.

## 7. Documentation
* **Docstrings:** Enforce Google Style (Args, Returns, Raises).
* **DRY Types:** Do not duplicate type definitions in the docstring if they are already in the type hints.
* **Focus:** In-line comments must explain the *why* (domain logic/math), not the *what* (mechanics).