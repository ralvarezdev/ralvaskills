---
name: logic-cleaner
description: Language-agnostic refactoring engine that reduces complexity and enforces strict cleanliness standards.
---
# Universal Logic & Code Cleanliness Standards

## 1. Control Flow
* **Flattening:** Maximum 2 levels of indentation per function. Extract deeper logic into helpers.
* **Guard Clauses:** Handle edge cases and errors at the very top with early returns. Keep the happy path left-aligned.
* **Fail Fast:** Halt execution immediately on invalid states instead of passing nulls or error flags down the stack.

## 2. Conditionals
* **Phrasing:** Enforce positive phrasing; avoid double negatives (e.g., use `isPending`, not `isNotFinished`).
* **Extraction:** Extract conditions with more than two logical operators into well-named boolean variables or functions.
* **Simplification:** Apply De Morgan's Laws to simplify negated compound conditions (e.g., convert `!(A || B)` to `!A && !B`).

## 3. State & Mutation
* **Scope:** Declare variables as close to their first use as possible. Never shadow variables from an outer scope.
* **Immutability:** Treat function parameters as strictly read-only. Never reassign them; assign results to new variables instead.

## 4. Naming
* **Rules:** Avoid abbreviations (`req`, `idx`) unless they are universal language standards (e.g., `ctx` in Go). Use symmetric pairs for related concepts (`start`/`stop`, `open`/`close`).
* **Verbs:** Function names must begin with strong, specific action verbs (e.g., `CalculateTotal`, not `ProcessData` or `Handle`).

## 5. Magic Values & Hardcoding
* **Constants:** Extract raw strings and numbers (except 0 and 1) into named constants that explain *why* the value exists.
* **Config:** Inject all environment-dependent values (timeouts, retry limits, paths) via configuration. Never hardcode them.

## 6. Duplication (Rule of Three)
* **WET vs DRY:** Logic may be duplicated once without abstraction (Write Everything Twice). Refactor into a shared, generic function only upon the *third* occurrence (Don't Repeat Yourself).