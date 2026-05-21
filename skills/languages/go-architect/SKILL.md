---
name: go-architect
description: Defines strict architectural patterns for Go structures, enums, constants, and domain modeling.
---

# Go Architecture & Domain Modeling

## 1. Constants & Enums
* **Enums:** Use custom types with `iota` and implement `String()`. Avoid bare primitives. Use `1 << iota` for bitmasks.
* **Errors:** Export package-level sentinel errors (`var ErrNotFound = ...`) to allow `errors.Is()`.
* **Grouping:** Group related constants together in `const` blocks.

## 2. Structures & Memory
* **Design:** Use explicit struct tags, standard audit fields, and composition (embedding).
* **Optimization:** Order fields from largest to smallest type to minimize memory padding.
* **Semantics:** Use pointers for mutation or Mutexes; use values for small, immutable payloads.
* **Validation:** Implement `Validate() error` for structs receiving external data.

## 3. Instantiation
* **Constructors:** Always provide constructors (e.g., `New...`) to safely initialize maps, slices, and channels.
* **Configuration:** Use the Functional Options pattern for complex setups instead of large config structs.

## 4. Interfaces
* **Design:** Define interfaces where they are *used*, not where implemented. Keep them small (1-3 methods).
* **Signatures:** Accept interfaces (for easy mocking), return concrete structs.

## 5. Concurrency
* **State:** Embed `sync.RWMutex` directly above the fields it protects. Use `sync/atomic` for simple counters.
* **Context:** Pass `context.Context` as the first parameter for any blocking or I/O operations.

## 6. Packages
* **Naming:** Short, lowercase, single-word names. Never use `util`, `common`, or `helpers`.
* **Layout:** Organize code by domain/feature, not by technical layer.

## 7. Testing
* **Patterns:** Use table-driven tests and black-box testing (`package mypkg_test`).
* **Helpers:** Call `t.Helper()` as the very first line in custom test assertions.

## 8. Errors & Panics
* **Checking:** Use `errors.As()` for custom error types and `errors.Is()` for sentinels.
* **Panics:** Restrict `panic` exclusively to application initialization (e.g., `MustCompile`).

## 9. Dependencies & Logging
* **Injection:** Pass dependencies explicitly via constructors. Avoid global state and `init()` setups.
* **Logging:** Enforce structured logging (e.g., `slog`, `zap`). Do not use `fmt.Printf`.

## 10. Generics
* **Usage:** Limit to custom data structures and utility functions. Do not use if an interface suffices.

## 11. Documentation
* **Format:** Exported names need full-sentence comments starting with the object's name. Include package-level comments.
* **Style:** Weave parameter/return names naturally into prose (no `@param`). Document specific error conditions.
* **Focus:** Explain *why* the code exists (business rules, edge cases), not *what* it does.