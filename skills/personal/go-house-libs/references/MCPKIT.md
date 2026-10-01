# mcpkit

`github.com/ralvarezdev/mcpkit` — the helpers an MCP server built on the official `modelcontextprotocol/go-sdk` needs and the SDK does not provide. Extracted from one MCP server; it stays v0.x until a second server adopts it. Current tag v0.4.1 (go-sdk v1.8.0, Go 1.27.1).

## What it gives you

| API | Use |
|---|---|
| `RecoverMiddleware(next)` | turns a handler panic into an error response instead of crashing the process; logs via stdlib `log` |
| `RecoverWith(RecoverConfig{Logger})` | same, as an `mcp.Middleware` that logs the method, panic value and `debug.Stack()` to a `*slog.Logger` (nil means `slog.Default()`) at error level; the caller gets only a generic `internal error handling <method>`, never the stack |
| `TokenCheck` / `BackendVerifier(check, extraKey)` | `auth.TokenVerifier` that validates a bearer token against your backend (for example `GET /auth/me`), wraps failures with `auth.ErrInvalidToken`, and stores the verified token in `TokenInfo.Extra[extraKey]` |
| `CallerToken(ctx, extraKey, fallback)` | the verified caller's token from ctx, else a locally configured fallback |
| `ErrorMsg{Public, Debug}` / `DetailError` | tool error text that is safe by default: public mode gives `Public` (or a `DetailError`'s own message), debug mode adds the cause |
| `Callers[V]` / `NewCallers[V]()` | concurrency-safe per-caller state keyed by the verified `TokenInfo.UserID`; `Get(ctx)`, `Set(ctx, v)`, `Delete(ctx)` (call on logout or session end so state does not leak; no-op if unset) |

## Recipe

```go
server := mcp.NewServer(&mcp.Implementation{Name: "my-mcp", Version: "1.0.0"}, nil)
server.AddReceivingMiddleware(mcpkit.RecoverWith(mcpkit.RecoverConfig{Logger: slog.Default()}))

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

- **Never put a credential in a tool argument.** The model can read arguments; take the token from the verified caller (`CallerToken`), not from a tool argument.
- **Do not hand raw errors to an agent.** Wrap tool errors in `ErrorMsg` so SQL, paths and driver text stay out of the tool result unless debug mode is on. A client error type that is fit for an agent implements `DetailError`.
- **Delegate auth to the backend.** The MCP server holds no user store: `BackendVerifier` asks the API who the token belongs to, so revocation there applies here.
- **An unauthenticated `/mcp` is a decision, not a default.** Gate it with `BackendVerifier` via `auth.RequireBearerToken` unless the endpoint is deliberately open, and log that when it is.
- **Backend calls use hand-written `restkit/client`**, not a generated client.

## Removed in v0.4.0 (breaking)

`ToolOutput[T]`, `ToolStatus`, `ToolError`, `ToolSuccess`, `BearerToken`, `RequireAPIKey` and `ErrEmptyAPIKey` are gone: no consumer used them (servers render results natively and verify with `BackendVerifier`). They live on in v0.2.0 and in some projects' local copies, and return only when a second server adopts them. Do not recommend them; return plain tool results and let the go-sdk fill `Content`. The repo README still lists them; the code is the truth.

## Not included

`JSONResult` (the go-sdk fills tool `Content` from the output value when a handler returns a nil result), pointer helpers (Go 1.26 `new(expr)`), and the tool-output envelope and API-key gate above.

Stays `v0.x`: one consumer, so the API may move until a second MCP server adopts it. Extend it only when a second consumer imports it. See [../STACK.md](../STACK.md).
