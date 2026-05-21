---
name: go-architect
version: 1.0.0
description: >
  Enforces strict architectural standards for Go 1.26 — memory-aligned structs,
  typed enums, interface design philosophy, goroutine safety, iterator patterns,
  idiomatic error handling, and the sqlx + `//go:embed` SQL pattern. Use when
  writing, reviewing, or refactoring Go code, scaffolding a new Go service,
  or auditing an existing codebase against modern idioms.
---

# Go Architecture & Domain Modeling

Targets **Go 1.26**. See [STACK.md](STACK.md) for pinned dependency versions.

## 1. Constants & Enums

- **Enums:** Custom types with `iota` and an implemented `String()` method. Avoid bare primitives. Use `1 << iota` for bitmasks.
- **Errors:** Export package-level sentinel errors (`var ErrNotFound = ...`) for `errors.Is()`. Use `errors.Join` to aggregate multiple errors.
- **Grouping:** Group related constants in `const` blocks.

## 2. Structures & Memory

- **Design:** Explicit struct tags, standard audit fields, composition via embedding.
- **Optimization:** Order fields largest → smallest to minimize padding.
- **Semantics:** Pointers for mutation or mutex-protected state; values for small, immutable payloads.
- **Validation:** `Validate() error` for structs taking external input; tag-based validation via `go-playground/validator`.
- **Ordering:** Use `cmp` (`cmp.Compare[T]`, `cmp.Or`, `cmp.Less`) as the default ordering toolkit.

## 3. Instantiation

- **Constructors:** Always provide `New...` to initialize maps, slices, channels safely.
- **Configuration:** Functional Options pattern for complex setup; never large config structs passed by value.

## 4. Interfaces

- **Design:** Define interfaces where they're *used*, not where implemented. Keep them small (1–3 methods).
- **Signatures:** Accept interfaces (for easy mocking); return concrete structs.

## 5. Stdlib defaults

Prefer stdlib over hand-rolled or third-party when stdlib now covers it.

- `slices`, `maps` — collection helpers (`Sort`, `Contains`, `Index`, `Clone`, `Equal`, `Concat`, etc.)
- `cmp` — `Compare`, `Less`, `Or`, `Ordered` — the default ordering toolkit.
- `iter` — `iter.Seq[T]`, `iter.Seq2[K,V]` for sequence-shaped returns (see §6).
- `log/slog` — structured logging; **default for new code**. Reach for `zap`/`zerolog` only when slog's allocator perf is provably insufficient.
- `errors` — `errors.Is`, `errors.As`, `errors.Join`.

## 6. Iterators (Go 1.23+)

- **Range-over-func:** `for x := range seq { ... }` where `seq` is `iter.Seq[T]`.
- **Return `iter.Seq[T]` when:** the sequence is large/unknown-size, consumer might short-circuit, or it's backed by an underlying cursor (DB pagination, HTTP pages).
- **Return `[]T` when:** the result is small, bounded, and the caller almost always wants the whole thing.
- **Anti-pattern:** wrapping a `[]T` in `slices.Values` just to look modern. Round-tripping back to `slices.Collect` is wasted ceremony.

## 7. Concurrency

- **State:** Embed `sync.RWMutex` directly above the fields it protects. `sync/atomic` for simple counters.
- **WaitGroup:** Use `sync.WaitGroup.Go(fn)` (Go 1.25+) — replaces `Add(1) / go func() { defer Done(); ... }()` boilerplate.
- **Context:** Pass `context.Context` as the first parameter for any blocking or I/O. Use `signal.NotifyContext` for shutdown; the cancellation cause now records which signal fired (Go 1.26).
- **Containers:** `GOMAXPROCS` respects cgroup CPU limits since Go 1.25 — don't hardcode in containerized deploys.

## 8. Errors & Panics

- **Checking:** `errors.As` for custom types; `errors.Is` for sentinels.
- **Aggregation:** `errors.Join(errs...)` when collecting multiple errors (validation, parallel fan-out).
- **Wrapping:** `fmt.Errorf("op: %w", err)` to add context while preserving the chain.
- **Panics:** Restrict to initialization (`MustCompile` style). Never as control flow.

## 9. Testing

- **Patterns:** Table-driven tests and black-box testing (`package mypkg_test`).
- **Helpers:** `t.Helper()` as the first line in any custom assertion.
- **Concurrent code:** Use `testing/synctest` (GA in Go 1.25) — virtualizes time, waits for goroutines to quiesce, makes async tests deterministic.
- **Library:** `stretchr/testify` for assertions/mocks/suites when stdlib alone is awkward.

## 10. Packages

- **Naming:** Short, lowercase, single-word. Never `util`, `common`, `helpers`.
- **Layout:** Organize by domain/feature, not by technical layer.

## 11. Dependencies & Logging

- **Injection:** Pass dependencies via constructors. No global state, no package `init()` setups. For larger graphs use `uber-go/fx` or `google/wire`.
- **Config:** `spf13/viper` for layered env + file + flag configuration.
- **Logging:** `log/slog` (stdlib). Structured, context-aware. `zap`/`zerolog` only when slog is provably the bottleneck — start with slog.

## 12. Database access — SQL files + `//go:embed`

Strong preference: **raw SQL in `.sql` files**, embedded at compile time, executed via `jmoiron/sqlx`. Avoids ORM magic, keeps queries auditable in git, gives editors full SQL syntax highlighting and linting.

```go
//go:embed queries/get_user_by_id.sql
var getUserByIDSQL string

func (r *UserRepo) GetByID(ctx context.Context, id int64) (User, error) {
    var u User
    err := r.db.GetContext(ctx, &u, getUserByIDSQL, id)
    return u, err
}
```

Layout:

```
internal/userrepo/
├── repo.go
└── queries/
    ├── get_user_by_id.sql
    ├── insert_user.sql
    └── list_users.sql
```

- **Migrations:** `golang-migrate` — versioned up/down pairs in `migrations/`.
- **Dynamic queries:** If a query needs runtime composition (optional filters), still write the static fragments in `.sql` files and join them in Go. Avoid building SQL by string concatenation against user input — use parameter binding.

## 13. Generics

- **Usage:** Data structures and utility functions only. If an interface would serve, prefer it.
- **Self-referential constraints (Go 1.26):** `type Adder[A Adder[A]] interface { Add(A) A }` is valid — useful for fluent-style domain types.
- **Type aliases with type params (Go 1.24):** `type IntMap[V any] = map[int]V`.

## 14. Tooling

- **Lint:** `golangci-lint` **v2** — a single binary that aggregates `staticcheck`, `govet`, `gofmt`, `goimports`, `revive`, and dozens more. Pin via `.golangci.yml` (v2 config schema). Run on every commit and in CI. Treat warnings as errors.
- **Modernization:** Run `go fix` periodically (revamped in Go 1.26 as push-button modernizers). It rewrites legacy patterns toward current stdlib APIs and respects `//go:fix inline` directives for local refactors. Ongoing hygiene, not a one-off.
- **Format:** `gofmt` (built-in) plus `goimports` (via `golangci-lint`) — no other formatter, no debate.
- **Build/run:** `go build`, `go test`, `go vet`, `go work` for multi-module repos. Stdlib only — avoid Make/Task wrappers unless multi-language CI demands it.

## 15. Documentation

- **Format:** Exported names get full-sentence comments starting with the identifier name. Include a package-level comment.
- **Style:** Weave parameter and return names into prose (no `@param`). Document specific error conditions.
- **Focus:** Explain *why* (business rules, edge cases) — code already shows *what*.

## Canonical libraries

See [STACK.md](STACK.md) for the full pinned list — viper, validator, cobra/pflag, gin, sqlx, grpc, protobuf-go, testify, fx, golang-migrate, buf.
