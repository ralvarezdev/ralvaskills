# Stack Versions

| Dependency | Pinned version | Purpose |
|---|---|---|
| go | 1.26 | Language |
| gin-gonic/gin | 1.12 | HTTP framework |
| go-playground/validator | 10.30 | Struct-tag request validation |
| jmoiron/sqlx | 1.4 | DB access — see [sql-architect](../../databases/sql-architect/STACK.md) and [go-architect](../../languages/go-architect/STACK.md) |
| golang-jwt/jwt/v5 | 5.3 | JWT issuance + verification (auth pattern A) |
| golang.org/x/crypto | 0.51 | Argon2id password hashing (`argon2.IDKey`) |
| swaggo/swag | 1.16 | OpenAPI generation via annotation comments (default) |
| getkin/kin-openapi | 0.138 | Code-first OpenAPI 3.1 + runtime validation (alternative) |

## Notes

- **Inherits the `go-architect` stack.** This file pins what is *specific* or *additionally required* for Gin services. Shared canonical libraries (viper, sqlx, slog, golangci-lint, etc.) live in [go-architect/STACK.md](../../languages/go-architect/STACK.md).
- **Authentication library by pattern** (see [rest-api-architect §11](../../protocols/rest-api-architect/SKILL.md#11-authentication-patterns)):
  - Pattern A (in-house JWT): `golang-jwt/jwt/v5` + `golang.org/x/crypto/argon2`
  - Pattern B (external IdP): `golang-jwt/jwt/v5` only — keys come from JWKS
- **OpenAPI library choice:** `swaggo/swag` is the default for speed; switch to `getkin/kin-openapi` once you need OpenAPI 3.1 features (oneOf, discriminators, runtime validation).
- **Data access:** `sqlx + //go:embed` per [go-architect §12](../../languages/go-architect/SKILL.md#12-database-access--sql-files--goembed). No ORM.

_Last reviewed: 2026-05-21_
_Skill version at last review: 1.0.0_
