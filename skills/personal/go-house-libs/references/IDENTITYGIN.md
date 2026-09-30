# identitygin

`github.com/ralvarezdev/identitygin` — the gin adapter for identity: the auth middleware and cookie handling every gin app on identity would otherwise copy. `*app.App` from identity satisfies its `Authenticator` directly.

## What it gives you

| API | Use |
|---|---|
| `RequireAuth(app, opts...)` | middleware accepting either `Authorization: Bearer <PAT>` or the access-token cookie; stores the resolved `auth.Claims` under `ContextClaimsKey` |
| `Claims(c)` | reads those claims in a handler; `claims.UserID` is the same for a session and a PAT |
| `Cookies{Domain, Secure, AccessTTL, RefreshTTL, RefreshPath, ...}` | `Set` writes the access and refresh cookies (HttpOnly, SameSite=Strict); `Clear` expires them |
| `RefreshToken(c)` | the raw refresh token from its cookie |
| `WithUnauthorized(fn)` | replaces the default `401 {"status":401,"message":"unauthorized"}`, so an app can answer in its own envelope |

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
api.Use(identitygin.RequireAuth(identityApp, identitygin.WithUnauthorized(writeMyEnvelope401)))
api.GET("/me", func(c *gin.Context) {
    claims, _ := identitygin.Claims(c)
    // claims.UserID
})
```

## Rules

- **A Bearer header is authoritative.** When present it is the only credential considered: a bad or revoked PAT is a 401 even if a valid session cookie rides along, so a client cannot mask a rejected token behind a session.
- **One 401 for every failure.** Unknown, revoked, expired, or an inactive owner all answer 401; the middleware never says which credential was wrong.
- **Clear cookies after anything that ends the session** (logout, password change): call `Cookies.Clear`, or the browser keeps sending a token identity already revoked.
- **Use the same `Cookies` value for `Set` and `Clear`.** A cookie only clears when its Path and Domain match the one it replaces. `RefreshPath` defaults to `/api/auth`, so the refresh cookie only reaches the auth endpoints.
- **Rate limiting keys off the claims.** Put `ginkit.RateLimit` after `RequireAuth` and key it by `claims.UserID` ([GINKIT.md](GINKIT.md)).

## Not here

Login or registration handlers (the app writes those over `identity/app`), authorization decisions, non-gin adapters (a different framework gets its own `<framework>` adapter module).
