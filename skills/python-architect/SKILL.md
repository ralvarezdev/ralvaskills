---
name: python-architect
description: Advanced universal Python standards for strict typing, concurrency, testing, and enterprise architecture.
---
# Python Architecture Standards

## 1. Advanced Typing & Domain Safety
- Type Hints: Mandatory. Use modern syntax (list[str], dict[str, int], X | None). Do not use the legacy typing module for built-ins.
- Domain Types: Use typing.NewType to prevent mixing primitive strings/ints that represent different domain concepts (e.g., UserId = NewType('UserId', int) vs OrderId = NewType('OrderId', int)).
- Literals: Use typing.Literal for string-based flags instead of raw strings to enforce compile-time checks.
- Enums: Use enum.Enum or enum.IntEnum. Define __str__ methods for clean serialization.

## 2. Data Structures & Memory Optimization
- Immutability: Default to @dataclass(slots=True, frozen=True) for data transfer objects (DTOs) to make them hashable, thread-safe, and explicitly clear they should not mutate state.
- Pydantic Boundaries: Use Pydantic BaseModel strictly at the application boundaries (API endpoints, DB parsing) for validation. Do not leak Pydantic models deep into the core domain logic where simple dataclasses suffice.
- Memory Allocation: For high-volume standard classes (non-dataclass), explicitly define __slots__ = (...) to dict__dict__ creation and significantly reduce RAM usage.

## 3. Interfaces & Dependency Injection
- Protocols over Inheritance: Define interfaces using typing.Protocol (structural subtyping/duck typing) where the interface is *consumed*, not where it is implemented. Avoid deep, Java-style abc.ABC inheritance trees.
- Constructors: Pass all dependencies (DB clients, loggers, external APIinit__init__. Never instantiate I/O clients inside the class itself.
- Global State Ban: Ban the use of global variables. Use dependency injection. If request-scoped state is strictly required, use contextvars.

## 4. Concurrency & Resource Management
- AsyncIO Boundaries: Never block the event loop. Offload CPU-bound tasks or synchronous I/O (like requests.get() or time.sleep()) to a ThreadPoolExecutor using asyncio.to_thread().
- Task Management (Python 3.11+): Use asyncio.TaskGroup() for managing multiple concurrent coroutines. This ensures proper cancellation and error propagation if one task in the group fails.
- Context Managers: Wrap all file, socket, and DB connections in with blocks. Use @contextlib.contextmanager or contextlib.AsyncExitStack for custom setup/teardown logic.

## 5. Package Structure & Imports
- Import Order: Group imports strictly: 1. Standard Library, 2. Third-Party Packages, 3. Local Application Imports. Separate each group with a blank line.
- Absolute Imports: Use absolute imports for everything outside the current package. Avoid relative imports (from .. import x) beyond one directory level deep.
- Init Fileinit__init__.py files minimal. Only use them to expose the public API of a package by explicitly dall __all__ = [...].

## 6. Error Handling & Testing Standards
- Custom Exceptions: Define a base Exception class for the module, and inherit from it for specific domain errors (e.g., class DatabaseConnectionError(AppBaseError)). Never use a bare except:.
- Error Chaining: When raising custom errors from built-in errors, preserve the traceback by using raise NewError from original_error.
- Pytest: Default to pytest conventions. Use conftest.py for shared fixtures. Do not use the legacy unittest module.

## 7. Documentation & Docstrings
- Google Style Docstrings: Mandatory for all public modules, classes, and functions. Must include Args:, Returns:, and Raises: sections where applicable.
- No Type Redundancy: Do not duplicate type definitions in the docstring if they are already defined in the Python 3.10+ type hints. Use the docstring to explain the *purpose* and *constraints* of the arguments.
- Why, Not What: In-line # comments must explain the domain logic, workaround, or mathematical reasoning, not the mechanical operation of the code.