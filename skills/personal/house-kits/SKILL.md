---
name: house-kits
version: 0.1.0
description: Maps a need to the ralvarezdev shared Go kits — restkit (REST envelope, RFC 9457 problem details, cursor codec, HTTP client core, net/http adapter, idempotency contract), ginkit (gin responders, validation, request-id, logger, recovery, idempotency, rate limit), pgkit (pgx/sqlc helpers, transactions, idempotency store, Postgres test container), mcpkit (MCP server helpers), identitygin (identity auth for gin). Use when building or extending a Go REST API, CLI client, MCP server or Postgres store, before hand-writing response envelopes, cursor tokens, pgtype conversions or MCP plumbing.
---

# House Kits

Small, reusable Go modules extracted from finance. Design: `finance-platform/docs/2026-09-29-house-kits-design.md`; what each kit gained since v0.1.0 and how other repos compare: `finance-platform/docs/2026-09-29-kit-adoption.md`. Each follows [go-library-builder](../go-library-builder/SKILL.md): stdlib-only core, one adapter module per framework, config structs over options, no globals. The cross-cutting rules live in the architect skills; the kits are how this house implements them.

## 1. Pick the kit

| Need | Kit | Rule it implements |
|---|---|---|
| Response envelope, RFC 9457 problem details (`problem`: typed `Catalog`, correlation id, 5xx redaction), list `Meta`, cursor token, typed HTTP client for a CLI/MCP | `restkit` |
| `net/http` service without gin: problem writer, request-id, panic recovery (`httpx`); the store contract for Idempotency-Key (`idempotency`) | `restkit` | [rest-api-architect](../../protocols/rest-api-architect/SKILL.md) |
| Gin handlers, validation errors, request-id, request logger, panic recovery, body cap, `Idempotency-Key`, per-user rate limit (CORS: use gin-contrib/cors) | `ginkit` | [gin-architect](../../frameworks/gin-architect/SKILL.md) |
| pgx/sqlc value conversion, `ErrNoRows` mapping, `RunInTx`, Postgres `Idempotency-Key` store, container-backed DB tests | `pgkit` | [sql-architect](../../databases/sql-architect/SKILL.md) |
| MCP panic recovery, backend-delegated auth, per-caller state | `mcpkit` | [mcp-architect](../../protocols/mcp-architect/SKILL.md) |
| Cookie/PAT auth middleware on identity | `identitygin` | identity |

## 2. Dependency direction

`ginkit -> restkit`; never the reverse. `restkit` imports the standard library only, including its `httpx` and `idempotency` packages. `ginkit` uses an idempotency store through the `restkit/idempotency` interface, so it never imports `pgkit`. A different framework gets its own `<framework>kit` adapter, not a branch inside `restkit`.

## 3. Rules

- **Do not re-declare a kit type locally.** A second copy of `Response[T]`, a cursor codec or `pgtype` helpers is a bug; import the kit.
- **A project keeps one wire contract.** Finance uses the `restkit.Response` envelope; `Meta` stays camelCase while bodies stay snake_case, and the frontend, CLI and MCP depend on it. Other repos return bare resources with problem+json errors: adopt the kit's middleware and `problem` there, not the envelope. New projects start on RFC 9457 problems.
- **Do not leak internals on 5xx.** Use `problem.RedactServerError` (or `httpx.Config.WriteError`) so a 5xx `detail` is generic outside debug mode.
- **Idempotency fails closed.** A store outage must refuse the request, never run the handler; release the key on 5xx or panic.
- **Generated code is types only.** oapi-codegen never generates client or server; hand-write calls on `restkit/client.Do`.
- **Adapters own their dependency.** gin lives in `ginkit`, testcontainers in `pgkit/pgtest`, the MCP SDK in `mcpkit`. Consumers that do not use one must not pull it in.
- **Promote only with a second consumer.** Code moves into a kit when a second project needs it. The exception is `mcpkit`, extracted up front for planned reuse and kept on v0.x until a second MCP server adopts it.
- **Verify a kit release against its previous tag** (`git diff <prev-tag>`) before labelling it additive.

## 4. Adding to a kit

1. Confirm a second consumer exists and the code imports no project domain package.
2. Add it to the narrowest module that fits (§1, §2), with a test.
3. Tag, then migrate the consumer in one commit and delete the local copy.
4. Update this skill's table and [STACK.md](STACK.md).

See [STACK.md](STACK.md) for kit versions and their pinned dependencies.
