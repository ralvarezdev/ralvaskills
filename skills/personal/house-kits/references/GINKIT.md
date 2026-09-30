# ginkit

`github.com/ralvarezdev/ginkit` — the gin glue for restkit: responders, validation, request-id, logger, recovery, body cap, ETag replies, idempotency, CORS and rate limiting. Its own module so restkit stays free of gin.

## What it gives you

| Area | API |
|---|---|
| Responders | `OK`, `OKList(c, data, meta)`, `OKMessage`, `Err(c, status, msg)`, `Abort` (same body, for middleware, no log), `Problem`, `AbortProblem` |
| Validation | `RegisterValidators()` (once, before routes: JSON-tag names, `iso_date`, `gte_date=Field`), `Validation(c, err)`, `ValidationProblem`/`ValidationProblemList`, `FormatValidationErrors` |
| Request identity | `RequestID()` / `CorrelationID(c)` (`X-Request-Id`, shared with `restkit/httpx`) |
| Observability | `Logger(cfg)` (one slog record per request, never the query string; `UserID`, `SkipPaths`, `SkipPrefixes`), `Recovery(cfg)`, `RecoveryWith(cfg, respond)` |
| Body cap | `MaxBodySize(n)`, `MaxBodySizeFunc(limit)` (per request, by `c.FullPath()`), `IsBodyTooLarge(err)` |
| Caching | `NotModified(c, tag, cacheControl)` sets `ETag` and answers 304 for a matching `If-None-Match`; build the tag with `restkit/etag` |
| Idempotency | `Idempotency(IdempotencyConfig[C])` over any `restkit/idempotency.Store` (`pgkit/idempotency`) |
| Browser | `CORS(cfg)` and `RequireOrigin(origins)` (CSRF-style gate for cookie-authenticated APIs) |
| Throttling | `RateLimit(limiter, key)` over a `ratelimit.Limiter` |
| Tests | `ginkit/testkit`: `NewRequest`, `WithBearer`, `WithHeader`, `Serve`, `DecodeJSON[T]`, `DecodeProblem` |

## Recipe

```go
ginkit.RegisterValidators() // before routes

r.Use(ginkit.RequestID(), ginkit.Logger(ginkit.LoggerConfig{ /* Logger, UserID, SkipPaths */ }))
r.Use(ginkit.RecoveryWith(httpx.Config{Logger: logger}, func(c *gin.Context) {
    ginkit.Abort(c, http.StatusInternalServerError, "internal server error")
}))

cors, err := ginkit.CORS(ginkit.CORSConfig{AllowOrigins: origins, AllowCredentials: true})
if err != nil { return err } // startup error, never "allow everyone"

api.Use(ginkit.RateLimit(limiter, func(c *gin.Context) (string, bool) {
    claims, ok := identitygin.Claims(c)
    if !ok { ginkit.Abort(c, http.StatusUnauthorized, "unauthorized"); return "", false }
    return claims.UserID, true
}))
```

Finance's request pipeline (`internal/api/router/pipeline.go`) is `RequestID`, `Logger`, `RecoveryWith`, then the body cap.

## Rules

- **`Err` logs and returns the same message at 5xx.** Never pass `err.Error()` for a 500: it ships SQL and constraint names to the client. Log the real error, answer with a generic message, and show the raw text only when `gin.IsDebugging()`. Finance does this in `respondInternalError` (`internal/api/handler/internal_error.go`); a problem-details service uses `problem.RedactServerError`.
- **Order matters.** `Idempotency` runs after `RequestID` and after auth (`Owner` returning `ok=false` is a 401). `CORS` goes before any middleware that sets `Vary`.
- **Rate limit contract.** When `key` returns `ok=false` it must already have responded (use `Abort`). A denied request gets 429 with `Retry-After` rounded up to whole seconds. A limiter error lets the request through and logs a warning: a limiter outage must not take the API down.
- **Idempotency fails closed.** A store outage answers 503 without running the handler; 5xx and panics release the key; replays carry `Idempotency-Replayed: true`; 409 while in flight; 422 for a reused key with a different request.
- **CORS is validated at startup.** No origins, a wildcard with credentials, a wildcard mixed with origins, or a malformed origin is an error. Set the allowed origins in every deployed environment, because an empty list refuses to start.
- **Bind failures.** `Validation` answers 413, not 400, when the bind failed on a `MaxBodySize` cap.
- **`RecoveryWith`, not `Recovery`, for an envelope API**: `Recovery` writes a problem; `RecoveryWith` lets `respond` write your own 5xx body.

## Not here

Response shape and problem types (restkit), Postgres idempotency storage (pgkit), authentication (identitygin), throttling algorithms (ratelimit).
