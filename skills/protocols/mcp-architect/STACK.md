# Stack Versions

This skill pins **protocol-level** spec revisions and the **Python + Go** SDK runtimes used in [RECIPES.md](RECIPES.md). Other-language SDKs (TypeScript, C#, Java, Kotlin, Ruby, Swift, Rust) implement the same protocol; pin their runtimes in the consuming project's manifest.

| Dependency | Pinned version | Purpose |
|---|---|---|
| MCP spec | 2025-11-25 | Protocol revision targeted by SKILL.md examples |
| MCP spec (upcoming) | 2026-07-28 RC | Locked 2026-05-21; ten-week validation window. Notable for stateless core, mandatory `Mcp-Method`/`Mcp-Name` routing headers, `ttlMs`/`cacheScope` on list/read, W3C Trace Context in `_meta` |
| JSON-RPC | 2.0 | Wire protocol — fixed by spec |
| Python — `mcp` | 1.27.1 | Official Python SDK (PyPI `mcp`); ships FastMCP server, stdio + Streamable HTTP, OAuth helpers |
| Python — `pydantic` | 2.x | Tool input/output schema + validation; matches [python-architect](../../languages/python-architect/STACK.md) |
| Python — `httpx` | 0.28+ | Client HTTP for tools that call upstream APIs |
| Python — `pytest` + `pytest-asyncio` | 9.x / 0.24+ | Test framework |
| Go — `github.com/modelcontextprotocol/go-sdk` | 1.4.0+ | Official Go SDK (maintained with Google); supports specs 2024-11-05 through 2025-11-25; packages: `mcp`, `jsonrpc`, `auth`, `oauthex` |
| Go runtime | 1.26 | Matches [go-architect](../../languages/go-architect/STACK.md) |
| MCP Inspector | 0.10+ | Visual + CLI tester. **Pin ≥0.10** — older versions have CVE-2025-49596 (RCE) |
| OAuth 2.1 | draft-ietf-oauth-v2-1 | Authorization spec for remote servers |
| RFC 8707 | published | Resource Indicators — MANDATORY for MCP clients and servers |
| RFC 9728 | published | OAuth 2.0 Protected Resource Metadata — `/.well-known/oauth-protected-resource` |
| RFC 7591 | published | Dynamic Client Registration — practical onboarding path |

## Notes

- **Stable spec target is 2025-11-25.** SKILL.md and RECIPES.md examples are written against this revision. Where 2026-07-28 introduces something useful (e.g. `ttlMs` / `cacheScope` on resources), it's called out inline and degrades gracefully on older clients.
- **Python SDK note.** The `mcp` package ships FastMCP as `mcp.server.fastmcp.FastMCP`. The standalone `fastmcp` PyPI package (a community fork) is **not** the canonical path — recipes use the official `mcp` package only.
- **Go SDK note.** `github.com/modelcontextprotocol/go-sdk` (Google-maintained, official since mid-2025) supersedes the earlier community `github.com/mark3labs/mcp-go`. New code should use the official SDK; only stay on `mark3labs/mcp-go` if you're maintaining an existing server.
- **Transport scope.** Recipes cover Streamable HTTP (default for remote) and stdio (default for local subprocess). Plain SSE is deprecated as of spec 2025-03-26 and intentionally omitted.
- **Auth scope.** OAuth 2.1 + RFC 8707 only; no legacy implicit / ROPC flows.

_Last reviewed: 2026-05-23_
_Skill version at last review: 1.0.0_
