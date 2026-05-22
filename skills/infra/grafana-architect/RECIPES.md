# Grafana — Reference Templates

Layouts, panel reference, and Grizzly alert YAML for the rules in [SKILL.md](SKILL.md). Loaded on demand.

## 1. Observability repo layout

```
observability/
├── grizzly.yaml                       # Grizzly config (Grafana URL, API key from env)
├── datasources/
│   └── prometheus.yaml                # data source provisioning
├── folders/
│   └── orders.yaml                    # folder definition
└── dashboards/
    └── orders/
        ├── orders-api-overview.json
        ├── orders-database.json
        └── orders-slo.json
```

- **Sync on merge to main.** A GitHub Action runs `grr apply` against staging on every push; against production on tagged releases.
- **Never edit in the UI for production dashboards.** Grizzly will refuse to apply if the in-Grafana version is newer (drift detection); resolve by pulling the change into JSON and committing it.
- **Exploration dashboards are different.** Build them in the UI; once they earn a permanent home, export and commit.

## 2. Service folder layout

```
Orders                    ← service folder
├── API overview         ← entry-point dashboard for the service
├── Database             ← specific subsystem
├── Queue                ← specific subsystem
├── SLO                  ← burn rate + error budget
└── Runbook links        ← links to runbook docs
```

- **Folder permissions match service ownership.** Editors of `Orders/` = the team that owns the Orders service.
- **General dashboards** (infrastructure overview, fleet-wide cost) live in a top-level `Platform` folder owned by SRE.
- **No personal folders for production.**

## 3. Panel-type reference

| Type | Use for |
|---|---|
| **Time series** | Default. Rates, latencies, gauges over time. |
| **Stat (single value)** | One number that matters right now: current error rate, in-flight requests, last-deploy timestamp. |
| **Gauge** | A single value against a known scale (CPU%, error-budget %). Use sparingly. |
| **Heatmap** | Distribution over time (latency histogram). When line-per-bucket would be unreadable. |
| **Bar chart / Table** | Top-N: top 10 slow endpoints, top 10 noisy errors. |
| **Logs panel** | Live tail of structured logs filtered by query. Pairs with traces panel for incident investigation. |
| **Traces panel** | Recent traces matching a service/operation filter. |
| **Text / Markdown** | Section headers, runbook links, oncall context. |

## 4. Grizzly alert rule

```yaml
# observability/dashboards/orders/orders-slo.alerts.yaml — Grizzly format
apiVersion: grizzly.grafana.com/v1alpha1
kind: AlertRuleGroup
metadata:
  name: orders-slo-burn
spec:
  folder: Orders
  interval: 1m
  rules:
    - title: "Orders API — fast burn (2% in 1h)"
      condition: B
      data: [...]
      noDataState: NoData
      execErrState: Error
      for: 2m
      labels:
        severity: page
        service: orders
      annotations:
        summary: "Orders API error budget burning at 2%+/hour"
        runbook_url: "https://wiki/runbooks/orders-api-burn"
```

- A panel and its alert rule are version-controlled together.
- `for:` is mandatory — minimum `2m` on noisy signals; `5m`+ for slow burns.
- `runbook_url` annotation is mandatory — the first thing on-call clicks.

## 5. Common pitfalls

- **Dashboard sprawl.** Every engineer creates a "my service" dashboard; six versions exist; nobody knows which is current. Per-service folder owned by the service team is the structural fix; one canonical dashboard per question.
- **Unactionable alerts.** Alerts that fire without a clear response erode trust. Review every page-severity alert quarterly; delete or downgrade what no one actioned.
- **Dashboards built on raw labels.** Direct PromQL on `request_id` (unbounded cardinality) is slow and wrong. Build on templatized labels per [observability-architect §2](../observability-architect/SKILL.md#2-metrics-prometheus).
- **Mixing units silently.** A panel showing both `_seconds` and `_milliseconds` lies. Convert at query time (`* 1000`) or set the unit explicitly.
- **Static-threshold alerts.** A threshold that worked at 100 req/s breaks at 1000. Use SLO-based alerts (burn rate) for traffic-relative signals.
- **Edit-in-UI drift.** Production dashboards edited in the Grafana UI bypass review. Detect with `grr diff`; fix by pulling the change into JSON.
- **Too many panels.** A scroll-forever dashboard means nothing gets read. Soft cap: ~15 panels per dashboard. Beyond that, split.
