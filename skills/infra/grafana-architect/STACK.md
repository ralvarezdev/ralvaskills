# Stack Versions

| Dependency | Pinned version | Purpose |
|---|---|---|
| Grafana | 13.0 | Dashboards + unified alerting + explore |
| Grizzly (grr) | 0.7 | Dashboards-as-code CLI (YAML/JSON apply, drift detection, kind=Dashboard/Folder/Datasource/AlertRuleGroup) |
| Prometheus | 3.11 | Default metrics data source (see [observability-architect](../observability-architect/STACK.md)) |
| Loki | 3.7 | Default logs data source (when used) |
| Tempo | 2.10 | Default traces data source (when used) |

## Notes

- **Dashboards-as-code via Grizzly** is the default. The Grafana UI is for exploration; production dashboards live in git as JSON synced by `grr apply`.
- **Single Grafana instance with per-environment data sources** (rather than per-env Grafana instances). Data source names: `prometheus-prod`, `prometheus-staging`, `loki-prod`. Cuts operational surface.
- **Unified alerting** (built-in since Grafana 8) — not Alertmanager. Alert rules live alongside dashboards in Grizzly. Same syncing pipeline.
- **No anonymous access**, even in staging. SAML/OIDC for humans; service-account API keys for Grizzly + automation.
- **Folder permissions** match service ownership. Editors of a service's folder = the team that owns the service. SRE / Platform admin handles cross-service concerns.
- **Pin Grafana to a specific minor version** (e.g. 13.0.x) in deployment manifests; bump via Renovate per [repo-tooling-architect](../../tooling/repo-tooling-architect/SKILL.md). Major upgrades go through a staging run + dashboard render diff before production.

_Last reviewed: 2026-05-21_
_Skill version at last review: 1.0.0_
