---
name: logic-cleaner
description: Language-agnostic refactoring engine. Reduces cyclomatic complexity, flattens control flow, and enforces strict naming and readability standards.
---
# Universal Logic & Code Cleanliness Standards

## 1. Control Flow & Guard Clauses
- Flattening: Code must never exceed 2 levels of indentation within a function. Extract deeper loops or conditionals into private helper functions.
- Guard Clauses (Early Returns): Do not use if/else wrappers for the "happy path". Handle edge cases, validations, and errors at the very top of the function and return immediately. Keep the core logic aligned to the left margin.
- Fail Fast: If an invalid state is detected, halt execution immediately rather than passing nulls or error flags down the call stack.

## 2. Boolean Algebra & Conditionals
- Positive Phrasing: Never use double negatives. Rename variables like isNotFinished to isPending or isComplete. Avoid if !notReady.
- Condition Extraction: If an if statement contains more than two logical operators (&&, ||), extract the entire condition into a well-named boolean variable or a separate function (e.g., isValidRequest = hasToken && !isExpired).
- De Morgan's Laws: Simplify negated compound conditions. Convert !(A || B) into !A && !B for better readability.

## 3. State & Mutation
- Variable Scope: Declare variables as close to their first use as possible. Do not declare all variables at the top of a function.
- Immutability by Default: Treat function parameters as strictly read-only. Never reassign a parameter variable (e.g., input_string = input_string.trim()). Assign the result to a new variable instead.
- Shadowing: Never shadow variables from an outer scope. Always use distinct, descriptive names for inner loop variables.

## 4. Naming & Semantics
- No Abbreviations: Do not use req, res, err, ctx, idx, or ptr unless they are the universally accepted standard in that specific language (e.g., ctx in Go). Prefer request, index, pointer.
- Action-Oriented Functions: Function names must start with a strong verb that describes the exact action being performed (e.g., CalculateTotal(), ParseConfig()). Avoid generic names like ProcessData() or Handle().
- Symmetric Naming: Use opposite pairs for related concepts: start/stop, open/close, add/remove, push/pop.

## 5. Magic Values & Hardcoding
- Zero Magic: No raw strings, numbers, or boolean flags should be hardcoded deep within logic. If a number other than 0 or 1 appears, it must be extracted to a named constant explaining *why* it is that value.
- Configuration over Code: Any value that might change based on the environment (timeouts, retry limits, file paths) must be injected via configuration, never hardcoded.

## 6. The Rule of Three (DRY vs WET)
- WET (Write Everything Twice): It is acceptable to have duplicated logic twice if extracting it would create a premature or forced abstraction.
- DRY (Don't Repeat Yourself): If the exact same logic appears a *third* time, you must refactor it into a shared, generic function.