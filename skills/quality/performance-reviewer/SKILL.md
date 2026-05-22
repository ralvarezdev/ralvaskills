---
name: performance-reviewer
version: 1.0.0
description: Cross-language performance review — N+1, missing indexes, blocking I/O in async, allocation hot paths, unbounded memory, slow algorithms. Grounds findings in EXPLAIN, pprof, py-spy, or observability metrics. Produces Critical/High/Medium/Low fixes. Use when reviewing a PR for perf or investigating a slow endpoint.
---

# Performance Reviewer

Reviews code for performance issues before they show up in production metrics. **Measurement-grounded** — every finding either references a profiler output, an `EXPLAIN` plan, or a metric from [observability-architect](../../infra/observability-architect/SKILL.md). Guesses don't go in the report.

## 1. When to invoke

- PR touches a hot path (request handler, DB query, loop over large data, queue consumer).
- Endpoint shows up in latency p99 / error-rate dashboards.
- Before a planned load test.
- After a metric regression — dashboard burn, SLO degradation.
- New code with `for` + `await` / `for` + DB call patterns (the N+1 detector).

## 2. Output format

Same shape as the other reviewers — findings table + summary.

```markdown
| Severity | Rule | Location | Evidence | Fix |
|---|---|---|---|---|
| Critical | N+1 query | `orders/service.py:42` | Loop over orders calls `customer_repo.get_by_id` per order | Batch — `customer_repo.get_by_ids(ids)`; one SELECT for the set |
| High | Blocking I/O in async | `users/handler.go:18` | `time.Sleep(d)` inside an HTTP handler | Use `time.After(d)` + `select` with ctx, or remove the sleep |
| Medium | Missing index | `EXPLAIN orders.status` | Seq Scan on 12M rows, filter status='pending' | `CREATE INDEX ix_orders_status_pending ON orders(status) WHERE deleted_at IS NULL` |
| Low | Hot-path allocation | `pprof: alloc_objects` | `[]byte` allocated per request in error path | Reuse a `sync.Pool`-managed buffer |
| Info | Pagination grows linearly | `customer_list` | OFFSET 100000 takes 2s | Switch to cursor pagination per sql-architect §5 |
```

Severity guide:

- **Critical** — current latency / throughput already violates SLO, or will under modest load.
- **High** — clear perf bug that compounds with traffic (N+1 with growing collection size).
- **Medium** — inefficient pattern with measurable cost but not yet breaking.
- **Low** — micro-optimization; defensible-fix only when other Critical/High are addressed.
- **Info** — observation about scalability that may matter at higher load.

**Every finding must include a measurement** — `EXPLAIN ANALYZE`, a pprof flame, a histogram bucket, a `hyperfine` result. *"Looks slow"* is not a finding.

## 3. Review approach

1. **Measure first.** Without a measurement, there's nothing to review. Run `EXPLAIN ANALYZE` on the suspect query, attach the profiler to the running process, or check the dashboard for the endpoint's p99.
2. **Read the diff with measurement in hand.** The measurement narrows where to look.
3. **Propose fixes that produce another measurement.** A fix without "and the new latency is X" is incomplete.

## 4. What to check — by category

### Database — N+1 and unindexed scans

The single most common perf bug. Per [sql-architect §5](../../databases/sql-architect/SKILL.md#5-joins--n1-prevention):

- **N+1 detection.** A list endpoint that fires one query for parents then one query *per parent* for children. Look for `for x in xs: y = repo.get(x.id)`.
- **Fix:** batch with `IN (?, ?, ?)`, or single JOIN, or DataLoader-style batcher.
- **Detect in tests:** log every SQL statement; any handler emitting > 3 queries per request is suspect.
- **Missing indexes.** Per [sql-architect §3](../../databases/sql-architect/SKILL.md#3-indexing-strategy): index every FK; index columns used in `WHERE`; composite index column order is most-selective-first.
- **`EXPLAIN ANALYZE` shows the truth.** Look for `Seq Scan` on large tables, `Sort` operations that spill to disk, `actual rows` >> `plan rows` (stats stale — run `ANALYZE`).
- **Pagination by OFFSET is O(N).** Switch to cursor per [sql-architect §4](../../databases/sql-architect/SKILL.md#4-query-patterns) — `WHERE (created_at, id) < ($1, $2) ORDER BY ... LIMIT N`.
- **Connection pooling** — every handler that opens a connection per request is a bug. Pool in the lifespan (per [fastapi-architect §5](../../frameworks/fastapi-architect/SKILL.md#5-lifespan--startup) / [gin-architect §6](../../frameworks/gin-architect/SKILL.md#6-lifespan--graceful-shutdown)).

### Async — blocking I/O in event loop

The Python async equivalent of N+1.

- **`time.sleep` inside `async def`** — blocks the event loop. Use `await asyncio.sleep(d)`.
- **`requests.get(...)` inside `async def`** — blocking HTTP call. Use `httpx.AsyncClient`.
- **Sync DB driver (`psycopg2`) inside `async def`** — blocking. Use `psycopg 3` async per [sql-architect](../../databases/sql-architect/SKILL.md).
- **Detection:** enable `asyncio.get_event_loop().slow_callback_duration = 0.1` in dev; logs every callback over 100 ms.
- **`asyncio.to_thread(blocking_fn, args)`** when you genuinely must call a sync API; never call it directly.
- **`asyncio.TaskGroup`**, not bare `asyncio.gather` — per [python-architect §4](../../languages/python-architect/SKILL.md#4-concurrency--resources). TaskGroup handles cancellation and exception aggregation properly.

### Go — goroutine and allocation issues

- **Unbounded goroutine spawn.** A loop that fires off `go func()` per item without backpressure. Use a worker pool or `errgroup` with `SetLimit(n)`.
- **`sync.WaitGroup.Go(fn)`** per [go-architect §7](../../languages/go-architect/SKILL.md#7-concurrency) — newer (Go 1.25) and cleaner than `Add(1) / defer Done()`.
- **Channels with wrong direction or buffer.** Unbuffered channel + goroutine on each side = deadlock waiting to happen.
- **`context.Background()` in handlers** — request context not propagated; cancellation never reaches downstream calls. Per [grpc-architect §6](../../protocols/grpc-architect/SKILL.md#6-deadlines--context-propagation): propagate `ctx` everywhere.
- **Allocation hot paths.** Run `pprof` heap profile; look for `runtime.makeslice`, `runtime.newobject` at the top of `alloc_objects`. Reuse buffers via `sync.Pool` where it's cheap to do so.
- **Slice growth** — appending to a slice without `make([]T, 0, capacity)` causes reallocations. If you know the size, set it.

### Algorithms and data structures

- **O(N²) loops over collections that grow.** Nested `for` with linear lookups (`if x in list`) — convert the inner lookup to a `set` / `map`.
- **`sort.Slice(...)` in a hot path** when items arrive sorted from upstream. Trust upstream order if you can.
- **Eager materialization of iterators.** `list(big_generator)` defeats the generator's purpose. Stay lazy unless something downstream actually needs the materialized list.
- **String concatenation in loops** — quadratic in Python and Go (when using `+`). Use `"".join(parts)` (Python) or `strings.Builder` (Go).
- **Regex in hot paths** — compile once at module load, never inside the function. Both Python's `re` and Go's `regexp` cache poorly under high churn.

### Memory

- **Unbounded caches.** A `dict` / `map` that grows without an eviction policy is a memory leak. Use an LRU with a max size, or a TTL.
- **Long-lived references to short-lived data.** Closure captures the parent function's whole frame; if the closure outlives the function, the frame stays alive.
- **Large struct copies.** Returning a 2 KB struct by value is fine; passing it through 10 layers by value isn't. Use a pointer once the size justifies it (~64 bytes is a rough cutoff in Go).
- **Goroutine leaks** — a goroutine waiting on a channel that's never closed is a leak. The Go 1.26 experimental `goroutineleak` profile catches them.

### Network / external calls

- **Missing timeouts** — every outbound HTTP / DB / queue / RPC call has a context deadline. Per [grpc-architect §6](../../protocols/grpc-architect/SKILL.md#6-deadlines--context-propagation), [fastapi-architect §12](../../frameworks/fastapi-architect/SKILL.md#12-performance), [nethttp-architect §6](../../frameworks/nethttp-architect/SKILL.md#6-lifespan--graceful-shutdown): unbounded waits cause cascading failures.
- **No retry with backoff** — flaky downstream gets retried 100 times in a tight loop; backoff is mandatory. Use jittered exponential backoff.
- **HTTP keep-alive disabled** — TLS handshake per request. Reuse the client (e.g. `httpx.AsyncClient` in the lifespan).
- **Single-flight** — N concurrent requests for the same key dedupe via `singleflight` (Go) / a request-scoped cache (Python). Otherwise a cache miss for a hot key floods downstream.

### Sampling and observability cost

- **High-cardinality labels in metrics** — per [observability-architect §2](../../infra/observability-architect/SKILL.md#2-metrics-prometheus): metric explosion is a perf issue for the Prometheus server too.
- **Logging in tight loops** — `log.Info` inside a 1M-iteration loop dominates the runtime. Log once at the end with a summary.
- **Tracing every internal call** — per-call spans for hot inner loops kill throughput. Trace request boundaries + major sub-ops, not every function.

## 5. Tooling

Tools that grounded the findings. Always reference the tool output in the finding.

| Tool | Language | What it shows |
|---|---|---|
| `EXPLAIN (ANALYZE, BUFFERS)` | SQL | Query plan + actual rows + IO cost |
| `pprof` (built-in) | Go | CPU + heap + goroutine + block profile |
| `pyspy` / `py-spy record` | Python | Sampling profiler for live processes |
| `hyperfine` | All (CLI) | Benchmarking; statistical comparison of two implementations |
| `go test -bench=.` | Go | Microbenchmarks |
| `pytest-benchmark` | Python | Microbenchmarks |
| `wrk` / `k6` | All | HTTP load testing |
| `pg_stat_statements` | Postgres | Production query profiling (top-N by total time) |
| Grafana dashboards | All | Per [observability-architect](../../infra/observability-architect/SKILL.md) — RED+USE signals show the symptom |

## 6. What this skill does NOT do

- **Load testing.** Generating load is operational work — `k6` / `wrk` runs against staging environments, not in a code review.
- **Capacity planning.** "How many instances do we need at 2× traffic" is a different conversation; this skill catches code-level issues that make capacity planning harder.
- **Premature optimization.** A finding requires a measurement; gut-feel optimizations are explicitly rejected. If the code isn't measurably slow, don't change it.

## 7. Cross-skill ties

- [sql-architect §3–§5, §9](../../databases/sql-architect/SKILL.md) — indexing, query patterns, N+1, `EXPLAIN ANALYZE`. The most common source of findings.
- [observability-architect](../../infra/observability-architect/SKILL.md) — metrics that reveal where to look (RED for request-driven, USE for resource-driven). Findings often start with "this dashboard shows..."
- [grafana-architect](../../infra/grafana-architect/SKILL.md) — burn rate / SLO dashboards trigger reviews.
- [go-architect §7](../../languages/go-architect/SKILL.md#7-concurrency) — concurrency primitives; goroutine + allocation discipline.
- [python-architect §4](../../languages/python-architect/SKILL.md#4-concurrency--resources) — async discipline; blocking-call detection.
- [rest-api-architect §5](../../protocols/rest-api-architect/SKILL.md#5-pagination--cursor-not-offset) / [§14](../../protocols/rest-api-architect/SKILL.md#14-rate-limiting) — cursor pagination, rate limiting affect perf at the API layer.
- [improve-codebase-architecture](../../workflows/improve-codebase-architecture/SKILL.md) — when a finding requires structural change (the hot path is shaped wrong, not just inefficient), promote to architecture review.
