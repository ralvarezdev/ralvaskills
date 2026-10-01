# identitygin

`github.com/ralvarezdev/identitygin` — the gin adapter for identity: the auth middleware and cookie handling every gin app on identity would otherwise copy. `*app.App` from identity satisfies its `Authenticator` directly.

## What it gives you

| API | Use |
|---|---|
| `RequireAuth(app, opts...)` | middleware accepting either `Authorization: Bearer <PAT>` or the access-token cookie; stores the resolved `auth.Claims` under `ContextClaimsKey` |
| `Claims(c)` | reads those claims in a handler; `claims.UserID` is the same for a session and a PAT |
| `AuthMethod(c)` | `MethodSession`, `MethodPAT` or `MethodNone` (not behind `RequireAuth`); reads an unexported key, so `c.Set(ContextMethodKey, ...)` cannot forge it |
| `RequireSession(opts...)` | after `RequireAuth`: 403 for PAT-authenticated requests, for endpoints a PAT must never reach (password change, PAT management) |
| `RequirePermission(app, codes...)` / `RequirePermissionWith(app, PermissionConfig, codes...)` | 403 unless the caller holds every code (`GetUserPermissions`, cached in identity); a resolution failure is a 500, never a grant. `PermissionConfig{Unauthorized, Forbidden, InternalError}` swaps in the app's envelope |
| `RequireSameOrigin(SameOriginConfig{AllowedOrigins}, opts...)` | `(gin.HandlerFunc, error)`: opt-in CSRF guard for cookie-authenticated unsafe methods (`Sec-Fetch-Site` / `Origin`); `*` or a non-origin returns `ErrSameOriginConfig` |
| `DeviceInfo(c)` | user agent and `c.ClientIP()` as `auth.DeviceInfo` for login, refresh, 2FA and WebAuthn calls |
| `ErrorProblem(err)` / `AbortError(c, err, opts...)` | map identity sentinels to a `restkit` problem: lockout 429, bad credentials or tokens 401, inactive 403, missing 404, taken 409, invalid input 422, anything else a sanitized 500 |
| `Cookies{Domain, Secure, AccessTTL, RefreshTTL, RefreshPath, Partitioned, HostPrefix, Logger}` | `Set` writes the access and refresh cookies (HttpOnly, SameSite=Strict); `Clear` expires them; `AccessName()` and `Validate()` below |
| `RefreshToken(c)` | the raw refresh token from its cookie |

Options (`Option`, shared by the middleware and `AbortError`; each only affects what it names):

| Option | Effect |
|---|---|
| `WithUnauthorized(fn)` / `WithForbidden(fn)` / `WithInternalError(fn)` | replace the default 401 `{"status":401,"message":"unauthorized"}`, 403 and 500 so an app can answer in its own envelope; the handler writes the response and the middleware aborts |
| `WithProblem()` | default responses become RFC 9457 `application/problem+json` through `ginkit.AbortProblem` |
| `WithLogger(l)` | records auth infrastructure failures and unmapped `AbortError` errors (default `slog.Default()`) |
| `WithAccessCookieName(name)` | read the access token from another cookie; pass `Cookies.AccessName()` when `HostPrefix` is on |
| `WithBearerSession()` | also accept a session JWT as `Authorization: Bearer` (`MethodSession`); a JWT-shaped token is tried as a session first, then as a PAT only if it is not a valid JWT. Off by default: a JWT in Bearer is a PAT lookup and fails |
| `WithRetryAfter(d)` | `Retry-After` on the 429 for `ErrLockedOut` (identity does not report the remaining lockout, pass your throttle window) |
| `WithErrorMapper(fn)` | `AbortError` tries `fn(err) (problem.Problem, ok)` first; use it for a too-short password, which has no sentinel |

## Recipe

```go
cookies := identitygin.Cookies{
    Domain:     cfg.CookieDomain,
    Secure:     cfg.CookieSecure,
    AccessTTL:  int(cfg.AccessTTL.Seconds()),
    RefreshTTL: int(cfg.RefreshTTL.Seconds()),
}

api := r.Group("/api")
api.POST("/auth/login", func(c *gin.Context) {
    result, err := identityApp.Login(c.Request.Context(), email, password, device)
    // ...
    cookies.Set(c, result.Pair)
})
api.Use(identitygin.RequireAuth(identityApp,
    identitygin.WithProblem(),
    identitygin.WithAccessCookieName(cookies.AccessName()),
))
csrf, err := identitygin.RequireSameOrigin(identitygin.SameOriginConfig{AllowedOrigins: []string{"https://app.example.com"}})
// handle err (startup error), then mount AFTER RequireAuth
api.Use(csrf)
api.GET("/me", func(c *gin.Context) {
    claims, _ := identitygin.Claims(c)
    // claims.UserID
})
api.POST("/auth/password", identitygin.RequireSession(), func(c *gin.Context) {
    // ... err := identityApp.ChangePassword(...)
    // if err != nil { identitygin.AbortError(c, err, identitygin.WithProblem()); return }
})
```

Call `cookies.Validate()` once at startup. The refresh and logout routes sit outside `RequireAuth` (they are cookie-authenticated by the refresh token): give them `RequireSameOrigin` too.

## Rules

- **A Bearer header is authoritative.** When present it is the only credential considered: a bad or revoked PAT is a 401 even if a valid session cookie rides along, so a client cannot mask a rejected token behind a session.
- **One 401 for every rejected credential, 500 for an outage.** Unknown, revoked, expired, or an inactive owner all answer 401 and never say which was wrong; a database or Valkey failure behind identity is logged and answered 500 (`WithInternalError`), never 401, so an outage does not log users out. A malformed Bearer header (no token, extra fields) stays authoritative and is rejected; the scheme is case-insensitive, whitespace-separated, and only the first `Authorization` header is read.
- **Mount `RequireSameOrigin` after `RequireAuth`.** The Bearer exemption applies only to requests `RequireAuth` actually authenticated from the header (a PAT, or a JWT under `WithBearerSession`); a bare `Authorization: Bearer` on an unauthenticated route gets no exemption. A request with neither `Sec-Fetch-Site` nor an allowed `Origin` is rejected 403. A CORS policy must not allow the `Authorization` header from untrusted origins.
- **Gate sensitive endpoints with `RequireSession`**, not just `RequireAuth`: a PAT passes `RequireAuth` and resolves to the same claims.
- **`DeviceInfo` trusts `c.ClientIP()`**, which honours `X-Forwarded-For` from any client unless the app calls `engine.SetTrustedProxies`.
- **Cookie hardening.** `Cookies.HostPrefix` names the access cookie `__Host-access_token` (needs `Secure` and an empty `Domain`; the refresh cookie cannot take the prefix because it is path-scoped). `Partitioned` (CHIPS) needs `Secure`. `Set` warns when `Secure` is off on an HTTPS request.
- **Clear cookies after anything that ends the session** (logout, password change): call `Cookies.Clear`, or the browser keeps sending a token identity already revoked.
- **Use the same `Cookies` value for `Set` and `Clear`.** A cookie only clears when its Path and Domain match the one it replaces. `RefreshPath` defaults to `/api/auth`, so the refresh cookie only reaches the auth endpoints.
- **`Claims` is the only supported read.** `RequireAuth` stores claims and method under unexported typed keys; the `ContextClaimsKey` string is kept for tests and legacy middleware that seed claims with `c.Set`.
- **Rate limiting keys off the claims.** Put `ginkit.RateLimit` after `RequireAuth` and key it by `claims.UserID` ([GINKIT.md](GINKIT.md)).

## Not here

Login or registration handlers (the app writes those over `identity/app`), authorization decisions beyond the `RequirePermission` gate, WebAuthn/TOTP endpoints (see [IDENTITY.md](IDENTITY.md) for the hardening opt-ins they depend on), non-gin adapters (a different framework gets its own `<framework>` adapter module).
