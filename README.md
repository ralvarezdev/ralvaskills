# ralvaskills

**Note:** This is my personal toolkit for daily development, made public for anyone who wants to use, fork, or learn from it.

A growing collection of opinionated, Staff-level skills for Claude Code and OpenCode.

By default, AI coding agents produce functional but generic code. They nest logic too deeply, ignore architectural patterns, and write verbose commit messages. These skills enforce strict standards — functioning as an automated Senior Engineer that demands clean, maintainable, and highly optimized code.

See [`docs/SPECS.md`](docs/SPECS.md) for the full technical specification, authoring guide, and roadmap.

---

## Skill Library

Skills are grouped by what they shape. Every folder with a `SKILL.md` is a skill; folders containing other folders are categories.

### `languages/` — language standards

- **`go-architect`** — Go 1.26 standards: memory-aligned structs, typed enums, interface design, goroutine safety, iterators, idiomatic errors, sqlx + `//go:embed` SQL pattern.
- **`python-architect`** — Python 3.14 enterprise standards: modern typing (PEP 649), immutable dataclasses, Protocol-based DI, asyncio discipline, pytest 9, psycopg + `.sql` files via `importlib.resources`.

### `frameworks/` — framework-specific patterns

- **`fastapi-architect`** — FastAPI 0.136 on Python 3.14: feature layout, Pydantic v2 request/response separation, async DI with lifespan, URL-prefix versioning, RFC 7807 errors, OAuth2+JWT.
- **`gin-architect`** — Gin 1.12 on Go 1.26: feature layout, struct-tag validation, RFC 7807 errors, JWT/IdP auth, route groups for versioning, OpenAPI.
- **`nethttp-architect`** — Go stdlib `net/http` (1.22+ ServeMux, no router): feature layout, struct-tag validation, RFC 7807, JWT, graceful shutdown, OpenAPI via kin-openapi.
- **`nextjs-architect`** — Next.js 16 with React 19, App Router only: server components by default with explicit `"use client"` boundaries, server actions, streaming Suspense, edge vs node runtime.
- **`react-architect`** — React 19 + TypeScript strict: feature-based components, hooks-first composition, TanStack Query for server state, zustand for cross-tree client state, Suspense + ErrorBoundary, Radix for a11y.

### `protocols/` — wire-level conventions

- **`rest-api-architect`** — Cross-language REST conventions: resource URLs, method semantics, URL-prefix versioning, cursor pagination, snake_case JSON, ISO 8601 timestamps, RFC 7807, Idempotency-Key, ETag/If-Match, OpenAPI as source of truth.
- **`grpc-architect`** — Vanilla gRPC: `.proto` services, `status.Error` with standard codes, domain→code mapping, interceptor chain (auth/log/recovery/validation/metrics), client deadlines, context propagation, bufconn testing.

### `encoding/` — schemas and wire formats

- **`protobuf-architect`** — Protocol Buffers (proto3): Buf-style package naming, protovalidate (CEL), buf toolchain for lint/breaking/generation, field-number reservation, well-known types.

### `databases/` — persistence

- **`sql-architect`** — SQL standards: UUID v7 PKs, snake_case, soft delete, forward-only migrations, parameter binding, N+1 prevention, EXPLAIN-driven indexing. PostgreSQL 18 primary; MySQL 9 / SQLite 3.53 noted.

### `messaging/` — event-driven systems

- **`event-driven-architect`** — Broker-agnostic event taxonomy (NATS/Kafka/RabbitMQ): event/command schemas in Protobuf, topic naming, mandatory outbox, partitioning, idempotency, DLQs, schema evolution.

### `frontend/` — UI and accessibility

- **`ui-ux-architect`** — WCAG 2.2 AA, Radix + Tailwind 4 + shadcn/ui, design-token theming, mandatory loading/error/empty/success states, mobile-first responsive, keyboard parity, contrast checked in CI.

### `design/` — architectural patterns

- **`ddd-architect`** — Domain-Driven Design: strategic-first (bounded contexts, context mapping) with aggregates, value objects, domain events, repositories. Event sourcing out of scope.
- **`hexagonal-arch`** — Ports & adapters: dependencies point inward; domain declares ports, adapters implement them; domain never imports framework/DB/HTTP.

### `infra/` — runtime, packaging, observability

- **`docker-architect`** — Multi-stage builds, per-language base defaults (distroless Go, slim Python/Node), BuildKit cache mounts, non-root, multi-arch amd64+arm64, digest-pinned bases, Trivy scanning, Compose v2.
- **`observability-architect`** — Application-side signal production: structured logs, Prometheus metrics, OTel traces, signal correlation, head sampling, PII discipline, RED+USE.
- **`grafana-architect`** — Signal consumption: dashboards-as-code (Grizzly), per-service folders, one-question-per-panel, unified alerting with runbooks, low-cardinality discipline.

### `robotics/` — robotics platforms

- **`ros2-architect`** — ROS2 across Jazzy/Kilted/Lyrical: colcon + ament_cmake/ament_python, lifecycle nodes, explicit QoS, services/actions, parameters, Python launch DSL. C++ (`rclcpp`) and Python (`rclpy`) equal first-class. Cross-platform dev env via Pixi + RoboStack.

### `quality/` — reviewers

- **`api-contract-reviewer`** — Reviews REST + gRPC contracts for stability, versioning, completeness, backwards compatibility. Runs `buf breaking` / `openapi-diff`. Severity-keyed findings.
- **`performance-reviewer`** — Cross-language perf review: N+1, missing indexes, blocking I/O in async, allocation hot paths, unbounded memory. Findings grounded in EXPLAIN / pprof / py-spy / metrics.
- **`security-reviewer`** — Cross-language security review: injection, auth/authz, secrets, insecure defaults, deserialization, CSRF/SSRF/IDOR, dep vulns. Critical/High/Medium/Low with file:line + fixes.

### `tooling/` — developer tooling

- **`cli-tool-architect`** — Cross-language CLI standards: subcommand structure, flag/env/config/default precedence, TOML in XDG, stdout-data/stderr-logs split, `--output json|yaml`, exit codes, NO_COLOR, completions.
- **`repo-tooling-architect`** — Repo-root productivity files: `.editorconfig`, `.gitignore`, version pinning (mise), task runner (Task), minimal pre-commit, dotenv + external secret manager, Renovate.

### `workflows/` — process

- **`feature-planner`** — Upstream feature planning: requirement clarification, constraint discovery, design outline, vertical-slice task breakdown ordered by risk. Hands off to `grill-with-docs` and `tdd`.
- **`grill-with-docs`** — Stress-tests plans against the project's domain model, sharpens terminology, updates CONTEXT.md and ADRs inline as decisions crystallise.
- **`tdd`** — Red-green-refactor loop emphasizing behavior verification through public interfaces.
- **`commit-author`** — Generates concise Conventional Commits messages from a staged diff. Enforces full type set, imperative subject lines, no AI co-author attribution.

### `refactoring/` — code cleanup ladder

Three rungs from expression-level polish to system-level architecture, plus a skeptical pattern catalog. Each rung explicitly defers up or down when scope changes.

- **`logic-cleaner`** — Expression-level cleanup: guard clauses, naming, magic values, conditional simplification, duplication. Language-agnostic.
- **`code-design-refactor`** — Design-level refactoring: extraction, decoupling, SRP, encapsulation, primitive obsession. Sits between `logic-cleaner` and `improve-codebase-architecture`.
- **`improve-codebase-architecture`** — System-level architectural friction analysis with deepening opportunities informed by CONTEXT.md and ADRs.
- **`design-patterns`** — Skeptical reference catalog for modern Go and Python: keeps repository, adapter, strategy, decorator, observer, builder; flags the rest as anti-patterns. Consult before applying a pattern in the other refactoring skills.

### `meta/` — ralvaskills ecosystem

Skills about authoring/operating the toolkit itself, plus communication modes that aren't really workflows.

- **`skill-builder`** — Meta-skill that scaffolds new ralvaskills skills per SPECS.md (SKILL.md skeleton, optional STACK/RECIPES, folder placement, SPECS updates). Interview-first.
- **`rsk-guide`** — Quick reference for the `rsk` CLI (in progress): discover, install, update, check ralvaskills bundles and skills.
- **`caveman`** — Ultra-compressed communication mode that reduces token usage by ~75% while preserving technical accuracy.

### `personal/` — author-specific, not bundle-installable

These are personal to me and excluded from bundle installs. Listed here for transparency; not intended for reuse.

- **`demo-presentation-architect`** — Slide-deck specifications in markdown: content distribution across slides, within-slide organization, layout from a fixed catalog. Outputs `.md` only.
- **`demo-script-architect`** — Presenter-centric demo scripts with narrative flow, visual-cue mapping, progressive capability reveals.
- **`uru-thesis-reviewer`** — Continuous-feedback reviewer for URU (Universidad Rafael Urdaneta) Trabajos Especiales de Grado. Produces ordered `.md` diff files; never edits the source `.docx`.
- **`work-report-generator`** — Generates formal daily work reports from unstructured input; persistent project catalog, per-day `LOG.md` + `REPORT.md`.

---

## `rsk` CLI (in progress)

A Go CLI is in development to manage skill installation across projects without copy-pasting folders. Tracked in [`docs/SPECS.md`](docs/SPECS.md).

```
go install github.com/ralvarezdev/ralvaskills/cmd/rsk@latest
rsk init
rsk install <bundle>
rsk update
```

Until shipped, use the manual installation steps below.

---

## Recommended Official Claude Code Skills

These official Claude Code skills complement the ones above and are highly recommended:

- **`frontend-design`** — Distinctive, production-grade frontend interfaces with high design quality.
- **`xlsx`** — Spreadsheet files (`.xlsx`, `.xlsm`, `.csv`, `.tsv`) for data manipulation, analysis, formatting, conversion.

---

## Attribution

Several skills are adapted from [Matt Pocock's skills](https://github.com/mattpocock/skills):
- `tdd`
- `caveman`
- `improve-codebase-architecture`

The original implementations have been customized and integrated into this toolkit.

---

## Installation

These skills are drop-in modules for Claude Code and OpenCode.

### Global installation

Clone this repository and copy the skill folders to your global skills directory.

**Windows (PowerShell):**

```powershell
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.config\opencode\skills\"
Copy-Item -Path .\skills\* -Destination "$env:USERPROFILE\.config\opencode\skills\" -Recurse -Force

New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.claude\skills\"
Copy-Item -Path .\skills\* -Destination "$env:USERPROFILE\.claude\skills\" -Recurse -Force
```

**macOS / Linux:**

```bash
mkdir -p ~/.config/opencode/skills/
cp -R ./skills/* ~/.config/opencode/skills/

mkdir -p ~/.claude/skills/
cp -R ./skills/* ~/.claude/skills/
```

### Per-project installation

Drop individual skill folders into your project's `.claude/skills/` (Claude Code) or `.opencode/skills/` (OpenCode) directory. Each skill is self-contained — copy only what you need.
