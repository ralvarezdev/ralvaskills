# Taskfile Reference

Companion to [SKILL.md](SKILL.md) §4. This pins the naming grammar, the canonical
verb and module vocabulary, the aggregate/atomic rule, include rules, and the
standard Go-library task set. Read it before adding or renaming a task.

## 1. Naming grammar

A task name is `segment(:segment)*`. The first segment picks one of **two
namespace families**, and a namespace belongs to exactly one:

| Family | Grammar | Examples |
|---|---|---|
| **Domain** — a module owns the tasks | `module[:submodule…]:verb[:target\|mode]` | `git:update:recursive`, `docker:up:dev`, `codesys:templates:check:exec` |
| **Action** — one action spans modules | `verb[:target[:mode]]` | `test:backend`, `lint:python`, `scan:fs`, `clean:docker` |

Rules:

- The **verb sits immediately after the module path**, before any target or mode.
  `docker:up:dev`, never `docker:dev:up`. `docker:build:codesys`, never
  `docker:codesys:build`.
- **Environment and service are targets, not submodules.** `dev`, `prod`, `db`,
  `openplc` follow the verb: `docker:up:dev`, `docker:up:db`.
- A **true submodule** is part of the domain hierarchy and precedes the verb:
  `codesys:templates:check`, `chaos:firecracker:node:netem:apply`,
  `terraform:do:apply`, `security:certs:sync`.
- **Never mix orderings inside one namespace.**
- **Unprefixed names are reserved for repo-wide entry points** that belong to no
  module: `default`, `build`, `test`, `lint`, `fmt`, `check`, `tidy`, `tag`,
  `clean`.

`task --list` sorts alphabetically, so this grammar keeps every task for a module
adjacent to its siblings and reads as a namespace even before the file is split.

## 2. Canonical verb vocabulary

One word per concept. Do not invent a synonym for a concept that already has one.

| Concept | Canonical | Rejected |
|---|---|---|
| Generate code | `generate` | `gen`, `codegen` |
| Test coverage | `coverage` | `cover` |
| `go mod tidy` | `tidy` | `mod:tidy`, `go:tidy`, `dev:tidy` |
| Format | `fmt` (`fmt:check`, `fmt:fix`) | `format:check`, `format:fix` |
| Pre-commit gate | `check` | `quality` |
| Style: one-line bodies | `oneliners` | `oneline` |
| Clean built binaries | `binaries` | `binary` |
| Dependency check | `deps:check` | bare `deps` |
| Submodule at pins | `git:init` | `git:init:submodules` |
| Submodule at remote | `git:update` / `git:update:recursive` | `submodules:update` |
| Local CI via act | `ci:local` | — |
| CI gate aggregate | `ci` | — |

## 3. Canonical module vocabulary

| Domain | Canonical namespace | Rejected |
|---|---|---|
| Containers | `docker` | `compose` |
| Submodules | `git` | `submodules` |
| Protobuf | `proto` | — |
| OpenAPI | `openapi` | `api` (for spec codegen) |
| sqlc | `sqlc` | `db:sqlc`, `db:generate` |
| Goose migrations | `migrate` | `db:migrate` |
| Database lifecycle | `db` (backup/up/down/status/reset) | — |
| Web app | `frontend` | `ui`, `web` |
| API service | `backend` | `be` |
| Docs | `docs` | — |
| Benchmarks | `bench` | — |
| Struct alignment | `align` | — |
| Security scan | `scan` | — |

Include keys use the **full** module name (`backend`, not `be`).

## 4. Canonical Go-library task set

Copy from `ralvaskills/skills/personal/go-library-builder/assets/Taskfile.yml`.
Every ralvarezdev Go library carries exactly these tasks:

| Task | Runs |
|---|---|
| `default` | `task --list` |
| `build` | `go build ./...` |
| `test` | `go test -race -count=1 ./...` |
| `test:short` | `go test -short -count=1 ./...` |
| `tidy` | `go mod tidy` + `go mod verify` |
| `fmt` | `goimports -w -local {{.MODULE}} .` + `go fmt ./...` |
| `lint` | `golangci-lint run ./...` |
| `lint:fix` | `golangci-lint run --fix ./...` |
| `lint:doccheck` | `golangci-lint run -c .golangci.doccheck.yml ./...` |
| `modernize` | `go fix ./...` |
| `align:check` | `betteralign ./...` |
| `align:fix` | `betteralign -apply ./...` |
| `scan` | delegates to `scan:fs` + `scan:secrets` |
| `scan:fs` | `trivy fs --scanners vuln .` |
| `scan:secrets` | `trivy fs --scanners secret .` |
| `deps:check` | `go mod download` from an empty module cache |
| `tag` | validate semver, `git tag`, `git push origin` |
| `check` | `fmt` → `align:fix` → `lint` → `scan` |

A library that adds a module (e.g. `sqlc:*` in a data layer) keeps the base set
intact and adds the module on top.

## 5. Aggregate vs atomic

- An **aggregate** keeps the bare namespace name and only delegates:
  `scan` = `scan:fs` + `scan:secrets`.
- An aggregate must **never add behaviour its subtasks lack**. `check` must run
  the same set everywhere (`fmt` → `align:fix` → `lint` → `scan`); a repo that
  silently drops a step from `check` is a bug, not a local choice.
- Prefer atomic subtasks plus one aggregate over one monolithic task: it keeps
  the CI step and the local step identical.

## 6. Includes and nesting

- **Stay flat** until a module earns its own file (5+ tasks, or config distinct
  enough to warrant separation). Splitting a two-task module is ceremony the flat
  `module:verb` key already avoids.
- **Include keys are full module names** — `backend`/`frontend`, not `be`/`fe`.
- **Prefer `flatten: true`** for domain files so each key carries its own
  namespace. When flattened, the sub-Taskfile must use fully-qualified keys
  (`hailo:lint`, not `lint`), so include order can never decide a winner.
- **Never mix** flattened and namespaced includes for the same domain in one repo.
- **Alternate filenames:** `Taskfile.local.yml` (git-ignored) for machine-local
  overrides, with a committed `Taskfile.local.yml.example`; include it with
  `optional: true`. Generated-repo templates use `Taskfile.yml.tmpl`.
- A sub-Taskfile in a nested directory that is invoked directly (not included)
  must still use the same grammar.

## 7. Cross-platform, vars, env, deprecation

- **Cross-platform:** use `platforms:` for OS-specific commands, `{{.EXE}}` for
  binary suffixes. On Windows, scripts run under Git Bash, never the WSL
  `bash.exe` launcher (the WSL launcher resolves a different `$HOME`, so SSH
  aliases and `.local` mDNS break). Never assume a POSIX-only shell.
- **Vars:** shared vars live at the root. A var a caller may override is written
  as its own default (`VAR: '{{.VAR | default "…"}}'`) so a CLI `VAR=x` still
  wins over an included file's global.
- **Env:** `dotenv:` only at the root — Task forbids it in included files. `.env`
  is committed (safe defaults); `.env.local` is git-ignored. Real secrets never
  enter either; they come from the external secret manager (§6 of SKILL.md).
- **Helpers:** mark with `internal: true` and prefix `_` (e.g. `_dev:parallel`).
- **Renames:** keep the old name as an alias for one cycle with
  `desc: "Deprecated — use <new>"`, then delete it in the next sweep.

## 8. Anti-patterns

| Anti-pattern | Fix |
|---|---|
| `git:init:submodules` alongside `git:init` | one name: `git:init` |
| `submodules:update` | `git:update` |
| `docker:dev:up` (env before verb) | `docker:up:dev` |
| `docker:up:db` / `docker:db:up` mixed | `docker:up:db` |
| `proto:gen` next to `proto:generate` | `generate` everywhere |
| `test:cover` next to `test:coverage` | `coverage` everywhere |
| `lint:oneline` next to `lint:oneliners` | `oneliners` everywhere |
| `clean:binary` next to `clean:binaries` | `binaries` everywhere |
| `mod:tidy` / `go:tidy` / `dev:tidy` | `tidy` |
| `quality` for the pre-commit gate | `check` |
| `format:fix` next to `fmt:fix` | `fmt:fix` |
| `ui:*` / `web:*` for the frontend | `frontend:*` |
| `be`/`fe` include keys | `backend`/`frontend` |
| `db:sqlc` / `db:generate` | `sqlc:generate` |
| `check` that omits `scan` | one `check` definition everywhere |
| bare `deps` | `deps:check` |

---

_Last reviewed: 2026-10-14_
_Skill version at last review: 1.4.0_
