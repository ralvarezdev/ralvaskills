# Gin Recipes

Reference implementations for the patterns in [SKILL.md](SKILL.md).

## Project structure

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
```

## Route registration

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

## Graceful shutdown

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

## Auth middleware — Pattern A (in-house JWT)

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

**Pattern B — external IdP**: `jwt.ParseWithClaims` with a `jwt.Keyfunc` resolving keys via a JWKS client. Cache the JWKS in-process with TTL.

## Authorization per route

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

## Problem-details renderer (RFC 7807)

```go
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

## Testing

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
