# gRPC Recipes (Go)

Reference implementations for the patterns in [SKILL.md](SKILL.md).

## Service definition

```proto
// acme/shop/users/v1/users.proto
service UserService {
  rpc GetUser(GetUserRequest) returns (User);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);   // cursor-paginated
  rpc UpdateUser(UpdateUserRequest) returns (User);
  rpc DeleteUser(DeleteUserRequest) returns (google.protobuf.Empty);
  rpc WatchUsers(WatchUsersRequest) returns (stream UserEvent);  // server-stream
}
```

## Domain → code mapping

```go
// internal/errors/grpc.go
func toGRPC(err error) error {
    switch {
    case errors.Is(err, users.ErrNotFound):
        return status.Error(codes.NotFound, err.Error())
    case errors.Is(err, users.ErrEmailExists):
        return status.Error(codes.AlreadyExists, err.Error())
    case errors.Is(err, auth.ErrInvalidToken):
        return status.Error(codes.Unauthenticated, "invalid token")
    case errors.Is(err, auth.ErrForbidden):
        return status.Error(codes.PermissionDenied, "")
    case errors.Is(err, db.ErrStale):
        return status.Error(codes.Aborted, "resource was modified")
    default:
        slog.Error("unexpected error", "err", err)
        return status.Error(codes.Internal, "internal error")  // never leak details
    }
}
```

## Interceptor chain

```go
import (
    "google.golang.org/grpc"
    "buf.build/go/protovalidate"
    protovalidate_grpc "buf.build/go/protovalidate/grpc"
)

server := grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        recovery.UnaryServerInterceptor(),         // 1. panic → INTERNAL
        requestid.UnaryServerInterceptor(),        // 2. assign correlation id
        slogger.UnaryServerInterceptor(),          // 3. structured access log
        auth.UnaryServerInterceptor(authSvc),      // 4. authenticate (sets user in ctx)
        protovalidate_grpc.UnaryServerInterceptor( // 5. validate request per protovalidate
            mustNewValidator(),
        ),
        metrics.UnaryServerInterceptor(),          // 6. instrument latency/codes
    ),
    grpc.ChainStreamInterceptor(/* analogous stream interceptors */),
)
```

## Client-side deadline

```go
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
resp, err := client.GetUser(ctx, &pb.GetUserRequest{Id: id})
```

## Reflection (dev only)

```go
import "google.golang.org/grpc/reflection"

if cfg.Env != "prod" {
    reflection.Register(server)
}
```

## In-process testing — `bufconn`

```go
func TestGetUser(t *testing.T) {
    lis := bufconn.Listen(1024 * 1024)
    srv := grpc.NewServer(/* interceptors */)
    pb.RegisterUserServiceServer(srv, newTestHandler(t))
    go func() { _ = srv.Serve(lis) }()
    t.Cleanup(srv.Stop)

    conn, err := grpc.NewClient("passthrough://bufnet",
        grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) {
            return lis.Dial()
        }),
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    require.NoError(t, err)
    t.Cleanup(func() { _ = conn.Close() })

    client := pb.NewUserServiceClient(conn)
    resp, err := client.GetUser(t.Context(), &pb.GetUserRequest{Id: "test"})
    require.NoError(t, err)
    require.Equal(t, "test", resp.Id)
}
```
