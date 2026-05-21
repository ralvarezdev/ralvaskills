# Stack Versions

| Dependency | Pinned version | Purpose |
|---|---|---|
| OpenTelemetry spec | 1.x stable | Vendor-neutral signal spec — logs, metrics, traces |
| go.opentelemetry.io/otel | 1.43 | Go OTel SDK (traces, baggage, propagation) |
| go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp | 0.68 | Auto-instrumentation for net/http + Gin handlers |
| prometheus/client_golang | 1.23 | Go Prometheus client (metrics) |
| opentelemetry-api (Python) | 1.42 | Python OTel API |
| opentelemetry-sdk (Python) | 1.42 | Python OTel SDK |
| opentelemetry-instrumentation (Python) | 0.63b1 | Python auto-instrumentation hub |
| prometheus-client (Python) | 0.25 | Python Prometheus client |
| structlog | 25.5 | Python structured logging |
| log/slog | stdlib (Go 1.21+) | Go structured logging — see [go-architect](../../languages/go-architect/STACK.md) |
| Prometheus server | 3.11 | TSDB + scraping; canonical metrics backend |

## Notes

- **Metrics: Prometheus, not OTel metrics.** OTel metrics are catching up but Prometheus client libraries are universally mature in 2026. Reassess when OTel metrics reach feature parity for histograms + exemplars.
- **Logs + traces: OTLP** via OTel collectors. Vendor-neutral; swap backends without touching code.
- **Naming convention:** Prometheus/OpenMetrics `<namespace>_<subsystem>_<name>_<unit>_<type>` with `_seconds`, `_bytes`, `_total`, `_ratio` suffixes (never `_ms`, never `_count`).
- **Cardinality cap:** `< 100` distinct values per label, soft cap. No per-request unique IDs as labels (`user_id`, `request_id` belong in logs/traces, not metric labels).
- **Sampling:** head sampling at 10% (`OTEL_TRACES_SAMPLER=parentbased_traceidratio`, `OTEL_TRACES_SAMPLER_ARG=0.1`) with 100% on errors. Tail sampling via OTel Collector is the upgrade.
- **Redaction:** SDK-level filters for PII (email, name, phone, address) and secrets (passwords, tokens, cookies). Configured at startup; reviewed on every deploy.
- **Resource attributes:** `service.name`, `service.version`, `deployment.environment` on every signal — drives correlation across deploys.

_Last reviewed: 2026-05-21_
_Skill version at last review: 1.0.0_
