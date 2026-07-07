# New Project Docker Checklist

Use this checklist when scaffolding a new project with Docker and Compose. Answer each question before creating your Dockerfile or docker-compose.yaml.

## Dockerfile & Base Image

- [ ] **Multi-stage build?** (builder + runtime; never ship toolchain to final image)
- [ ] **Language selected?** (Go → distroless, Python/Node → slim debian)
- [ ] **Base image pinned by digest** in production? (not just tag)
  - Go: `gcr.io/distroless/static-debian12:nonroot@sha256:...`
  - Python: `python:3.14-slim@sha256:...`
  - Node: `node:22-slim@sha256:...`
- [ ] **Non-root USER** in final stage? (UID ≥ 10000, or distroless `:nonroot` tag)
- [ ] **`.dockerignore` created** with `.git/`, `node_modules/`, `.venv/`, secrets, IDE files?
- [ ] **BuildKit features used?**
  - [ ] Cache mounts for package managers (`--mount=type=cache`)
  - [ ] Secret mounts, never `COPY` secrets
  - [ ] Multi-arch ready (`--platform linux/amd64,linux/arm64`)

## Compose & Services

- [ ] **File names correct?**
  - [ ] `docker-compose.yaml` (production baseline)
  - [ ] `docker-compose.override.yaml` (dev-only, auto-merged)
- [ ] **No `version:` key** in Compose file (v2 ignores it)
- [ ] **`depends_on` with `condition: service_healthy`** for ordered startup?
  - Requires `healthcheck` on all service dependencies
- [ ] **Named volumes for stateful data** (`db-data:`), not bind mounts?
- [ ] **Networks declared explicitly** (don't rely on default)?

## Runtime Configuration

- [ ] **Restart policy decided?** ⚠️
  - [ ] `restart: unless-stopped` for long-running services (app, db, cache, etc.)
  - [ ] `restart: no` for batch jobs / one-off tasks
  - [ ] `restart: on-failure` with max retries for transient services
- [ ] **Resource limits set** on every service?
  - [ ] `deploy.resources.limits.memory` (e.g., `512M`)
  - [ ] `deploy.resources.limits.cpus` (e.g., `0.5`)
- [ ] **Logging configured**?
  - [ ] `json-file` with size + count rotation, or `journald`
- [ ] **TZ environment variable set** (`TZ=Etc/UTC`)?
- [ ] **HEALTHCHECK on every long-running service**?
  - [ ] HTTP: `curl -f http://localhost:3000/healthz`
  - [ ] Database: `pg_isready`, `mysqladmin ping`, `redis-cli ping`

## Security

- [ ] **Digest-pinned base images** in production Dockerfiles?
- [ ] **No secrets in `.dockerignore` or baked into layers**?
- [ ] **Distroless or non-root USER** enforced in Compose?
- [ ] **Read-only root filesystem** for stateless services (`--read-only`)?
- [ ] **Drop unnecessary capabilities** (`--cap-drop=ALL --cap-add=NET_BIND_SERVICE`)?
- [ ] **Single-process containers** (or `--init` / `tini` if spawning children)?

## Scanning & Registry

- [ ] **Trivy scanning in CI** on every push?
  - [ ] `trivy image --severity HIGH,CRITICAL --exit-code 1 ...`
  - [ ] SBOM generation: `trivy image --format spdx-json --output sbom.json ...`
- [ ] **`.trivyignore` file created** for documented exceptions (not silent allowlists)?
- [ ] **Production images tagged by digest** (`@sha256:...`), not mutable tags?
  - [ ] Human-facing tags follow semver (1.2.3) + moving `latest`
  - [ ] CI also pushes `:<short-sha>` for traceability

## Dev vs Prod

- [ ] **`docker-compose.yaml`** = production-shaped baseline?
  - Digest-pinned images, no source bind mounts, prod env defaults
- [ ] **`docker-compose.override.yaml`** = dev-only additions?
  - Bind-mount source for hot reload, expose debugger ports, `.env.local` for secrets
- [ ] **Test: `docker compose -f docker-compose.yaml up`** skips override and runs prod-only?

---

**Before first deployment, review [SKILL.md](SKILL.md) § 1–11 and language-specific recipes in [RECIPES.md](RECIPES.md).**
