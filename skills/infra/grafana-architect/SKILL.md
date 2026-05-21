---
name: grafana-architect
version: 1.0.0
description: Grafana dashboard + alert standards — dashboards-as-code via Grizzly, per-service folders, one-question-per-panel design, unified alerting with actionable runbooks, data source provisioning, low-cardinality discipline. Use when designing Grafana dashboards, writing alert rules, organizing folders/permissions, or auditing existing dashboards.
---

# Grafana Architecture — Signal Consumption

How an operator *uses* what [observability-architect](../observability-architect/SKILL.md) emits. Dashboards, alerts, data sources, exploration. **Dashboards-as-code via Grizzly** is the default; the Grafana UI is for exploration, not authoring.

## 1. Dashboards-as-code workflow

Dashboards live in the application repo (or a sibling `observability/` repo), versioned in git, applied via [Grizzly](https://grafana.github.io/grizzly/) (`grr`).

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

- **Sync on merge to main.** A GitHub Action runs `grr apply` against staging on every push; against production on tagged releases. Dashboards drift only via PRs that get reviewed.
- **Never edit in the UI for production dashboards.** Grizzly will refuse to apply if the in-Grafana version is newer (drift detection); resolve by pulling the change into JSON and committing it.
- **Exploration dashboards are different.** Build them in the UI; once they earn a permanent home, export and commit. The UI is for "is there a signal here?"; the JSON is for "this dashboard is part of how we operate."

## 2. Folder organization — per service

One folder per service. Operators on call find dashboards by what they own, not by which team made them (teams reshuffle; services don't).

```
Orders                    ← service folder
├── API overview         ← entry-point dashboard for the service
├── Database             ← specific subsystem
├── Queue                ← specific subsystem
├── SLO                  ← burn rate + error budget
└── Runbook links        ← links to runbook docs
```

- **Folder permissions match service ownership.** Editors of the `Orders` folder = the team that owns the Orders service.
- **General dashboards** (infrastructure overview, fleet-wide cost, etc.) live in a top-level `Platform` folder owned by SRE / platform.
- **No personal folders for production.** Personal exploration is fine in the UI scratch space; production dashboards belong to a service.

## 3. Panel design — one question per panel

Every panel answers exactly one question. If the title is "stuff," redesign.

- **Panel title is a question or a noun phrase:** `"Request rate (req/s)"`, `"p95 latency by route"`, `"5xx errors (last 1h)"`. Not `"Stats"`, not `"Metrics"`.
- **Y-axis unit is mandatory.** Set it explicitly — `seconds`, `bytes`, `req/s`, `percent`. Auto-formatting hides unit confusion.
- **Time range is consistent across the dashboard.** Don't mix "last 1h" panels with "last 24h" panels unless the dashboard's purpose is the comparison.
- **Thresholds where they exist.** SLO at 99.9% gets a red line at 99.9% on the panel. Visual context > narrative.
- **Legends are bounded.** A legend with 50 lines is unreadable; templatize the query (`sum by (route)`) to bound it. If the natural cardinality is high, switch to a heatmap or top-N.

### Panel types — when each earns its place

| Type | Use for |
|---|---|
| **Time series** | Default. Rates, latencies, gauges over time. |
| **Stat (single value)** | One number that matters right now: current error rate, in-flight requests, last-deploy timestamp. |
| **Gauge** | A single value against a known scale (CPU%, error-budget %). Use sparingly — most things are time series. |
| **Heatmap** | Distribution over time (latency histogram). When line-per-bucket would be unreadable. |
| **Bar chart / Table** | Top-N: top 10 slow endpoints, top 10 noisy errors. |
| **Logs panel** | Live tail of structured logs filtered by query. Pairs with traces panel for incident investigation. |
| **Traces panel** | Recent traces matching a service/operation filter. |
| **Text / Markdown** | Section headers, runbook links, oncall context. |

## 4. Variables and templating

Variables let one dashboard serve many slices.

- **Standard variables on every service dashboard:**
  - `service` (the service name, often single-value on a service dashboard)
  - `environment` (`prod`, `staging`)
  - `region` if multi-region
- **Variables come from label queries**, not hand-maintained lists: `label_values(up{job="$service"}, environment)`.
- **`All` is dangerous** on metrics with many series — selecting "All environments" can fan out into millions of series. Set `Include All option = false` for high-cardinality variables, or set `All` to a curated regex.
- **Variables must NOT be the only difference between dashboards.** If you have an `Orders Prod` and `Orders Staging` dashboard, those should be one dashboard with an `environment` variable.

## 5. Alerting — Grafana unified alerting

Alerts live alongside dashboards in Grizzly. A panel and its alert rule are version-controlled together.

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

### Alert hygiene

- **Every alert is actionable.** If the response is "I'll look in the morning," it's a dashboard, not an alert. **Pages must matter.**
- **Two-tier severity:** `severity=page` (wakes someone) and `severity=ticket` (queues for the next business day). No middle ground; clarity matters under pressure.
- **Multi-window, multi-burn-rate for SLOs:** `fast_burn = (2% in 1h)` AND `slow_burn = (5% in 6h)` together — page only when both fire. Catches real issues without flapping on momentary spikes.
- **`for:` is mandatory.** A spike that lasts 30s doesn't deserve a page. `for: 2m` minimum on noisy signals; `for: 5m`+ for slow burns.
- **Annotations include a `runbook_url`.** The first thing on-call clicks is the runbook; if it doesn't exist, the alert is half-built.
- **Group related alerts** under labels (`team`, `service`) so silencing can target a service without touching others.

### Alerts NOT to write

- **CPU > 80%.** Useless without context — is it serving load or saturated? Use saturation metrics (run queue, GC time) or RED (latency/errors) instead.
- **"Disk full" without `for:`** — `for: 15m` so a temporary spike from log rotation doesn't page.
- **Per-instance alerts on horizontally scaled services.** Fleet-level signal; one bad pod isn't a page.
- **Static thresholds on metrics that grow with traffic.** Use rate-of-change or anomaly instead.

## 6. Data sources

- **Provisioned via Grizzly**, not the UI. Datasources `.yaml` files live next to dashboards.
- **One data source per signal type per environment:** `prometheus-prod`, `prometheus-staging`, `loki-prod`, `tempo-prod`. Dashboards reference by data source name; environment switching is via Grafana data source variables or per-organization scoping.
- **Service account API keys**, not personal tokens. Rotate quarterly.
- **Read-only data sources for dashboards.** Write access (Alertmanager mute rules, etc.) goes through dedicated service accounts with audit logging.

## 7. Permissions

- **Editor at the folder level.** A service team gets `Editor` on `Orders/`; everyone else has `Viewer`.
- **Admin sparingly.** Two or three admins per Grafana instance — they manage data sources, plugins, and global settings. Service teams don't need it.
- **`Anonymous` access off** in any environment with non-public data — including staging.
- **Single sign-on (SAML/OIDC) for human access**, service accounts for automation. No shared passwords.

## 8. Common pitfalls

- **Dashboard sprawl.** Every engineer creates a "my service" dashboard; six versions exist; nobody knows which is current. **Per-service folder owned by the service team** is the structural fix; one canonical dashboard per question.
- **Unactionable alerts.** Alerts that fire but have no clear response erode trust. Review every page-severity alert quarterly; delete or downgrade what no one actioned.
- **Dashboards built on raw labels.** Direct PromQL on `request_id` (unbounded cardinality) is slow and wrong. Build on templatized labels per [observability-architect §2](../observability-architect/SKILL.md#2-metrics-prometheus).
- **Mixing units silently.** A panel showing both `_seconds` and `_milliseconds` lies. Convert at query time (`* 1000`) or set the unit explicitly.
- **Static-threshold alerts.** A threshold that worked at 100 req/s breaks at 1000. Use SLO-based alerts (burn rate of error budget) for traffic-relative signals.
- **Edit-in-UI drift.** Production dashboards edited in the Grafana UI bypass review. Detect with `grr diff`; fix by pulling the change into JSON and committing it.
- **Too many panels per dashboard.** A scroll-forever dashboard means nothing on it gets read. **Soft cap: ~15 panels per dashboard.** Beyond that, split.

## 9. SLO dashboards — the special case

Every service has exactly one **SLO dashboard** showing:

- **Current SLO compliance** — percentage over the rolling window (last 30 days).
- **Error budget remaining** — `(SLO_target - current_failure_rate) * total_requests` as a budget number.
- **Burn rate** — instantaneous burn rate, with fast/slow window comparison (multi-window alert thresholds visible as red lines).
- **Top contributors** — table of endpoints/operations driving the failure rate.

This dashboard is the single source of truth during an incident. Link to it from the runbook; from the page; from the post-incident review.

## 10. Cross-skill ties

- [observability-architect](../observability-architect/SKILL.md) — produces what Grafana consumes. Naming conventions and cardinality discipline established there must hold for queries here to work.
- [docker-architect §10](../docker-architect/SKILL.md#10-vulnerability-scanning--trivy) — Trivy scan results can be dashboards too (security metric over time).
- [rest-api-architect](../../protocols/rest-api-architect/SKILL.md) — SLO burn rate alerts reference REST status codes (`5xx` → error budget) and latency (`http_request_duration_seconds`).
- [improve-codebase-architecture](../../workflows/improve-codebase-architecture/SKILL.md) — friction visible in dashboards (slow endpoints, error hot spots) is input to architecture review.
- [grpc-architect §2](../../protocols/grpc-architect/SKILL.md) — gRPC status codes feed equivalent metrics; same SLO mechanics apply.
