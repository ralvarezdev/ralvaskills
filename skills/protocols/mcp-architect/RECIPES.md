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
12. [Schema-driven tool contracts (Go)](#12-schema-driven-tool-contracts-go)

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
- **In Go, hand-write the schema rather than deriving it from the output type.** The SDK will happily infer one from `Out`, but that schema describes the Go struct — which may not match the wire. See [§12](#12-schema-driven-tool-contracts-go) for the three cases where they diverge (money-as-string aliases, re-wrapped envelopes, `omitempty` fields).

---

## 12. Schema-driven tool contracts (Go)

The Go SDK derives `inputSchema` from the `In` struct by reflection and the structured-output
schema from `Out`. That is the right default for a handful of tools. Past ~10, hand-written
structs drift from each other and from the API they wrap, and nothing catches it. This is the
pipeline that fixes that: JSON Schema as the source, Go types generated, tool bodies
hand-written.

Worked example: `finance-platform`, 42 tools — arguments at `158fc32`, outputs at `eadd08c`,
shared enums at `7b7f21d`.

### Decide first: authored or derived

**Author the schemas when the tool surface is not the API surface.** An MCP tool is a curated view
for a model, not an endpoint. In finance, `profile_id` does not exist in the client types at all —
it lives in the URL path and the tool resolves it from the caller's session. `list_movements`
omits the API's `sort` and `future`. A tool that mirrors its endpoint exactly has nothing to
curate.

**Derive them when the tool is a thin pass-through** — same fields, same names, nothing added or
dropped. Then generation is right and you should not hand-maintain a parallel spec.

**Do not share MCP enums with API enums.** An MCP enum is the set a tool *accepts*; an API enum is
the domain's. `create_movement` takes 2 of the 9 movement types the API knows. One schema listing
nine values of which six error is worse than useless — the model reads it as available. Author
both and let them differ.

The failure mode to watch: reaching for the API's enum to avoid duplication, then shipping a tool
whose schema advertises nine values it cannot use.

### Layout

Keep schemas at the repo root beside the OpenAPI tree, with generated Go under the module:

```
openapi/contexts/*.yaml    → backend/internal/api/handler/*.gen.go
json/mcpserver/*/*.json    → backend/internal/mcpserver/schemas/*.gen.go
```

Putting `json/` inside the module inverts the repo's own convention — no other schema tree lives
there.

### Arguments

`github.com/atombender/go-jsonschema` generates the Go struct. These flags are load-bearing:

```
go run github.com/atombender/go-jsonschema@latest \
  -p movement -o list.gen.go \
  --disable-omitzero \                  # drop the omitzero tag
  --capitalization ID \                 # profile_id → ProfileID, not ProfileId
  --struct-name-from-title \            # title → struct name, not the filename
  --tags json                           # drop yaml/mapstructure
```

Add `--only-models` only when nothing deserializes these types outside the SDK — see below.

- **`--struct-name-from-title` is not optional.** Without it the type is named after the *file*:
  `list.json` gives `ListJson`, `list_arguments.json` gives `ListArgumentsJson`. The schema's
  `title` is what names the type, so every schema needs one.
- The generated file says `DO NOT EDIT`. Commit it (same as the OpenAPI's generated types) so a
  plain `go build` works without the generator, and add a check task that regenerates and diffs.

#### `--only-models`: decide by who deserializes the struct

The default emits a custom `UnmarshalJSON` per enum. It is **not** redundant — it is the only
validation these types get when unmarshaled outside the SDK. It enforces `required`, `minimum`,
and the enum set:

```go
func (j *Req) UnmarshalJSON(value []byte) error {
    if _, ok := raw["name"]; raw != nil && !ok {
        return fmt.Errorf("field name in Req: required")
    }
    ...
    if plain.N != nil && 5 > *plain.N {
        return fmt.Errorf("field %s: must be >= %v", "n", 5)
    }
```

`--only-models` removes all of it: a `n: 3` against `minimum: 5` then deserializes cleanly. The
choice is not "duplicate types or clean types":

| Who unmarshals the struct | Flag | Why |
|---|---|---|
| Only `mcp.AddTool` | `--only-models` | The SDK already validates the input against the `inputSchema` it derived from this same schema. The rest is dead code. |
| Also a CLI, a test, a config loader | omit it | There it is the only validation, and dropping it silently accepts bad data. |

Ask who consumes the types before choosing — it is not knowable from the tool code.

The flag also removes a duplicate type. Without it, a shared enum referenced from a consumer file
yields a *third* type named after that consumer's filename (`usea.json` → `UseADialect`) carrying
the same values as the shared one. If you cannot use `--only-models`, give each enum its own file
so the derived name cannot collide.

### Outputs

Do **not** generate Go for outputs. Declare the schema and hand it to the SDK:

```go
mcp.AddTool(server, &mcp.Tool{
    Name:         "get_movement",
    OutputSchema: outputs["MovementOutputSchema"],
}, handler)
```

The SDK validates the result against it before it reaches the wire (`mcp/server.go`, "validating
tool output"), so a wrong shape fails at the call instead of confusing the model.

**Describe the wire, not the Go type.** Three cases where they differ, each of which had to be
fixed in the reference implementation:

- **Money as string.** `Decimal` was `type Decimal = string` — a deliberate alias so no float
  rounding enters. A `"type": "number"` schema makes the server reject its own output.
- **The re-wrapped envelope.** `list_movements` returns `movements`, not the API envelope's `data`,
  because the HTTP client re-wraps after decoding.
- **`omitempty` fields are optional.** The paging meta serializes with `omitempty`, so a first
  empty page omits every counter. Declaring them `required` rejects the server's own empty
  response.

Read the actual type and its JSON tags; do not assume the Go field name is the wire key
(`per_page` is `perPage`).

### Three things that will bite you

**1. The SDK resolves `OutputSchema` with no loader.** A `$ref` to another document panics at
`AddTool` — this is startup, not a validation error:

```go
panic: AddTool "list_movements": output schema: loading ./x.json:
  cannot resolve remote schemas: no loader passed to Schema.Resolve
```

Inline every `$ref` at generation time; the resolver then only ever sees one self-contained
document.

**2. `go-jsonschema` does not follow `$ref` at all.** A referenced property silently becomes
`interface{}` — the type and enum gone, with no error:

```go
Type interface{} `json:"type"`   // was {"type":"string","enum":["income","expense"]}
```

Same fix, different reason. Both paths should share one inliner so there is a single definition
of "resolved".

**3. `go-jsonschema` does not descend into `allOf`.** Relevant only when you use `allOf` to
*override* a shared definition's constraint — flatten the single-element wrapper before
generating, or the property degrades to `interface{}` like any unresolved ref.

Note what is **not** a trap: a `$ref` with a sibling `description` needs no wrapper. In draft
2020-12 `go-jsonschema` reads it directly and keeps the comment on the generated field.

### Sharing

Three strategies, chosen by how many tools share the vocabulary and whether a generator sits
between the schema and the code:

1. **One shared `$defs` file** — a single `shared/enums.json` every tool `$ref`s. Best when several
   tools share the vocabulary and JSON Schema is the source of truth. Detailed below.
2. **One file per enum** — spread each enum into its own file. The workaround when you cannot use
   `--only-models`: a shared enum referenced from a consumer file otherwise yields a *third* type
   named after that consumer (`usea.json` → `UseADialect`) carrying the same values.
3. **No shared schema files** — define each tool's input as a Go struct with `json` tags and let the
   SDK derive the schema (work-hour-reports). No `$ref`, no resolved copies, no generator. The
   accepted set lives as typed Go constants checked in a `switch`; the model learns the values only
   if the tool `Description` spells them out, so keep them there.

The rest of this section details strategy 1.

Put reusable definitions in one file and reference them:

```
json/mcpserver/shared/enums.json
  $defs: movement_type_input, movement_type_output, currency, decimal, date

json/mcpserver/movement/create_arguments.json
  "type": { "$ref": "../shared/enums.json#/$defs/movement_type_input" }
```

Commit both the sources (with refs) and the resolved copies — the sources so a reader sees the
sharing, the resolved ones because that is what the generators read. Resolve relative refs against
the containing file's directory per RFC 3986, not the schema root.

Every `$defs` entry needs a `description`: it is inlined into every schema that references it, so
it is the only text the model reads about those values.

On nullability: `anyOf: [{$ref}, {type: null}]` degrades to `interface{}` like any other
indirection. Keep optional properties a plain `$ref` and let `null` live in the definition when the
field is genuinely nullable — then cover it with a test, because an explicit `null` against an enum
that does not list `null` is exactly what your optional fields will hit.

### Verify

Test both directions, and prove the rejection tests are not vacuous — mutate the schema and watch
the test fail:

```go
func TestOutputSchemaRejectsNumericAmount(t *testing.T) {
    // amount: "42.50" passes; amount: 42.5 must be refused
    // flipping the schema's amount to "number" makes this test fail
}
```

Also assert the schema is an object with `additionalProperties: false`, that every property has a
description, and that a cross-file `$ref` was inlined rather than left dangling. A schema linter
over the source tree catches malformed JSON before generation does — go-jsonschema is lenient and
will happily emit Go from a broken schema.

Expect existing tests to need conforming fixtures. A fake response missing a field the schema
requires is exactly the bug the schema exists to catch.

---



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
        raise FileNotFoundError(sku)   # → JSON-RPC -32602 Resource not found (moved off -32002 in spec 2026-07-28)
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

mcp = FastMCP("acme-shop")  # spec 2026-07-28: always stateless — no session flag to set
# ... register tools / resources / prompts ...

# FastMCP exposes a Starlette app at .streamable_http_app()
app = mcp.streamable_http_app()  # mounts /mcp endpoint (POST only — no session GET/DELETE)

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
```

Every request is independent and can be routed to any replica by a plain round-robin load balancer — no sticky sessions, no shared session store. `Mcp-Method` and `Mcp-Name` request headers and `MCP-Protocol-Version` are handled by the SDK; you don't set them by hand server-side.

Need state across calls? Use the explicit-handle pattern ([SKILL.md §8](SKILL.md#8-request-model--the-stateless-core)): a tool returns an ID, the model passes it back as an argument on later calls. For "push something to the client mid-call" (confirmation, missing param), return an `InputRequiredResult` instead of completing — see [SKILL.md §8](SKILL.md#8-request-model--the-stateless-core) for the MRTR shape. Resource subscriptions register via `resources/subscribe` as before, but updates now arrive over `subscriptions/listen` rather than a held-open per-connection stream.

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
    // Spec 2026-07-28: no handshake, no session — the handler factory is
    // invoked per request. Client identity/capabilities arrive via _meta on
    // each call, not a stored session; read them off the request if you need them.
    handler := mcp.NewStreamableHTTPHandler(
        func(r *http.Request) *mcp.Server {
            s := mcp.NewServer(&mcp.Implementation{
                Name:    "acme-shop",
                Version: "1.0.0",
            }, nil)
            register(s)   // see §2
            return s
        },
        nil, // no stateful/stateless toggle to set — every deployment is stateless now
    )

    mux := http.NewServeMux()
    mux.Handle("/mcp", handler) // POST only; Mcp-Method/Mcp-Name/MCP-Protocol-Version handled by the SDK

    srv := &http.Server{Addr: ":8000", Handler: mux}
    slog.Info("listening", "addr", srv.Addr)
    _ = srv.ListenAndServe()
}
```

Don't pre-build a single `mcp.Server` and reuse it across requests — build fresh per invocation (or keep it fully stateless internally) since nothing pins a client to a particular instance anymore. Need cross-call state? Return a handle from a tool and have the model pass it back — see [SKILL.md §8](SKILL.md#8-request-model--the-stateless-core).

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

**Client onboarding (2026-07-28):** prefer Client ID Metadata Documents (CIMD) over Dynamic Client Registration — DCR (RFC 7591) is deprecated but still functions during its grace window. If you still register clients via DCR, require `application_type` on registration so `localhost` redirect URIs from desktop/CLI clients aren't rejected. On the authorization-code exchange, validate the `iss` parameter (RFC 9207) against the authorization server you initiated the flow with before redeeming the code — this closes an AS-mix-up hole where a code minted by one AS gets replayed against another.

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
        # No session.initialize() handshake — spec 2026-07-28 is stateless.
        # Client identity/capabilities ride in _meta on every call; the SDK sets this.
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

Client responsibilities the spec puts on you (2026-07-28):

- Send `Mcp-Method` and `Mcp-Name` headers plus `MCP-Protocol-Version` on every Streamable HTTP request (SDKs handle this — don't hand-roll requests without them).
- Include `resource=<canonical URI>` in OAuth `/authorize` and `/token` requests (RFC 8707), and validate the `iss` parameter on the authorization response before redeeming a code (RFC 9207).
- If a call returns `InputRequiredResult`, resolve every entry in `inputRequests` and re-issue the *same* call with `inputResponses` plus the echoed `requestState` — don't start a new call.
- There is no `initialize`/`notifications/initialized` exchange and no `Mcp-Session-Id` to carry — a client written against 2025-11-25 will not interoperate with a 2026-07-28 server without an SDK upgrade.

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
