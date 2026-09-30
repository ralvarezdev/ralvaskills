# Stack Versions

Inherits the Go stack from [go-architect/STACK.md](../../languages/go-architect/STACK.md) and [go-library-builder/STACK.md](../go-library-builder/STACK.md).

## Modules

| Module | Latest tag | Go | Status | Key dependencies |
|---|---|---|---|---|
| restkit | v0.6.0 | 1.26 | released, pre-1.0 | stdlib only (`problem`, `httpx`, `etag`, `client`, `idempotency`) |
| ginkit | v0.6.0 | 1.26.4 | released | gin 1.12, gin-contrib/cors 1.7, go-playground/validator 10, restkit, ratelimit |
| pgkit | v0.5.0 | 1.26.0 | released | pgx v5, shopspring/decimal, restkit; `migrate`: goose 3; `pgmetrics`: prometheus client; `pgtest`: testcontainers-go |
| mcpkit | v0.2.0 | 1.26 | released, stays v0.x until a second MCP server adopts it | modelcontextprotocol/go-sdk 1.8 |
| identity | v1.6.4 | 1.26.4 | released | pgx v5, goose 3, pquerna/otp, go-webauthn, golang-jwt v5, valkey-go, x/crypto, email, pgkit, ratelimit |
| identitygin | v0.1.3 | 1.26 | released | gin 1.12, identity |
| ratelimit | v0.3.0 | 1.26.4 | released | stdlib root; `valkey`: valkey-go; `valkeytest`: testcontainers valkey |
| email | v0.2.0 | 1.26 | released | stdlib root; `smtp`: gomail v2 |
| webpush | v0.4.3 | 1.26.4 | released; `Send` status handling merged, not yet tagged (below) | webpush-go 1.4, pgx v5, goose 3, pgkit |
| termkit | v0.55.0 | 1.27.1 | released, pre-1.0 | bubbletea 1.3, huh 1, lipgloss, cobra, viper |

## Sibling pins

Each module pins the siblings it builds on at the version below (its `go.mod`, 2026-09-29). A consumer's `go.mod` resolves the highest version any module requires, so upgrading a module does not always upgrade its siblings.

| Module | Pins |
|---|---|
| ginkit | restkit v0.6.0, ratelimit v0.2.1 |
| pgkit | restkit v0.4.1 |
| identity | email v0.2.0, pgkit v0.3.0, ratelimit v0.3.0 |
| identitygin | identity v1.6.2 |
| webpush | pgkit v0.3.0 |

## Consumers

`go.mod` scan of `~/Dev/active` (depth 5), 2026-09-29. A "second consumer" for the promotion rule means a project in this list other than finance.

| Module | Used by |
|---|---|
| restkit, ginkit, pgkit, mcpkit, identitygin, ratelimit | finance-platform |
| identity, email | finance-platform, repuestos-edge |
| termkit | finance-platform (CLI), devtrack, rsk (ralvaskills) |
| webpush | none yet |

## Notes

- Tags as of 2026-09-29. Update this file whenever a module is tagged.
- **webpush**: `master` (`ebede8f`) makes `vapid.Sender.Send` return an error for any non-2xx from the push service and adds `webpush.ErrSubscriptionGone` (404 or 410). It is a behavior break for callers and is not tagged yet; tag it (v0.5.0 is the natural number for a pre-1.0 break) and update this row.
- **Go versions differ** across the modules (1.26, 1.26.0, 1.26.4; termkit on 1.27.1). No policy has been chosen; a consumer builds with the highest.
- **ginkit** carries `CORS`, so the earlier advice to use `gin-contrib/cors` directly is obsolete.

_Last reviewed: 2026-09-29_
_Skill version at last review: 0.3.0_
