# mcpkit

`github.com/ralvarezdev/mcpkit` — the helpers an MCP server built on the official `modelcontextprotocol/go-sdk` needs and the SDK does not provide. Extracted from finance-platform's MCP server, with the tool-result and auth helpers taken from the plc-platform and uns-platform gateways.

## What it gives you

| API | Use |
|---|---|
| `RecoverMiddleware(next)` | turns a handler panic into an error response instead of crashing the process |
| `TokenCheck` / `BackendVerifier(check, extraKey)` | `auth.TokenVerifier` that validates a bearer token against your backend (for example `GET /auth/me`), wraps failures with `auth.ErrInvalidToken`, and stores the verified token in `TokenInfo.Extra[extraKey]` |
| `CallerToken(ctx, extraKey, fallback)` | the verified caller's token from ctx, else a locally configured fallback |
| `BearerToken(req)` | the bearer token from the tool call's HTTP `Authorization` header, so a capability token is never a JSON argument the model can see |
| `RequireAPIKey(key, next)` | `http.Handler` gate for the MCP endpoint: 401 with `WWW-Authenticate: Bearer` unless the key matches (constant-time); an empty key is `ErrEmptyAPIKey` |
| `ToolOutput[T]`, `ToolStatus`, `ToolSuccess`, `ToolError` | one status/message/errors envelope for every tool; `ToolError` sets `IsError`; `T`'s `String()` must cope with its zero value |
| `ErrorMsg{Public, Debug}` / `DetailError` | tool error text that is safe by default: public mode gives `Public` (or a `DetailError`'s own message), debug mode adds the cause |
| `Callers[V]` / `NewCallers[V]()` | concurrency-safe per-caller state keyed by the verified `TokenInfo.UserID`; `Get(ctx)`, `Set(ctx, v)` |

## Recipe

```go
server := mcp.NewServer(&mcp.Implementation{Name: "my-mcp", Version: "1.0.0"}, nil)
server.AddReceivingMiddleware(mcpkit.RecoverMiddleware)

verifier := mcpkit.BackendVerifier(func(ctx context.Context, token string) (string, error) {
    user, err := backend.Me(ctx, token)
    return user.ID, err
}, "backend_token")
protected := auth.RequireBearerToken(verifier, nil)(mcpHandler)

// In a tool handler: call the backend as the caller, keep per-caller state.
token := mcpkit.CallerToken(ctx, "backend_token", cfg.LocalToken)
activeProfile := mcpkit.NewCallers[string]() // shared; Get/Set(ctx) per request
```

## Rules

- **Never put a credential in a tool argument.** The model can read arguments; take the token from the request header (`BearerToken`) or the verified caller.
- **Do not hand raw errors to an agent.** Wrap tool errors in `ErrorMsg` so SQL, paths and driver text stay out of the tool result unless debug mode is on. A client error type that is fit for an agent implements `DetailError`.
- **Delegate auth to the backend.** The MCP server holds no user store: `BackendVerifier` asks the API who the token belongs to, so revocation there applies here.
- **An unauthenticated `/mcp` is a decision, not a default.** Use `RequireAPIKey` (or `BackendVerifier`) unless the endpoint is deliberately open, and log that when it is.
- **Backend calls use hand-written `restkit/client`**, not a generated client.

## Not included

`JSONResult` (the go-sdk fills tool `Content` from the output value when a handler returns a nil result) and pointer helpers (Go 1.26 `new(expr)`).

Stays `v0.x`: it is extracted from few consumers, and the API is expected to move until a second MCP server adopts it. See [../STACK.md](../STACK.md).
