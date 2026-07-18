# Stack Versions

Inherits the full Go stack from [go-architect/STACK.md](../../languages/go-architect/STACK.md). Listed here are the toolchain versions the scaffolded `mise.toml`/`Taskfile.yml` pin and the libraries the house style reaches for by default.

## Toolchain (mise-pinned)

| Dependency | Pinned version | Purpose |
|---|---|---|
| go | 1.26 (go.mod); toolchain 1.26.5 | Language; `.golangci.yml` targets 1.26.3 |
| task | 3.52.0 | Task runner (`Taskfile.yml`) |
| golangci-lint | 2.12.2 | Aggregated lint gate (v2 config schema) |
| goimports | 0.48.0 | Formatting with `-local` module prefix |
| betteralign | 0.14.0 | Struct field alignment (`align:check/fix`) |
| trivy | 0.72.0 | FS + secret scanning (`scan`) |
| sqlc | 1.31.1 | Typed SQL codegen (pgx/v5) — DB adapters only |
| goose | v3 | SQL migrations (`//go:embed`) — DB adapters only |

## Default libraries

| Dependency | Pinned version | Purpose |
|---|---|---|
| google/uuid | 1.6 | Only third-party import allowed in the domain package |
| stretchr/testify | 1.11 | Assertions/require in tests (optional; stdlib-only is also fine) |
| testcontainers/testcontainers-go | 0.42 | Container-backed integration tests (one per package) |
| jackc/pgx | v5 | Postgres driver for `postgres` adapters |
| golang-jwt/jwt | v5 | Token adapters |
| gopkg.in/gomail.v2 | 2 | SMTP adapter (email-style libraries) |
| valkey-io/valkey-go | 1.x | Valkey adapters |

## Notes

- **No functional-options, no ORM, no in-library config resolution** — architectural opinions enforced by this skill, not just the linters (see SKILL §5, §11).
- **`exhaustruct` scoped to `Config` structs** and `interfacebloat` capped at 3 are the two golangci settings this skill tunes beyond go-architect's template.
- **`.golangci.yml`** is copied from [go-architect/assets/golangci.yml](../../languages/go-architect/assets/golangci.yml); only `goimports.local-prefixes` and the `exhaustruct` scope change per repo. The doc-coverage config ships separately as `.golangci.doccheck.yml`.

_Last reviewed: 2026-07-17_
_Skill version at last review: 1.0.0_
