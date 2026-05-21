---
name: docker-architect
version: 1.0.0
description: >
  Enforces strict Docker standards: multi-stage builds, per-language base-image
  defaults (distroless for Go, debian slim for Python/Node), BuildKit cache mounts,
  non-root runtime, multi-arch (amd64 + arm64) builds, digest-pinned bases,
  Trivy scanning, and Compose v2 with `docker-compose.yaml` + override files.
  Use when writing or reviewing Dockerfiles and Compose files, scaffolding a new
  service's container layer, or hardening an existing one.
---

# Docker Architecture & Container Standards

Targets **Docker Engine 29**, **Compose v2**, **BuildKit** (default). Commands use `docker compose` (v2 plugin, no hyphen). File names use `docker-compose.yaml`. See [STACK.md](STACK.md) for pinned tool versions.

## 1. Dockerfile fundamentals

- **Always multi-stage.** A build stage (toolchain + sources) and a final runtime stage that copies only the artifacts. Never ship the toolchain in the runtime image.
- **Layer order = least → most volatile.** Pin OS deps first, then language deps, then source. Source code changes invalidate the fewest layers possible.
- **`.dockerignore` is mandatory.** Excludes `.git/`, `node_modules/`, `.venv/`, build outputs, secrets, IDE files. Bad `.dockerignore` is the most common cause of bloated images and accidentally-leaked secrets.
- **Non-root USER.** Final stage runs as a dedicated, non-root user (UID ≥ 10000). Distroless `:nonroot` tag handles this; for Debian-based, `useradd -u 10001 -r app && USER 10001`.
- **`HEALTHCHECK`** on every long-running service. Use the simplest possible probe (HTTP `/healthz`, `pg_isready`, etc.). Compose `depends_on` conditions depend on healthchecks being correct.
- **No `RUN apt-get update` without install + cleanup in the same layer:** `RUN apt-get update && apt-get install -y --no-install-recommends X && rm -rf /var/lib/apt/lists/*`.

## 2. Base image selection

Per-language defaults:

| Language | Default base | Why |
|---|---|---|
| **Go** | `gcr.io/distroless/static-debian12:nonroot` | Static binary, ~2 MB image, non-root by default, no shell (smaller attack surface). Reach for `scratch` only after auditing CA certs + tzdata yourself. |
| **Python** | `python:3.14-slim` (debian slim) | `uv` in a builder stage, copy `.venv` to runtime. Avoid alpine — musl breaks several scientific wheels. |
| **Node** | `node:22-slim` (debian slim) | Alpine breaks too many native modules. LTS only in production. |

**Always pin by digest in production:** `FROM python:3.14-slim@sha256:abc...`. Tags are mutable; digests aren't. Refresh digests via Renovate / Dependabot.

## 3. BuildKit features

BuildKit is the default builder in Docker 29 — use its features deliberately.

- **Cache mounts** for package managers — they survive layer invalidation:
  ```dockerfile
  RUN --mount=type=cache,target=/root/.cache/uv \
      uv sync --frozen
  ```
- **Secret mounts** — never `COPY` secrets into layers:
  ```dockerfile
  RUN --mount=type=secret,id=npmrc,target=/root/.npmrc \
      npm ci
  ```
- **Bind mounts** to read source without `COPY` when the artifact alone is needed:
  ```dockerfile
  RUN --mount=type=bind,source=.,target=/src go build -o /out/app ./cmd/app
  ```
- **Here-docs** for multi-line scripts:
  ```dockerfile
  RUN <<EOF
  apt-get update
  apt-get install -y --no-install-recommends ca-certificates
  rm -rf /var/lib/apt/lists/*
  EOF
  ```

## 4. Image security

- **Digest-pin bases** in production Dockerfiles. Renovate updates them automatically; review the diff.
- **No secrets baked in.** Build-time secrets go through `--mount=type=secret`. Runtime secrets come from the orchestrator (env, mounted file, secrets manager).
- **Drop capabilities** at runtime when possible (`--cap-drop=ALL --cap-add=NET_BIND_SERVICE`).
- **Read-only root filesystem** for stateless services (`--read-only` + tmpfs for `/tmp`).
- **Distroless `:nonroot` or explicit `USER`** — never run as root in the final stage.
- **Single-process containers.** No init system unless the app forks (then use `--init` / `tini`).
- **Scan every image in CI** (see §10).

## 5. Multi-arch builds

Always build `linux/amd64` + `linux/arm64`. Cloud is largely arm64-friendly now (Graviton, Ampere); local dev on Apple Silicon is arm64-native.

```bash
docker buildx create --use --name multiarch
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag ghcr.io/org/app:1.2.3 \
  --push \
  .
```

Cache the build via `--cache-to type=gha` (GitHub Actions) or `type=registry,ref=ghcr.io/org/app:buildcache` for cross-runner reuse.

## 6. Compose patterns (v2)

- **File name:** `docker-compose.yaml` (long extension), with `docker-compose.override.yaml` for dev-only additions. Compose auto-merges them.
- **Command:** `docker compose up` (v2 plugin, integrated). The legacy `docker-compose` standalone binary is deprecated — don't use it.
- **No `version:` key.** Compose v2 ignores it; remove from any file you touch.
- **`depends_on` with conditions:**
  ```yaml
  services:
    app:
      depends_on:
        db:
          condition: service_healthy
  ```
  This requires `db` to define a working `healthcheck`. `depends_on` without `condition:` only orders startup — doesn't wait for readiness.
- **Named volumes for stateful data** (`db-data:`), bind mounts only for source-code hot-reload in dev.
- **Networks:** declare them explicitly; don't rely on the default network for anything non-trivial.
- **Secrets and configs:** use `secrets:` and `configs:` top-level blocks for production-shaped local runs.

## 7. Runtime defaults

- **Resource limits** on every service (`deploy.resources.limits.memory`, `cpus`). Unbounded containers eat hosts.
- **Logging driver:** `json-file` with size + count rotation, or `journald` on Linux hosts. Production typically forwards to a log aggregator.
- **Restart policy:** `unless-stopped` for long-running services; `no` for batch jobs.
- **Init process:** add `--init` (or `init: true` in Compose) when the app spawns child processes — prevents zombie processes.
- **TZ:** set `TZ=Etc/UTC` explicitly in the image; never rely on host timezone.

## 8. Dev vs prod

- `docker-compose.yaml` — production-shaped baseline (images by digest, no source bind mounts, prod env defaults).
- `docker-compose.override.yaml` — dev-only additions: bind-mount source for hot reload, expose ports for debuggers, pull secrets from `.env.local`.
- Compose loads both automatically with `docker compose up`. To run *prod-only*, use `docker compose -f docker-compose.yaml up` (skip the override).
- **Never bake dev conveniences into the main file.** That's how mounting `./` into prod ships.

## 9. Registry & tagging

- **Production deploys reference image digests** (`@sha256:...`), not tags. Tags are for humans; digests are for machines.
- **Human-facing tags follow semver** (`1.2.3`) plus a moving `latest` and `1.2` major/minor aliases for local convenience.
- **CI also pushes `:<short-sha>`** for traceability — easy to roll back to a specific commit.
- **Sign images** with Cosign for prod-bound registries. Verify on pull in the deployment platform.

## 10. Vulnerability scanning — Trivy

- **Default scanner: `aquasecurity/trivy`.** Open-source, fast, scans images + filesystems + IaC.
- **CI step on every push:**
  ```bash
  trivy image --severity HIGH,CRITICAL --exit-code 1 ghcr.io/org/app:${SHA}
  ```
- **Ignore file** (`.trivyignore`) for documented, accepted exceptions — never silent allowlists.
- **SBOM:** `trivy image --format spdx-json --output sbom.json ...` — attach to releases. Required for supply-chain compliance.

## 11. Language-specific recipes

### Go (multi-stage, distroless)

```dockerfile
# syntax=docker/dockerfile:1.10
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/app

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/app /app
USER nonroot:nonroot
ENTRYPOINT ["/app"]
```

### Python (multi-stage, uv, debian slim)

```dockerfile
# syntax=docker/dockerfile:1.10
FROM python:3.14-slim AS build
COPY --from=ghcr.io/astral-sh/uv:latest /uv /bin/uv
WORKDIR /app
ENV UV_LINK_MODE=copy
RUN --mount=type=cache,target=/root/.cache/uv \
    --mount=type=bind,source=uv.lock,target=uv.lock \
    --mount=type=bind,source=pyproject.toml,target=pyproject.toml \
    uv sync --frozen --no-install-project --no-dev
COPY . .
RUN --mount=type=cache,target=/root/.cache/uv uv sync --frozen --no-dev

FROM python:3.14-slim
RUN useradd -u 10001 -r app
COPY --from=build --chown=app:app /app /app
WORKDIR /app
USER 10001
ENV PATH="/app/.venv/bin:$PATH"
ENTRYPOINT ["python", "-m", "myapp"]
```

### Node (multi-stage, debian slim, npm ci)

```dockerfile
# syntax=docker/dockerfile:1.10
FROM node:22-slim AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci
COPY . .
RUN npm run build

FROM node:22-slim
RUN useradd -u 10001 -r app
WORKDIR /app
COPY --from=build --chown=app:app /app/dist /app/dist
COPY --from=build --chown=app:app /app/node_modules /app/node_modules
USER 10001
ENTRYPOINT ["node", "dist/index.js"]
```
