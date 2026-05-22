# Performance Reviewer — Reference Tables

Tooling reference, example findings, and severity guide. Loaded on demand.

## 1. Example findings report

```markdown
| Severity | Rule | Location | Evidence | Fix |
|---|---|---|---|---|
| Critical | N+1 query | `orders/service.py:42` | Loop over orders calls `customer_repo.get_by_id` per order | Batch — `customer_repo.get_by_ids(ids)`; one SELECT for the set |
| High | Blocking I/O in async | `users/handler.go:18` | `time.Sleep(d)` inside an HTTP handler | Use `time.After(d)` + `select` with ctx, or remove the sleep |
| Medium | Missing index | `EXPLAIN orders.status` | Seq Scan on 12M rows, filter status='pending' | `CREATE INDEX ix_orders_status_pending ON orders(status) WHERE deleted_at IS NULL` |
| Low | Hot-path allocation | `pprof: alloc_objects` | `[]byte` allocated per request in error path | Reuse a `sync.Pool`-managed buffer |
| Info | Pagination grows linearly | `customer_list` | OFFSET 100000 takes 2 s | Switch to cursor pagination per sql-architect §5 |
```

## 2. Severity guide

- **Critical** — current latency / throughput already violates SLO, or will under modest load.
- **High** — clear perf bug that compounds with traffic (N+1 with growing collection size).
- **Medium** — inefficient pattern with measurable cost but not yet breaking.
- **Low** — micro-optimization; defensible only when other Critical/High are addressed.
- **Info** — observation about scalability that may matter at higher load.

End the report with a one-sentence summary: *"2 Critical, 1 High, 3 Medium; address Critical/High before next release."*

## 3. Tooling reference

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

Always reference the tool output in the finding's Evidence column.
