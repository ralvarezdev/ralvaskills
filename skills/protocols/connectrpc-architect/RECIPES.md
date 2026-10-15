# ConnectRPC Recipes (Go)

Reference implementations for the patterns in [SKILL.md](SKILL.md).

## Service definition

```proto
// acme/shop/users/v1/users.proto
service UserService {
  rpc GetUser(GetUserRequest) returns (User);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);        // cursor-paginated
  rpc UpdateUser(UpdateUserRequest) returns (User);
  rpc DeleteUser(DeleteUserRequest) returns (google.protobuf.Empty);
  rpc WatchUsers(WatchUsersRequest) returns (stream UserEvent);       // server-stream
}
```

```yaml
# buf.gen.yaml — paths=source_relative keeps generated files beside the proto tree
version: v2
plugins:
  - remote: buf.build/protocolbuffers/go
    out: gen
    opt: paths=source_relative
  - remote: buf.build/connectrpc/go          # or: local: protoc-gen-connect-go
    out: gen
    opt: paths=source_relative
```

## Mounting handlers with the interceptor chain

```go
import (
    "net/http"

    "connectrpc.com/connect"
    "connectrpc.com/otelconnect"
    "connectrpc.com/validate"

    usersv1 "acme/shop/gen/users/v1"
    "acme/shop/gen/users/v1/usersv1connect"
)

func newMux(svc UserService, authSvc AuthService) (http.Handler, error) {
    otelInterceptor, err := otelconnect.NewInterceptor()
    if err != nil {
        return nil, err
    }

    chain := connect.ChainUnaryInterceptors(
        recoveryInterceptor(),            // 1. panic → CodeInternal
        requestIDInterceptor(),           // 2. correlation id in ctx + response header
        loggingInterceptor(),             // 3. structured access log
        authInterceptor(authSvc),         // 4. authenticate (principal in ctx)
        validate.NewInterceptor(),        // 5. protovalidate rules from the .proto
        otelInterceptor,                  // 6. metrics + traces
    )

    mux := http.NewServeMux()
    path, handler := usersv1connect.NewUserServiceHandler(
        &userServer{svc: svc},
        connect.WithInterceptors(chain),
        connect.WithSendGzip(),
    )
    mux.Handle(path, handler)

    // Health is always on; reflection is dev-only.
    mux.Handle(grpchealth.NewHandler(grpchealth.NewStaticChecker(usersv1connect.UserServiceName)))
    if cfg.Dev {
        reflector := grpcreflect.NewStaticReflector(usersv1connect.UserServiceName)
        mux.Handle(grpcreflect.NewHandlerV1(reflector))
        mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
    }
    return mux, nil
}
```

## Domain → `connect.Code` mapping

```go
// internal/errors/connect.go
func toConnect(err error) error {
    switch {
    case errors.Is(err, users.ErrNotFound):
        return connect.NewError(connect.CodeNotFound, err)
    case errors.Is(err, users.ErrEmailExists):
        return connect.NewError(connect.CodeAlreadyExists, err)
    case errors.Is(err, auth.ErrInvalidToken):
        return connect.NewError(connect.CodeUnauthenticated, errors.New("invalid token"))
    case errors.Is(err, auth.ErrForbidden):
        return connect.NewError(connect.CodePermissionDenied, nil)
    case errors.Is(err, db.ErrStale):
        return connect.NewError(connect.CodeAborted, errors.New("resource was modified"))
    default:
        slog.Error("unexpected error", "err", err)
        return connect.NewError(connect.CodeInternal, errors.New("internal error")) // never leak
    }
}

// Attach machine-readable detail (read with connect.DetailsFromError on the client).
func validationError(requestID string, fields []*errdetails.BadRequest_FieldViolation) error {
    cerr := connect.NewError(connect.CodeInvalidArgument, errors.New("invalid request"))
    detail, derr := connect.NewErrorDetail(&errdetails.BadRequest{FieldViolations: fields})
    if derr == nil {
        cerr.AddDetail(detail)
    }
    return cerr
}
```

## Interceptors — recovery + request-id

```go
type ctxKey int

const requestIDKey ctxKey = iota

var errInternal = errors.New("internal error")

func recoveryInterceptor() connect.UnaryInterceptorFunc {
    return func(next connect.UnaryFunc) connect.UnaryFunc {
        return func(ctx context.Context, req connect.AnyRequest) (resp connect.AnyResponse, err error) {
            defer func() {
                if r := recover(); r != nil {
                    slog.Error("panic", "method", req.Spec().Procedure, "panic", r)
                    resp, err = nil, connect.NewError(connect.CodeInternal, errInternal)
                }
            }()
            return next(ctx, req)
        }
    }
}

func requestIDInterceptor() connect.UnaryInterceptorFunc {
    return func(next connect.UnaryFunc) connect.UnaryFunc {
        return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
            id := req.Header().Get("X-Request-Id")
            if id == "" {
                id = newID()
            }
            resp, err := next(context.WithValue(ctx, requestIDKey, id), req)
            if resp != nil {
                resp.Header().Set("X-Request-Id", id)
            }
            return resp, err
        }
    }
}
```

## Streaming server handler

```go
func (s *userServer) WatchUsers(
    ctx context.Context,
    req *connect.Request[usersv1.WatchUsersRequest],
    stream *connect.ServerStream[usersv1.UserEvent],
) error {
    after := req.Msg.GetStartAfterId() // resume token — reconnect from a known point
    for {
        select {
        case <-ctx.Done():
            return connect.NewError(connect.CodeCanceled, ctx.Err())
        case ev := <-s.events(after):
            after = ev.GetId()
            if err := stream.Send(ev); err != nil {
                return err // Send already carries the right code (e.g. client canceled)
            }
        }
    }
}
```

## Typed client with deadline

```go
client := usersv1connect.NewUserServiceClient(httpClient, baseURL) // Connect protocol by default

ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

resp, err := client.GetUser(ctx, connect.NewRequest(&usersv1.GetUserRequest{Id: id}))
if err != nil {
    if connect.CodeOf(err) == connect.CodeNotFound {
        // branch on the code; details via connect.DetailsFromError
    }
    return err
}
_ = resp.Msg
```

- Force a protocol with `connect.WithGRPC()` / `connect.WithGRPCWeb()`; wrap a call in the house [resilience](../../personal/go-house-libs/references/RESILIENCE.md) retry policy instead of a hand-rolled loop.

## `httptest` + conformance-style test

```go
// One client factory per transport of the same Go port.
type transport struct {
    name  string
    build func(t *testing.T) usersv1connect.UserServiceClient
}

func newConnectClient(t *testing.T, svc UserService) usersv1connect.UserServiceClient {
    t.Helper()
    mux := http.NewServeMux()
    path, h := usersv1connect.NewUserServiceHandler(&userServer{svc: svc})
    mux.Handle(path, h)
    srv := httptest.NewServer(mux)
    t.Cleanup(srv.Close)
    return usersv1connect.NewUserServiceClient(srv.Client(), srv.URL)
}

func TestUserServiceConformance(t *testing.T) {
    t.Parallel()
    transports := []transport{
        {"connect", newConnectClient},
        {"grpc", newGRPCBufconnClient},  // bufconn — see grpc-architect
        {"grpcweb", newGRPCWebClient},
    }
    for _, tc := range transports {
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            client := tc.build(t)
            t.Run("not_found", func(t *testing.T) {
                t.Parallel()
                _, err := client.GetUser(t.Context(),
                    connect.NewRequest(&usersv1.GetUserRequest{Id: "missing"}))
                var cerr *connect.Error
                require.ErrorAs(t, err, &cerr)
                require.Equal(t, connect.CodeNotFound, cerr.Code())
            })
        })
    }
}
```
