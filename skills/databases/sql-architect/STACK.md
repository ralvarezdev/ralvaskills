# Stack Versions

| Dependency | Pinned version | Purpose |
|---|---|---|
| postgresql | 18.4 | **Primary** target engine; native `uuidv7()`, declarative partitioning, RLS |
| mysql | 9.7 | Secondary — covered with inline notes where it diverges |
| sqlite | 3.53 | Secondary — embedded / single-process use |
| golang-migrate | 4.19 | Migration runner for Go projects — raw `.sql` files, forward-only |
| alembic | 1.18 | Migration runner for Python projects — raw `.sql` files supported |
| pgbouncer | 1.24 | Connection pooler in front of Postgres (transaction mode for OLTP) |

## Notes

- **Soft delete is the default.** Hard delete only for GDPR/compliance requirements or append-only event tables.
- **Surrogate UUID v7 primary keys** + `UNIQUE` on natural identifiers. Avoid `BIGSERIAL` unless storage cost is genuinely the bottleneck on FK-heavy tables.
- **Forward-only migrations.** No down migrations in production.
- **`timestamptz` over `timestamp`** on PostgreSQL. Always store with timezone.
- **JSONB only for genuinely schemaless data.** Anything structured belongs in normalized columns.

_Last reviewed: 2026-05-20_
_Skill version at last review: 1.0.0_

Plan for `docker-architect`. Five quick decisions, then I'll execute.

## Plan: `docker-architect`

**Location:** `skills/infra/docker-architect/`
**Files:** SKILL.md + STACK.md (infra/ requires STACK.md)
**Target:** ~1000 tokens, 10 sections

### Section outline

1. Dockerfile fundamentals (multi-stage, layer order, `.dockerignore`, non-root USER, HEALTHCHECK)
2. Base image selection
3. BuildKit features (cache mounts, secrets, no `apt` cache bloat)
4. Image security (digest pinning, no baked secrets, supply-chain)
5. Compose patterns (v2, healthchecks + `depends_on` conditions, networks, volumes)
6. Runtime defaults (resource limits, logging, restart policy, init)
7. Language-specific recipes (Go, Python, Node)
8. Dev vs prod (compose override files / profiles)
9. Registry & tagging (sha digests for prod, semantic for humans)
10. Vulnerability scanning

### Five decisions

**1. Base image strategy per language?**
- **A. Per-language defaults** (recommended):
  - Go → **distroless static** (`gcr.io/distroless/static-debian12:nonroot`) for static binaries; `scratch` only when you've audited CA certs / tzdata manually
  - Python → **debian slim** with `uv` multi-stage (cleaner than alpine for Python because musl breaks several scientific wheels)
  - Node → **debian slim** (alpine breaks native modules too often)
- **B. Alpine everything** — small but musl pitfalls
- **C. Debian slim everything** — predictable, slightly bigger

I'd pick **A**.

**2. Multi-arch builds by default?**
- **A.** `linux/amd64` + `linux/arm64` (covers cloud + Apple Silicon dev)
- **B.** `linux/amd64` only

I'd pick **A** in 2026 — arm64 is mainstream.

**3. Vulnerability scanning tool?**
- **A. Trivy** (open-source, fast, broad coverage)
- **B. Grype** (anchore, similar)
- **C. Snyk** (commercial, broader integration)
- **D. Skip — leave to project**

I'd pick **A** (Trivy) — pin it in STACK.md, mention as the default CI step.

**4. Compose dev/prod split?**
- **A.** `compose.yaml` (prod baseline) + `compose.override.yaml` (dev-only) — the standard pattern, auto-merged
- **B.** Use **profiles** (`profiles: [dev]`) inside one file
- **C.** No opinion — leave to project

I'd pick **A** — fewer surprises, cleaner separation.

**5. `compose` vs `docker compose` vs `docker-compose`?**
Standard since 2024 has been `docker compose` (the v2 plugin, single binary, integrated). The standalone `docker-compose` is deprecated. Confirm we lock in v2-only?

---

Quick A/B/C for each (1-5), then I'll execute.

docker-architect` v1.0.0 done.

## Summary

| File | Size | Detail |
|---|---|---|
| `skills/infra/docker-architect/SKILL.md` | ~1500 tokens | 11 sections — Dockerfile fundamentals, base images, BuildKit, security, multi-arch, Compose v2, runtime, dev/prod, registry/tags, Trivy, language recipes |
| `skills/infra/docker-architect/STACK.md` | ~250 tokens | 8 pinned items (Docker 29.5, Compose v2, Buildx 0.34, Trivy 0.70, dockerfile syntax 1.10, distroless / debian-slim / node-slim bases) |
| `docs/SPECS.md` | — | Folder structure marker + roadmap entry updated to ✅ v1.0.0 |

## What it enforces (all per your choices)

- **Multi-stage always.** Builder stage with toolchain, runtime stage with only artifacts.
- **Per-language base defaults:** Go → distroless static; Python → `python:3.14-slim` + `uv`; Node → `node:22-slim`.
- **Digest-pinned bases** in production (`@sha256:...`).
- **BuildKit cache mounts and secret mounts** as the default; never `COPY` secrets.
- **Multi-arch:** `linux/amd64 + linux/arm64` for every build.
- **Compose v2 only:** `docker compose` subcommand, `docker-compose.yaml` + `docker-compose.override.yaml` for dev.
- **Non-root USER** in every final stage; `HEALTHCHECK` on every long-running service.
- **Trivy** scanning in CI; fail on HIGH/CRITICAL.
- **Three concrete recipes** in §11 (Go, Python, Node) — ready-to-copy Dockerfiles.

## Body size: ~1500 tokens

In line with other architect skills. Could trim by moving §11 (recipes) to a `RECIPES.md` reference (-400 tokens), but the inline recipes are genuinely the most useful part for a working engineer — I'd leave them.

---

Next per the order: **`fastapi-architect`** (Python context still warm). Ready?