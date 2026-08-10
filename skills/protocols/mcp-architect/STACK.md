# Stack Versions

This skill pins **protocol-level** spec revisions and the **Python + Go** SDK runtimes used in [RECIPES.md](RECIPES.md). Other-language SDKs (TypeScript, C#, Java, Kotlin, Ruby, Swift, Rust) implement the same protocol; pin their runtimes in the consuming project's manifest.

| Dependency | Pinned version | Purpose |
|---|---|---|
| MCP spec | 2026-07-28 | Protocol revision targeted by SKILL.md examples — stable, published final 2026-07-28. Stateless core: no `initialize` handshake, no `Mcp-Session-Id`; mandatory `Mcp-Method`/`Mcp-Name` routing headers; `ttlMs`/`cacheScope` mandatory on list/read; MRTR (`InputRequiredResult`) replaces server-initiated elicitation/sampling/roots pushes |
| MCP spec (previous) | 2025-11-25 | Superseded. Servers still speaking it use `Mcp-Session-Id` + `initialize`/`initialized` — a materially different transport contract, not just an additive change |
| JSON-RPC | 2.0 | Wire protocol — fixed by spec |
| Python — `mcp` | ≥2.0.0 (first PyPI release targeting 2026-07-28; verify exact minor against the package changelog before pinning) | Official Python SDK (PyPI `mcp`); ships FastMCP server, stdio + Streamable HTTP, OAuth helpers. FastMCP 4.0 adds first-class stateless interactivity (MRTR) and the Tasks extension |
| Python — `pydantic` | 2.x | Tool input/output schema + validation; matches [python-architect](../../languages/python-architect/STACK.md). Schemas now support full JSON Schema 2020-12 (`oneOf`/`anyOf`/`allOf`, `$ref`) per spec 2026-07-28 |
| Python — `httpx` | 0.28+ | Client HTTP for tools that call upstream APIs |
| Python — `pytest` + `pytest-asyncio` | 9.x / 0.24+ | Test framework |
| Go — `github.com/modelcontextprotocol/go-sdk` | ≥2.0.0 (first tag targeting 2026-07-28; verify exact minor against the module's release notes before pinning) | Official Go SDK (maintained with Google); Tier-1 SDK for spec 2026-07-28; packages: `mcp`, `jsonrpc`, `auth`, `oauthex` |
| Go runtime | 1.26 | Matches [go-architect](../../languages/go-architect/STACK.md) |
| MCP Inspector | 0.10+, spec-2026-07-28-aware build | Visual + CLI tester. **Pin ≥0.10** — older versions have CVE-2025-49596 (RCE). Use a release that understands header-based routing and MRTR when testing against this spec target |
| OAuth 2.1 | draft-ietf-oauth-v2-1 | Authorization spec for remote servers |
| RFC 8707 | published | Resource Indicators — MANDATORY for MCP clients and servers |
| RFC 9728 | published | OAuth 2.0 Protected Resource Metadata — `/.well-known/oauth-protected-resource` |
| RFC 9207 | published | OAuth 2.0 Authorization Server Issuer Identification — `iss` validation, MANDATORY as of spec 2026-07-28 |
| CIMD | draft, per MCP spec 2026-07-28 | Client ID Metadata Documents — supersedes RFC 7591 DCR as the standard onboarding path |
| RFC 7591 (DCR) | published, **deprecated** by MCP spec 2026-07-28 | Dynamic Client Registration — keeps working for backward compatibility; migrate new servers to CIMD |

## Notes

- **Stable spec target is 2026-07-28.** SKILL.md and RECIPES.md examples are written against this revision. It is a breaking rewrite of the transport contract versus 2025-11-25 — don't treat it as additive when migrating an existing server; audit every `Mcp-Session-Id` reference, every server-initiated `elicitation/create`/`sampling/createMessage`/`roots/list` call, and every hardcoded `-32002`.
- **SDK version numbers above are floors, not exact pins** — the Tier-1 SDKs (TypeScript, Python, Go, C#) all began shipping 2026-07-28 support at the spec's stable release; check each package's changelog for the first release that dropped 2025-11-25-only session handling before pinning an exact version in a real project.
- **Python SDK note.** The `mcp` package ships FastMCP as `mcp.server.fastmcp.FastMCP`. The standalone `fastmcp` PyPI package (a community fork) is **not** the canonical path — recipes use the official `mcp` package only.
- **Go SDK note.** `github.com/modelcontextprotocol/go-sdk` (Google-maintained, official since mid-2025) supersedes the earlier community `github.com/mark3labs/mcp-go`. New code should use the official SDK; only stay on `mark3labs/mcp-go` if you're maintaining an existing server.
- **Transport scope.** Recipes cover Streamable HTTP (default for remote) and stdio (default for local subprocess). Legacy HTTP+SSE is deprecated as of spec 2026-07-28 (twelve-month offramp) and intentionally omitted.
- **Auth scope.** OAuth 2.1 + RFC 8707 + RFC 9207 issuer validation; CIMD over DCR for new servers; no legacy implicit / ROPC flows.
- **Tasks scope.** `io.modelcontextprotocol/tasks` is now an opt-in extension (negotiated via `ClientCapabilities`/`ServerCapabilities.extensions`), not a core method group. Not covered in RECIPES.md — treat it as a separate reference lookup if you need it.

_Last reviewed: 2026-08-10_
_Skill version at last review: 2.0.0_
