# ralvaskills — Technical Specification

> Personal, ever-growing collection of Staff-level AI skills for Claude Code and OpenCode.  
> Enforces strict clean architecture and professional standards across all projects.

---

## Table of Contents

1. [Overview](#overview)
2. [Repository & Folder Structure](#repository--folder-structure)
3. [Skill Authoring Guide](#skill-authoring-guide)
4. [CLI Design & Commands (`rsk`)](#cli-design--commands-rsk)
5. [Bundle & Skill Catalog](#bundle--skill-catalog)
6. [Roadmap](#roadmap)

---

## Overview

### Identity

| Thing | Value |
|---|---|
| GitHub repo | `ralvarezdev/ralvaskills` |
| Go module | `github.com/ralvarezdev/ralvaskills` |
| CLI binary | `rsk` |
| Install | `go install github.com/ralvarezdev/ralvaskills@latest` |

### Skill Sources

The CLI manages skills from two independent sources:

| Source ID | Origin | Update mechanism |
|---|---|---|
| `local` | This repo (`ralvarezdev/ralvaskills`) | `git pull` |
| `official` | Anthropic's repo (`anthropics/skills`) | Cached clone, re-fetched on demand |

- **Local skills** are symlinked from the cloned repo — updates propagate instantly across all projects on `git pull`.
- **Official skills** are fetched and cached at `~/.ralvaskills/cache/anthropic/` — never copied into this repo.

### Design Principles

- Skills are **focused** — each does one thing well, under ~500 tokens in the SKILL.md body.
- Skills are **composable** — they chain into workflows (see [Workflow Pipelines](#workflow-pipelines)).
- Skills that reference specific language, framework, or library versions **must** include a `STACK.md` file — the single source of truth for version tracking.
- Bundles are **declarative** — a bundle is a named list of skill refs; no duplication.
- The CLI is **config-driven** — repo paths and preferences live in `~/.config/rsk/config.json`, set once via `rsk init`.
- **Personal skills** (any skill whose path contains a `personal/` segment) are excluded from bundle installs by default — they are self-contained, non-reusable, and not intended for other developers.

---

## Repository & Folder Structure

```
ralvaskills/                          # current state — (📋) marks planned additions/moves
│
├── cli/                              # 📋 to be created — Go CLI source (binary: rsk)
│   ├── main.go
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   ├── root.go                   # cobra root, version flag
│   │   ├── init.go                   # rsk init [flags]
│   │   ├── install.go                # rsk install [bundle...] [flags]
│   │   ├── uninstall.go              # rsk uninstall [bundle|skill] [flags]
│   │   ├── update.go                 # rsk update [bundle|skill] [flags]
│   │   ├── status.go                 # rsk status [flags]
│   │   └── list.go                   # rsk list [flags]
│   └── internal/
│       ├── config/
│       │   ├── config.go             # load/save ~/.config/rsk/config.json
│       │   └── bundles.go            # bundle & skill ref definitions
│       ├── skill/
│       │   ├── skill.go              # Skill, Bundle, SkillRef structs
│       │   ├── registry.go           # walks skills/ root; any folder with SKILL.md is a skill (recursion stops); parses SKILL.md + STACK.md
│       │   ├── linker.go             # symlink create/remove logic
│       │   └── version.go            # semver comparison helpers
│       ├── source/
│       │   ├── local.go              # resolves skills from local repo clone
│       │   └── official.go           # fetches/caches from anthropics/skills
│       └── ui/
│           └── output.go             # colored terminal output helpers
│
├── skills/                           # all reusable local skills
│   ├── languages/
│   │   ├── go-architect/             # ✅ exists — Go 1.26 audit complete (v1.0.0); has STACK.md
│   │   └── python-architect/         # ✅ exists — Python 3.14 audit complete (v1.0.0); has STACK.md
│   ├── databases/
│   │   └── sql-architect/            # ✅ exists (v1.0.0 — PG 18 primary, MySQL/SQLite notes)
│   ├── frameworks/
│   │   ├── fastapi-architect/        # ✅ exists (v1.0.0 — FastAPI 0.136, feature-based, RFC 7807)
│   │   ├── gin-architect/            # ✅ exists (v1.0.0 — Gin 1.12 on Go 1.26)
│   │   ├── nethttp-architect/        # ✅ exists (v1.0.0 — stdlib net/http + Go 1.22+ enhanced ServeMux)
│   │   ├── react-architect/          # ✅ exists (v1.0.0 — React 19, TS strict, TanStack Query for server state, zustand when justified)
│   │   └── nextjs-architect/         # ✅ exists (v1.0.0 — Next 16, App Router only, RSC defaults, server actions, hybrid data access)
│   ├── protocols/
│   │   ├── rest-api-architect/       # ✅ exists (v1.0.0 — snake_case JSON, ISO 8601, RFC 7807 errors, mandatory Idempotency-Key & ETag)
│   │   └── grpc-architect/           # ✅ exists (v1.0.0 — vanilla gRPC, status codes, interceptor chain, deadlines, bufconn testing)
│   ├── encoding/
│   │   └── protobuf-architect/       # ✅ exists (v1.0.0 — proto3, Buf-style packages, protovalidate, breaking-change discipline)
│   ├── messaging/
│   │   └── event-driven-architect/   # ✅ exists (v1.0.0 — Protobuf schemas, outbox mandatory, broker-agnostic; NATS/Kafka/RabbitMQ)
│   ├── tooling/
│   │   ├── cli-tool-architect/       # ✅ exists (v1.0.0 — Go cobra+pflag+viper, Python typer; TOML/XDG config, --output, exit codes, NO_COLOR)
│   │   ├── repo-tooling-architect/   # ✅ exists (v1.0.0 — .editorconfig/.gitignore, mise|proto, Task|just, pre-commit, Renovate)
│   │   ├── rsk-guide/                # ✅ exists (v0.1.0 — tracks the planned rsk CLI)
│   │   └── skill-builder/            # ✅ exists (v1.0.0 — meta-skill: scaffolds new skills following ralvaskills standards; no STACK.md)
│   ├── design/
│   │   ├── ddd-architect/            # ✅ exists (v1.0.0 — strategic-first DDD, bounded contexts, aggregates; event sourcing OOS)
│   │   └── hexagonal-arch/           # ✅ exists (v1.0.0 — ports & adapters; dependency direction inward; no folder rule)
│   ├── ai-ml/
│   │   ├── llm-app-architect/        # 📋 planned
│   │   ├── agent-architect/          # 📋 planned
│   │   └── ml-pipeline-architect/    # 📋 planned
│   ├── infra/
│   │   ├── docker-architect/         # ✅ exists (v1.0.0 — Docker 29, Compose v2, distroless/slim defaults, multi-arch, Trivy)
│   │   ├── ci-cd-architect/          # 📋 planned
│   │   ├── grafana-architect/        # ✅ exists (v1.0.0 — dashboards-as-code via Grizzly, unified alerting, per-service folders)
│   │   └── observability-architect/  # ✅ exists (v1.0.0 — Prometheus metrics + OTel logs/traces, RED+USE, head sampling at 10%)
│   ├── robotics/
│   │   └── ros2-architect/           # ✅ exists (v1.0.0 — Jazzy/Kilted/Lyrical, Pixi + RoboStack env, C++/Python equal, lifecycle nodes, explicit QoS)
│   ├── quality/
│   │   ├── security-reviewer/        # ✅ exists (v1.0.0 — injection, auth, secrets, insecure defaults, OWASP-flavored)
│   │   ├── api-contract-reviewer/    # ✅ exists (v1.0.0 — REST + gRPC contract stability, buf breaking + openapi-diff)
│   │   └── performance-reviewer/     # ✅ exists (v1.0.0 — N+1, blocking I/O, allocation hot paths; measurement-grounded)
│   ├── frontend/
│   │   └── ui-ux-architect/          # ✅ exists (v1.0.0 — WCAG 2.2 AA, Radix + Tailwind 4, design tokens, four-state UI discipline)
│   └── workflows/
│       ├── commit-author/            # ✅ exists
│       ├── tdd/                      # ✅ exists
│       ├── grill-with-docs/          # ✅ exists (formerly grill-me; absorbed ubiquitous-language)
│       ├── caveman/                  # ✅ exists
│       ├── logic-cleaner/            # ✅ exists
│       ├── code-design-refactor/     # ✅ exists (v1.0.0 — design-level refactor; sits between logic-cleaner and improve-codebase-architecture)
│       ├── feature-planner/          # ✅ exists (v1.0.0 — requirement → constraints → design → vertical-slice tasks)
│       ├── improve-codebase-architecture/  # ✅ exists
│       └── design-patterns/          # ✅ exists (v1.0.0 — skeptical catalog; modern Go/Python; defers to ddd-architect & hexagonal-arch)
│
│   └── personal/                     # 📋 new category — personal skills, never auto-bundled
│       └── demo-script-architect/    # ✅ exists (moved from workflows/)
│
├── .github/
│   └── workflows/
│       └── release.yml               # 📋 to be created — builds rsk binaries on tag push
│
├── README.md
└── LICENSE
```

**Legend:** ✅ exists · 📋 planned/to be created · 🔀 needs to be moved from current location

### Discovery rule

**Skills are auto-discovered by `SKILL.md` presence.** The CLI walks the `skills/` root recursively; any folder containing `SKILL.md` is a skill and recursion stops there. Folders without `SKILL.md` are pure organizational hierarchy.

This means:

- **Adding a category** = `mkdir skills/<category>` — no spec change required.
- **Category labels** are inferred from `path.Dir` relative to the skills root; not centrally registered.
- **Skills can be nested at any depth** under `skills/`. The shape shown above is a recommended convention, not a constraint.
- **Nested `SKILL.md` is forbidden** — a skill folder must not contain a subfolder that also has its own `SKILL.md`. The CLI errors on detection: `✗ skills/<path>: nested SKILL.md at <subpath>; move the inner skill outside its parent`.
- **`personal/` opt-in is path-segment-based** — any skill whose path contains a `personal/` segment requires `--personal`. The literal location of `personal/` doesn't matter (currently `skills/personal/`).

**Pending migrations from current repo state:**
- _(none — all prior migrations applied)_

### Local Cache (outside repo)

All paths shown below are **defaults** — every one is configurable via `config.json` (see [`rsk init`](#configuration-rsk-init)).

```
~/.ralvaskills/                       # default base — overridable via config
  cache/
    anthropic/                        # default value of  official_cache
      skills/                         # cloned anthropics/skills repo
        docx/
        xlsx/
        pdf/
        pptx/
        frontend-design/
        find-docs/
      last-updated.json
    versions.json                     # default value of  versions_cache  (TTL: 24h)
~/.config/rsk/
  config.json                         # CLI config — path is fixed (XDG convention)
~/.claude/skills/                     # default  global_targets.claude-code
~/.config/opencode/skills/            # default  global_targets.opencode
```

> **Note:** Any skill whose path contains a `personal/` segment is never symlinked by `rsk install`. Personal skills must be installed explicitly via `rsk install --skill demo-script-architect --personal`. The rule is path-based, so the literal location of `personal/` doesn't matter — it currently lives at `skills/personal/`, but could be anywhere.

---

## Skill Authoring Guide

### Anatomy of a Skill

Every skill is a folder containing a required `SKILL.md` and optional supporting files:

```
skill-name/
├── SKILL.md          # required — frontmatter + instructions
├── STACK.md          # required for version-sensitive skills — dependency version registry
├── RECIPES.md        # optional — extracted code blocks / reference implementations (keeps SKILL.md lean)
├── scripts/          # optional — executable scripts (bash, python)
├── references/       # optional — docs loaded into context on demand
└── assets/           # optional — templates, examples
```

`RECIPES.md` is an optional side file used by code-heavy skills. The SKILL.md body keeps the rules, decisions, and one- or two-line snippets; long reference implementations (full Dockerfiles, middleware skeletons, test scaffolds, multi-file project trees) move to `RECIPES.md` and are linked from the relevant SKILL.md section. Goal: SKILL.md stays loadable and skim-friendly; the recipes load on demand when the user actually needs an implementation. Skip `RECIPES.md` for skills that are mostly conceptual prose — adding it just to split prose loses value.

### SKILL.md Format

The `SKILL.md` stays clean and focused on instructions. Stack version metadata lives in `STACK.md` — never in the skill body. Long reference implementations live in `RECIPES.md` — never inlined when they exceed a small snippet.

```markdown
---
name: skill-name
version: 1.0.0
description: One line. WHAT the skill enforces + WHEN to invoke (trigger phrases). This sits in Claude's always-loaded skill index, so length costs every turn.
---

# Skill Name

[Instructions Claude will follow when this skill is active.]
```

**Body discipline:**

- **Rules first, code second.** Each section explains the rule and the *why*; code is illustrative, not exhaustive. If a code block grows past ~15 lines or appears as a full reference implementation (a Dockerfile, an auth middleware, a project tree, a test scaffold), move it to `RECIPES.md` and replace it in `SKILL.md` with a one-line pointer like *"Skeleton in [RECIPES.md](RECIPES.md)."*
- **No duplicated cross-skill content.** If a pattern is canonically explained in another skill (auth in `rest-api-architect §11`, status codes in `rest-api-architect/STATUS_CODES.md`), link to it instead of restating. The link is the contract; restatement drifts.
- **Body target:** keep the SKILL.md body under ~10 KB / ~250 lines where the topic allows. Foundational skills (`go-architect`, `rest-api-architect`) may run longer because the conventions themselves are dense; that's fine. Framework / infra / encoding skills should almost always factor recipes out.

### STACK.md Format

Version-sensitive skills **must** include a `STACK.md` alongside `SKILL.md`. This file is for `rsk` tooling only — Claude never loads it unless explicitly asked. It is the single source of truth for what versions a skill was authored against.

> The HTML comment (`<!-- skill-name/STACK.md -->`) in the examples below is a documentation label only — **omit it in your actual STACK.md**.

```markdown
# Stack Versions

<!-- go-architect/STACK.md -->
| Dependency | Pinned version |
|---|---|
| go | 1.24 |
| viper | 2.0 |
| validator | 10.22 |
| zap | 1.27 |
| fx | 1.22 |
| cobra | 1.8 |
| sqlx | 1.3 |
| migrate | 4.18 |

_Last reviewed: 2026-05-20_
_Skill version at last review: 1.0.0_
```

```markdown
# Stack Versions

<!-- fastapi-architect/STACK.md -->
| Dependency | Pinned version |
|---|---|
| python | 3.13 |
| fastapi | 0.115 |
| pydantic | 2.9 |
| pydantic-settings | 2.6 |
| sqlalchemy | 2.0 |
| alembic | 1.14 |
| uvicorn | 0.32 |
| httpx | 0.28 |
| ruff | 0.8 |
| mypy | 1.13 |
| uv | 0.5 |

_Last reviewed: 2026-05-20_
_Skill version at last review: 1.0.0_
```

`STACK.md` serves two purposes:
- **`rsk status --stack`** reads it and fetches latest versions from **`proxy.golang.org`** (Go modules) and **`pypi.org`** (Python packages), surfacing a `⚠ stack may be outdated` warning per skill. Fetched data is cached for 24 h in `~/.ralvaskills/cache/versions.json`. **Fetching only happens when the user explicitly passes `--stack`** — `rsk` never reaches the network in the background. Pass `--refresh` to bypass the cache and force a re-fetch.
- **You**, when auditing — `_Last reviewed_` and `_Skill version at last review_` tell you exactly when it was last checked and whether it has drifted from the current skill version.

When updating a skill due to a stack change, always update `STACK.md` to match, refresh both metadata lines, and add a note in the skill body describing what changed.

### Which Skills Need a `STACK.md`

This table is a **heuristic keyed on the skill's parent folder name** — it's not enforced. Override per-skill when the skill genuinely has or lacks version-sensitive content.

| Parent folder | Needs `STACK.md`? | Reason |
|---|---|---|
| `languages/` | ✅ Yes | Language version drives idioms and stdlib usage |
| `databases/` | ✅ Yes | DB engine version drives query syntax, JSON support, index types — pin to target engine (PostgreSQL, MySQL, SQLite) |
| `frameworks/` | ✅ Yes | Framework APIs change significantly between versions |
| `protocols/` | ✅ Yes | Tooling versions matter (e.g. `buf`, gRPC spec) |
| `encoding/` | ✅ Yes | Proto spec version, `buf`, `protovalidate` |
| `messaging/` | ✅ Yes | Kafka/RabbitMQ client library versions |
| `tooling/` | ✅ Yes | Cobra, Click, Typer versions — **exception:** meta-skills with no library stack (e.g. `skill-builder`) are exempt |
| `ai-ml/` | ✅ Yes | Library APIs change rapidly (LangChain, LlamaIndex) |
| `infra/` | ✅ Yes | Docker, compose spec, GitHub Actions runner versions |
| `robotics/` | ✅ Yes | ROS2 distro (Jazzy, Kilted, etc.) |
| `frontend/` | ✅ Yes | React, Next.js versions evolve quickly |
| `design/` | ❌ No | Architectural patterns are version-agnostic |
| `quality/` | ❌ No | Security/performance principles don't change with versions — add one if the skill references framework-specific vulnerability patterns |
| `workflows/` | ❌ No | Workflow skills are tool/language agnostic |
| `personal/` | Depends | Add if the skill references specific tooling |

### Naming Convention

All skill folder names follow `kebab-case` and end in a meaningful suffix:

| Suffix | Meaning | Examples |
|---|---|---|
| `-architect` | Enforces structure/standards for a language, framework, or protocol | `go-architect`, `fastapi-architect`, `grpc-architect` |
| `-arch` | Enforces an architectural pattern | `hexagonal-arch` |
| `-reviewer` | Reviews existing code/contracts for issues | `security-reviewer`, `api-contract-reviewer` |
| `-refactor` | Restructures existing code | `code-design-refactor` |
| `-planner` | Designs something before implementation | `feature-planner` |
| `-patterns` | Catalog of patterns or principles for a domain | `design-patterns` |
| `-builder` | Scaffolds new artifacts of a given type (meta-tooling) | `skill-builder` |
| `-guide` | Documentation or usage guide for a specific tool | `rsk-guide` |
| (none) | Workflow tools that don't fit above | `tdd`, `commit-author`, `caveman` |

### Version Field

Every `SKILL.md` must include a `version` field in semver format:

| Bump | When |
|---|---|
| `patch` (1.0.x) | Clarifications, wording fixes, minor library version update with no behavioral change |
| `minor` (1.x.0) | New rules or sections added, library upgraded with new conventions enforced |
| `major` (x.0.0) | Breaking change to skill behavior, or major language/framework version upgrade (e.g. Python 3.12 → 3.13, FastAPI 0.x → 1.x) |

### Description Quality

The `description` field is the most critical part of `SKILL.md`. Claude reads **only** this (plus the name) at startup and decides whether to load the full skill body based solely on it. **It is loaded into every turn's context** — verbose descriptions cost tokens *forever*, not just when the skill is invoked.

**Rules:**

- **One line.** Use the YAML scalar form (no `>`/`|`). The line can be long; line breaks aren't necessary.
- **WHAT then WHEN.** Lead with what the skill enforces; close with trigger phrases the user might say or commands they might invoke.
- **Mention version for version-sensitive skills.** "Go 1.26", "Python 3.14", "Gin 1.12 on Go 1.26" — signals upgrade-sensitivity to both Claude and you.
- **Drop marketing copy.** "Enforces strict X standards" is the same as "X standards". "Comprehensive guide to" is filler.

**Bad:**
```yaml
description: Go best practices skill.
```

**Bad (verbose):**
```yaml
description: >
  Enforces strict architectural standards for Go 1.26 — memory-aligned structs,
  typed enums, interface design philosophy, goroutine safety, iterator patterns,
  idiomatic error handling, and the sqlx + `//go:embed` SQL pattern. Use when
  writing, reviewing, or refactoring Go code, scaffolding a new Go service,
  or auditing an existing codebase against modern idioms.
```

**Good:**
```yaml
description: Go 1.26 architectural standards — memory-aligned structs, typed enums, interface design, goroutine safety, iterators, idiomatic errors, sqlx + //go:embed SQL pattern. Use when writing, reviewing, or scaffolding Go code.
```

### Supporting Files

- `RECIPES.md` — extracted code blocks / reference implementations referenced from `SKILL.md`. Loaded on demand by the user / Claude when an implementation is needed; never auto-loaded. See [Body discipline](#skillmd-format) above for when to factor recipes out.
- `scripts/` — scripts Claude can execute via bash. Only their **output** enters context, not the source.
- `references/` — markdown docs Claude loads on demand when the task requires deeper context.
- `assets/` — templates or boilerplate Claude uses to generate output.

All `.md` files inside a skill folder **must** use `UPPER_SNAKE_CASE` (e.g. `DEEP_MODULES.md`, `ADR_FORMAT.md`, `RECIPES.md`). `README.md` at the repo root is the only exception — it follows the universal tool convention.

### Canonical Libraries Reference

Language and framework architect skills must explicitly reference the canonical libraries for their stack. Claude should enforce their use over reinventing equivalent functionality.

#### Go canonical libraries

| Concern | Library |
|---|---|
| Config & env | `spf13/viper` |
| Struct validation | `go-playground/validator` |
| CLI | `spf13/cobra` + `spf13/pflag` |
| HTTP router | `gin-gonic/gin` (REST) |
| ORM / query builder | `jmoiron/sqlx`, `uptrace/bun`, or `go-gorm/gorm` |
| gRPC | `google.golang.org/grpc` |
| Protobuf | `google.golang.org/protobuf`, `buf` toolchain |
| Logging | `log/slog` (stdlib, **default**); `uber-go/zap` or `rs/zerolog` only when slog's allocator perf is provably insufficient |
| Testing | `testify/suite`, `testify/mock` |
| DI | `uber-go/fx` or `google/wire` |
| Migrations | `golang-migrate/migrate` |
| Linting | `golangci-lint` v2 (aggregates staticcheck, govet, revive, goimports, etc.) |

#### Python canonical libraries

| Concern | Library |
|---|---|
| Data validation & settings | `pydantic` v2, `pydantic-settings` |
| Web framework | `fastapi` |
| ASGI server | `uvicorn` |
| DB driver (recommended) | `psycopg` 3 — sync + async, used with `.sql` files via `importlib.resources` (no ORM by default) |
| ORM (when justified) | `sqlalchemy` 2.x (Core or ORM); record the choice in an ADR |
| Migrations | `alembic` (works with raw SQL — no ORM required) |
| HTTP client | `httpx` |
| Testing | `pytest` 9, `pytest-asyncio` |
| Type checking | `mypy` 2 (`--strict` baseline) |
| Linting / formatting | `ruff` (replaces black / isort / flake8 / pyupgrade) |
| Environment + packaging | `uv` (replaces pip / virtualenv / pyenv) |
| CLI | `typer` |
| Task queue (situational) | `arq` (asyncio-native, Redis) or `celery` (mature, broker-agnostic) — add when a real use case appears |

---

## CLI Design & Commands (`rsk`)

### Configuration (`rsk init`)

Run once per machine. Writes `~/.config/rsk/config.json`.

```bash
rsk init
```

Prompts for:
- Path to local `ralvaskills` repo clone
- Which AI tools to support (multi-select: `claude-code`, `opencode`)
- For each selected tool, the global skills directory (defaults shown):
  - `claude-code` → `~/.claude/skills/`
  - `opencode` → `~/.config/opencode/skills/`
- Default behavior for `--global` (`all` enabled tools, or a single tool)

`config.json` shape:
```json
{
  "repo_path": "/home/user/ralvaskills",
  "global_targets": {
    "claude-code": "~/.claude/skills/",
    "opencode": "~/.config/opencode/skills/"
  },
  "default_target_scope": "all",
  "official_cache": "~/.ralvaskills/cache/anthropic/",
  "versions_cache": "~/.ralvaskills/cache/versions.json"
}
```

- `global_targets` — a map of every AI tool the user wants to support, with its skills directory. Add/remove keys to enable/disable tools.
- `default_target_scope` — when `--global` is passed without `--for`, install to `all` configured targets, or pin to a single tool name (e.g. `"claude-code"`).

**Per-command targeting:** all commands that act on global skills (`install`, `update`, `uninstall`, `status`) accept a `--for <tool>` flag to scope the operation to a single configured target, overriding `default_target_scope`. Examples:
```bash
rsk install global --global                            # uses default_target_scope (all by default)
rsk install go-grpc --global --for claude-code         # Claude Code only
rsk status --global --for opencode                     # OpenCode global skills only
```

> **CLI implementation note:** use `spf13/viper` to load this config (supports env var overrides via `RSK_*` prefixed vars) and `go-playground/validator` to validate required fields on load. This mirrors the canonical Go stack enforced by `go-architect`.

---

### Install

```bash
rsk install [bundle...] [flags]

Flags:
  --global            Install to global skills dir(s) instead of ./.claude/skills/
  --for <tool>        Scope --global to a single configured tool (claude-code|opencode)
  --skill <name>      Install a single skill by name (skip bundle)
  --personal          Allow installing from personal/ folder (opt-in)
  --version <v>       Pin to a specific repo tag (local skills only — official skills are not co-versioned)
  --dry-run           Show what would be installed without doing it
```

Examples:
```bash
rsk install global --global                            # global skills, machine-wide
rsk install go-grpc                                    # stack bundle for current project
rsk install design docs                                # multiple bundles at once
rsk install --skill go-architect                       # single skill by name
rsk install go-grpc --version v1.1.0                  # pin to specific version
rsk install --skill demo-script-architect --personal   # personal skill, explicit opt-in
```

**Behavior when a bundled skill is not yet created:**
Skills marked as planned in a bundle are silently skipped with a warning — the install continues for all available skills:
```
⚠ llm-app-architect is not yet available (planned). Skipping — it will be included automatically once released.
✓ python-architect installed
✓ agent-architect skipped — not yet available (planned)
```

**Behavior of `--version` with official skills:**
`--version` is only valid for local skills/bundles (the `ralvaskills` repo is co-versioned). Official skills come from `anthropics/skills`, where each skill is independently versioned, so a single tag does not apply. Attempting to use `--version` with anything that resolves to an official skill errors out:
```
$ rsk install docs --version v1.0.0
✗ --version is only supported for local skills/bundles.
  Official skills (from anthropics/skills) are not co-versioned and always install at HEAD.
  Drop --version to install the latest of each.
```

---

### Update

```bash
rsk update [bundle|skill] [flags]

Flags:
  --global          Target global skills dir(s)
  --for <tool>      Scope --global to a single configured tool (claude-code|opencode)
  --official        Also re-fetch official (anthropics/skills) cache
  --personal        Include personal/ skills in update
  --dry-run         Show what would change
```

Examples:
```bash
rsk update                          # git pull + re-symlink all installed skills
rsk update --official               # re-fetch anthropics/skills cache
rsk update docs                     # re-fetch only the docs bundle
rsk update --skill grpc-architect   # update one skill
```

Since local skills are symlinks, `rsk update` is essentially a `git pull` on the repo — all projects using symlinks get the update for free.

---

### Status

```bash
rsk status [flags]

Flags:
  --global          Show global skills only
  --for <tool>      Scope --global to a single configured tool (claude-code|opencode)
  --project         Show project skills only
  --stack           Fetch latest versions (proxy.golang.org / PyPI) and show per-skill STACK.md drift (opt-in; no network calls otherwise)
  --refresh         With --stack: bypass the 24h cache and force a re-fetch
  --personal        Include personal/ skills in output
```

Example output:
```
ralvaskills v1.2.0  (github.com/ralvarezdev/ralvaskills)

Global (~/.claude/skills/)
  [ralva]  commit-author        v1.1.0  ✓ up to date
  [ralva]  tdd                  v1.0.0  → v1.1.0 available
  [anthr]  docx                 v2.3.1  ✓ up to date     [docs]
  [anthr]  xlsx                 v2.3.1  ✓ up to date     [docs]
  [anthr]  frontend-design      v1.0.0  ✓ up to date     [design]

Project (./.claude/skills/)  [bundles: go-grpc]
  [ralva]  go-architect         v1.2.0  ✓ up to date     [stack: go=1.24]
  [ralva]  grpc-architect       v1.0.0  → v1.2.0 available
  [ralva]  protobuf-architect   v1.0.0  ✓ up to date     [stack: buf=1.47]
  [ralva]  docker-architect     v1.1.0  ✓ up to date
  [ralva]  sql-architect        v1.0.0  ✓ up to date

⚠  1 skill may reference outdated stack versions. Run 'rsk status --stack' for details.
Run 'rsk update' to update all outdated skills.
```

---

### List

```bash
rsk list [flags]

Flags:
  --bundle <name>   Show skills in a specific bundle
  --source <s>      Filter by source: local | official
  --installed       Show only installed skills
  --personal        Include personal/ skills in listing
```

---

### Uninstall

```bash
rsk uninstall [bundle|skill] [flags]

Flags:
  --global          Target global skills dir(s)
  --for <tool>      Scope --global to a single configured tool (claude-code|opencode)
  --personal        Allow uninstalling from personal/ folder (must match install opt-in)
```

---

## Bundle & Skill Catalog

### Bundle Naming Convention

Bundle names follow `kebab-case`. Stack-specific bundles are named after their primary framework or protocol, not the language (e.g. `fastapi`, not `python-web`).

---

### Official Bundles (source: `official`)

Skills fetched from `github.com/anthropics/skills`. Never stored in this repo.

#### `docs`
Document creation, editing, and technical reference retrieval skills.

| Skill | Description |
|---|---|
| `docx` | Create and edit Word documents |
| `xlsx` | Work with spreadsheets and data |
| `pdf` | Create, fill, and manipulate PDFs |
| `pptx` | Build presentation slide decks |
| `find-docs` | Retrieve authoritative API references, library docs, and code examples for any developer technology |

#### `design` (mixed: official + local)
UI/UX and frontend design skills.

| Skill | Source | Description |
|---|---|---|
| `frontend-design` | official | Production-grade UI, avoids generic AI aesthetics |
| `react-architect` | local | Component structure, hooks, state, performance |
| `nextjs-architect` | local | App router, server/client boundaries, data fetching |
| `ui-ux-architect` | local | Accessibility, responsive layout, interaction feedback |

---

### Local Bundles (source: `local`)

#### `global`
Universal skills installed machine-wide. Apply to every project regardless of stack.

| Skill | Category | Description |
|---|---|---|
| `commit-author` | workflows | Conventional Commits from git diffs |
| `logic-cleaner` | workflows | Guard clauses, boolean simplification, magic numbers |
| `code-design-refactor` | workflows | Extract/encapsulate, reduce coupling, SRP, naming |
| `tdd` | workflows | Red-green-refactor loop, behavior-focused tests |
| `grill-with-docs` | workflows | Stress-tests plans through relentless questioning; also extracts and maintains the domain glossary (CONTEXT.md) |
| `caveman` | workflows | ~75% token reduction mode, preserves technical accuracy |
| `feature-planner` | workflows | Requirement clarification → design → task breakdown |
| `security-reviewer` | quality | Injection, auth issues, secret leakage, insecure defaults |
| `ddd-architect` | design | Bounded contexts, aggregates, value objects, domain events |
| `hexagonal-arch` | design | Ports & adapters enforcement, dependency direction |
| `design-patterns` | workflows | When to apply, anti-patterns, Go & Python examples |
| `improve-codebase-architecture` | workflows | ADR-aware architectural friction analysis |
| `rsk-guide` | tooling | How to use the `rsk` CLI — discover, install, update, status, uninstall, bundle catalog |
| `repo-tooling-architect` | tooling | Repo-level productivity files — `.editorconfig`, `.gitignore`, tool version pinning (mise/proto), task runner (Task/just), pre-commit, Renovate |
| `skill-builder` | tooling | Scaffolds new skills following ralvaskills naming, format, and STACK.md standards |

---

#### `go-grpc`
Go service exposing a gRPC API.

| Skill | Source |
|---|---|
| `go-architect` | local |
| `grpc-architect` | local |
| `protobuf-architect` | local |
| `docker-architect` | local |
| `sql-architect` | local |

#### `gin`
Go service exposing a REST API via the Gin framework.

| Skill | Source |
|---|---|
| `go-architect` | local |
| `gin-architect` | local |
| `rest-api-architect` | local |
| `docker-architect` | local |
| `sql-architect` | local |

#### `nethttp`
Go service exposing a REST API via the stdlib `net/http` package (Go 1.22+ enhanced ServeMux — no framework dependency).

| Skill | Source |
|---|---|
| `go-architect` | local |
| `nethttp-architect` | local |
| `rest-api-architect` | local |
| `docker-architect` | local |
| `sql-architect` | local |

#### `go-cli`
Go command-line tool.

| Skill | Source |
|---|---|
| `go-architect` | local |
| `cli-tool-architect` | local |
| `docker-architect` | local |

#### `fastapi`
Python service exposing a REST API with FastAPI.

| Skill | Source |
|---|---|
| `python-architect` | local |
| `fastapi-architect` | local |
| `rest-api-architect` | local |
| `docker-architect` | local |
| `sql-architect` | local |

#### `llm-app`
Python-based LLM application or RAG pipeline.

| Skill | Source |
|---|---|
| `python-architect` | local |
| `fastapi-architect` | local |
| `llm-app-architect` | local |
| `agent-architect` | local |
| `hexagonal-arch` | local |
| `docker-architect` | local |

#### `ros2`
ROS2 robotics application.

| Skill | Source |
|---|---|
| `python-architect` | local |
| `ros2-architect` | local |
| `docker-architect` | local |

---

### Personal Skills (not bundled)

Currently located in `skills/personal/`. Any skill whose path contains a `personal/` segment is never installed by `rsk install` unless `--personal` is passed explicitly.

| Skill | Description |
|---|---|
| `demo-script-architect` | Presenter-centric demo scripts with narrative flow and progressive capability reveals |

---

### Composable Installs

Instead of creating a bundle for every combination, compose installs:

```bash
# Full-stack: Next.js + FastAPI
rsk install design fastapi

# Full-stack: Next.js + Go REST (Gin)
rsk install design gin

# Full-stack: Next.js + Go REST (stdlib net/http)
rsk install design nethttp

# Go gRPC service with document generation
rsk install go-grpc docs

# LLM app with full design and document capability
rsk install llm-app design docs

# First machine setup — global + docs
rsk install global docs --global
```

---

### Workflow Pipelines

Skills chain into engineering pipelines. There are two canonical flows — pick the one that matches the work you're doing. Skills that don't appear in either pipeline (`caveman`, `skill-builder`) are utilities used independently as needed.

#### New-feature pipeline

Entry point for building something new.

```
feature-planner             ← clarify requirements, design, task breakdown
    ↓
grill-with-docs             ← stress-test the plan, capture domain terms
    ↓
ddd-architect               ← bounded contexts, aggregates, domain events
    ↓
hexagonal-arch              ← validate ports & adapters are correctly placed
    ↓
[stack architect]           ← go-architect / python-architect / etc.
    ↓
[framework architect]       ← fastapi-architect / gin-architect / etc.
    ↓
tdd                         ← implement with tests first
    ↓
logic-cleaner               ← polish expressions, guard clauses, magic numbers
    ↓
security-reviewer           ← catch vulns before shipping
    ↓
api-contract-reviewer       ← validate API surface (if applicable)
    ↓
performance-reviewer        ← N+1s, missing indexes, blocking I/O
    ↓
commit-author               ← generate conventional commit message
```

#### Refactor pipeline

Entry point for restructuring existing code.

```
improve-codebase-architecture  ← ADR-aware friction analysis, identify hotspots
    ↓
grill-with-docs                ← stress-test the proposed refactor
    ↓
code-design-refactor           ← extract, decouple, SRP, naming, primitive obsession
    ↓
logic-cleaner                  ← guard clauses, boolean simplification, magic numbers
    ↓
[stack architect]              ← re-validate idioms after restructuring
    ↓
security-reviewer              ← regressions introduced by the refactor
    ↓
performance-reviewer           ← regressions introduced by the refactor
    ↓
commit-author                  ← generate conventional commit message
```

---

## Roadmap

Status legend: ✅ exists · 🔨 in progress · 📋 planned

### Skills

#### Languages
| Skill | Status | Notes |
|---|---|---|
| `go-architect` | ✅ | v1.0.0 — Go 1.26 audit complete; `STACK.md` pinned (viper, validator, cobra, gin, sqlx, grpc, protobuf-go, testify, fx, migrate, buf); enforces `sqlx + //go:embed` over ORM, `log/slog` over zap |
| `python-architect` | ✅ | v1.0.0 — Python 3.14 audit complete; `STACK.md` pinned (pydantic, fastapi, uvicorn, httpx, pytest 9, mypy 2, ruff, uv, typer, psycopg 3, alembic); enforces `psycopg + .sql files via importlib.resources` over ORM, `mypy --strict` baseline |

#### Databases
| Skill | Status | Notes |
|---|---|---|
| `sql-architect` | ✅ | v1.0.0 — PostgreSQL 18 primary, MySQL 9 & SQLite 3.53 notes. Enforces UUID v7 PKs + UNIQUE naturals, soft delete by default, forward-only migrations, parameter binding, cursor pagination, EXPLAIN-driven indexing |

#### Frameworks
| Skill | Status | Notes |
|---|---|---|
| `fastapi-architect` | ✅ | v1.0.0 — FastAPI 0.136 on Python 3.14. Feature-based structure, Pydantic v2 request/response separation, URL-prefix versioning (`/v1/...`), async DI with lifespan, RFC 7807 problem-details errors, in-house OAuth2+JWT (pyjwt+argon2-cffi) or external IdP with switching criterion |
| `gin-architect` | ✅ | v1.0.0 — Gin 1.12 on Go 1.26. Feature-based structure, struct-tag validation, RFC 7807 middleware, in-house JWT (golang-jwt+argon2) or external IdP, route groups for URL versioning, OpenAPI via swaggo/swag or kin-openapi |
| `nethttp-architect` | ✅ | v1.0.0 — stdlib `net/http` on Go 1.26 (no router framework). Enhanced ServeMux method patterns, middleware function-wrap chain, RFC 7807, mandatory production timeouts, kin-openapi for spec |
| `react-architect` | ✅ | v1.0.0 — React 19 with TS strict. Feature-based folders, hooks-first composition, TanStack Query for server state, local state + Context default + zustand when justified, Suspense + ErrorBoundary at every async boundary, Vitest + RTL + Playwright + axe-core. Pairs with `nextjs-architect` for server-side concerns |
| `nextjs-architect` | ✅ | v1.0.0 — Next.js 16, App Router only. Server components default + explicit `"use client"` boundaries, server actions for mutations, hybrid data access (RSC direct for reads, API routes for writes), streaming Suspense, edge vs node runtime, `next/image`/`next/font`/Metadata APIs, `output: "standalone"` Docker deploys |

#### Protocols
| Skill | Status | Notes |
|---|---|---|
| `rest-api-architect` | ✅ | v1.0.0 — framework-agnostic REST conventions. Plural-noun URLs, URL-prefix versioning (`/v1/`), cursor pagination, `snake_case` JSON + ISO 8601 timestamps, RFC 7807 problem-details errors, **mandatory `Idempotency-Key`** on POST/PATCH, **mandatory `ETag` + `If-Match`** on PUT/PATCH, OpenAPI 3.1 generated from code |
| `grpc-architect` | ✅ | v1.0.0 — vanilla gRPC over HTTP/2. Service definitions, `status.Error` + standard codes, domain→code mapping, mandatory interceptor chain (recovery, request-id, slog, auth, protovalidate, metrics), client-side deadlines, context propagation, reflection off in prod, `bufconn` testing. Language-agnostic protocol; Go-specific examples |

#### Encoding
| Skill | Status | Notes |
|---|---|---|
| `protobuf-architect` | ✅ | v1.0.0 — proto3, Buf-style hierarchical packages (`<org>.<product>.<resource>.<version>`), `protovalidate` CEL validation, `buf lint` / `buf breaking` CI, `buf generate` + `buf.gen.yaml` v2, field-number reservation discipline, well-known-type conventions. Language-agnostic schema layer |

#### Messaging
| Skill | Status | Notes |
|---|---|---|
| `event-driven-architect` | ✅ | v1.0.0 — event/command/message taxonomy, Protobuf schemas with `buf breaking`, hierarchical topic naming, **mandatory outbox pattern**, per-key partitioning for ordering, idempotent consumers with dedupe, DLQs with retry policy, schema evolution, saga choreography vs orchestration. Broker-agnostic; NATS lightweight default, Kafka log-based, RabbitMQ command queues |

#### Tooling
| Skill | Status | Notes |
|---|---|---|
| `cli-tool-architect` | ✅ | v1.0.0 — cross-language CLI conventions. Root + subcommands, **flag > env > config > default** precedence, TOML config in XDG location, stdout/stderr separation, `--output text\|json\|yaml` mandatory, standard exit codes, `NO_COLOR` respect, shell completions, multi-arch distribution. Recipes for Go (cobra+pflag+viper) and Python (typer+rich) |
| `repo-tooling-architect` | ✅ | v1.0.0 — cross-language repo productivity layer. `.editorconfig` + `.gitignore` always; `mise` (default) / `proto` (alternative) for tool version pinning; `Task` (default) / `just` (alternative) for task running + dotenv; minimal pre-commit hooks; Renovate for dependency updates; explicit "when to skip" guidance |
| `rsk-guide` | ✅ | v0.1.0 draft — quick-reference for the `rsk` CLI. Tracks the planned CLI design; promote to 1.0 once `rsk` ships. **Exempt from STACK.md** (meta-skill mirroring SPECS.md) |
| `skill-builder` | ✅ | v1.0.0 — interview-first scaffolder. References SPECS.md as source of truth; generates frontmatter, body skeleton, STACK.md if needed, atomic SPECS updates (folder structure + roadmap + bundles + convention table); validation checklist before hand-off. **Exempt from STACK.md** (meta-skill, no library stack) |

#### Design
| Skill | Status | Notes |
|---|---|---|
| `ddd-architect` | ✅ | v1.0.0 — strategic-first DDD. Bounded contexts + context mapping (consumes `grill-with-docs`/CONTEXT.md), tactical patterns (aggregates, value objects, domain events, repositories, ACL), small-aggregate discipline, when DDD is overkill. Event sourcing out of scope |
| `hexagonal-arch` | ✅ | v1.0.0 — ports & adapters. Dependency direction always inward; ports declared by consumer (Go interface-where-used / Python Protocol); primary vs secondary adapters; testing benefits; one-vs-two-adapter rule; Go + Python examples; no folder-layout prescription |

#### AI/ML
| Skill | Status | Notes |
|---|---|---|
| `llm-app-architect` | 📋 | RAG patterns, prompt management, eval pipelines, LangChain/LlamaIndex |
| `agent-architect` | 📋 | Tool definitions, memory patterns, orchestration loops, observability |
| `ml-pipeline-architect` | 📋 | Ingestion → feature engineering → training → serving structure |

#### Infra
| Skill | Status | Notes |
|---|---|---|
| `docker-architect` | ✅ | v1.0.0 — Docker 29, Compose v2 (`docker compose` + `docker-compose.yaml`), per-language base defaults (distroless Go, debian-slim Python/Node), BuildKit cache+secret mounts, multi-arch (amd64+arm64), Trivy scanning, language-specific recipes |
| `ci-cd-architect` | 📋 | GitHub Actions structure, test gates, deployment strategies |
| `observability-architect` | ✅ | v1.0.0 — application-side signal production. Prometheus for metrics (naming convention, cardinality cap), OTLP for logs+traces, `log/slog` (Go) + `structlog` (Python), trace_id correlation across all three pillars, head sampling at 10% + 100% on errors, PII/secret redaction at SDK, RED + USE golden signals, SLOs per critical user journey |
| `grafana-architect` | ✅ | v1.0.0 — dashboard + alert consumption layer. Dashboards-as-code via Grizzly (`grr`), one-folder-per-service ownership, one-question-per-panel design, unified alerting with multi-window/multi-burn-rate SLOs, mandatory `runbook_url` annotations, drift detection on UI edits. Pairs with `observability-architect` |

#### Robotics
| Skill | Status | Notes |
|---|---|---|
| `ros2-architect` | ✅ | v1.0.0 — covers **Jazzy** (current LTS, EOL 2029), **Kilted** (latest non-LTS, EOL Dec 2026), and **Lyrical** (next LTS, May 2026, EOL 2031). Workspace layout, `ament_cmake` / `ament_python` packages, lifecycle nodes, explicit QoS profiles, services vs actions, parameter discipline, Python launch DSL. **Pixi + RoboStack** for cross-platform reproducible env (replaces `apt install ros-*`). Both C++ (`rclcpp`) and Python (`rclpy`) first-class |

#### Quality
| Skill | Status | Notes |
|---|---|---|
| `security-reviewer` | ✅ | v1.0.0 — cross-language code-level security review. Critical/High/Medium/Low findings on injection, auth, secrets, insecure defaults, deserialization, CSRF/SSRF/IDOR, rate limiting, dependency hygiene. Tool pass (gitleaks, semgrep, gosec, bandit, Trivy) + read pass. Penetration testing & threat modeling explicitly out of scope |
| `api-contract-reviewer` | ✅ | v1.0.0 — REST + gRPC contract stability. Versioning (URL prefix, package path), field/method hygiene (no reuse, no type changes), error shape (RFC 7807, gRPC codes), idempotency, ETag, OpenAPI completeness. Mechanical pass via `buf breaking` + `openapi-diff`, then conventions vs `rest-api-architect` / `protobuf-architect` / `grpc-architect` |
| `performance-reviewer` | ✅ | v1.0.0 — measurement-grounded perf review. N+1, missing indexes, blocking I/O in async, allocation hot paths, unbounded memory, missing timeouts, high-cardinality labels. Every finding cites `EXPLAIN ANALYZE` / `pprof` / `py-spy` / `hyperfine` / dashboard signal. Load testing & capacity planning out of scope |

#### Frontend
| Skill | Status | Notes |
|---|---|---|
| `ui-ux-architect` | ✅ | v1.0.0 — WCAG 2.2 AA mandatory, Radix primitives + Tailwind 4 + shadcn/ui recipes, OKLCH-based design tokens with semantic naming + dark mode, mobile-first + container queries, **four required states** (loading/error/empty/success) on every async surface, `prefers-reduced-motion` respected, `axe-core` + Lighthouse a11y ≥ 95 in CI |

#### Workflows
| Skill | Status | Notes |
|---|---|---|
| `commit-author` | ✅ | |
| `tdd` | ✅ | |
| `grill-with-docs` | ✅ | Stress-tests plans through relentless questioning; absorbed the former `ubiquitous-language` skill (domain glossary extraction into `CONTEXT.md`) |
| `caveman` | ✅ | |
| `logic-cleaner` | ✅ | Expression-level only — intentionally narrow |
| `code-design-refactor` | ✅ | v1.0.0 — design-level refactoring rules (extract, decouple, SRP, encapsulation, primitive obsession). Sits between `logic-cleaner` (expression) and `improve-codebase-architecture` (system). One move at a time, tests first, commit each move |
| `feature-planner` | ✅ | v1.0.0 — upstream planning pipeline. Requirement → constraints → design outline → vertical-slice task breakdown ordered by risk (tracer bullet first, riskiest assumption next). Hands off to `grill-with-docs` and `tdd`. When NOT to plan: trivial changes, spikes, bug fixes with reproductions |
| `improve-codebase-architecture` | ✅ | |
| `design-patterns` | ✅ | v1.0.0 — skeptical catalog for modern Go/Python. Keeps the few that still earn their place (Repository, Adapter, Strategy, Decorator, Observer, Builder/options, Factory); names anti-patterns (Singleton, Abstract Factory, Visitor, Chain-of-Responsibility, Template Method); defers depth to `ddd-architect` / `hexagonal-arch`; rule-of-three before introducing |

#### Personal
| Skill | Status | Notes |
|---|---|---|
| `demo-script-architect` | ✅ | Personal use only — not bundled, requires `--personal` flag |

---

### CLI (`rsk`)

| Feature | Status | Notes |
|---|---|---|
| `rsk init` | 📋 | Config file setup, prompts for repo path and AI tool target |
| `rsk install` | 📋 | Bundle + skill install via symlinks, `--personal` opt-in flag |
| `rsk update` | 📋 | `git pull` + re-symlink, `--official` flag for Anthropic cache |
| `rsk status` | 📋 | Version comparison, source labels `[ralva]`/`[anthr]`, bundle tags |
| `rsk status --stack` | 📋 | Reads each skill's `STACK.md`, fetches latest versions from `proxy.golang.org` (Go) and `pypi.org` (Python), highlights stale skills; results cached 24 h; opt-in only — no background network calls; `--refresh` bypasses cache |
| `rsk list` | 📋 | Full catalog with filters by bundle, source, installed state |
| `rsk uninstall` | 📋 | Remove symlinks cleanly |
| Official skill cache | 📋 | Clone and cache `anthropics/skills` at `~/.ralvaskills/cache/` |
| GitHub Actions release | 📋 | Cross-platform binaries (linux/mac/windows) on tag push |
| Homebrew tap | 📋 | `brew install ralvarezdev/tap/rsk` |

---

### Existing Skills to Audit

Both `go-architect` and `python-architect` audits are complete.

#### `go-architect` → target Go 1.26 — ✅ complete (v1.0.0, 2026-05-20)

Completed in this audit:
- `slices`, `maps`, `cmp`, `iter`, `log/slog` stdlib conventions
- `errors.Join`, `sync.WaitGroup.Go`, container-aware GOMAXPROCS (Go 1.25), `signal.NotifyContext` cause (Go 1.26)
- `testing/synctest` for deterministic concurrent tests
- Range-over-func iterator patterns + when to return `iter.Seq[T]` vs `[]T`
- Self-referential generics (Go 1.26), type-aliased generics (Go 1.24)
- `go fix` modernizers as ongoing hygiene
- Canonical libraries: viper, validator, cobra/pflag, gin, sqlx, grpc, protobuf-go, testify, fx, migrate, buf
- Architectural opinion: `sqlx + //go:embed` over any ORM; `log/slog` over zap/zerolog as default
- `STACK.md` created with all pinned library versions

#### `python-architect` → target Python 3.14 — ✅ complete (v1.0.0, 2026-05-20)

Completed in this audit:
- PEP 649 deferred annotations (3.14) — no more quoted forward refs; use `annotationlib.get_annotations()`
- PEP 758 bracketless `except`, PEP 765 SyntaxWarning for `finally` control flow (3.14)
- PEP 734 `concurrent.interpreters` for CPU-bound parallelism (3.14)
- PEP 703 free-threaded build awareness; asyncio introspection (`python -m asyncio ps/pstree`, 3.14)
- `map(strict=True)`, `compression.zstd`, `Path.copy/move` (3.14)
- `typing.override` (3.12+), `TypedDict` over `dict[str, Any]`
- Tooling: `uv` (env+packaging), `ruff` (lint+format), `mypy 2 --strict`, `pytest 9`
- Architectural opinion: `psycopg 3 + .sql files via importlib.resources` over any ORM; SQLAlchemy 2.x only when justified by ADR
- `STACK.md` created with all pinned library versions
