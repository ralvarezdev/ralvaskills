# Stack Versions

Inherits the Go stack from [go-architect/STACK.md](../../languages/go-architect/STACK.md) and [go-library-builder/STACK.md](../go-library-builder/STACK.md).

## Modules

| Module | Latest tag | Go | Status | Key dependencies |
|---|---|---|---|---|
| restkit | v0.10.1 | 1.27.1 | released, pre-1.0 | stdlib only (`problem`, `httpx`, `etag`, `client`, `idempotency`) |
| ginkit | v0.11.0 | 1.27.1 | released | gin 1.12, gin-contrib/cors 1.7, go-playground/validator 10, restkit, ratelimit |
| pgkit | v0.7.0 | 1.27.1 | released | pgx v5, shopspring/decimal, restkit; `migrate`: goose 3; `pgmetrics`: prometheus client; `pgtest`: testcontainers-go |
| mcpkit | v0.4.1 | 1.27.1 | released, stays v0.x until a second MCP server adopts it | modelcontextprotocol/go-sdk |
| identity | v2.3.0 | 1.27.1 | released; module path `github.com/ralvarezdev/identity/v2` | pgx v5, goose 3, pquerna/otp, go-webauthn, golang-jwt v5, valkey-go, x/crypto, email, pgkit, ratelimit |
| identitygin | v0.3.3 | 1.27.1 | released | gin 1.12, identity |
| ratelimit | v0.4.0 | 1.27.1 | released | stdlib root; `valkey`: valkey-go; `valkeytest`: testcontainers valkey |
| email | v0.4.0 | 1.27.1 | released | stdlib root; `smtp`: go-mail |
| webpush | v0.6.0 | 1.27.1 | released, no consumer yet | webpush-go 1.4, pgx v5, goose 3, pgkit |
| termkit | v0.64.1 | 1.27.1 | released, pre-1.0 | bubbletea 1.3, huh 1, lipgloss, cobra, viper |
| svckit | v0.1.0 | 1.27.1 | released 2026-09-30, private | stdlib only |
| tick | v0.1.0 | 1.27.1 | released 2026-09-30, private | stdlib only |
| sqlitekit | v0.1.0 | 1.27.1 | released 2026-09-30, private | modernc.org/sqlite 1.60.1; `migrate`: goose 3 |
| resilience | v0.1.0 | 1.27.1 | released 2026-09-30, private | stdlib only (`backoff`, `retry`, `supervise`) |
| viperkit | v0.1.0 | 1.27.1 | released 2026-10-10, private | stdlib root; `viperconfig`: viper 1.21 |
| grpckit | v0.8.2 | 1.27.1 | released | corekit v0.3.0, resilience v0.2.0 (`grpcretry`), google.golang.org/grpc 1.84, google.golang.org/protobuf 1.36.11, google.golang.org/genproto/googleapis/rpc |
| corekit | v0.3.0 | 1.27.1 | released | stdlib root; `xconv`: shopspring/decimal 1.4.0 |
| contracts | v0.1.0 | 1.27.2 | released 2026-10-15, private; Go is generated and committed | google.golang.org/protobuf 1.36.12, connectrpc.com/connect 1.21, google.golang.org/grpc 1.84; buf 1.71 (generation) |
| connectrpckit | v0.1.0 | 1.27.2 | released 2026-10-15, private | connectrpc.com/connect 1.21, connectrpc.com/grpchealth, connectrpc.com/grpcreflect, corekit 0.3.0, google.golang.org/genproto/googleapis/rpc |

`contracts` and `connectrpckit` are on Go 1.27.2, so a consumer must be on 1.27.2 before importing them.

## Sibling pins

Each module pins the siblings it builds on at the version below (its `go.mod`, 2026-09-30). A consumer's `go.mod` resolves the highest version any module requires, so upgrading a module does not always upgrade its siblings.

| Module | Pins |
|---|---|
| ginkit | restkit v0.9.2, ratelimit v0.4.0 |
| pgkit | restkit v0.10.0 |
| identity | email v0.4.0, pgkit v0.7.0, ratelimit v0.4.0 |
| identitygin | ginkit v0.11.0, identity v2.0.0, restkit v0.10.1, email v0.4.0, ratelimit v0.4.0 |
| webpush | pgkit v0.6.0 |
| grpckit | corekit v0.3.0, resilience v0.2.0 |
| corekit | none |
| connectrpckit | corekit v0.3.0 |
| contracts | none |
| svckit, tick, sqlitekit, resilience | none |

## Consumers

Consumers are tracked in each project's own `go.mod`, not here. A "second consumer" for the promotion rule means another project that needs the same code; check with a `go.mod` and code search across your projects before promoting.

Every current tag needs Go 1.27.1 (`contracts` and `connectrpckit` need Go 1.27.2), so a project still on an older `go` directive must bump it before importing any module here. `grpckit` and `connectrpckit` are the gRPC and Connect transports (both map `corekit.Kind` to their own code), and `corekit` is consumed by both and by every service that returns `*corekit.Error` from its handlers. `contracts` carries the shared contracts and platform-service stubs; `connectrpckit` is the Connect transport.

## Notes

- Tags as of 2026-10-15. Update this file whenever a module is tagged.
- **Go policy:** every module is on Go 1.27.1 (golangci-lint 2.14.0), except `contracts` and `connectrpckit` on Go 1.27.2; a consumer still on 1.26 must bump its `go` directive before it can import any current tag.
- **corekit v0.3.0** is the transport-neutral error vocabulary (`Kind`, `Issue`, `Classifier`); its root is stdlib-only and only `xconv` carries a dependency (`shopspring/decimal`). `grpckit` and `connectrpckit` both pin v0.3.0 and render its `Kind` as their own code.
- **contracts** is also a Buf proto module (`buf build`/`buf generate`, remote plugins, buf 1.71, `breaking` `FILE`); generated Go is committed under `gen/go/` and drift-gated by `task proto:check`. It is consumed as a Go module; there is no registry dependency.
- **connectrpckit** is the Connect sibling of `grpckit` over the same `corekit` vocabulary; `NewClient` takes the full procedure URL, not a base URL.
- **resilience v0.1.0 gaps** found against real call sites: no linear backoff (a CAS retry loop wanted one), no `OnStart`/`OnGiveUp` hooks or consecutive-failure budget (a supervisor that publishes state needed them). Planned for v0.2.0.
- **sqlitekit** requires modernc v1.60.1; modernc v1.52.0 silently ignores mattn-style DSN options (a session recorder was affected and fixed with `_pragma=`).
- **mcpkit v0.4.0 removed** `ToolOutput`, `ToolStatus`, `ToolError`, `ToolSuccess`, `BearerToken`, `RequireAPIKey` and `ErrEmptyAPIKey`; the repo README still lists them.
- **identity v1.10.0 behavior (unchanged in v2)** `App.EnableTOTP`/`GenerateTOTP` do not apply TOTP replay protection or secret sealing (the login second factor does); see [IDENTITY](references/IDENTITY.md).
- **ginkit** carries `CORS`, so the earlier advice to use `gin-contrib/cors` directly is obsolete.

_Last reviewed: 2026-10-15_
_Skill version at last review: 0.6.0_
