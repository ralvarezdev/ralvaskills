# net/http Recipes

Reference implementations for the patterns in [SKILL.md](SKILL.md). Same conventions; the SKILL.md body keeps only the *what* and *why*.

## Project structure

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
```

## Routing

```go
// internal/server/mux.go
func newMux(deps *Deps) http.Handler {
    mux := http.NewServeMux()
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

## Graceful shutdown — `http.Server` with all timeouts

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

## Middleware chain

```go
type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, mws ...Middleware) http.Handler {
    for i := len(mws) - 1; i >= 0; i-- {
        h = mws[i](h)
    }
    return h
}

func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := uuid.NewString()
        ctx := context.WithValue(r.Context(), correlationKey{}, id)
        w.Header().Set("X-Request-ID", id)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

handler := Chain(mux,
    Recover, RequestID, SLog, CORS(cfg.CORS), GZip,
)
```

## Auth — in-house JWT middleware (Pattern A)

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

## Per-route authorization

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

## Problem-details writer (RFC 7807)

```go
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

## Testing — `httptest.NewRecorder`

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
