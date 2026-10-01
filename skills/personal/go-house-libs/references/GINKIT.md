# ginkit

`github.com/ralvarezdev/ginkit` — the gin glue for restkit: responders, validation, request-id, logger, recovery, body cap, ETag replies, idempotency, CORS, rate limiting, health probes, page params and a bearer gate. Its own module so restkit stays free of gin.

## What it gives you

| Area | API |
|---|---|
| Responders | `OK`, `OKList(c, data, meta)`, `OKMessage`, `Err(c, status, msg)` (5xx log carries `request_id`), `Abort` (same body, for middleware, no log), `Problem` (zero status becomes 500), `AbortProblem` |
| Validation | `RegisterValidators()` (once, before routes: wire names from `json`, then `form`, then `uri` tag, else snake_case; `iso_date`, `gte_date=Field`), `Validation(c, err)` (400), `ValidationStatus(c, err, status)` (e.g. 422), `ValidationProblem`/`ValidationProblemList` (RFC 6901 pointers, sorted), `FormatValidationErrors`, `MaxValidationErrors` (100, cap on reported failures) |
| Page params | `QueryPageSize(c, name, def, max)`, `QueryPageNumber(c, name, def)` (lenient: fall back to `def`); `QueryPageSizeStrict`, `QueryPageNumberStrict` return an error on malformed or out-of-range input (wraps `restkit.PageSizeStrict`/`PageNumberStrict`) |
| Probes | `Health()` (200 `ok`, no checks), `Readiness(map[string]Check)` (503 with failures), `ReadinessWithTimeout(d, checks)`; `Check` is `httpx.Check`, so net/http checks are reusable |
| Request identity | `RequestID()` / `CorrelationID(c)` (`X-Request-Id`, shared with `restkit/httpx`) |
| Observability | `Logger(cfg)` (one slog record per request, never the query string; `UserID`, `SkipPaths`, `SkipPrefixes`), `Recovery(cfg)`, `RecoveryWith(cfg, respond)` |
| Body cap | `MaxBodySize(n)`, `MaxBodySizeFunc(limit)` (per request, by `c.FullPath()`), `IsBodyTooLarge(err)` |
| Caching | `NotModified(c, tag, cacheControl)` sets `ETag` and answers 304 for a matching `If-None-Match`; build the tag with `restkit/etag` |
| Idempotency | `Idempotency(IdempotencyConfig[C])` over any `restkit/idempotency.Store` (`pgkit/idempotency`) |
| Browser | `CORS(cfg)` and `RequireOrigin(origins)` (CSRF-style gate for cookie-authenticated APIs) |
| Throttling | `RateLimit(limiter, key)` over a `ratelimit.Limiter`; `RateLimitWith(limiter, key, RateLimitOptions{Headers, PolicyName, FailClosed})` |
| Service auth | `RequireBearer(secret) (gin.HandlerFunc, error)`: constant-time shared-secret gate (metrics, internal hooks) |
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

```go
// brute-force-sensitive route: reject with 503 + Retry-After when the limiter is down,
// and expose draft RateLimit / RateLimit-Policy headers
login.Use(ginkit.RateLimitWith(limiter, ipKey, ginkit.RateLimitOptions{
    FailClosed: true, Headers: true, PolicyName: "login",
}))

// shared-secret gate; keep the handler even on error (it is a deny-all stand-in)
gate, err := ginkit.RequireBearer(os.Getenv("METRICS_TOKEN"))
if err != nil { return err } // ErrEmptyBearerSecret
r.GET("/metrics", gate, metricsHandler)

r.GET("/healthz", ginkit.Health())
r.GET("/readyz", ginkit.Readiness(map[string]ginkit.Check{"db": pingDB}))
```

Finance's request pipeline (`internal/api/router/pipeline.go`) is `RequestID`, `Logger`, `RecoveryWith`, then the body cap.

## Rules

- **`Err` logs and returns the same message at 5xx.** Never pass `err.Error()` for a 500: it ships SQL and constraint names to the client. Log the real error, answer with a generic message, and show the raw text only when `gin.IsDebugging()`. Finance does this in `respondInternalError` (`internal/api/handler/internal_error.go`); a problem-details service uses `problem.RedactServerError`.
- **Order matters.** `Idempotency` runs after `RequestID` and after auth (`Owner` returning `ok=false` is a 401). `CORS` goes before any middleware that sets `Vary`.
- **Rate limit contract.** When `key` returns `ok=false` it must already have responded (use `Abort`); if it did not, the middleware aborts with a 500 problem instead of letting the request through. A denied request gets 429 with `Retry-After` rounded up to whole seconds. A limiter error fails open by default (logs a warning); set `FailClosed` for login and other brute-force-sensitive routes (503 + `Retry-After: 5`).
- **Rate limit headers are opt-in** (`Headers: true`): `RateLimit` / `RateLimit-Policy` per draft-ietf-httpapi-ratelimit-headers -11 (an Internet-Draft, not an RFC), written only when the limiter reports `Result.Limit > 0`; `w=` appears only if the limiter has a `Policy()` (memory and valkey do). `PolicyName` must be printable ASCII, otherwise it falls back to `"default"`.
- **`RequireBearer` fails closed.** An empty secret returns `ErrEmptyBearerSecret` plus a handler that 401s everything; comparison is over SHA-256 digests, so neither content nor length leaks. Answers 401 problem with `WWW-Authenticate: Bearer`.
- **Two error families, do not mix**: `Err`/`Validation`/`ValidationStatus` write the restkit envelope; `Problem`/`AbortProblem`/`ValidationProblem*` write RFC 9457 problems.
- **Validation keys.** `FormatValidationErrors` keys are full wire paths (`items[0].name`, `addr.zipCode`), used verbatim, human-oriented and ambiguous if a name contains `.`, `[` or `]`; use `ValidationProblemList` for unambiguous JSON Pointers. Output is capped at `MaxValidationErrors`. `gte_date` no longer panics on a missing or non-string field (fails validation).
- **Idempotency fails closed.** A store outage answers 503 without running the handler; 5xx and panics release the key; replays carry `Idempotency-Replayed: true`; 409 while in flight; 422 for a reused key with a different request.
- **CORS is validated at startup.** No origins, a wildcard with credentials, a wildcard mixed with origins, or a malformed origin is an error. Set the allowed origins in every deployed environment, because an empty list refuses to start.
- **Bind failures.** `Validation` answers 413, not 400, when the bind failed on a `MaxBodySize` cap.
- **`RecoveryWith`, not `Recovery`, for an envelope API**: `Recovery` writes a problem; `RecoveryWith` lets `respond` write your own 5xx body.

## Not here

Response shape and problem types (restkit), Postgres idempotency storage (pgkit), authentication (identitygin), throttling algorithms (ratelimit).
