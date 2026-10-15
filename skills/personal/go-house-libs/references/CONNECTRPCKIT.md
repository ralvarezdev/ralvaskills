# connectrpckit

`github.com/ralvarezdev/connectrpckit` — the ConnectRPC counterpart of [GRPCKIT](GRPCKIT.md): house-default `connect.HandlerOption`s, an `http.Handler` you mount, and a typed client. One generated service, one `http.Handler`, serves the **Connect, gRPC and gRPC-Web** protocols at once. Released as v0.1.0 (pre-1.0, so expect the API to move). Reach for it before hand-rolling recovery/logging/validation interceptors, a `corekit.Kind` → `connect.Code` table, request-id propagation, or a health/reflection registration.

## Packages

One flat package: `NewServer` + `ServerConfig` (`Options`, `NewServeMux`, the `Register` seam), `NewClient[Req,Resp]` + `ClientConfig`, the standalone interceptors (`Recovery`, `Logging`, `RequestID`, `RequestIDClient`, `Validate`, `TimeoutClient`, `BearerClient`, `MapErrors`), `Validator`, `ErrorMapper`, `KindErrorMapper`, `KindCode`, `KindConnectError`, `ConnectError`, `InternalMessage`, `Reflector`, `NewHealthHandle` + `HealthHandle`, `BearerToken`, `RequestIDFromContext`/`WithRequestID`. Depends on `connectrpc.com/connect`, `connectrpc.com/grpchealth`, `connectrpc.com/grpcreflect`, `github.com/ralvarezdev/corekit` and `google.golang.org/genproto/googleapis/rpc`; it never imports a service, `svckit` or `net/http` beyond the handler it returns.

## Recipe

```go
srv := connectrpckit.NewServer(connectrpckit.ServerConfig{
    Logger:       logger,
    RequestID:    true,
    Validate:     func(m proto.Message) error { return validator.Validate(m) }, // protovalidate
    MapError:     connectrpckit.KindErrorMapper(logger),
    Interceptors: []connect.Interceptor{authInterceptor, metricsInterceptor}, // auth, metrics
})

health := connectrpckit.NewHealthHandle("ralvarezdev.identity.v1.PoolService")
mux := srv.NewServeMux(
    func(opts ...connect.HandlerOption) (string, http.Handler) {
        return identityv1connect.NewPoolServiceHandler(pools, opts...)
    },
    health.Register,
    connectrpckit.Reflector("ralvarezdev.identity.v1.PoolService"), // dev only
)
// mux is an http.Handler: hand it to svckit.Serve (see SVCKIT) or net/http.

client := connectrpckit.NewClient[GetPoolRequest, GetPoolResponse](
    "https://identity.internal"+identityv1connect.PoolServiceGetPoolProcedure, // base URL + procedure
    connectrpckit.ClientConfig{
        Bearer:    func(ctx context.Context) string { return tokenFrom(ctx) },
        RequestID: true,
        Timeout:   5 * time.Second,
    },
)
resp, err := client.CallUnary(ctx, connect.NewRequest(&GetPoolRequest{Id: id}))
```

## Rules

- **Chain order, outermost first:** recovery, logging, request id, `ServerConfig.Interceptors`, validation, `MapError`. Recovery is outermost so it also catches panics from your interceptors. Put auth in `Interceptors` (it runs before validation, so unauthenticated callers never reach the validator).
- **A panic never reaches the client.** Recovery answers `CodeInternal` with the constant `InternalMessage`; the panic value and stack are logged server side only.
- **Logs carry no payloads, error text or metadata:** `procedure`, `kind`, `code`, `duration_ms`, `peer`; Warn when the call failed.
- **One error-mapping table.** Handlers return `*corekit.Error`; `KindErrorMapper` renders it via `KindConnectError` (`KindCode` + a `google.rpc.ErrorInfo` reason + a `google.rpc.BadRequest` detail from the error-severity issues), passes an already-built `*connect.Error` through unchanged, and logs then answers `CodeInternal`/`InternalMessage` for anything else. `KindUnknown` maps to `CodeInternal`. `MapErrors` keeps the original error when a mapper returns nil, so a misbehaving mapper cannot turn a failure into a silent success.
- **`NewClient` takes the full procedure URL** — base URL plus procedure path, exactly what a generated client passes to `connect.NewClient` (for example `".../ralvarezdev.identity.v1.PoolService/GetPool"`). A base URL alone calls the wrong path. The client is cheap; reuse it.
- **`ClientConfig`'s zero value is valid.** `GRPC`/`GRPCWeb` force a protocol and are mutually exclusive (Connect is the default), `ProtoJSON`/`SendGzip` change the wire, and `Timeout` only fills a call with no deadline — a caller's own deadline always wins. Streaming calls are given no leaked deadline; set one yourself.
- **`Register` is the composition seam.** A generated `NewXServiceHandler`, `Reflector` and `HealthHandle.Register` share the shape `func(...connect.HandlerOption) (string, http.Handler)`, so `NewServeMux` mounts them uniformly with the house options.
- **Reflection is dev-only.** Add `Reflector` in development; leave it out of production.
- **Bearer is a header helper, not an auth policy.** `BearerToken(header)` reads the incoming `Authorization: Bearer` token; `BearerClient(supply)` writes the outgoing one (an empty token adds no header). `RequestID`/`RequestIDClient` do the same for `X-Request-Id` via `WithRequestID`/`RequestIDFromContext`. Policy belongs in your own interceptor.
- **Validation is `func(proto.Message) error`** (adapt protovalidate); its error text is returned to the client in a `CodeInvalidArgument` (`validation failed: ...`), so it must be safe to expose. Nil disables it; non-proto messages are skipped.
- **Health is the standard `grpc.health.v1.Health` service** over Connect: `NewHealthHandle(names...)` starts them SERVING and composes with `NewServeMux`; call `SetServing`/`Shutdown` as dependencies come up or drain.
- **The kit never imports `svckit` or a concrete service.** It returns a plain `http.Handler`; the deployable owns the lifecycle ([SVCKIT](SVCKIT.md)).

## Not here

The service lifecycle (svckit), auth policy, metrics/OpenTelemetry, retry (use [RESILIENCE](RESILIENCE.md)), and the error vocabulary itself (`corekit` owns `Kind`; this kit only maps it). Native gRPC from the same proto is the deliberate fallback: keep the port mountable behind a [GRPCKIT](GRPCKIT.md) adapter too, so the protocol can change without rewriting the service. See [../STACK.md](../STACK.md) for the tag and the `corekit` pin.
