# svckit

`github.com/ralvarezdev/svckit` — server and service lifecycle for Go: a bind failure fails the process, shutdown is graceful, sibling services stop together, and the exit code says why. Standard library only; gRPC goes through a three-method interface that `*grpc.Server` satisfies, so svckit never imports grpc. Reach for it before writing `go srv.ListenAndServe()` plus `<-ctx.Done()`.

## Packages

One flat package: `Serve`/`ServeListener` (net/http, optional TLS), `ServeGRPC`/`ServeGRPCListener`, `Group` (several services, first failure cancels the rest), `NotifyContext`/`Finish`/`ErrInterrupted`/`ExitCode` (signals and exit codes).

## Recipe

```go
func main() {
    ctx, stop := svckit.NotifyContext(context.Background()) // SIGINT, SIGTERM
    defer stop()

    err := run(ctx)
    if err != nil && !errors.Is(err, svckit.ErrInterrupted) {
        fmt.Fprintln(os.Stderr, err)
    }
    os.Exit(svckit.ExitCode(err)) // 0 clean, 1 failure, 130 after a signal
}

func run(ctx context.Context) error {
    g := svckit.NewGroup(ctx)
    g.Go("api", func(ctx context.Context) error {
        return svckit.Serve(ctx, apiServer, svckit.Config{Name: "api", ShutdownTimeout: 10 * time.Second})
    })
    return g.Wait()
}
```

One server: `return svckit.Finish(ctx, svckit.Serve(ctx, srv, svckit.Config{}))` turns a clean signal-driven stop into `ErrInterrupted` (exit 130).

## Rules

- **Never `go srv.ListenAndServe()` and wait on the context.** The bug is idle-forever: the port is taken, the error is only logged, and the process looks healthy while serving nothing. `Serve` binds synchronously and returns `listen on <addr>: ...` immediately.
- **`Serve` returns nil after a clean graceful stop** and never reports `http.ErrServerClosed`. Do the post-shutdown work (draining mail or queues) after it returns.
- **Shutdown is bounded.** `ShutdownTimeout` (default 10s) waits for in-flight requests, then closes forcibly and returns an error wrapping `context.DeadlineExceeded`.
- **Tests use `:0`.** Read the port from `Config.OnListening`, or hand your own listener to `ServeListener`.
- **An already-cancelled context returns nil without binding.**
- **`Group` joins every failure, first first,** and does not treat a service that returns nil or only `context.Canceled` as failed.
- **Wrap the returned error** at the call site if your linter requires it (for example `serve api: %w`); svckit already prefixes the service name inside a `Group`.
- **Log lines changed from the hand-rolled versions** ("listening" with a `service` field); check alerts that match the old text.

## Not here

Health and readiness endpoints, a pre-stop drain delay, restart supervision (use [RESILIENCE](RESILIENCE.md) `supervise`), start and stop ordering, socket activation.
