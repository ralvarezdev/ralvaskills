# Stack Versions

| Dependency | Pinned version | Purpose |
|---|---|---|
| go | 1.26 | Language |
| log/slog | stdlib (tracks Go) | Structured logging — **default** for new code |
| spf13/viper | 1.21 | Layered config (env + file + flag) |
| go-playground/validator | 10.30 | Struct-tag validation |
| spf13/cobra | 1.10 | CLI command framework |
| spf13/pflag | 1.0 | POSIX-style flags |
| gin-gonic/gin | 1.12 | HTTP router (REST) |
| jmoiron/sqlx | 1.4 | SQL extensions — used with `//go:embed` for `.sql` files |
| google.golang.org/protobuf | 1.36 | Protocol Buffers runtime |
| google.golang.org/grpc | 1.81 | gRPC |
| stretchr/testify | 1.11 | Assertions, mocks, suites |
| uber-go/fx | 1.24 | DI graph wiring |
| golang-migrate/migrate | 4.19 | SQL migrations (versioned up/down) |
| buf | 1.69 | Proto toolchain (linter, breaking-change detector) |
| golangci-lint | 2.12 | Aggregated linter (staticcheck + govet + revive + goimports + dozens more); v2 config schema |

## Notes

- **`log/slog` is the logging default.** Reach for `uber-go/zap` or `rs/zerolog` only when slog's allocator performance is provably insufficient — start with slog.
- **No ORM.** Database access goes through `sqlx` with raw SQL in `.sql` files embedded via `//go:embed`. This is an architectural opinion enforced by `go-architect`.
- **Dependency injection:** `uber-go/fx` is the default for larger graphs. `google/wire` is acceptable when compile-time-only wiring is preferred.

_Last reviewed: 2026-07-05_
_Skill version at last review: 1.2.0_
