# MCP Recipes (Python + Go)

Reference implementations for [SKILL.md](SKILL.md). Python uses the official `mcp` package's FastMCP server; Go uses `github.com/modelcontextprotocol/go-sdk`.

Section index — for skinnier loads, open only the sections you need:

1. [Tool — Python (FastMCP)](#1-tool--python-fastmcp)
2. [Tool — Go (official SDK)](#2-tool--go-official-sdk)
3. [Structured tool output](#3-structured-tool-output)
4. [Resource + template + subscription](#4-resource--template--subscription)
5. [Streamable HTTP server — Python](#5-streamable-http-server-python)
6. [Streamable HTTP server — Go](#6-streamable-http-server-go)
7. [stdio server (both languages)](#7-stdio-server-both-languages)
8. [OAuth 2.1 remote server](#8-oauth-21-remote-server)
9. [Testing — in-process + MCP Inspector](#9-testing--in-process--mcp-inspector)
10. [Client quick reference](#10-client-quick-reference)
11. [SSRF guard for URL-taking tools](#11-ssrf-guard-for-url-taking-tools)

---

## 1. Tool — Python (FastMCP)

```python
# server.py
from mcp.server.fastmcp import FastMCP
from pydantic import BaseModel, Field

mcp = FastMCP("acme-shop", instructions="Operations on the Acme shop catalog.")


class SearchProductsArgs(BaseModel):
    query: str = Field(..., min_length=1, description="Free-text search over product titles")
    limit: int = Field(20, ge=1, le=100, description="Max results")


@mcp.tool(
    title="Search products",
    annotations={"readOnlyHint": True, "openWorldHint": False, "idempotentHint": True},
)
async def search_products(args: SearchProductsArgs) -> list[dict]:
    """Search the product catalog by title. Read-only; safe to call repeatedly."""
    # ... call into your service layer ...
    return [{"id": "sku-1", "title": "Widget", "price_cents": 1999}]
```

Notes:

- The function docstring becomes the tool `description` — the most important field for model selection.
- `annotations` map to the MCP spec's `ToolAnnotations`. Set them explicitly.
- Returning a list/dict auto-serializes to a JSON `content[]` block. For structured output, see [§3](#3-structured-tool-output).
- Pydantic models for args give you free JSON-Schema generation and validation. Validation errors auto-convert to tool errors with `isError: true`.

---

## 2. Tool — Go (official SDK)

```go
// server.go
package main

import (
    "context"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

type SearchProductsArgs struct {
    Query string `json:"query" jsonschema:"required,minLength=1,description=Free-text search over product titles"`
    Limit int    `json:"limit,omitempty" jsonschema:"minimum=1,maximum=100,default=20"`
}

type Product struct {
    ID         string `json:"id"`
    Title      string `json:"title"`
    PriceCents int    `json:"price_cents"`
}

func searchProducts(ctx context.Context, req *mcp.CallToolRequest, args SearchProductsArgs) (
    *mcp.CallToolResult, []Product, error,
) {
    // Business logic — return (result, structuredOutput, err)
    products := []Product{{ID: "sku-1", Title: "Widget", PriceCents: 1999}}
    return nil, products, nil
}

func register(s *mcp.Server) {
    mcp.AddTool(s, &mcp.Tool{
        Name:        "search_products",
        Title:       "Search products",
        Description: "Search the product catalog by title. Read-only; safe to call repeatedly.",
        Annotations: &mcp.ToolAnnotations{
            ReadOnlyHint:    true,
            IdempotentHint:  true,
            OpenWorldHint:   mcp.Ptr(false),
        },
    }, searchProducts)
}
```

Notes:

- `mcp.AddTool[In, Out]` is the generic registration entrypoint — it derives the input JSON schema from `In` (struct tags) and the structured-output schema from `Out`. No hand-written schemas.
- Returning `(nil, structuredValue, nil)` produces a tool result with `structuredContent` populated; the SDK also emits a `content[]` text fallback automatically.
- For tool-level failures (bad input, downstream 404), return `(&mcp.CallToolResult{IsError: true, Content: [...]}, zeroOut, nil)`. The `error` return is reserved for *protocol* failures.

---

## 3. Structured tool output

Two channels coexist on every result: unstructured `content[]` (always populated; what the model reads) and `structuredContent` (populated when `outputSchema` is declared; what the client app parses). See [SKILL.md §3a](SKILL.md#3a-tool-output--unstructured-content-vs-structuredcontent).

### Structured + auto fallback (Python)

```python
from pydantic import BaseModel

class Product(BaseModel):
    id: str
    title: str
    price_cents: int

@mcp.tool(title="Search products")
async def search_products(query: str, limit: int = 20) -> list[Product]:
    # Returning Pydantic models populates structuredContent and emits a JSON
    # fallback in content[] automatically — both channels in one return.
    return [Product(id="sku-1", title="Widget", price_cents=1999)]
```

### Structured + auto fallback (Go)

```go
// The `Out` type parameter in mcp.AddTool[In, Out] drives outputSchema.
// Returning (nil, structuredValue, nil) populates structuredContent and the
// SDK emits a JSON content[] fallback automatically. See §2.
```

### Mixed unstructured blocks — hand-rolled `content[]`

When you need image/audio/resource_link blocks alongside text — no `outputSchema`, just `content[]`.

```python
from mcp.types import (
    TextContent, ImageContent, EmbeddedResource, ResourceLink,
    CallToolResult, TextResourceContents,
)

@mcp.tool(title="Render chart")
async def render_chart(query: str) -> CallToolResult:
    png_b64 = await render_to_png(query)
    return CallToolResult(content=[
        TextContent(type="text", text=f"Rendered {query}: 7-day trend ↑12%."),
        ImageContent(type="image", data=png_b64, mimeType="image/png"),
        ResourceLink(type="resource_link", uri="acme://reports/weekly-2026-W21",
                     name="Full weekly report", mimeType="text/markdown"),
        EmbeddedResource(type="resource", resource=TextResourceContents(
            uri="acme://queries/chart.sql",
            mimeType="text/x-sql",
            text="SELECT date, value FROM metrics WHERE ...",
        )),
    ])
```

### When to declare `outputSchema` (and when not to)

- **Declare it** when the client app or downstream agent code will programmatically parse the result.
- **Skip it** for free-form prose tools (summarizers, explainers) — one `TextContent` block is the right answer.
- **Always emit a `text` fallback** in `content[]` even when `structuredContent` is set. Most 2026 chat clients still render only `content[]`; without a text block they show nothing. Both Python and Go SDKs do this automatically when you return typed values — only worry about it when you build `CallToolResult` by hand.

---

## 4. Resource + template + subscription

**Python:**

```python
from mcp.server.fastmcp import FastMCP

mcp = FastMCP("acme-shop")

@mcp.resource("acme://catalog/categories", mime_type="application/json")
async def categories() -> dict:
    return {"categories": ["widgets", "gizmos"]}

# Template — clients can read acme://catalog/product/{sku}
@mcp.resource("acme://catalog/product/{sku}", mime_type="application/json")
async def product(sku: str) -> dict:
    p = await load_product(sku)
    if p is None:
        raise FileNotFoundError(sku)   # → JSON-RPC -32002 Resource not found
    return p.model_dump()
```

Subscriptions are managed by FastMCP automatically when you enable them in the server constructor; emit changes via `mcp.send_resource_updated(uri)`.

**Go:**

```go
mcp.AddResource(s, &mcp.Resource{
    URI:      "acme://catalog/categories",
    Name:     "Categories",
    MIMEType: "application/json",
}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
    return mcp.NewReadResourceResultJSON(req.Params.URI, map[string]any{
        "categories": []string{"widgets", "gizmos"},
    }), nil
})

mcp.AddResourceTemplate(s, &mcp.ResourceTemplate{
    URITemplate: "acme://catalog/product/{sku}",
    Name:        "Product",
    MIMEType:    "application/json",
}, readProduct)
```

---

## 5. Streamable HTTP server — Python

```python
# main.py — `python main.py` listens on :8000
import uvicorn
from mcp.server.fastmcp import FastMCP

mcp = FastMCP("acme-shop", stateless_http=False)  # stateful: enables subscriptions
# ... register tools / resources / prompts ...

# FastMCP exposes a Starlette app at .streamable_http_app()
app = mcp.streamable_http_app()  # mounts /mcp endpoint with GET/POST/DELETE

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
```

For stateless deployments (horizontally scalable behind a plain HTTP LB), pass `stateless_http=True`. You lose subscriptions and server-initiated requests; each POST is independent.

Mount behind a reverse proxy that strips client-supplied `Mcp-Session-Id` headers on `initialize` — the server is the only entity allowed to mint session IDs.

---

## 6. Streamable HTTP server — Go

```go
package main

import (
    "context"
    "log/slog"
    "net/http"

    "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
    handler := mcp.NewStreamableHTTPHandler(
        func(r *http.Request) *mcp.Server {
            s := mcp.NewServer(&mcp.Implementation{
                Name:    "acme-shop",
                Version: "1.0.0",
            }, &mcp.ServerOptions{
                InitializedHandler: func(ctx context.Context, req *mcp.InitializedRequest) {
                    slog.InfoContext(ctx, "client initialized", "client", req.Params.ClientInfo)
                },
            })
            register(s)   // see §2
            return s
        },
        &mcp.StreamableHTTPOptions{Stateless: false},
    )

    mux := http.NewServeMux()
    mux.Handle("/mcp", handler)

    srv := &http.Server{Addr: ":8000", Handler: mux}
    slog.Info("listening", "addr", srv.Addr)
    _ = srv.ListenAndServe()
}
```

The handler factory is invoked **per request** in stateless mode and **per session** in stateful mode. Don't pre-build a single `mcp.Server` and reuse it across requests in stateless mode — it'll leak handler state.

---

## 7. stdio server (both languages)

**Python:**

```python
# entrypoint declared as a console script in pyproject.toml
from mcp.server.fastmcp import FastMCP

mcp = FastMCP("acme-local")
# ... register tools ...

if __name__ == "__main__":
    mcp.run()   # defaults to stdio transport
```

**Go:**

```go
s := mcp.NewServer(&mcp.Implementation{Name: "acme-local", Version: "1.0.0"}, nil)
register(s)

if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
    slog.Error("server exited", "err", err)
    os.Exit(1)
}
```

Stdio discipline:

- **stdout = JSON-RPC frames only.** No `print()` / `fmt.Println` for debugging — it corrupts the framing.
- **stderr = logs.** Structured (NDJSON or `slog`), captured by the client and shown in Claude Desktop logs.
- **One process per client.** No concurrency model. Crash on init failure rather than half-running.

Client config example (Claude Desktop `~/Library/Application Support/Claude/claude_desktop_config.json` on macOS, `%APPDATA%\Claude\claude_desktop_config.json` on Windows):

```json
{
  "mcpServers": {
    "acme-local": {
      "command": "uvx",
      "args": ["acme-mcp-server"]
    }
  }
}
```

---

## 8. OAuth 2.1 remote server

**Discovery — `/.well-known/oauth-protected-resource`** (RFC 9728):

```json
{
  "resource": "https://mcp.example.com",
  "authorization_servers": ["https://auth.example.com"],
  "bearer_methods_supported": ["header"],
  "scopes_supported": ["acme.read", "acme.write"]
}
```

**Token validation pseudocode (server-side, every request):**

```
1. Extract Bearer token from Authorization header.
2. Verify signature against the auth server's JWKS.
3. Verify exp, nbf, iss.
4. Verify aud equals "https://mcp.example.com"  (RFC 8707 audience binding).
5. Verify scope claim covers what this tool/resource requires.
6. Reject with 401 + WWW-Authenticate (RFC 6750) on failure.
```

**Python — minimal middleware sketch:**

```python
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.responses import JSONResponse

class BearerAuthMiddleware(BaseHTTPMiddleware):
    def __init__(self, app, audience: str, jwks_client):
        super().__init__(app)
        self.audience = audience
        self.jwks = jwks_client

    async def dispatch(self, request, call_next):
        if request.url.path == "/.well-known/oauth-protected-resource":
            return await call_next(request)
        auth = request.headers.get("authorization", "")
        if not auth.startswith("Bearer "):
            return JSONResponse({"error": "unauthorized"}, status_code=401,
                headers={"WWW-Authenticate": f'Bearer resource_metadata="https://mcp.example.com/.well-known/oauth-protected-resource"'})
        token = auth.removeprefix("Bearer ")
        claims = self.jwks.verify(token, audience=self.audience)  # RFC 8707 enforcement
        request.state.user = claims["sub"]
        return await call_next(request)

app.add_middleware(BearerAuthMiddleware, audience="https://mcp.example.com", jwks_client=jwks)
```

**Go — use `auth` / `oauthex` packages from the SDK:**

```go
import "github.com/modelcontextprotocol/go-sdk/auth"

verifier := auth.NewJWTVerifier(auth.JWTVerifierConfig{
    JWKSURL:  "https://auth.example.com/.well-known/jwks.json",
    Issuer:   "https://auth.example.com",
    Audience: "https://mcp.example.com",          // RFC 8707 audience binding
})

mux.Handle("/mcp", auth.Require(verifier, handler))   // wraps the StreamableHTTPHandler
mux.Handle("/.well-known/oauth-protected-resource", auth.ProtectedResourceMetadata(metadata))
```

Don't hand-roll token parsing. Use the SDK's `auth`/`oauthex` packages (Go) or a hardened JWT lib like `PyJWT` with `audience=` enforced (Python) — the audience check is the load-bearing step.

---

## 9. Testing — in-process + MCP Inspector

**Python — in-process via `mcp.shared.memory`:**

```python
import pytest
from mcp.shared.memory import create_connected_server_and_client_session
from server import mcp

@pytest.mark.asyncio
async def test_search_products():
    async with create_connected_server_and_client_session(mcp._mcp_server) as (client, _):
        result = await client.call_tool("search_products", {"query": "widget", "limit": 5})
        assert not result.isError
        assert len(result.structuredContent) >= 1
```

**Go — in-process via `mcp.NewInMemoryTransports`:**

```go
func TestSearchProducts(t *testing.T) {
    clientT, serverT := mcp.NewInMemoryTransports()

    server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "test"}, nil)
    register(server)
    go func() { _ = server.Run(t.Context(), serverT) }()

    client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil)
    sess, err := client.Connect(t.Context(), clientT, nil)
    require.NoError(t, err)
    t.Cleanup(func() { _ = sess.Close() })

    res, err := sess.CallTool(t.Context(), &mcp.CallToolParams{
        Name:      "search_products",
        Arguments: map[string]any{"query": "widget", "limit": 5},
    })
    require.NoError(t, err)
    require.False(t, res.IsError)
}
```

**MCP Inspector — interactive:**

```bash
# Visual UI on http://localhost:6274
npx @modelcontextprotocol/inspector node ./build/index.js
npx @modelcontextprotocol/inspector uvx acme-mcp-server
npx @modelcontextprotocol/inspector --url http://localhost:8000/mcp  # Streamable HTTP
```

**MCP Inspector — CLI mode (CI-friendly):**

```bash
npx @modelcontextprotocol/inspector --cli uvx acme-mcp-server --method tools/list
npx @modelcontextprotocol/inspector --cli uvx acme-mcp-server \
    --method tools/call --tool-name search_products --tool-arg query=widget
```

Pin Inspector to `≥0.10` in your `package.json` devDeps to dodge CVE-2025-49596.

---

## 10. Client quick reference

You'll rarely build a client from scratch — but useful for integration tests, agents, or thin wrappers.

**Python:**

```python
from mcp.client.streamable_http import streamablehttp_client
from mcp import ClientSession

async with streamablehttp_client("https://mcp.example.com/mcp",
                                  headers={"Authorization": f"Bearer {token}"}) as (read, write, _):
    async with ClientSession(read, write) as session:
        await session.initialize()
        tools = await session.list_tools()
        result = await session.call_tool("search_products", {"query": "widget"})
```

**Go:**

```go
transport := mcp.NewStreamableClientTransport("https://mcp.example.com/mcp",
    &mcp.StreamableClientTransportOptions{
        HTTPClient: oauthClient,   // *http.Client carrying Bearer token (RFC 8707 audience)
    })

client := mcp.NewClient(&mcp.Implementation{Name: "agent", Version: "1.0.0"}, nil)
sess, err := client.Connect(ctx, transport, nil)
// ... sess.ListTools, sess.CallTool, sess.ReadResource ...
```

Client responsibilities the spec puts on you:

- Include `Mcp-Session-Id` header on every request after `initialize` (SDKs handle this).
- Include `resource=<canonical URI>` in OAuth `/authorize` and `/token` requests (RFC 8707).
- Send `notifications/initialized` after receiving the `initialize` response — server-initiated requests may not flow before this.

---

## 11. SSRF guard for URL-taking tools

Any tool that takes a URL or fetches an LLM-derived host needs this. Drop into your fetch path.

```python
import ipaddress, socket
from urllib.parse import urlparse

BLOCKED_NETS = [
    ipaddress.ip_network(n) for n in (
        "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16",
        "127.0.0.0/8", "169.254.0.0/16",         # loopback, link-local (IMDS)
        "::1/128", "fc00::/7", "fe80::/10",
    )
]

def assert_safe_url(url: str, *, allowed_hosts: set[str] | None = None) -> None:
    parsed = urlparse(url)
    if parsed.scheme not in {"http", "https"}:
        raise ValueError(f"scheme not allowed: {parsed.scheme}")
    host = parsed.hostname or ""
    if allowed_hosts is not None and host not in allowed_hosts:
        raise ValueError(f"host not in allowlist: {host}")
    for family in (socket.AF_INET, socket.AF_INET6):
        try:
            infos = socket.getaddrinfo(host, None, family)
        except socket.gaierror:
            continue
        for *_, sockaddr in infos:
            addr = ipaddress.ip_address(sockaddr[0])
            if any(addr in net for net in BLOCKED_NETS):
                raise ValueError(f"host resolves to blocked range: {addr}")
```

Same shape in Go: parse with `net/url`, resolve with `net.LookupIP`, check each result against `net.IPNet` instances built from the same CIDRs above. Use a dedicated `*http.Client` with `Transport.DialContext` that re-checks at dial time (defense against DNS rebinding).
