# Stack Versions

| Dependency | Pinned version | Purpose |
|---|---|---|
| docker-engine | 29.5 | Container runtime + builder |
| docker-compose | 2 (v2 plugin) | Service orchestration — `docker compose` subcommand |
| buildx | 0.34 | Multi-arch builds, advanced BuildKit features |
| trivy | 0.70 | Vulnerability + IaC + SBOM scanner — CI default |
| dockerfile syntax | `docker/dockerfile:1.10` | Latest BuildKit syntax (cache mounts, secrets, here-docs) |
| Go runtime base | `gcr.io/distroless/static-debian12:nonroot` | Static binary; ~2 MB; non-root |
| Python runtime base | `python:3.14-slim` | Debian slim — pairs with `uv` multi-stage |
| Node runtime base | `node:22-slim` | Debian slim — LTS only in production |

## Notes

- **`docker compose` (v2 plugin)** is the standard. The standalone `docker-compose` binary is deprecated.
- **File name:** `docker-compose.yaml` + `docker-compose.override.yaml` for dev (auto-merged).
- **Multi-arch default:** `linux/amd64` + `linux/arm64`.
- **Digest-pin base images** in production — refresh via Renovate / Dependabot.
- **Trivy is the default scanner** — fail builds on `HIGH,CRITICAL` severities by default.
- **`HEALTHCHECK` on every service** — Compose `depends_on` conditions rely on them.

_Last reviewed: 2026-07-07_
_Skill version at last review: 1.1.0_
