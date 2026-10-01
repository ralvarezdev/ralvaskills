# resilience

`github.com/ralvarezdev/resilience` — retry with backoff, a stateful backoff and a panic-safe restart supervisor. Standard library only; import the subpackage you need, the root exports nothing. Reach for it before hand-writing a retry loop, an exponential delay or a "restart the goroutine" supervisor.

## Packages

| Package | Owns |
|---|---|
| `resilience/backoff` | `Strategy`, `Exponential` (initial, max, factor, jitter; saturates instead of overflowing), `Constant`, jitter modes `JitterNone`/`JitterFull`/`JitterEqual`, the stateful `Backoff` (`Next`, `Reset`, `Attempts`) |
| `resilience/retry` | `Do[T]`, `DoErr`, `Policy`, `Permanent`, `After`, the typed `*retry.Error` |
| `resilience/supervise` | `Run` (restart on error or panic, backoff reset after a healthy run, restart budget), `Group` (supervise several) |

## Recipe

```go
v, err := retry.Do(ctx, retry.Policy{
    MaxAttempts: 4, // zero = 3, negative = until ctx ends
    Backoff:     backoff.Exponential{Initial: time.Second, Max: 8 * time.Second, Jitter: backoff.JitterEqual},
    Retryable:   isTransient, // nil = everything but context.Canceled
}, func(ctx context.Context, attempt int) (Result, error) {
    res, err := call(ctx)
    if isBadRequest(err) {
        return res, retry.Permanent(err) // never retry
    }
    return res, err
})

err = supervise.Run(ctx, "consumer", consume, supervise.Config{
    Backoff:       backoff.Exponential{Initial: time.Second, Max: time.Minute, Jitter: backoff.JitterEqual},
    HealthyAfter:  time.Minute, // a run this long resets the backoff
    MaxRestarts:   5,
    RestartWindow: 10 * time.Minute,
})
```

## Rules

- **Classify errors.** Return `retry.Permanent(err)` for 4xx and config errors; retrying everything loops on a bug until the context dies. Use `retry.After(err, d)` to honor a server's `Retry-After`.
- **Add jitter** (`JitterEqual` or `JitterFull`) wherever many clients retry together; lockstep retries stampede the dependency.
- **Context first.** Retry checks the context before the first attempt and between attempts; the sleep is context-aware. `*retry.Error` unwraps to the last error and joins the context error when the context ended, so `errors.Is` works for both.
- **Callers that expect the bare last error** must unwrap `*retry.Error` (`errors.As`) to read `Attempts` or the cause.
- **`supervise.Run` returns nil on context cancellation** and on a clean exit (unless `RestartOnExit`); it returns `ErrRestartBudget` when it gives up. `tick` and `svckit` follow the same cancel-is-nil rule.
- **A panic is recovered,** logged with its stack and treated as a failure (`*PanicError`).
- **`HealthyAfter` resets the backoff, not the restart count;** the budget is `MaxRestarts` within `RestartWindow` (zero window = whole lifetime).
- **Tests are instant:** `Policy.Sleep`, `supervise.Config.Sleep`/`Now` and the strategies' `Rand` are injectable.
- **Known gaps (v0.1.0):** no linear backoff, no `OnStart`/`OnGiveUp` hooks, no consecutive-failure budget. Check the repo before working around them.

## Not here

Circuit breaker, rate limiter, bulkhead, hedged requests, retry budgets, a `restkit` adapter feeding `Retry-After` into `retry.After`, supervising processes.
