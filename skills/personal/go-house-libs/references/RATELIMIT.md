# ratelimit

`github.com/ralvarezdev/ratelimit` — throttling of how often a key (user, IP, API key) may act. A port plus token-bucket adapters; nothing in the root package knows about gin, `net/http` or a backing store.

## Packages

| Package | Owns |
|---|---|
| `ratelimit` | the `Limiter` port, `Config{Rate, Burst}`, `Config.Policy()`, `Policy{Quota, Window}`, `Result{Allowed, Remaining, RetryAfter, Limit, Reset}`; no third-party imports |
| `ratelimit/memory` | `memory.New(cfg)`: in-process token bucket behind a mutex, for tests and single-instance deployments; idle keys are evicted lazily during `Allow` (no goroutine, nothing to `Close`) |
| `ratelimit/valkey` | `valkey.New(client, cfg)` (token bucket) and `valkey.NewGCRA(client, cfg)` (GCRA): one atomic Lua script per call, for several instances sharing one Valkey; both limiters and `memory.Limiter` expose `Policy()` |
| `ratelimit/valkeytest` | `Start` / `StartClient`: a throwaway Valkey container for integration tests (`Config` zero value works; image defaults to `valkey/valkey:8-alpine`) |

## Recipe

```go
cfg := ratelimit.Config{Rate: 10, Burst: 20} // 10 req/s steady, bursts up to 20

limiter, err := memory.New(cfg)               // single instance / tests
limiter, err := valkey.NewGCRA(client, cfg)   // several instances; same behavior, cheaper than valkey.New

res, err := limiter.Allow(ctx, key)
if err != nil { /* the limiter itself failed: decide fail open or closed */ }
if !res.Allowed { /* reject; wait res.RetryAfter */ }
```

`res.Limit` is the bucket capacity and `res.Reset` the time until it is full again; zero means unknown.

For gin, do not write the middleware: `ginkit.RateLimit(limiter, key)` ([GINKIT.md](GINKIT.md)) already returns 429 with a rounded-up `Retry-After` and fails open when the limiter errors. `ginkit.RateLimitWith` adds `FailClosed` (503) and the draft `RateLimit` / `RateLimit-Policy` headers, fed by `Result.Limit`/`Reset` and `Policy()`.

## Rules

- **Prefer `valkey.NewGCRA` over `valkey.New`** unless a token bucket implementation is specifically wanted: it admits and denies identically and stores one value per key.
- **The caller owns the key.** A user id, an IP or an API key; the port is identity-agnostic. Finance keys the `/api` group by the authenticated caller: `api.Use(ginkit.RateLimit(limiter, middleware.MustCallerID))` in `internal/api/router/router.go` ([GINKIT.md](GINKIT.md)).
- **Choose fail-open or fail-closed on error, explicitly.** `Allow` returns an error only when the limiter could not evaluate (store unreachable); denial is `Allowed == false` with no error. `ginkit.RateLimit` fails open; for a security-sensitive path use `RateLimitOptions{FailClosed: true}`.
- **Token bucket, not a fixed window**: a fixed window lets a key spend twice its budget around the boundary.
- **`Policy.Window` is `Burst / Rate`** (time to refill an empty bucket), not a fixed window.
- **Custom `Limiter` implementations** that want `RateLimit` headers should fill `Result.Limit`/`Reset` and may add a `Policy() ratelimit.Policy` method.
- **Valkey adapters use the server's `TIME`,** not the caller's clock, so clock skew between instances cannot desynchronize a bucket. The Lua scripts now return a fourth element (reset seconds); the parser still accepts the old three-element reply.
- **Not login-attempt counting.** `identity/auth.AttemptLimiter` counts failed logins per account ([IDENTITY.md](IDENTITY.md)); this is API throughput.
- **Test the Valkey adapters against a real container** (`valkeytest`), not a mock.

## Not here

Middleware for any framework (ginkit), algorithms it does not implement (see the module's `docs/2026-07-17-algorithms.md`), per-route policy (the app picks a `Config` per limiter).
