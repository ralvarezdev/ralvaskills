# Stack Versions

Inherits the Go stack from [go-architect/STACK.md](../../languages/go-architect/STACK.md) and [go-library-builder/STACK.md](../go-library-builder/STACK.md).

## Kits

| Kit | Latest tag | Status | Key dependencies |
|---|---|---|---|
| restkit | v0.4.1 | released | stdlib only (`problem`, `httpx`, `idempotency`, `client`) |
| ginkit | v0.3.1 | released | gin 1.12, go-playground/validator 10, restkit, ratelimit |
| pgkit | v0.2.1 | released | pgx v5, shopspring/decimal, restkit; `pgtest`: testcontainers-go |
| mcpkit | v0.1.0 | released, stays v0.x until a second MCP server adopts it | modelcontextprotocol/go-sdk |
| identitygin | v0.1.3 | released | gin 1.12, identity v1.6.2 |

## Notes

- Tags as of 2026-09-29. `restkit` v0.4.0 has a lint-ordering issue fixed in v0.4.1; use v0.4.1.
- Update this table whenever a kit is tagged.
