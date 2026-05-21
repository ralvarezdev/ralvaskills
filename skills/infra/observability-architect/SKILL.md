---
name: observability-architect
version: 1.0.0
description: Application-side observability — structured logging, Prometheus metrics, OpenTelemetry tracing, trace/log/metric correlation, head sampling, PII discipline, RED+USE golden signals. Prometheus for metrics, OTLP for logs+traces. Use when instrumenting code, naming metrics, designing log structure, or auditing what an application emits.
---

# Observability Architecture — Signal Production

What an application emits about itself: **logs**, **metrics**, **traces**. Pairs with [grafana-architect](../grafana-architect/SKILL.md) which owns the consumption side (dashboards, alerts). This skill enforces what gets emitted, how it's named, and how the three pillars correlate.

## 1. Three pillars + correlation

Each pillar answers a different question:

| Pillar | Question | Cardinality | Cost |
|---|---|---|---|
| **Logs** | "What happened on *this* request?" | High (per event) | High storage |
| **Metrics** | "What is the *rate* / *aggregate*?" | Low (aggregated) | Low storage |
| **Traces** | "What was the *path*?" | One per request | Medium |

The *design goal* is **correlation**: from any signal you can pivot to the others. A trace shows you a slow request → click → see the logs from that request → see the metric spike around the same time. This is what makes observability worth the cost.

**Correlation mechanism: trace_id everywhere.**

- Every log line carries the current `trace_id` and `span_id`.
- Every metric exposes the current `trace_id` as an exemplar (Prometheus exemplars feature).
- Every trace span carries the operation name, attributes, and status.

If you skip the correlation, you have three independent data lakes — useful in isolation, painful to cross-reference.

## 2. Metrics: Prometheus

**Why Prometheus over OTel metrics in 2026:** Prom client libraries are mature in every language, the exposition format is universal, and ops engineers already know `rate()` / `histogram_quantile()`. OTel metrics are catching up but not yet at parity for the ergonomics.

### Naming

Follow the [Prometheus / OpenMetrics convention](https://prometheus.io/docs/practices/naming/):

- **Pattern:** `<namespace>_<subsystem>_<name>_<unit>_<type>`
- `<namespace>` = application name (`orders`, `payments`).
- `<unit>` is **always present** for sample values: `_seconds`, `_bytes`, `_total`, `_ratio`, `_celsius`. Never `_ms`, never bare units.
- `<type>` suffix for cumulative counters: `_total`. Gauges and histograms don't take a type suffix.

```
orders_http_requests_total{method="POST", route="/v1/orders", status="201"}
orders_http_request_duration_seconds_bucket{le="0.5", route="/v1/orders"}
orders_db_connections_active        # gauge, no suffix
orders_queue_messages_inflight       # gauge, no suffix
```

### Types

- **Counter** — monotonically increasing; use `_total` suffix; reset on process restart. Use `rate()` to query.
- **Gauge** — point-in-time value, can go up or down. Active connections, queue depth, in-flight requests.
- **Histogram** — distribution of values; choose buckets explicitly (don't accept defaults). For latency: `[0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]` seconds.
- **Summary** — pre-computed quantiles; avoid unless you can't aggregate across instances (you usually can — use histogram).

### Cardinality

The single most expensive observability mistake: high-cardinality labels.

- **Bad:** `user_id`, `email`, `request_id`, `correlation_id`, anything unique per request. These create one time series per unique value — millions of series, ruined retention.
- **Good:** `method`, `route` (templatized, not the raw path), `status_class` (`2xx`, `4xx`, `5xx`), `tenant`, `region`. Bounded sets.
- **Templatize routes** to bounded patterns: `/v1/users/{id}` not `/v1/users/01J9X...`.
- **`< 100` distinct values per label** as a soft cap; review labels with more.

## 3. Logging: structured, leveled, correlated

- **Structured JSON** to stderr (per [cli-tool-architect §6](../../tooling/cli-tool-architect/SKILL.md#6-output-discipline--stdout-for-data-stderr-for-logs) and language conventions). One log line = one JSON object.
- **Stdlib first:**
  - Go: `log/slog` per [go-architect §5](../../languages/go-architect/SKILL.md#5-stdlib-defaults). `slog.JSONHandler` in production.
  - Python: `structlog` configured to emit JSON. `python-architect`-canonical for app-side structured logs.
- **Mandatory fields on every line:**
  - `timestamp` (RFC 3339, UTC, with timezone)
  - `level` (`debug` / `info` / `warn` / `error`)
  - `msg` (a short, fixed string — not a sprintf)
  - `trace_id`, `span_id` (when inside a traced request)
  - `service` (the application name)
- **Don't log full payloads.** A request body field can be PII; the bytes are dead weight even when it isn't. Log the **shape**: `request_size_bytes`, `field_count`, `customer_id`. Never the contents.
- **`msg` is a fixed string** for grep-ability: `msg="user created"` with `user_id` and `email_domain` as separate fields, not `msg=f"created user {email}"`.
- **One level per environment:**
  - Production: `info`
  - Staging: `info` or `debug`
  - Dev: `debug`
- **Errors include the operation that failed + the inputs that mattered** (per [cli-tool-architect §10](../../tooling/cli-tool-architect/SKILL.md#10-logging--structured-to-stderr)). No stack traces in `msg`; stack traces are a separate `error.stack` field.

## 4. Tracing: OpenTelemetry

OTel for traces is unambiguously the right choice — vendor-neutral, well-instrumented per language, supports all the major backends (Jaeger, Tempo, Honeycomb, Datadog, etc.).

- **Auto-instrumentation first.** OTel instrumentation libs for net/http, gin, FastAPI, requests, psycopg, grpc — they cover 80% of useful spans for free. Add manual spans only where business logic boundaries deserve them.
- **Span naming:** verb + resource. `POST /v1/orders`, `db.query`, `kafka.publish`. **Templatized**, low cardinality — same discipline as metric labels.
- **Attribute keys follow [OTel semantic conventions](https://opentelemetry.io/docs/specs/semconv/):** `http.method`, `http.status_code`, `db.system`, `messaging.system`. Don't invent your own.
- **Set span status on errors:** `span.SetStatus(codes.Error, msg)` — backends colorize error spans in the trace view.
- **Don't trace everything.** Per-call spans for hot loops kill performance. Trace the request boundary, the major sub-operations (DB call, external API call, queue publish), and the failure paths.

## 5. Correlation

Three rules to make signals jump between each other:

1. **`trace_id` in every log line.** Get it from the OTel SDK's current context — every modern logging integration supports this.
2. **Prometheus exemplars** on histograms (especially latency). When a slow bucket increments, the exemplar records the `trace_id`. Grafana lets you click from a histogram bucket directly to the trace.
3. **`service.name` and `service.version`** as resource attributes on traces + as labels on metrics + as fields on logs. Ties signals across deploys and versions.

```python
# Python — slog-style with structlog + OTel context
import structlog, opentelemetry.trace

def add_trace_context(_, __, event):
    span = opentelemetry.trace.get_current_span()
    if span.is_recording():
        ctx = span.get_span_context()
        event["trace_id"] = f"{ctx.trace_id:032x}"
        event["span_id"]  = f"{ctx.span_id:016x}"
    return event

structlog.configure(processors=[add_trace_context, structlog.processors.JSONRenderer()])
```

```go
// Go — slog handler that pulls trace_id from context
func (h *traceHandler) Handle(ctx context.Context, r slog.Record) error {
    if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
        r.AddAttrs(
            slog.String("trace_id", sc.TraceID().String()),
            slog.String("span_id", sc.SpanID().String()),
        )
    }
    return h.inner.Handle(ctx, r)
}
```

## 6. Sampling

- **Head sampling at 10%** in production. Deterministic per `trace_id` so a request is either fully sampled or fully dropped — no half-traces.
- **100% sampling on errors** — always keep the trace when the span status is error. Cheap, decisive ROI.
- **100% sampling in non-prod** (dev, staging) — keep everything; the volume is low and you want full visibility.
- **Tail sampling via the OTel Collector** is the upgrade when 10% head sampling misses interesting low-volume endpoints. Adds Collector infrastructure. Defer until needed.
- **Configure via env vars:** `OTEL_TRACES_SAMPLER=parentbased_traceidratio` + `OTEL_TRACES_SAMPLER_ARG=0.1`.

## 7. What NOT to emit

**PII and secrets** never appear in any signal:

- No email addresses, names, addresses, phone numbers in attributes or fields. Use `email_domain` (`acme.com`) if you need to slice by tenant.
- No passwords, tokens, API keys, session cookies. Not in headers, not in payloads. Use redaction filters at the SDK level — defense in depth, since one careless `log.Info("creating user", user)` can leak it all.
- No full payloads. Log size and shape, never contents.
- No internal IDs that could be re-identified. UUID v7 IDs are fine in metric labels because they don't leak meaning (per [sql-architect §1](../../databases/sql-architect/SKILL.md#1-schema-design)); raw `auth_token_id` is not fine.

Configure your SDK redaction list at startup; review it on every deploy.

## 8. SLOs and golden signals

What to measure for *every* service.

### RED — for request-driven services

| Signal | Question | Metric |
|---|---|---|
| **R**ate | How many requests/sec? | `<svc>_http_requests_total` over time |
| **E**rrors | How many fail? | `<svc>_http_requests_total{status=~"5.."}` |
| **D**uration | How long do they take? | `<svc>_http_request_duration_seconds` histogram |

### USE — for resource-driven services

| Signal | Question | Metric |
|---|---|---|
| **U**tilization | What % of capacity is used? | CPU%, memory%, queue%, connection pool% |
| **S**aturation | How queued / blocked is it? | Run queue, GC pause, lock waits |
| **E**rrors | What's failing at this layer? | Resource-specific error counters |

### SLOs

- **One SLO per critical user journey.** Not per endpoint — per *journey*. "User can place an order in < 500ms 99% of the time."
- **Error budget** = `1 - SLO`. Track burn rate; alert on fast burn.
- **Don't write more than 3–5 SLOs per service.** More than that and no one watches any of them.
- **Multi-window, multi-burn-rate alerts** (Google SRE handbook): 2% in 1h + 5% in 6h = page. Avoids both alert noise and slow detection.

## 9. Language recipes

### Go

```go
import (
    "log/slog"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics
var httpRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
    Namespace: "orders",
    Subsystem: "http",
    Name:      "requests_total",
    Help:      "HTTP requests by method, route, status.",
}, []string{"method", "route", "status"})

// Tracing — auto-instrument outbound HTTP and the server
tp := initTracerProvider()
otel.SetTracerProvider(tp)
client := &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
mux := http.NewServeMux()
handler := otelhttp.NewHandler(mux, "orders-api")

// Logging — slog with trace context (see §5)
logger := slog.New(traceHandler{inner: slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})})
slog.SetDefault(logger)
```

### Python

```python
import structlog
from opentelemetry import trace
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from prometheus_client import Counter, Histogram

# Metrics
http_requests = Counter(
    "orders_http_requests_total",
    "HTTP requests by method, route, status",
    ["method", "route", "status"],
)

# Tracing
FastAPIInstrumentor().instrument_app(app)

# Logging — structlog with trace context (see §5)
structlog.configure(
    processors=[
        structlog.contextvars.merge_contextvars,
        add_trace_context,  # from §5
        structlog.processors.add_log_level,
        structlog.processors.TimeStamper(fmt="iso", utc=True),
        structlog.processors.JSONRenderer(),
    ],
    wrapper_class=structlog.make_filtering_bound_logger(logging.INFO),
)
```

## 10. Cross-skill ties

- [grafana-architect](../grafana-architect/SKILL.md) — consumes everything this skill produces. Dashboards query Prometheus metrics; explorations follow trace_id from logs to traces.
- [go-architect §5](../../languages/go-architect/SKILL.md#5-stdlib-defaults) — `log/slog` as the logging default.
- [python-architect](../../languages/python-architect/SKILL.md) — `structlog` for structured logging.
- [fastapi-architect](../../frameworks/fastapi-architect/SKILL.md) / [gin-architect](../../frameworks/gin-architect/SKILL.md) / [nethttp-architect](../../frameworks/nethttp-architect/SKILL.md) — OTel auto-instrumentation libs cover the framework HTTP boundary.
- [rest-api-architect §10](../../protocols/rest-api-architect/SKILL.md#10-auth--security-headers) — `correlation_id` ties to `trace_id`; expose it in error responses for support workflows.
- [docker-architect §7](../docker-architect/SKILL.md#7-runtime-defaults) — logging driver routes structured logs to stderr → aggregator.
