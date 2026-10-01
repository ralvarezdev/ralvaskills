# grpckit

`github.com/ralvarezdev/grpckit` — the gRPC counterpart of [GINKIT](GINKIT.md): a `*grpc.Server` and `*grpc.ClientConn` built with house defaults. **Untagged**: no release yet (CHANGELOG is all `[Unreleased]`, two commits), so pin a commit or wait for the first tag; the API may still move. Reach for it before hand-rolling recovery/logging/validation interceptors, a `reflection.Register` toggle, or an `insecure.NewCredentials()` dial.

## Packages

One flat package: `NewServer` + `ServerConfig`, `NewClient` + `ClientConfig`, standalone interceptors (`RecoveryUnary/Stream`, `LoggingUnary/Stream`, `ValidateUnary/Stream`), `Validator`, `InternalMessage`, and the errors `ErrNoTransportSecurity`, `ErrConflictingTransport`, `ErrInsecureRemote`. Depends only on `google.golang.org/grpc` and protobuf.

## Recipe

```go
srv := grpckit.NewServer(grpckit.ServerConfig{
    Logger:            logger,
    Validate:          func(m proto.Message) error { return validator.Validate(m) }, // protovalidate
    UnaryInterceptors: []grpc.UnaryServerInterceptor{authUnary, metricsUnary},
    EnableReflection:  cfg.Dev,
})
pb.RegisterThingServer(srv, handler)

// *grpc.Server satisfies svckit.GRPCServer; see SVCKIT.
err := svckit.ServeGRPC(ctx, srv, ":50051", svckit.Config{})

conn, err := grpckit.NewClient("dns:///thing:50051", grpckit.ClientConfig{TLS: tlsCfg})
conn, err = grpckit.NewClient("127.0.0.1:50051", grpckit.ClientConfig{Insecure: true})
```

## Rules

- **Chain order, outermost first:** recovery, logging, `UnaryInterceptors`/`StreamInterceptors`, validation, then any `grpc.ChainUnaryInterceptor` passed in `opts`. Recovery is outermost so it also catches panics from your interceptors. Put auth in `UnaryInterceptors` (it runs before validation, so unauthenticated callers never reach the validator).
- **A panic never reaches the client.** Recovery answers `codes.Internal` with the constant `InternalMessage`; the panic value and stack are logged server side only.
- **Logs carry no payloads, error text or metadata:** `method`, `kind`, `code`, `duration_ms`, `peer`; Warn when the call failed.
- **`Validator` is `func(proto.Message) error`,** so grpckit does not import protovalidate; adapt it as above. Its error text is returned to the client in an `InvalidArgument` status (`validation failed: ...`), so it must be safe to expose. Stream requests are validated per message inside `RecvMsg`; non-proto messages are skipped. Nil `Validate` disables it.
- **Reflection is off by default;** enable only in dev (`EnableReflection`).
- **`NewClient` is lazy and fails closed.** Set exactly one of `TLS` or `Insecure`: neither gives `ErrNoTransportSecurity`, both `ErrConflictingTransport`. `Insecure` is accepted only for unix sockets, `localhost` and loopback IPs (`unix:`, `dns:///`, `passthrough:///` or bare host:port forms); any other scheme counts as remote and returns `ErrInsecureRemote` unless `AllowInsecureRemote` is also set (trusted private network only).
- **Config structs, zero value is safe;** nil `Logger` means `slog.Default()`. `ClientConfig.DialOptions` are appended after grpckit's own.
- The standalone interceptors are exported for hand-built servers; use `NewServer` otherwise.

## Not here

Bearer/auth interceptors and client credentials, error-to-status mapping, request ids, keepalive and message-size defaults, health service, client timeout/retry (use [RESILIENCE](RESILIENCE.md) in your handler or call site), rate limiting, metrics/OpenTelemetry. All are on the README Roadmap, deferred until a second consumer needs them. Serving and graceful shutdown belong to [SVCKIT](SVCKIT.md).
