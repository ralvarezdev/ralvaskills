# restkit

`github.com/ralvarezdev/restkit` — the framework-free half of the house REST conventions. Standard library only, so a gin service, a `net/http` service, a CLI client and an MCP server share one wire contract.

## Packages

| Package | Owns |
|---|---|
| `restkit` | `Response[T]` envelope, `Meta` (`NewCursorMeta`, `NewOffsetMeta`), keyset cursor (`EncodeCursor`/`DecodeCursor`, and the `…Sep` forms for an existing separator), page-param parsers (`PageSize`, `PageNumber`, and the `…Strict` forms that return an error) |
| `restkit/problem` | RFC 9457 problem details: `Problem`, `FromStatus`, typed `Catalog` of per-status type URIs, `WithErrors`/`WithFieldErrors`, `WithCorrelationID`, `RedactServerError` |
| `restkit/httpx` | `net/http` adapter: `WriteProblem`, `RequestID` middleware (validates inbound ids), `Config.Recovery`, `Config.WriteError`, `RequireBearer(secret)` static-secret gate, `Health`/`Readiness`/`ReadinessWithTimeout` probes, `QueryPageSize`/`QueryPageNumber` (+`Strict`) readers |
| `restkit/etag` | weak ETags: `Encode`/`Decode` from a last-modified time, `Weak(parts...)` hash of what determines a response, `Match` for `If-None-Match`/`If-Match` |
| `restkit/client` | HTTP client core for a CLI or MCP server: `Do[T]`, `Pages[T]` (cursor iterator), bearer auth, one `*APIError` (with `DebugString()`, satisfies `mcpkit.DetailError`) |
| `restkit/idempotency` | the `Store[C]` contract, `Record`, `Status` between `Idempotency-Key` middleware (ginkit) and a store (pgkit) |

## Reach for it when

- writing a response body, list metadata or a cursor token: use the envelope and the codec, never hand-roll them;
- returning an error from a `net/http` service: `problem` + `httpx`;
- building a typed API client: `client.Do` / `client.Pages`, with endpoint methods hand-written on top;
- parsing `limit`/`page` query params: `httpx.QueryPageSize`/`QueryPageNumber` (lenient: fall back to the default) or the `…Strict` forms (error on malformed, sub-1 or over-max);
- gating an internal endpoint behind a shared secret, or exposing probes, on `net/http`: `httpx.RequireBearer`, `httpx.Health`, `httpx.Readiness`;
- deciding what an ETag covers: `etag.Weak(parts...)` over the inputs of the response.

## Recipe

```go
// Server: the adapter writes the envelope.
resp := restkit.Response[[]Loan]{Status: 200, Data: loans, Meta: restkit.NewCursorMeta(next)}

// Client: hand-written endpoints on Do / Pages.
c, err := client.New(client.Config{
    BaseURL:  "https://api.example.com",
    BasePath: "/api",
    Token:    func(ctx context.Context) string { return tokenFor(ctx) },
})
loan, _, err := client.Do[Loan](ctx, c, http.MethodGet, "/loans/"+id, nil, nil)
for page, err := range client.Pages[Loan](ctx, c, "/loans", url.Values{"limit": {"50"}}) { ... }

// Problem details.
p := problem.FromStatus(http.StatusBadRequest, "validation failed").WithErrors(map[string]string{"email": "required"})
```

net/http probes and a service-to-service gate:

```go
mux.Handle("GET /healthz", httpx.Health()) // liveness, always 200 "ok"
mux.Handle("GET /readyz", httpx.Readiness(map[string]httpx.Check{
    "db": func(ctx context.Context) error { return pool.Ping(ctx) },
})) // 200 {"status":"ok"} or 503 {"status":"unavailable","failed":{"db":"..."}}
mux.Handle("/internal/", httpx.RequireBearer(secret)(internal))

size, err := httpx.QueryPageSizeStrict(r, "limit", 50, 200) // err on "abc", 0, 500
```

Wire shape:

```json
{"data": {...}, "meta": {"nextCursor": "...", "hasMore": true}, "status": 200}
{"data": null, "errors": {"email": "required"}, "message": "validation failed", "status": 400}
```

## Rules

- **`Meta` is camelCase, bodies are snake_case.** Deliberate, and pinned byte-for-byte by `response_test.go`; frontend, CLI and MCP depend on it.
- **One error type on the client.** Every non-2xx becomes `*client.APIError` (`Status`, `Message`, `Fields`, `Problem` when the server sent problem details). Do not add a second error path.
- **Cursors are opaque.** A service that already issues cursors with another separator keeps it via `EncodeCursorSep`/`DecodeCursorSep`; switching separators invalidates every cursor in flight.
- **5xx detail is redacted.** Use `problem.RedactServerError` or `httpx.Config.WriteError`, so a 5xx never carries driver or SQL text outside debug mode.
- **`PanicWriter` is for envelope APIs.** A service whose errors are not problem details replaces the recovered-panic 500 through `Config.PanicWriter`; `http.ErrAbortHandler` is always re-panicked.
- **`Meta` ints are `omitempty`.** A zero `total`, `page`, `perPage`, `totalPages` or `offset` is absent from the JSON; clients must read a missing number as 0.
- **`NewOffsetMeta` is overflow-safe.** Never negative `Offset` (page below 1 gives 0), zero `TotalPages` for non-positive `perPage` or negative total, `Offset` saturates at `math.MaxInt`.
- **Lenient vs strict parsers.** `PageSize`/`PageNumber` fall back to the default on bad input and `PageSize` clamps over `maxSize` (`maxSize <= 0` means no cap); the `…Strict` forms return the default only for an empty value and error otherwise. Pick strict when a bad value should be a 400.
- **`Readiness` runs checks in name order**, each under its own timeout (5s default; `ReadinessWithTimeout(d, checks)`, `d <= 0` uses the request context only). Failure messages go in the body, so return sanitized errors. `Health` runs no checks.
- **`RequireBearer` is a shared-secret gate, not user auth.** Constant-time over digests, an empty secret fails closed, scheme matched case-insensitively (v0.10.1) with any whitespace before the token, only the first `Authorization` header is read; rejects with 401 problem+json and `WWW-Authenticate: Bearer`. Same behavior as `ginkit.RequireBearer`.
- **Unknown problem members are kept** on decode (RFC 9457 requires consumers to ignore extensions they do not know).

## Not here

Framework glue (gin: `ginkit`), Postgres storage of idempotency keys (`pgkit/idempotency`), generated API clients (oapi-codegen produces types only; hand-write the calls).

Pre-1.0: the API may move until a second project adopts it. See [../STACK.md](../STACK.md) for the tag.
