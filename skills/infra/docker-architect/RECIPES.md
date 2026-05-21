# Docker Recipes

Reference Dockerfiles for the per-language patterns in [SKILL.md](SKILL.md).

## Go (multi-stage, distroless)

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

## Python (multi-stage, uv, debian slim)

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

## Node (multi-stage, debian slim, npm ci)

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

## BuildKit features (reference)

- **Cache mounts** survive layer invalidation:
  ```dockerfile
  RUN --mount=type=cache,target=/root/.cache/uv uv sync --frozen
  ```
- **Secret mounts** — never `COPY` secrets into layers:
  ```dockerfile
  RUN --mount=type=secret,id=npmrc,target=/root/.npmrc npm ci
  ```
- **Bind mounts** to read source without `COPY`:
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

## Multi-arch build

```bash
docker buildx create --use --name multiarch
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag ghcr.io/org/app:1.2.3 \
  --push \
  .
```

Cache via `--cache-to type=gha` (GitHub Actions) or `type=registry,ref=ghcr.io/org/app:buildcache`.

## Trivy scan (CI)

```bash
trivy image --severity HIGH,CRITICAL --exit-code 1 ghcr.io/org/app:${SHA}
trivy image --format spdx-json --output sbom.json ghcr.io/org/app:${SHA}
```
