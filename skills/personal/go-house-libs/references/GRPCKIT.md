# grpckit

`github.com/ralvarezdev/grpckit` — the gRPC counterpart of [GINKIT](GINKIT.md): a `*grpc.Server` and `*grpc.ClientConn` built with house defaults. It owns what several projects hand-rolled — panic recovery, structured request logging, optional protobuf validation and a `reflection.Register` toggle — and layers on the shared error mapping (`corekit.Kind` → `codes.Code`), bearer auth, request-id propagation and the standard health service. Current tag v0.8.2 (Go 1.27.1; depends on [COREKIT](COREKIT.md) v0.3.0 and [RESILIENCE](RESILIENCE.md) v0.2.0 through `grpcretry`). Reach for it before hand-rolling recovery/logging/validation interceptors, a `Kind` → code table, an `x-request-id` scheme, a health service, or an `insecure.NewCredentials()` dial.

## Packages

| Package | Owns |
|---|---|
| `grpckit` | `NewServer` + `ServerConfig`, `NewClient` + `ClientConfig`, the interceptor constructors, `Validator`/`ErrorMapper`/`Verifier`, `KindCode`/`KindStatusError`/`KindErrorMapper`/`StatusError`, `RegisterHealth`/`HealthHandle`, `RequestIDFromContext`/`WithRequestID`, and the sentinels `ErrNoTransportSecurity`/`ErrConflictingTransport`/`ErrInsecureRemote`/`ErrMissingBearer`; the constants `InternalMessage`, `RequestIDHeader`, `AuthorizationHeader` |
| `grpckit/grpcretry` | `UnaryClient(Config)`: retries idempotent unary calls over `resilience/retry`. Its own package so the root depends on grpc and protobuf only |

Server surface: `ServerConfig{Logger, Validate, MapError, UnaryInterceptors, StreamInterceptors, MaxRecvMsgSize, MaxSendMsgSize, RequestID, EnableReflection}`, `NewServer(cfg, opts...) *grpc.Server`; the interceptors `RecoveryUnary`/`RecoveryStream`, `LoggingUnary`/`LoggingStream`, `RequestIDUnary`/`RequestIDStream`, `ValidateUnary`/`ValidateStream`, `MapErrorsUnary`/`MapErrorsStream`, `BearerUnary`/`BearerStream`; `KindCode`, `KindStatusError`, `KindErrorMapper`; `RegisterHealth(srv) *HealthHandle`.

Client surface: `ClientConfig{TLS, Insecure, AllowInsecureRemote, Bearer, RequestID, Timeout, MaxRecvMsgSize, MaxSendMsgSize, DialOptions}`, `NewClient(target, cfg) (*grpc.ClientConn, error)`; the client interceptors `TimeoutClientUnary`, `BearerClientUnary`/`BearerClientStream`, `RequestIDClientUnary`/`RequestIDClientStream`; `StatusError`.

## Recipe

```go
srv := grpckit.NewServer(grpckit.ServerConfig{
    Logger:            logger,
    Validate:          func(m proto.Message) error { return validator.Validate(m) }, // protovalidate
    MapError:          grpckit.KindErrorMapper(logger),                              // corekit.Kind -> codes.Code
    UnaryInterceptors: []grpc.UnaryServerInterceptor{grpckit.BearerUnary(verify)},
    RequestID:         true,
    EnableReflection:  cfg.Dev,
})
pb.RegisterThingServer(srv, handler)
health := grpckit.RegisterHealth(srv) // *HealthHandle: SetServingStatus("", NOT_SERVING) before a drain

// *grpc.Server satisfies svckit.GRPCServer; see SVCKIT.
err := svckit.ServeGRPC(ctx, srv, ":50051", svckit.Config{})

conn, err := grpckit.NewClient("dns:///thing:50051", grpckit.ClientConfig{
    TLS:       tlsCfg,
    Bearer:    func(ctx context.Context) string { return tokenFrom(ctx) },
    RequestID: true,
    Timeout:   5 * time.Second,
})
conn, err = grpckit.NewClient("127.0.0.1:50051", grpckit.ClientConfig{Insecure: true})
```

Handlers return `*corekit.Error`; the mapper renders it as a status. Retries are opt-in per idempotent method:

```go
grpckit.NewClient(target, grpckit.ClientConfig{DialOptions: []grpc.DialOption{
    grpc.WithChainUnaryInterceptor(grpcretry.UnaryClient(grpcretry.Config{
        Idempotent: func(method string) bool { return strings.HasSuffix(method, "/Get") },
    })),
}})
```

## Rules

- **Chain order, outermost first:** recovery, logging, request-id (when `RequestID`), `UnaryInterceptors`/`StreamInterceptors`, validation, `MapError`, then any `grpc.ChainUnaryInterceptor` option passed to `NewServer` (innermost). Recovery is outermost so it also catches panics from your interceptors. Put auth in `UnaryInterceptors` (it runs before validation, so unauthenticated callers never reach the validator).
- **A panic never reaches the client.** Recovery answers `codes.Internal` with the constant `InternalMessage`; the panic value and stack are logged server side only.
- **Logs carry no payloads, error text or metadata:** `method`, `kind`, `code`, `duration_ms`, `peer`; Warn when the call failed.
- **`Validator` is `func(proto.Message) error`,** so grpckit does not import protovalidate; adapt it as above. Its error text is returned to the client in an `InvalidArgument` status (`validation failed: ...`), so it must be safe to expose. Stream requests are validated per message inside `RecvMsg`; non-proto messages are skipped. Nil `Validate` disables it.
- **`MapError` runs innermost and only on a non-nil handler error.** `KindErrorMapper` renders a `*corekit.Error` through `KindCode` and `Kind.DefaultMessage()` (error-severity `corekit.Issue`s become a `BadRequest` field-violation detail), passes an already-built status error through unchanged, and logs then answers `codes.Internal`/`InternalMessage` otherwise. `KindUnknown` — and any non-`*corekit.Error` — maps to `Internal`, so internals never leak. The standalone `MapErrorsUnary`/`MapErrorsStream` do the same for a hand-built server; `StatusError` is the client companion.
- **`BearerUnary`/`BearerStream(verify)` fail closed.** A call with no usable `authorization: Bearer <token>` gets `ErrMissingBearer` (`Unauthenticated`); `verify` returns the context the handler runs with (carrying the caller's identity) or the status to answer verbatim. On the client, `ClientConfig.Bearer` attaches the token to unary and stream calls; an empty token adds no header.
- **Reflection is off by default;** enable only in dev (`EnableReflection`).
- **Request ids** use the `x-request-id` metadata key: `RequestID` on either config reuses the inbound id or mints one; the server reads it with `RequestIDFromContext`, a caller sets the outgoing one with `WithRequestID`.
- **`NewClient` is lazy and fails closed.** Set exactly one of `TLS` or `Insecure`: neither gives `ErrNoTransportSecurity`, both `ErrConflictingTransport`. `Insecure` is accepted only for unix sockets, `localhost` and loopback IPs (`unix:`, `unix-abstract:`, `dns:///`, `passthrough:///` or bare host:port forms); any other scheme counts as remote and returns `ErrInsecureRemote` unless `AllowInsecureRemote` is also set (trusted private network only).
- **Config structs, zero value is safe;** nil `Logger` means `slog.Default()`. `ClientConfig.Timeout` bounds a unary call that has no deadline (a caller's own always wins) and wraps the whole call, retries included. `MaxRecvMsgSize`/`MaxSendMsgSize` raise the size bound on either side. `DialOptions` are appended after grpckit's own.
- **`HealthHandle` flips the standard health service.** `RegisterHealth` registers it and starts the overall server (`""`) SERVING, so a container probe works immediately; `SetServingStatus` handles readiness and `Shutdown` a drain. It is separate from svckit's HTTP probes.
- **`grpcretry` only retries what you mark idempotent.** `Config.Idempotent` is the gate (nil retries nothing, because a retry can repeat a write the server already applied) and `Config.Codes` defaults to `Unavailable`. It returns the server's own status, not a `*retry.Error`, so `status.Code` keeps working; pass it through `DialOptions`.
- The standalone interceptors are exported for hand-built servers; use `NewServer` otherwise.

## Not here

The REST and Connect transports ([GINKIT](GINKIT.md)/[RESTKIT](RESTKIT.md), [CONNECTRPCKIT](CONNECTRPCKIT.md)), the error vocabulary itself ([COREKIT](COREKIT.md) owns `Kind`), the retry policy ([RESILIENCE](RESILIENCE.md) `retry`; `grpcretry` only wires it to a status), rate limiting and metrics/OpenTelemetry (both on the README roadmap until a second consumer needs them). Serving and graceful shutdown belong to [SVCKIT](SVCKIT.md). grpckit is the native-gRPC sibling of [CONNECTRPCKIT](CONNECTRPCKIT.md): the same proto can be served either way. See [../SKILL.md](../SKILL.md) for where grpckit sits among the house kits, and [../STACK.md](../STACK.md) for its tag and pins.
