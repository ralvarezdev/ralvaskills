# tick

`github.com/ralvarezdev/tick` — one tested periodic runner for background jobs (reapers, pruners, refresh loops). Standard library only. Reach for it before writing `ticker := time.NewTicker(...)` with a `select` on `ctx.Done()` and `ticker.C`.

## Packages

One flat package: `Every` (blocking), `Start` (returns a `*Runner` with `Stop`, `Done`, `Err`), `Config`, `PanicError`, and an injectable `Clock` for tests.

## Recipe

```go
g.Go(func() error {
    return tick.Every(ctx, 24*time.Hour, tick.Config{
        Name:    "mv-refresh",
        Timeout: time.Minute,
        Logger:  logger,
    }, func(ctx context.Context) error {
        return db.RefreshMaterializedViews(ctx, database)
    })
})

r, err := tick.Start(ctx, time.Hour, tick.Config{Name: "prune"}, pruner.Run)
defer func() { _ = r.Stop(stopCtx) }() // cancels, then waits for the in-flight run
```

## Rules

- **Return the error; do not log it in the job.** `tick` logs a failed run at Error (with the job name) and calls `OnError`; a job that also logs double-reports. Log success yourself only if you want it (`Debug` completions are built in).
- **Runs never overlap.** A slow run drops ticks like `time.Ticker`; the cadence re-synchronizes afterwards. This is a sequential runner by design.
- **The first run is after one interval,** not at start. Set `RunImmediately` for a run on startup (the ticker starts first, so a long first run does not shift the schedule).
- **Panics are recovered** and reported as a failed run carrying `*PanicError` with the stack. `DisableRecover` lets them propagate.
- **Cancellation is not a failure.** `Every` returns nil when the context ends (after the in-flight run returns), so an errgroup does not report `context.Canceled` on a clean shutdown. A job must honor its context, because `tick` cannot abort one that ignores it.
- **`Jitter` must be smaller than the interval;** it spreads replicas sharing a schedule and does not accumulate. `Timeout` bounds each run.
- **`StopOnError` ends the loop** with the first error wrapped with the job name; the default is log and continue.
- **Metrics** go through `OnRun(name, duration, err)`; there is no Prometheus adapter yet.

## Not here

Cron schedules, backoff after failures or per-error retry (use [RESILIENCE](RESILIENCE.md) `retry` inside the job), concurrent runs.
