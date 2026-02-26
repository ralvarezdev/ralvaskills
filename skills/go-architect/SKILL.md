---
name: go-architect
description: Defines strict architectural patterns for Go structures, typed enums, constants, and standard domain modeling for any project.
---
# Go Architecture & Domain Modeling

## 1. Constants, Enums & Sentinel Errors
- Strict Typing: All enums must be defined using custom types and iota. Never use bare strings or integers for domain states.
- Stringer Implementation: Always generate or write a String() method for custom enums to ensure clean logging and monitoring.
- Constant Grouping: Group related constants in const blocks (e.g., default timeouts, configuration keys).
- Bitmasking: For complex, overlapping states, use iota with bitwise shifts (e.g., 1 << iota) to allow state combining.
- Sentinel Errors: Define package-level expected errors as exported variables (e.g., var ErrNotFound = errors.New("not found")) rather than returning inline strings. This allows the caller to use errors.Is().

## 2. Structures, Payloads & Memory
- Data Schemas: Structs designed for network transport, database models, or external APIs must use explicit tags (e.g., json, db). Always include standard audit fields where applicable (e.g., CreatedAt, ID).
- Composition over Inheritance: Build complex domain models by embedding smaller, focused structs rather than creating massive single structures.
- Memory Alignment: Order struct fields from largest data type to smallest (e.g., int64 down to bool) to minimize memory padding and optimize footprint.
- Value vs. Pointer Semantics: Use pointers (*Struct) when the struct contains a Mutex or needs to be mutated. Use values (copies) for small, immutable data payloads to reduce garbage collection overhead.
- Validation: Structures that receive external data must have a corresponding validation method (e.g., Validate() error) to check for missing or malformed fields before processing.

## 3. Instantiation & Configuration
- Constructors: Always provide a constructor function (e.g., NewService(...)) that initializes maps, slices, and channels safely to prevent panics.
- Functional Options Pattern: For complex structures (like servers or external clients), do not pass a massive config struct or multiple arguments into the constructor. Use functional options (e.g., func WithTimeout(t time.Duration) Option) to make configuration extensible and clean.

## 4. Interface Philosophy
- Consumer Defines the Interface: Define interfaces where they are *used*, not where the underlying struct is *implemented*. 
- Keep Interfaces Small: Aim for 1-3 methods per interface (e.g., type Reader interface { Read(p []byte) (n int, err error) }). Broad interfaces break abstractions.
- Accept Interfaces, Return Structs: Functions should take interfaces as arguments (for easy mocking/testing) but return concrete structs.

## 5. Concurrency & State Management
- Mutex Placement: If a struct has fields that require concurrent access, embed sync.RWMutex directly above the fields it protects, and group those fields together visually.
- Atomic Operations: For simple, highly contested numerical metrics (like a counter), use sync/atomic instead of wrapping a standard integer in a Mutex.
- Context Plumbing: Every method that touches the network, disk, database, or a blocking channel MUST take ctx context.Context as its first parameter to ensure graceful cancellation and timeout management.

## 6. Package Structure & Naming
- No Dump Packages: Never create packages named util, common, or helpers. Packages must be named after exactly what they provide (e.g., retry, stringset, jsonparser).
- Domain-Driven Layout: Organize code by domain/feature, not by technical layer. Prefer package invoice (containing its own models, DB logic, and handlers) over a global package models and package handlers.
- Package Names: Package names must be short, lowercase, and one word. Avoid underscores or mixedCaps.

## 7. Testing Standards
- Table-Driven Tests: All unit tests for business logic must use the table-driven test pattern (a slice of anonymous structs defining name, input, and want).
- t.Helper(): Any custom test assertion or helper function must call t.Helper() as its very first line to ensure test failure lines point to the actual test case, not the helper function.
- Black-Box Testing: Whenever possible, use the package mypkg_test naming convention for test files to enforce testing only the exported API of the package.

## 8. Advanced Error Handling & Panics
- Error Types: Use errors.As() when you need to extract and inspect the fields of a custom error struct, and errors.Is() when checking against a sentinel error.
- Panics: panic must *only* be used during application initialization (e.g., failing to load a critical config or compiling a regex via MustCompile). Never use panic for regular control flow or returning errors to a caller.

## 9. Dependency Injection & Observability
- Inversion of Control: Do not use global state (var DB *sql.DB) or init() functions to set up database connections or loggers. Pass dependencies explicitly into constructors.
- Structured Logging: Always use structured logging (e.g., log/slog or zap). Never use standard fmt.Printf or unstructured log.Println in production code. Pass contextual data as key-value pairs.

## 10. Generics (Go 1.18+)
- Rule of Thumb: Use type parameters ([T any]) for data structures (like custom caches or trees) and utility functions (like slices/maps manipulation). Do *not* use generics when an interface will solve the problem just as well.

## 11. Documentation & Comments
- Idiomatic Comments: All exported names (types, functions, constants, variables) MUST have a doc comment. The comment must be a complete sentence that begins with the exact name of the object being declared (e.g., // Server represents the main TCP server.).
- Package Docs: Every package must have a package-level comment immediately preceding the package clause to explain its overarching purpose.
- Parameters & Returns: Never use artificial tags (like @param, @return, or Args:). Weave the exact names of parameters and return values naturally into the descriptive prose of the comment (e.g., // Parse parses the string s and returns the extracted ID or an ErrInvalidFormat.).
- Error Documentation: Always explicitly document the conditions under which a function will return an error, referencing specific sentinel errors by name if applicable.
- Why, Not What: In-line comments must explain *why* a piece of code exists (the business rule, race condition, or edge case), never *what* the code is doing mechanically. Assume the reader knows Go.