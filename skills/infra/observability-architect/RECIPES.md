# Observability — Implementation Recipes

Wiring code for the signals enforced in [SKILL.md](SKILL.md). Loaded on demand.

## 1. Correlation: `trace_id` in log lines

### Python (structlog + OTel)

```python
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

### Go (slog wrapper handler)

```go
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

## 2. Full instrumentation stack — Go

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

// Logging — slog with trace context (see §1)
logger := slog.New(traceHandler{inner: slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})})
slog.SetDefault(logger)
```

## 3. Full instrumentation stack — Python

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

# Logging — structlog with trace context (see §1)
structlog.configure(
    processors=[
        structlog.contextvars.merge_contextvars,
        add_trace_context,  # from §1
        structlog.processors.add_log_level,
        structlog.processors.TimeStamper(fmt="iso", utc=True),
        structlog.processors.JSONRenderer(),
    ],
    wrapper_class=structlog.make_filtering_bound_logger(logging.INFO),
)
```

## 4. Default histogram buckets

For latency histograms, choose buckets explicitly — don't accept defaults.

```
[0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]   // seconds
```

Tune at the edges for the service's actual latency distribution; the middle bands are usually fine across services.

## 5. Sampling — env-var configuration

```
OTEL_TRACES_SAMPLER=parentbased_traceidratio
OTEL_TRACES_SAMPLER_ARG=0.1
```

- Deterministic per `trace_id` — a request is either fully sampled or fully dropped.
- Override to `1.0` in dev/staging.
- Pair with span processors that force-sample on error status for 100% error-trace retention.
