# Stack Versions

Inherits the Go stack from [go-architect/STACK.md](../../languages/go-architect/STACK.md) and [go-library-builder/STACK.md](../go-library-builder/STACK.md).

## Kits

| Kit | Latest tag | Status | Key dependencies |
|---|---|---|---|
| restkit | — | designed, not built | stdlib only |
| ginkit | — | designed, not built | gin 1.12, go-playground/validator 10, restkit, ratelimit |
| pgkit | — | designed, not built | pgx v5, shopspring/decimal; `pgtest`: testcontainers-go |
| mcpkit | — | designed, not built | modelcontextprotocol/go-sdk |
| identitygin | v0.1.0 | released | gin 1.12, identity |

## Notes

- Fill in the "Latest tag" column as each kit is tagged, and drop the draft banner in SKILL.md.
- identitygin v0.1.0 requires identity v1.5.0 in its own go.mod; finance currently builds against identity v1.6.0.
