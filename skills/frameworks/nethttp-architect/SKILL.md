---
name: nethttp-architect
version: 1.0.0
description: >
  Enforces strict standards for Go REST services built on the stdlib `net/http`
  (Go 1.22+ enhanced ServeMux with method + path variable matching — no router
  framework). Feature-based structure, struct-tag validation, RFC 7807
  problem-details middleware chain, in-house JWT or external IdP auth, graceful
  shutdown, and OpenAPI via getkin/kin-openapi. Targets Go 1.26. Use when
  scaffolding a stdlib REST service, writing or reviewing handlers / mux
  registrations / middleware, or auditing an existing net/http codebase.
---

# net/http Architecture

Targets the stdlib **`net/http`** package on **Go 1.26**. Uses the Go 1.22+ enhanced `ServeMux` (method matching, path variables, host matching) — no router framework needed. Companion to [go-architect](../../languages/go-architect/SKILL.md) (language idioms) and [rest-api-architect](../../protocols/rest-api-architect/SKILL.md) (REST conventions: versioning, errors, auth patterns, idempotency, ETag). Data access via [sql-architect](../../databases/sql-architect/SKILL.md)'s `sqlx + //go:embed` pattern. See [STACK.md](STACK.md) for pinned dependencies.

## 1. Project structure — feature-based

Same shape as `gin-architect` and `fastapi-architect`. One folder per bounded context.

```
cmd/api/
└── main.go                      # wire deps, build mux, run server
internal/
├── config/                      # viper-loaded settings
├── server/                      # mux assembly, middleware chain, lifecycle
├── errors/                      # RFC 7807 problem-details + error writer
├── auth/                        # JWT verify, current-user middleware
├── users/
│   ├── handlers.go              # http.Handler / HandlerFunc per route
│   ├── service.go               # business logic — no net/http imports
│   ├── repo.go                  # sqlx + queries/*.sql
│   ├── dto.go                   # request / response structs with validate tags
│   └── queries/
├── orders/
└── ...
```

- **`handlers.go`** depends on `service.go`; never reaches into `repo.go` directly.
- **`service.go`** is pure Go — no `net/http` imports. Easy to unit-test.
- **`dto.go`** holds wire types; never reuse domain structs as DTOs.

## 2. Routing — `http.ServeMux` with method patterns (Go 1.22+)

The stdlib `ServeMux` now supports method matching, path wildcards (`{id}`), and host matching. No third-party router needed.

```go
// internal/server/mux.go
func newMux(deps *Deps) http.Handler {
    mux := http.NewServeMux()

    // Versioned prefix is part of each pattern — no native "group" concept.
    users.Register(mux, "/v1/users", deps.UserSvc)
    orders.Register(mux, "/v1/orders", deps.OrderSvc)

    return mux
}

// internal/users/handlers.go
func Register(mux *http.ServeMux, prefix string, svc *Service) {
    h := &handler{svc: svc}
    mux.HandleFunc("GET "    + prefix + "/{user_id}", h.getByID)
    mux.HandleFunc("POST "   + prefix,                h.create)
    mux.HandleFunc("PATCH "  + prefix + "/{user_id}", h.update)
    mux.HandleFunc("DELETE " + prefix + "/{user_id}", h.delete)
}
```

- **Method patterns:** `"GET /v1/users/{id}"` — method + path in one string. Mismatched methods auto-return `405 Method Not Allowed`.
- **Path variables:** `id := r.PathValue("user_id")` — typed parsing happens in your handler.
- **Trailing slash:** `"GET /v1/users/"` matches everything under the prefix; `"GET /v1/users/{$}"` matches only the exact path. Be explicit.
- **No "group" abstraction in stdlib** — write the prefix per route or use a small helper. Don't reach for a router framework just for syntactic sugar.

## 3. Request validation

`go-playground/validator` (canonical per [go-architect](../../languages/go-architect/SKILL.md#11-dependencies--logging)) — bind via `json.NewDecoder(r.Body).Decode(&req)` then `validate.Struct(&req)`.

```go
type CreateUserReq struct {
    Email    string `json:"email"    validate:"required,email"`
    Password string `json:"password" validate:"required,min=12"`
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
    var req CreateUserReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        problem.Write(w, r, problem.BadRequest("malformed json"))
        return
    }
    if err := h.validate.Struct(&req); err != nil {
        problem.Write(w, r, problem.Validation(err))
        return
    }
    // ...
}
```

- **`json.Decoder.DisallowUnknownFields()`** at decoder construction — equivalent of Pydantic's `extra="forbid"`. Reject unknown fields rather than silently dropping them.
- **One shared `*validator.Validate`** stored on the handler, not per-request — instantiation is expensive.
- **Limit body size:** wrap `r.Body` with `http.MaxBytesReader(w, r.Body, 1<<20)` before decode — prevents memory exhaustion from oversized payloads.

## 4. Response shaping

A tiny helper to serialise JSON consistently — every handler uses it.

```go
func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(status)
    if v != nil {
        _ = json.NewEncoder(w).Encode(v)
    }
}
```

- **Set headers before `WriteHeader`** — once headers are sent, you can't change them.
- **`http.StatusNoContent`** for empty responses (skip `writeJSON` body).
- **Return DTOs, not domain types.** Same discipline as gin-architect.

## 5. Dependency injection

Explicit constructors. Dependencies wired in `main.go`; `*Deps` is passed where needed.

```go
type Deps struct {
    DB       *sqlx.DB
    HTTP     *http.Client
    UserSvc  *users.Service
    OrderSvc *orders.Service
    AuthSvc  *auth.Service
}

func main() {
    cfg := config.MustLoad()
    deps := mustBuildDeps(cfg)
    defer deps.Close()
    srv := server.New(cfg, deps)
    server.Run(srv, cfg)
}
```

For larger graphs, `uber-go/fx` (canonical per go-architect). Hand-wired is clearer for small services.

## 6. Lifespan & graceful shutdown

Identical pattern to gin-architect — the server abstraction is `http.Server`, not `*gin.Engine`. `signal.NotifyContext` for shutdown trigger; the cause records which signal (Go 1.26).

```go
func Run(handler http.Handler, cfg *config.Config) {
    srv := &http.Server{
        Addr:              cfg.Addr,
        Handler:           handler,
        ReadHeaderTimeout: 5 * time.Second,
        ReadTimeout:       30 * time.Second,
        WriteTimeout:      30 * time.Second,
        IdleTimeout:       120 * time.Second,
    }

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    go func() {
        if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            slog.Error("server error", "err", err)
            stop()
        }
    }()

    <-ctx.Done()
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    _ = srv.Shutdown(shutdownCtx)
}
```

- **All four timeouts** are mandatory in production. Defaults are "unlimited" — that's a DOS vector.

## 7. Middleware — function wrapping

stdlib has no middleware abstraction. The standard pattern: a function that takes `http.Handler` and returns `http.Handler`. Compose with a tiny helper.

```go
type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, mws ...Middleware) http.Handler {
    for i := len(mws) - 1; i >= 0; i-- {
        h = mws[i](h)
    }
    return h
}

// internal/server/middleware.go
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := uuid.NewString()
        ctx := context.WithValue(r.Context(), correlationKey{}, id)
        w.Header().Set("X-Request-ID", id)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// main / server wiring
handler := Chain(mux,
    Recover,        // 1. panic → 500 problem
    RequestID,      // 2. assign correlation id
    SLog,           // 3. structured access log
    CORS(cfg.CORS), // 4. preflight
    GZip,           // 5. compression
)
```

- **Outermost first** in `Chain` — Recover wraps everything so even middleware panics are caught.
- **Auth is per-route, not global** — apply `AuthRequired` only where needed:
  ```go
  mux.Handle("DELETE /v1/users/{id}", AuthRequired(http.HandlerFunc(h.delete)))
  ```
- **Never write your own logger middleware that calls `log.Printf`** — use `slog` (per go-architect).

## 8. Authentication & authorization — stdlib implementation

**Patterns** live in [rest-api-architect §11](../../protocols/rest-api-architect/SKILL.md#11-authentication-patterns). Stdlib specifics:

**Pattern A — in-house JWT** via `golang-jwt/jwt/v5` + `argon2`:

```go
func AuthRequired(secret []byte) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            tok := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
            if tok == "" {
                problem.Write(w, r, problem.Unauthorized("missing bearer token"))
                return
            }
            claims := &auth.Claims{}
            parsed, err := jwt.ParseWithClaims(tok, claims, func(*jwt.Token) (any, error) {
                return secret, nil
            }, jwt.WithValidMethods([]string{"HS256"}))
            if err != nil || !parsed.Valid {
                problem.Write(w, r, problem.Unauthorized("invalid token"))
                return
            }
            ctx := context.WithValue(r.Context(), userKey{}, claims.Subject)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

**Pattern B — external IdP**: `jwt.ParseWithClaims` with a `Keyfunc` that resolves keys via cached JWKS. Verify `aud` and `iss` explicitly.

**Authorization per route, never global:**

```go
func RequireScope(scope string) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            scopes, _ := r.Context().Value(scopesKey{}).([]string)
            if !slices.Contains(scopes, scope) {
                problem.Write(w, r, problem.Forbidden())
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

mux.Handle("DELETE /v1/users/{id}",
    AuthRequired(secret)(RequireScope("users:delete")(http.HandlerFunc(h.delete))))
```

When the wrap chain gets ugly, build a small helper that takes multiple `Middleware`s and the final handler.

## 9. Error handling — RFC 7807

Same `problem` package shape as `gin-architect`. Helpers take `http.ResponseWriter` + `*http.Request` instead of `*gin.Context`.

```go
// internal/errors/problem.go
type Problem struct {
    Type          string `json:"type"`
    Title         string `json:"title"`
    Status        int    `json:"status"`
    Detail        string `json:"detail,omitempty"`
    Instance      string `json:"instance,omitempty"`
    CorrelationID string `json:"correlation_id,omitempty"`
}

func Write(w http.ResponseWriter, r *http.Request, p Problem) {
    p.Instance = r.URL.String()
    if id, ok := r.Context().Value(correlationKey{}).(string); ok {
        p.CorrelationID = id
    }
    w.Header().Set("Content-Type", "application/problem+json")
    w.WriteHeader(p.Status)
    _ = json.NewEncoder(w).Encode(p)
}
```

The `Recover` middleware (see §7) catches panics and emits a 500 problem with the correlation id — never a stack trace.

## 10. Concurrency, not "background tasks"

Same guidance as gin-architect — Go has no FastAPI-style `BackgroundTasks`. Detach with `context.Background()`, log errors, set timeouts.

```go
func (h *handler) create(w http.ResponseWriter, r *http.Request) {
    // respond synchronously
    writeJSON(w, http.StatusCreated, resp)

    // detach
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    go func() {
        defer cancel()
        if err := h.svc.SendWelcomeEmail(ctx, user.ID); err != nil {
            slog.ErrorContext(ctx, "welcome email failed", "user_id", user.ID, "err", err)
        }
    }()
}
```

## 11. Testing

`httptest.NewRecorder` + the handler directly — no real socket.

```go
func TestCreateUser(t *testing.T) {
    deps := newTestDeps(t)
    srv := server.New(testConfig(), deps)

    body := mustJSON(CreateUserReq{Email: "a@b.com", Password: "secret-1234567"})
    req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Idempotency-Key", uuid.NewString())

    w := httptest.NewRecorder()
    srv.ServeHTTP(w, req)

    require.Equal(t, http.StatusCreated, w.Code)
}
```

- **`testing/synctest`** (per go-architect §9) for any test that depends on goroutine completion timing — deterministic without `time.Sleep`.
- **Table-driven** request-validation tests.

## 12. OpenAPI generation

stdlib `net/http` has no annotation-based generator like `swaggo/swag` (which is Gin-coupled). Two practical options:

- **`getkin/kin-openapi`** — write the OpenAPI 3.1 spec in code, serve it at `/v1/openapi.json`, optionally use it for runtime request validation. The intended choice for stdlib.
- **Hand-written `openapi.yaml`** committed to the repo, served as a static file. Cheapest, but drifts; pair with the OpenAPI snapshot test from [rest-api-architect §15](../../protocols/rest-api-architect/SKILL.md#15-openapi-as-the-source-of-truth).

CI snapshot-tests the spec either way.

## When to pick `net/http` over Gin

- **Minimal dependencies** matter (binary size, supply chain, audit).
- **Tight integration with stdlib middleware ecosystem** (third-party `func(http.Handler) http.Handler` chains work everywhere).
- **Forward compatibility** — Go's HTTP server gets faster every release; framework-coupled code lags.
- **Predictable behavior** — no framework-specific surprises (recovery semantics, body re-reads, etc.).

Pick **Gin** when: you want a richer ecosystem of middleware/plugins out of the box, your team is more familiar with framework patterns, or you need features Gin provides that aren't in stdlib (response writers with built-in marshaling, etc.).

Both bundles (`gin` and `nethttp`) are valid; this skill targets `nethttp` projects.
