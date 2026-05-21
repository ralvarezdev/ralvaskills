---
name: gin-architect
version: 1.0.0
description: >
  Enforces strict Gin standards: feature-based project structure, struct-tag
  validation with go-playground/validator, RFC 7807 problem-details error
  middleware, in-house OAuth2+JWT (golang-jwt+argon2) or external IdP, route
  groups for URL-prefix versioning, and OpenAPI generation. Targets Gin 1.12
  on Go 1.26. Use when scaffolding a Gin service, writing or reviewing
  handlers / routes / middleware, or auditing an existing Gin codebase.
---

# Gin Architecture

Targets **Gin 1.12** on **Go 1.26**. Companion to [go-architect](../../languages/go-architect/SKILL.md) (language idioms) and [rest-api-architect](../../protocols/rest-api-architect/SKILL.md) (REST conventions: versioning, errors, auth patterns, idempotency, ETag). Data access via [sql-architect](../../databases/sql-architect/SKILL.md)'s `sqlx + //go:embed` pattern. See [STACK.md](STACK.md) for pinned dependencies.

## 1. Project structure — feature-based

One folder per bounded context. Each feature owns its routes, service, repo, DTOs, and SQL files. Mirrors [fastapi-architect](../fastapi-architect/SKILL.md) so polyglot teams can navigate either side.

```
cmd/api/
└── main.go                      # wire dependencies, register routes, run engine
internal/
├── config/                      # viper-loaded settings (validator-validated struct)
├── server/                      # gin.Engine setup, middleware, lifecycle
├── errors/                      # RFC 7807 problem-details + handler middleware
├── auth/                        # JWT issue/verify, current-user middleware
├── users/
│   ├── handlers.go              # gin.HandlerFunc per route
│   ├── service.go               # business logic — no gin imports
│   ├── repo.go                  # sqlx + queries/*.sql
│   ├── dto.go                   # request / response structs with validate tags
│   └── queries/
│       ├── get_user_by_id.sql
│       └── insert_user.sql
├── orders/
│   └── ...
```

- **`handlers.go`** depends on `service.go`; never reaches into `repo.go` directly.
- **`service.go`** is pure Go — no `gin` imports. Easy to unit-test without a fake `*gin.Context`.
- **`dto.go`** holds wire types with `json:` and `binding:` (validator) tags. **Never reuse domain structs as DTOs** — that's how internal fields leak into the API.

## 2. Routing & versioning

URL-prefix versioning via route groups. One group per version, one sub-group per feature.

```go
func registerRoutes(r *gin.Engine, deps *Deps) {
    v1 := r.Group("/v1")
    {
        users.Register(v1.Group("/users"), deps.UserSvc)
        orders.Register(v1.Group("/orders"), deps.OrderSvc)
    }
}

// internal/users/handlers.go
func Register(rg *gin.RouterGroup, svc *Service) {
    h := &handler{svc: svc}
    rg.GET("/:user_id", h.getByID)
    rg.POST("", h.create)
    rg.PATCH("/:user_id", h.update)
}
```

- **One `Register(rg, deps)` function per feature** — keeps `main.go` thin.
- **Path params typed at parse time:** `id, err := uuid.Parse(c.Param("user_id"))` — return 400 on parse failure.
- **Use `gin.RouterGroup`, not bare `Engine.GET`,** so versioning + per-group middleware stays clean.

## 3. Request validation

Use struct tags with `go-playground/validator` (canonical per [go-architect](../../languages/go-architect/SKILL.md#11-dependencies--logging)). Bind via `c.ShouldBindJSON` / `c.ShouldBindUri` / `c.ShouldBindQuery` — never `c.MustBindWith` (panics; we don't panic in handlers).

```go
type CreateUserReq struct {
    Email    string `json:"email"    binding:"required,email"`
    Password string `json:"password" binding:"required,min=12"`
}

func (h *handler) create(c *gin.Context) {
    var req CreateUserReq
    if err := c.ShouldBindJSON(&req); err != nil {
        problem.Render(c, problem.Validation(err))  // RFC 7807
        return
    }
    ...
}
```

- **`binding:"required"`** on every field that isn't truly optional.
- **Custom validators** registered once at startup: `validate.RegisterValidation("uuid_v7", isUUIDv7)`.
- **`binding:"omitempty"`** on PATCH partial-update DTOs — let absent fields mean "leave unchanged."

## 4. Response shaping

- **Return DTOs, not domain types.** `c.JSON(200, dto.UserResponse{...})` — DTOs control what leaks.
- **`json:"-"`** on any DTO field that should never serialize (passwords, internal IDs).
- **Status code with `c.JSON(http.StatusCreated, ...)`** — explicit, not Gin's default 200.
- **Empty body uses `c.Status(http.StatusNoContent)`**, not `c.JSON(204, nil)` (sends `null`).

## 5. Dependency injection

Explicit constructors per [go-architect §3](../../languages/go-architect/SKILL.md#3-instantiation). No globals, no `init()` magic. Dependencies live in a `Deps` struct wired in `main.go`.

```go
type Deps struct {
    DB         *sqlx.DB
    HTTPClient *http.Client
    UserSvc    *users.Service
    OrderSvc   *orders.Service
    AuthSvc    *auth.Service
}

func main() {
    cfg := config.MustLoad()
    deps := mustBuildDeps(cfg)
    defer deps.Close()
    r := server.New(cfg, deps)
    server.Run(r, cfg)  // sets up graceful shutdown
}
```

For larger graphs use `uber-go/fx` (canonical per go-architect). For services with <10 dependencies, hand-wired is clearer.

## 6. Lifespan & graceful shutdown

Open shared resources in `main.go`, never per-request. Close them on shutdown signal.

```go
func Run(r *gin.Engine, cfg *config.Config) {
    srv := &http.Server{
        Addr:              cfg.Addr,
        Handler:           r,
        ReadHeaderTimeout: 5 * time.Second,
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

- **`signal.NotifyContext`** (Go 1.26: cause records which signal) for clean signal handling.
- **`ReadHeaderTimeout`** is mandatory — without it, Slowloris can pin connections forever.
- **Graceful shutdown timeout** > longest expected request duration.

## 7. Authentication & authorization — Gin implementation

**Patterns** (in-house JWT vs external IdP, Argon2id, JWT lifetimes, JWKS verification, switching criterion) live in [rest-api-architect §11](../../protocols/rest-api-architect/SKILL.md#11-authentication-patterns). This section covers Gin specifics.

**Pattern A — in-house JWT** via `golang-jwt/jwt/v5` + `argon2` from `golang.org/x/crypto`:

```go
func AuthRequired(secret []byte) gin.HandlerFunc {
    return func(c *gin.Context) {
        tok := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
        if tok == "" {
            problem.Render(c, problem.Unauthorized("missing bearer token"))
            c.Abort()
            return
        }
        claims := &auth.Claims{}
        parsed, err := jwt.ParseWithClaims(tok, claims, func(*jwt.Token) (any, error) {
            return secret, nil
        }, jwt.WithValidMethods([]string{"HS256"}))
        if err != nil || !parsed.Valid {
            problem.Render(c, problem.Unauthorized("invalid token"))
            c.Abort()
            return
        }
        c.Set("user_id", claims.Subject)
        c.Next()
    }
}
```

**Pattern B — external IdP**: use `jwt.ParseWithClaims` with a `jwt.Keyfunc` that resolves keys via a JWKS client (small custom client around `golang.org/x/oauth2/jwt`, or a community lib). Cache the JWKS in-process with TTL.

**Authorization per route, never global:**

```go
func RequireScope(scope string) gin.HandlerFunc {
    return func(c *gin.Context) {
        scopes := c.MustGet("scopes").([]string)
        if !slices.Contains(scopes, scope) {
            problem.Render(c, problem.Forbidden())
            c.Abort()
            return
        }
        c.Next()
    }
}

usersGroup.DELETE("/:id", AuthRequired(secret), RequireScope("users:delete"), h.delete)
```

## 8. Error handling — RFC 7807 middleware

A central `problem` package emits `application/problem+json` (per [rest-api-architect §7](../../protocols/rest-api-architect/SKILL.md#7-error-contracts--rfc-7807-problem-details)). Handlers either call `problem.Render(c, p)` directly or `c.Error(err)` and let the recovery middleware convert.

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

func Render(c *gin.Context, p Problem) {
    p.Instance = c.Request.URL.String()
    p.CorrelationID = c.GetString("correlation_id")
    c.Data(p.Status, "application/problem+json", mustJSON(p))
}
```

- **One handler per domain-error family** — map known sentinel errors to `Problem` types in a single switch.
- **`gin.Recovery`** with a custom `RecoveryHandler` that emits a 500 `Problem` with `correlation_id` — never a stack trace.

## 9. Middleware order

Outermost first; order matters.

```go
r := gin.New()  // not gin.Default — we set logging ourselves
r.Use(
    middleware.RequestID(),             // 1. assign correlation_id
    middleware.SLog(),                  // 2. structured access log
    gin.Recovery(),                     // 3. convert panics → 500 problem
    middleware.CORS(cfg.CORS),          // 4. preflight handling
    gzip.Gzip(gzip.DefaultCompression), // 5. response compression
)
```

- **Never `gin.Default()`** in production — its logger writes unstructured text to stdout. Use `slog` (per go-architect).
- **Auth is per-route middleware**, never global (see §7).
- **Recovery before CORS** — a panic that bypasses CORS handler returns no headers; the browser shows a misleading CORS error.

## 10. Concurrency, not "background tasks"

Go has no FastAPI-style `BackgroundTasks`. To run work after the response:

```go
func (h *handler) create(c *gin.Context) {
    // ...respond synchronously...
    c.JSON(http.StatusCreated, resp)

    // Detach from request context — request context is cancelled on response.
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    go func() {
        defer cancel()
        if err := h.svc.SendWelcomeEmail(ctx, user.ID); err != nil {
            slog.ErrorContext(ctx, "welcome email failed", "user_id", user.ID, "err", err)
        }
    }()
}
```

- **Always derive a new `context.Context`** from `context.Background()` — request context is cancelled the moment the response writes.
- **Capture errors** — goroutine errors are invisible by default; log them.
- **Anything serious** (retryable, distributed, scheduled) belongs in a real task queue, not a goroutine.

## 11. Testing

- **`httptest.NewRecorder` + the engine directly** — no real socket.

```go
func TestCreateUser(t *testing.T) {
    deps := newTestDeps(t)
    r := server.New(testConfig(), deps)

    body := mustJSON(CreateUserReq{Email: "a@b.com", Password: "secret-1234567"})
    req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Idempotency-Key", uuid.NewString())

    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    require.Equal(t, http.StatusCreated, w.Code)
}
```

- **Test-DB strategy:** mirror `sql-architect` — wrap each test in a rolled-back transaction, or use a per-test schema with `golang-migrate`.
- **Table-driven tests** (per go-architect §9) for request validation paths — one row per (input, expected status, expected error type).

## 12. OpenAPI generation

- **Default: `swaggo/swag`** — `// @` annotation comments above handlers; `swag init` generates `docs/swagger.json` and `swagger.yaml`. Predictable, mature.
- **For full OpenAPI 3.1 control: `getkin/kin-openapi`** — write the spec in code; serve it; use it to validate requests at runtime. More work but no annotation noise.
- **CI snapshot-tests the spec** (per [rest-api-architect §15](../../protocols/rest-api-architect/SKILL.md#15-openapi-as-the-source-of-truth)).
- **Internal endpoints excluded** with `// @Hidden` (swag) or by not registering them in the spec route group.
