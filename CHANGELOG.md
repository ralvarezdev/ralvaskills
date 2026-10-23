# Changelog

All notable changes to ralvaskills: the `rsk` CLI and the skill catalogue. CLI releases are tagged `vX.Y.Z`; each skill release is tagged `<skill>@vX.Y.Z`. Entries are summarised from the git history and tags; routine patch releases of a skill are folded into the release they follow.

Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## Unreleased

### Added
- **codebase-hygiene-reviewer:** new skill (1.0.0) for evidence-based declutter audits: dead code, unused exports per consumer, duplication and oversized files. It only proposes changes and writes its report outside the repo.

### Changed
- **go-architect** (1.11.0): the dead-code guidance now covers shared libraries with several consumers.
- **react-architect** (1.1.0): adds `tsc --noEmit` and `knip` as static checks.
- **improve-codebase-architecture** (1.1.0): gathers hygiene evidence before exploring.
- The docs now live in typed folders with frontmatter, and `AGENTS.md` is imported from `CLAUDE.md`.

# CLI releases

## v3.3.0 (2026-10-02)
### Added
- `rsk mcp` runs a stdio MCP server. It has a skill search tool, a project profile tool that detects project signals and proposes skills, and an `install_skills` tool that requires user approval. The skill catalogue is exposed as a resource.
- `rsk new` registers the MCP server in `.mcp.json` and `opencode.json`, and adds a workflow pointer to `CLAUDE.md`. Existing config is merged, writes are idempotent, and registration is skipped with a warning if `rsk` is not on `PATH`. `rsk destroy` removes both.
- `rsk new` offers to install and pin the rsk-guide skill (`--guide`).
- Skill descriptions are now parsed and shown in the catalog.

### Changed
- Re-running `rsk new` on a configured project now proceeds, so the MCP registration can be refreshed.

### Fixed
- Install approval for the MCP `install_skills` tool is enforced, and the search result shape was corrected.

## v3.2.1 (2026-10-01)
Maintenance release.

## v3.2.0 (2026-10-01)
### Fixed
- `rsk` refuses to write `CLAUDE.md` when it is a symlink. This avoids an import loop with `.rsk/CLAUDE.md` and avoids silently replacing the link. It fails with an error and leaves the link alone.
- Atomic file writes keep the existing file's permissions.
- SIGINT and SIGTERM cancel in-flight work: downloads abort and temp directories are removed. rsk exits silently with code 130.

### Changed
- go-house-libs skill references were updated to current tags and now cover svckit, tick, sqlitekit and resilience.

### Breaking
- Go module path changed to `.../ralvaskills/v3`. The previous `/v2` path made the v3 releases uninstallable, because `go install` resolved to v2.0.5. Migration: install with `go install github.com/ralvarezdev/ralvaskills/v3/cmd/rsk@latest`, and update any imports.

## v3.1.1 (2026-09-30)
Maintenance release.

## v3.1.0 (2026-09-29)
### Added
- go-library-builder skill covers the toolkit and adapter shapes. The go-house-libs references were renamed from house-kits.

## v3.0.0 (2026-09-29)
### Added
- Global flags: `--config` (overrides the config directory, takes precedence over `RSK_CONFIG_HOME`), `--verbose`, `--quiet` and `--no-color` (overrides `NO_COLOR` and TTY detection).
- `--yes`/`-y` on `install`, `update`, `uninstall` and `destroy` to skip confirmation in scripts.
- `rsk tools <allow|deny|list|remove>` with `--for` (default `claude-code`).
- `pin --remove`, plus `--include personal|official` on `install`, `uninstall`, `update`, `list`, `status` and `catalog`.
- `rsk catalog bundles` and `rsk catalog bundle <name>`.
- `--for`, `--source` and `--output` values are validated up front, with usage text and shell completion drawn from the same list. Pin and unpin actions were added to `list` rows.
- Help is grouped by command category, with `help` and `completion` under "Other". The interactive session confirms destructive commands, remembers recent commands and form values, and has Markdown help.

### Changed
- Terminal handling in the session now relies on termkit, and destructive prompts use its confirm dialog.

### Fixed
- `task tag` can now run unattended.

### Breaking
- `rsk catalog --bundle` and `rsk catalog <name>` are removed. Use `rsk catalog bundles` and `rsk catalog bundle <name>`.
- Legacy aliases removed: `rsk unpin` (use `rsk pin --remove`), `rsk claude tools` (use `rsk tools`) and `--bundles`. The `--personal` and `--official` flags are replaced by `--include`, and the unused `catalog --stack` flag is gone.

## v2.0.5 (2026-09-28)
### Added
- In the interactive session, `list`, `status`, `catalog` and the Claude tools view render as read-only tables. Rows have actions: install or uninstall from `catalog` and `list`, uninstall from `status`, and allow, deny or remove a tool.
- `-o json` in the session shows the raw payload of the result view.

### Changed
- Tables, banner and palette are now drawn with termkit, and the banner wordmark gradient follows the active palette. The Claude tools list and `destroy` plan are shown as tables.

## v2.0.4 (2026-09-26)
### Fixed
- `rsk --version` prints the same bare version number (for example `2.0.3`) whether rsk was built with `go install` or from a release build.

## v2.0.3 (2026-09-26)
### Fixed
- Module path is now `github.com/ralvarezdev/ralvaskills/v2`. Before this, `go install .../cmd/rsk@latest` resolved to v1.4.3 and stayed there through all v2.0.x releases.

### Breaking
- Go module path changed to `.../ralvaskills/v2`. Migration: install with `go install github.com/ralvarezdev/ralvaskills/v2/cmd/rsk@latest`, and update any code that imports the old path.

## v2.0.2 (2026-09-26)
### Added
- go-architect skill defaults to sqlc, with sqlx as the alternative (see v2.0.0).

### Changed
- Go toolchain requirement raised to 1.27.1.

## v2.0.1 (2026-09-22)
### Fixed
- The update-availability notice is now refreshed based on the actual outcome of an update, so it no longer lingers after you update.

## v2.0.0 (2026-09-22)
### Added
- Running `rsk` with no arguments shows a themed welcome banner and an interactive command picker. Picked commands run inline, with nested pickers and parameter forms, instead of exiting to the shell.
- A notice such as "N updates available" appears on bare `rsk` runs. It is read from a cache that is refreshed in the background at most daily, so commands are never delayed.
- `--output` is now enforced on read commands.
- New skills: latex-architect, ml-conference-paper-architect, go-library-builder, reveal-js-architect and a Spanish-language thesis defense slide skill, among others. The go-architect skill now defaults to sqlc for database access, with sqlx as the alternative.
- `registry-remove` helper tool for pruning skills or versions from the registry. The work-report-generator skill was removed from the registry.

### Fixed
- `list`, `catalog`, `uninstall` and `update` no longer panic with "flag not registered".
- `claude` no longer appears twice in `--help`, and `rsk new` success messages now name the target tool.
- `rsk status --global` now works outside a project.
- The official-skills cache and the version fallback were repaired.

## v1.4.3 (2026-07-04)
### Added
- `RSK_CONFIG_HOME` environment variable to point rsk at a separate config and cache directory, useful for sandboxed testing.
- `rsk new` supports incremental tool initialization, so tools can be added to an already-initialized project.

### Fixed
- `rsk install <bundle> --pin` now pins every skill the bundle resolves to, not just the bundle entry.

## v1.4.2 (2026-07-04)
### Fixed
- `rsk update` in registry mode now refreshes `rsk.lock`; previously the update succeeded but left the lock stale.
- `rsk update` with no arguments in local mode now refreshes every local skill listed in `rsk.mod` and updates the lock.
- `rsk update --official` now warns when it is used in registry mode, where the flag has no effect.

## v1.4.1 (2026-07-04)

### Fixed
- `rsk destroy` now reports errors that it previously discarded silently.

## v1.4.0 (2026-07-04)

### Changed
- `rsk new` on an existing project now adds newly requested tools to `rsk.mod` instead of failing, and applies the already pinned skills to them. It reports "updated" when it changes something and does nothing when already in sync.

### Fixed
- Interrupted or partial installs no longer leave an empty registry cache folder that is treated as valid. The cache is now checked for `SKILL.md` and extracted atomically.

## v1.3.2 (2026-06-10)

### Fixed
- Interrupted or partial installs no longer leave an empty registry cache folder that is treated as valid, which left skill links pointing at empty directories. The cache is now checked for `SKILL.md` and extracted atomically.

## v1.3.1 (2026-05-23)

Maintenance release.

## v1.3.0 (2026-05-23)

Maintenance release.

## v1.2.0 (2026-05-23)

### Added
- Interactive prompts: when a required argument is missing, `install`, `uninstall`, `pin`, `unpin`, `new` and `claude tools` ask for it instead of aborting. Prompts use rich widgets in a terminal and plain-text input in non-interactive environments such as CI.
- `rsk destroy` shows the skills it will remove and asks for confirmation first.

### Changed
- `rsk init` uses the shared prompt components. Aborting a prompt exits with status 130.

## v1.1.0 (2026-05-23)

### Added
- `rsk claude tools [list|allow|deny|remove]` manages tool permissions in a project's `.claude/settings.json`. It marks each tool as allowed, denied or not configured, and cleans up empty settings.

### Changed
- Project skill links now go directly into `<project>/.claude/skills/` instead of an intermediate `.rsk/skills/` folder. OpenCode reads the same folder.

### Fixed
- `rsk status` no longer misses Windows junction links.

## v1.0.5 (2026-05-23)

### Fixed
- On Windows, project skill links use directory junctions, so creating them no longer needs administrator privileges.

## v1.0.4 (2026-05-23)

### Fixed
- `rsk new` no longer fails with an empty `.rsk` directory path when the project folder cannot be resolved.
- Unpinning a skill now removes its import from the project's `CLAUDE.md` correctly.

## v1.0.3 (2026-05-23)

### Fixed
- Corrected the global Claude skills folder path.

## v1.0.2 (2026-05-23)

### Fixed
- Windows config home now falls back to `~/.config` instead of `%APPDATA%`, so it matches the other platforms.

## v1.0.1 (2026-05-23)

### Fixed
- Task-based build and registry steps now run natively on Windows (PowerShell) as well as POSIX shells.
- Release builds can be started manually if the tag-triggered run does not fire.

## v1.0.0 (2026-05-23)

### Added
- Initial release of `rsk`, with a skill catalogue covering backend, frontend, language, protocol, observability, design and workflow skills, plus bundles that group them.
- Declarative project manifest: `.rsk/rsk.mod` records the desired skills and pinned skills, and `.rsk/rsk.lock` records resolved versions, sources and cache paths.
- Claude and OpenCode adapters: pinned skills are imported into Claude Code through Markdown imports and into OpenCode through `opencode.json`.
- Coloured terminal output: the `rsk` name, source labels, bundle tags, versions and status marks are styled, and table columns stay aligned.
- `rsk status` reports installation state, and bundle listings show a full, partial or none status that also appears in JSON output.
- Token-count reporting for the catalogue now estimates the context cost of each install profile and bundle.

### Changed
- Reorganised the command surface into top-level commands: `rsk new` and `rsk destroy` manage the project, `rsk pin` and `rsk unpin` manage pinned skills, and `rsk install`, `rsk uninstall` and `rsk update` act on the project or on `--global`.
- `rsk catalog` now shows the published catalogue, and `rsk list` shows only what is installed.
- Config and manifest files are written atomically, and registry archive extraction rejects symlinks and paths that escape the cache directory.
- The skill catalogue was regrouped into domain folders.

### Breaking
- The `rsk skill` and `rsk project` command groups were removed. Use `rsk install`, `rsk pin` and `rsk unpin` instead of `rsk skill ...`, and `rsk new` and `rsk destroy` instead of `rsk project ...`.
- `rsk list` no longer shows the catalogue. Use `rsk catalog`.

# Skill releases

## api-contract-reviewer

Reviews REST and gRPC contracts for stability, versioning and backwards compatibility, running mechanical breaking-change checks and reporting severity-keyed findings.

### 1.0.0 (2026-05-22)
First release. The skill defers to the REST, Protobuf and gRPC architect skills for the rules and enforces them mechanically with `buf breaking` and `openapi-diff`, so schema regressions across REST and gRPC boundaries are caught before a breaking-change release.

Patch release 1.0.1: updated the RFC 9457 citation (replacing RFC 7807) in line with the related API skills.

## astro-architect

Standards for building Astro sites around a static-first, zero-JavaScript-by-default model.

### 1.0.0 (2026-09-30)
First release. Everything follows from one invariant: ship no JavaScript by default. Interactivity is a cost ladder (CSS, then native elements, then a processed script, then an island) so a content site does not ship a framework runtime for one dropdown. It also covers Content Layer collections with Zod validation, script processing and `is:inline`, routing and `getStaticPaths`, Actions, Sessions and Server Islands, rendering modes and adapters, and theming through custom properties so a theme picker never needs to be an island. It stays biased toward `output: 'static'` and documents the runtime cost of adapters wherever they become necessary.

## caveman

An ultra-compressed communication mode that drops filler and pleasantries while keeping technical accuracy.

### 1.0.0 (2026-05-22)
First release, brought into the catalogue from an external skill collection. It was then given a version field and author attribution, and its guidance was made language-agnostic.

## ci-cd-architect

CI/CD principles for designing, reviewing and scaffolding pipelines, with GitHub Actions recipes.

### 1.2.0 (2026-10-11)
#### Added
- A new section on running CI locally with `act`, with a matching recipe covering the `act` model, the caching that makes it viable, and the `--concurrent-jobs` race.
- Two caching pitfalls that silently disable caching: a tool-installer action that caches binaries but not the package store, and `cache-dependency-path` defaulting to the repo root when the module lives in a subdirectory (workflow-level `defaults.run.working-directory` does not apply to `uses:` steps).
- A note that the classic combined `repo` + `read:packages` token is a single-token alternative to two fine-grained tokens.

### 1.1.0 (2026-10-10)
#### Added
- Guidance that private Go modules and private npm packages need distinct, scoped tokens.

### 1.0.0 (2026-05-22)
First release. Covers pipeline taxonomy, trigger design, supply-chain hygiene, caching, matrix strategy, test gates, release automation and deployment strategies.

## cli-tool-architect

Cross-language standards for designing and reviewing command-line tools, with Go (cobra) and Python (typer) recipes.

### 1.0.0 (2026-05-22)
First release. Standardizes subcommand structure and a flag, then environment, then config, then default precedence. New user configs use TOML in XDG paths, data goes to stdout and logs to stderr, and user-facing errors follow a concise three-line remediation shape with no raw stack traces.

## code-design-refactor

Function-boundary refactoring moves (extract, decouple, single responsibility) that follow requirement slicing in the design workflow.

### 1.0.0 (2026-05-22)
First release. It completes the design ladder from vertical-sliced requirements down to concrete refactoring moves. A rule was added later that a directory's size is a signal, never a target: split folders on a real responsibility seam plus navigability, not on file count. It notes that in languages where a directory is a package, splitting changes import paths and the exported surface.

Patch release 1.0.1: added the folder-splitting-on-cohesion rule above, which the language and framework architect skills now link from their structure sections.

## commit-author

Generates concise Conventional Commits messages from a staged diff.

### 1.11.0 (2026-10-08)
#### Added
- Guidance for merge commit bodies: instead of a bare merge subject, the body now summarizes the changes the merge brings in, grouped by type or area.

### 1.0.0 (2026-05-22)
First release. Enforces conventional commit types and imperative subjects. A section requires explicit user permission before the agent commits, with GPG signing guidance when it does, and the commit-safety rules were later tightened so there are no blanket session-wide permissions.

## connectrpc-architect

Standards for ConnectRPC services that serve Connect, gRPC and gRPC-Web from a single handler.

### 1.0.0 (2026-10-15)
First release. Covers serving one `.proto` through `protoc-gen-connect-go`, `connect.Code` error mapping, a mandatory interceptor chain, streaming caveats (HTTP/2 for client and bidi streaming, server-stream only for gRPC-Web), browser callers without a proxy, generic typed clients and conformance testing.

## ddd-architect

Domain-Driven Design guidance, strategic first, for shaping domain models and aggregate boundaries.

### 1.0.0 (2026-05-22)
First release. Starts from bounded contexts and context mapping, then codifies tactical building blocks: small aggregates with clear transaction bounds, value objects, repositories and immutable, past-tense domain events.

## demo-presentation-architect

Designs slide-deck specifications as a slide-by-slide markdown file, never HTML, PDF or PPTX.

### 1.1.0 (2026-05-22)
#### Changed
- Slimmed the skill by moving the layout catalog, word budgets and narrative arcs into a more compact form, cutting its size by about a third while keeping the interview-first workflow and per-slide layout choices.

## demo-script-architect

Designs presenter-centric demo scripts with narrative flow, visual cues and a redundancy audit.

### 1.0.0 (2026-05-22)
First release. A six-step narrative engine reveals a system progressively, with strict mapping between each spoken claim and the UI element being shown. Anti-patterns for robotic jargon and redundancy are collected in a right/wrong table, and the skill can audit existing demos and walkthroughs.

## design-patterns

A skeptical design-patterns catalogue for Go and Python that flags most patterns as anti-patterns.

### 1.0.1 (2026-05-22)
- Maintenance release. The catalogue keeps repository, adapter, strategy, decorator, observer and builder, applies a "rule of three" before introducing any abstraction, and was trimmed as part of the repository-wide prompt-size reduction.

## docker-architect

Docker standards for Dockerfiles and Compose files: multi-stage builds, non-root, multi-arch and image scanning.

### 1.3.0 (2026-10-10)
#### Added
- Build secrets must come from the environment, never from a file outside the build context (BuildKit and Bake sandbox file reads), with the build-time versus runtime secret rule spelled out.
- A gotcha: copy `.npmrc` and `pnpm-workspace.yaml` before running `pnpm install`.

### 1.2.0 (2026-07-07)
#### Changed
- Version bump made alongside a task-naming convention update (`module:verb`) in the sibling repo-tooling skill; the Docker guidance itself only gained a version reference.

### 1.1.0 (2026-07-07)
#### Added
- A lifecycle checklist for bootstrapping new services, covering Dockerfile structure, Compose, runtime networking, container security contexts and vulnerability scanning.
- Explicit restart-policy rules: `unless-stopped` for long-running services, `no` for batch or sync jobs, and `on-failure` for transient workers.

### 1.0.0 (2026-05-22)
First release. Covers multi-stage builds, per-language base image defaults, BuildKit cache mounts, non-root users, multi-arch builds, digest-pinned bases, Trivy scanning and Compose v2.

## docs-architect

Sets up and maintains a project documentation system as a typed `docs/` tree.

### 1.0.0 (2026-10-20)
First release. Defines a folder-per-document-type layout (ADRs, backlog, ideas, plans, designs, reviews, changes, guides) with mandatory frontmatter, lifecycle by moving files, and ADR supersede mechanics. It offers scaffold, write-doc and migrate modes and ships a standard-library linter.

## event-driven-architect

Event-driven architecture guidance that is independent of any particular broker.

### 1.0.0 (2026-05-22)
First release. Provides a broker-agnostic taxonomy of events and commands, Protobuf event envelopes, consistent topic and routing-key naming, and a mandatory transaction-bound outbox pattern.

## fastapi-architect

Architectural standards for FastAPI services: feature-based layout, Pydantic v2 request/response boundaries and lifespan-managed connection pools.

### 1.0.3 (2026-10-12)
- Patch releases 1.0.1-1.0.3: boilerplate moved into a RECIPES.md side-card, problem-details guidance now cites RFC 9457 instead of RFC 7807, and the structure section links to a new rule to split folders on cohesion rather than file count.

## feature-planner

Upstream feature planning: turns a vague request into clarified requirements, a design outline and a risk-ordered vertical-slice task breakdown.

### 1.0.1 (2026-05-22)
- Patch release: joined the global design and workflow set as the planning step that feeds the refactoring ladder (vertical slicing into extract, decouple and single-responsibility moves); no change to the skill's guidance.

## gin-architect

Architectural standards for Go services built on Gin: router groups, strict request validation and structured problem-details middleware.

### 1.0.3 (2026-10-12)
- Patch releases 1.0.1-1.0.3: boilerplate moved into RECIPES.md, problem-details guidance now cites RFC 9457 instead of RFC 7807, and the structure section links to the new rule to split folders on cohesion rather than file count.

## go-architect

Architectural standards for writing, reviewing and scaffolding Go code, backed by a ready-to-use golangci-lint configuration and hygiene scripts.

### 1.10.0 (2026-10-13)
#### Added
- A new LAYOUTS.md catalogue of Go repo layouts (library, flat by concept, feature-first, escalated feature-first, grouped at the root, multi-module workspace, layer-first) with a selector table, a lightest-fit-first rule and a per-project review checklist.
- Guidance on module count and versioning, process types, generated code and SQL, test layout, ownership and CI scoping, and the smells that signal a layout needs review.
- A new "layout at scale" section: feature-first organisation, flat packages by default with layers only when earned, implementations as technology-named sibling subpackages, a composition root per binary, and an enforced rank rule between layers.

#### Changed
- Layout guidance now says there is no single correct layout and tells you to pick the lightest one that fits the project's shape, moving to a heavier one only when a concrete pain appears.

### 1.8.0 (2026-10-13)
#### Added
- First version of the layout-at-scale section: a project-shape table (library, flat, feature-first, escalated, grouped, multi-module, layer-first), the feature-first recommendation with flat packages by default, and the rank rule between layers.

### 1.7.0 (2026-09-26)
#### Changed
- sqlc is now the default database-access pattern (schema-first code generation); sqlx with embedded SQL remains documented for query shapes sqlc cannot express as a static named query.

#### Added
- deadcode documented as an opt-in companion to golangci-lint, with its caveats (needs `-test`, does not follow dynamic dispatch).
- Rule that a library exposes `New(deps...)` and never imports a DI container; the deployable owns the container at the composition root, with an optional `fxmod` subpackage for fx users.
- Cross-language rule linked from the structure section: split folders on cohesion, not file count.

#### Fixed
- The bundled golangci-lint template no longer references dead linters (stylecheck, looppointer), so it loads again; target Go version raised to 1.27.1 and golangci-lint to 2.14.0.

Patch releases 1.7.1-1.7.4: RFC 9457 citation updates, the folder-cohesion link, the linter-template fix, deadcode guidance and the DI-container rule.

### 1.5.0 (2026-07-08)
#### Added
- A multi-module workspace (go.work) section covering `pkg/*` cross-cutting utilities, a root shared-kernel `domain/`, and per-service `internal/<service>/` modules.
- Rules separating thin `cmd/` process wrappers from `app/` configuration roots, and distinguishing infrastructure "services" from DDD domain/application services.
- A RECIPES.md with a complete workspace blueprint for a multi-context system.

Patch release 1.6.2: testcontainers-go recommended for Docker-backed integration tests, and testcontainers-go and golangci-lint pins tightened.

### 1.4.0 (2026-07-06)
#### Added
- About 21 more linters in the bundled config: performance and safety (prealloc, reassign, makezero, predeclared), context rules (containedctx, fatcontext), type and time safety (durationcheck, forcetypeassert, mirror, exptostd), and test and structure checks (paralleltest, tparallel, thelper, testifylint, funcorder, wrapcheck, unparam, dogsled, usetesting).
- Guidance to use `net.JoinHostPort` instead of string formatting for addresses, checked type assertions, parameter-driven context boundaries, error-wrapping at package edges, and parallel tests.

#### Changed
- Declaration order is now type, then const, var and func, matching the decorder linter that enforces it.

### 1.3.0 (2026-07-06)
#### Added
- A "Code quality" subsection covering nolintlint (suppressions must name a linter and give a reason), varnamelen and errname, plus a "Testing linters" index for testifylint and thelper.
- An "Optional: scoped linters" block isolating adoption-sensitive linters (exhaustruct, tagliatelle, depguard) and forbidigo tuning notes.

### 1.2.0 (2026-07-05)
#### Added
- A 27-linter golangci-lint v2 template including decorder, mnd, gocognit, maintidx, ireturn, interfacebloat, exhaustive, forbidigo (bans `map[string]any` in favour of structs), gochecknoinits, copyloopvar, perfsprint, intrange and modernize.
- Two shell hygiene checks (no one-line functions, no decorative separator comments) and a secondary doc-coverage lint config.
- gocritic rules (rangeValCopy, hugeParam, rangeExprCopy, forcetypeassert) and Go 1.22+ `range n` guidance.

#### Changed
- Sections 1, 2 and 4 now map each written rule to the linter that enforces it, and the tooling section describes the Taskfile wiring.

#### Fixed
- Removed an invalid gocritic rule reference in favour of the intrange linter.

### 1.1.0 (2026-06-10)
#### Added
- A bundled opinionated golangci-lint v2 config (fieldalignment, testifylint, sqlclosecheck) with setup pointers in the tooling section.

#### Changed
- File-level `type`, `const` and `var` declarations must each be grouped into a single parenthesised block per file, aligned with gofmt conventions.

### 1.0.0 (2026-05-22)
Go standards for the 1.26 toolchain: memory-aligned structs, typed enums, interface design, goroutine safety (`sync.WaitGroup.Go`, `testing/synctest`, container-aware GOMAXPROCS), log/slog structured logging, and an anti-ORM persistence approach using sqlx with `//go:embed` SQL files, alongside golangci-lint v2 settings.

## go-house-libs

Maps a need to the author's shared Go modules (REST, Gin, Postgres, MCP, identity, rate limiting, email, web push, terminal UI and more), with a reference per module.

### 0.6.0 (2026-10-15)
#### Added
- New references for protos, connectrpckit and corekit, and for viperkit (file-sourced secrets that resolve `<VAR>_FILE` into `<VAR>`), registered in the package table and STACK.md.

#### Changed
- Identity reference now describes v2 tenancy, service accounts and asymmetric tokens.

Patch release 0.6.1: identity v2 documentation refresh.

### 0.5.0 (2026-09-30)
#### Added
- References for grpckit, svckit, tick, sqlitekit and resilience, with updated router table and rules.

#### Changed
- Every module reference and STACK.md refreshed to the current tags, Go 1.27.1 policy and sibling pins.

Patch release 0.5.1: removed project-specific names, consumers table and examples so the skill describes only the modules.

### 0.3.0 (2026-09-29)
#### Changed
- Renamed from house-kits to go-house-libs to match go-architect and go-library-builder; the intro no longer claims every module follows the library builder (only email, ratelimit, webpush and identity share its shapes; others are helper toolkits or framework adapters).
- Notes that the webpush Send status handling is merged but not yet tagged.

## go-library-builder

Blueprint for building new Go libraries with hexagonal structure, narrow ports, and shared tooling assets.

### 1.1.0 (2026-09-29)
#### Added
- Four library shapes (flat single-port, vertical-slice, toolkit libraries, framework-adapter modules), each with a folder tree.
- A manual-dispatch CI workflow, and a bootstrap that begins by checking go-house-libs and ends by adding the new module to it.

#### Changed
- Now follows go-architect on generics, uses sqlc rather than sqlx for small queries, and points Postgres and Valkey tests and goose-based migrations at the shared pgtest, valkeytest and migrate helpers instead of hand-written harnesses.

Patch release 1.1.1: libraries use constructor injection only and never import a DI container, with an optional `fxmod` subpackage for fx users.

### 1.0.0 (2026-07-17)
A 14-section guide to library design covering strict domain isolation (standard library plus uuid), narrow composed ports, "accept interfaces, return concrete structs", and sentinel error translation in technology-named adapter packages, with layout recipes for small single-port modules and multi-slice contexts, pinned Go and tooling versions, and copyable Taskfile, mise and Renovate assets.

## grafana-architect

Dashboards-as-code and alerting standards for Grafana, designed to pair with observability-architect.

### 1.0.0 (2026-05-22)
Covers Grizzly-managed dashboards-as-code with per-service folders, one question per panel, unified alerting with runbooks, low-cardinality discipline and SLO tracking panels.

## grill-with-docs

Grilling session that challenges a plan against the project's domain model and updates documentation (CONTEXT.md, ADRs) as decisions crystallise.

### 1.0.1 (2026-05-22)
- Patch release: repaired broken relative links across the format specifications and removed stray conversational residue.

## grpc-architect

Vanilla gRPC standards: service definitions, standard status codes, domain-to-code mapping and the interceptor chain.

### 1.0.0 (2026-05-22)
Covers strict request/response messages, an error-code map, and a six-stage runtime interceptor chain, with framework boilerplate kept in a RECIPES.md side-card.

Patch release 1.0.1: citation updates from RFC 7807 to RFC 9457.

## hexagonal-arch

Hexagonal (ports and adapters) architecture: dependencies point inward and the domain never imports framework, database or HTTP code.

### 1.1.0 (2026-07-08)
#### Added
- Common-mistakes rule forbidding an infrastructure folder named "ports"; ports are internal abstract interfaces owned by `domain/` or `app/`.
- Prescribed `adapters/primary/` (driving) and `adapters/secondary/` (driven) folder naming.
- Compile-time interface assertions (`var _ port.Interface = (*Adapter)(nil)`) as the standard check.
- Cross-context process management via sagas and orchestrators, with consumers defining outbound ports and anti-corruption layers over peer services' public APIs.

### 1.0.0 (2026-05-22)
Pure inward dependency flow with consumer-declared interface locations, warning against folder-as-architecture and anemic domain traps, with code recipes in a RECIPES.md side-card.

## house-kits

Earlier name of go-house-libs: maps a need to the author's shared Go modules.

### 0.2.0 (2026-09-29)
#### Changed
- Split into a router over ten modules (restkit, ginkit, pgkit, mcpkit, identity, identitygin, ratelimit, email, webpush, termkit), each with its own reference covering package layout, a recipe, and rules and gotchas.
- STACK.md now records current tags, Go versions and sibling pins.

#### Added
- Cross-cutting rules on never leaking 5xx details, keeping infrastructure in its own module, and treating a nil-to-error change as breaking.

#### Fixed
- Routing table no longer sends CORS needs to gin-contrib/cors when ginkit ships its own.

## hugo-architect

Standards for Hugo static sites: layout, front matter, templates, Hugo Modules, Page Bundles and the asset pipeline.

### 1.0.0 (2026-05-23)
Opinionated 12-section guide targeting Hugo Extended with Tailwind and PostCSS, preferring Hugo Modules over git submodules, Page Bundles, Hugo Pipes asset processing and TOML front matter, with recipes for layouts, asset pipelines and GitHub Pages deployment.

## improve-codebase-architecture

Finds deepening opportunities in a codebase, informed by CONTEXT.md domain language and docs/adr decisions.

### 1.0.0 (2026-05-22)
Produces its architecture findings as a self-contained HTML report (side-by-side cards and Mermaid dependency diagrams) written to a temporary directory, and ties the "Deepen" step to the rules in DEEPENING.md.

## latex-architect

LaTeX document standards on TeX Live for writing, building and structuring documents.

### 1.0.0 (2026-08-10)
Covers latexmk-based builds, load-bearing package order (mathtools, hyperref, cleveref), semantic markup and one-sentence-per-line sources, with recipes and a pinned stack; Beamer presentations are out of scope.

## logic-cleaner

Expression-level code cleanup: guard clauses, naming, magic values, conditional simplification and duplication removal, language-agnostic.

### 1.0.0 (2026-05-22)
Expression-level cleanup with explicit out-of-scope directives that send architecture and module-decomposition concerns to the appropriate sibling skills.

## mcp-architect

Standards for designing Model Context Protocol servers: tools, resources, prompts, transports, auth and schema contracts, with recipes for Python, Go and TypeScript.

### 2.3.0 (2026-10-13)
#### Added
- Guidance on MCP Apps panels as an enhancement only: every tool must return a complete text result without the extension, the panel adds value when present, content is identical either way, and every capability stays reachable through a model-visible tool.
- Rules that guards stay on the server and that capability detection happens per request on a stateless server, plus a Go skeleton and a with/without-extension test in the recipes.
- The three strategies for sharing enums across tool schemas: `$defs` in one shared file, one file per enum when the generator would emit a colliding per-consumer type, or no shared schema files (typed Go constants checked in a switch).

#### Changed
- Documented tool-schema tree is now `schema/` instead of `json/`, named for what it holds rather than the file format.

### 2.2.0 (2026-10-02)
#### Added
- Notes that `destructiveHint` alone does not make Claude Code prompt the user, whereas the `anthropic/requiresUserInteraction` meta does.

### 2.1.0 (2026-10-02)
#### Added
- A schema-driven tool-contract pipeline for Go for servers past roughly ten tools: author the schemas when the tool surface differs from the API surface, derive them when the tool is a thin pass-through, and never share MCP enums with API enums.
- Three generator traps: the SDK cannot resolve a cross-document `$ref` for output schemas (panics at registration), go-jsonschema silently degrades referenced properties to `interface{}`, and `allOf` is opaque to the generator.
- The rule that a schema must describe the wire format, covering money as a string alias, re-wrapped envelopes and `omitempty` fields.
- Stack notes for go-jsonschema, including its module-root invocation.

Patch releases 2.1.1-2.1.2: corrected the `--only-models` guidance (it removes validation, so choose by whether anything outside the SDK unmarshals the structs), reworded the `$ref`-with-siblings claim, removed a duplicated recipe section, and set the Go SDK floor to v1.7.0.

### 2.0.0 (2026-08-10)
#### Changed
- Retargeted the skill from the 2025-11-25 spec to the final 2026-07-28 revision.
- A stateless protocol core is now assumed: no initialize handshake and no session ID, with routing by `Mcp-Method` / `Mcp-Name` headers.
- Multi Round-Trip Requests replace server-initiated elicitation, sampling and roots; list results must carry `ttlMs` and `cacheScope`.
- OAuth guidance hardened (issuer validation per RFC 9207, client ID metadata documents preferred over dynamic registration); Tasks moved to an extension; Roots, Sampling and Logging are treated as deprecated under the new lifecycle policy.

#### Breaking
- Session- and handshake-based designs from 1.x no longer match the guidance.

### 1.0.0 (2026-05-23)
Covers server-side MCP design (tools, resources, prompts, transports) with a short client-connection section, favouring stdio and Streamable HTTP, JSON-RPC error handling, OAuth 2.1 boundaries and MCP Inspector checks, with recipes in Python, Go and TypeScript.

## ml-conference-paper-architect

Guidance for writing and preparing ML conference papers (NeurIPS, ICML, ICLR, AAAI) that survive desk-reject checks.

### 1.0.0 (2026-08-10)
Covers per-venue requirements and desk-reject checklists, anonymization rules, page-limit accounting and how to obtain venue templates, with a LaTeX-based recipe set and a stack note on the NeurIPS styling approximation.

## nethttp-architect

Architecture standards for Go services built on the standard-library `net/http` server, ServeMux and middleware.

### 1.0.1-1.0.3 (2026-05-22 to 2026-10-12)
Patch releases only: the initial content landed with the catalogue build-out (Go 1.22+ method-pattern routing, middleware chains, mandatory production socket timeouts), then problem-details references were updated to RFC 9457 (replacing 7807), and the structure section gained a rule to split folders on cohesion rather than file count.

## nextjs-architect

Standards for Next.js 16 apps: App Router only, server components by default and explicit client boundaries.

### 1.0.0 (2026-05-22)
Covers the App Router paradigm with a hybrid data-access model (direct reads from server components, REST routes for mutations), server/client boundary auditing and standalone Docker output, with framework boilerplate moved into a recipes file.

Patch releases 1.0.1-1.0.2: updated the RFC citation to 9457 and added a rule to split folders on cohesion, not file count.

## observability-architect

Application-side observability: structured logs, Prometheus metrics, OpenTelemetry traces and what a service should emit.

### 1.0.0 (2026-05-22)
Standardizes metrics via Prometheus and traces/logs via OTLP, requires trace ID propagation and log-to-trace correlation, applies a 10% deterministic head-sampling default, and covers PII discipline and RED/USE naming.

## performance-reviewer

Cross-language performance review grounded in measurements rather than speculation.

### 1.0.0 (2026-05-22)
Reviews for N+1 queries, missing indexes, blocking I/O in async code, allocation hot paths, unbounded memory and slow algorithms, requiring findings to be backed by profiling or query-plan evidence (pprof, py-spy, EXPLAIN ANALYZE, metrics).

## pi-iteration-workflow

An edit-test deploy workflow for developing a robot codebase on a Raspberry Pi from a Windows machine.

### 1.0.0 (2026-07-07)
Covers the Windows-to-Pi SSH deploy loop, when pixi/ROS 2 rebuilds are needed, surviving flaky WiFi with `screen`, PowerShell SSH quoting pitfalls, and sudo caching on a Pi Zero.

## protobuf-architect

Protocol Buffers (proto3) schema design standards, language-agnostic.

### 1.0.0 (2026-05-22)
Covers Buf-style hierarchical package naming, explicit `optional` for patch semantics, enum zero-value rules, protovalidate (CEL) constraints, the buf toolchain for lint, breaking-change checks and generation, field-number reservation and well-known types.

## python-architect

Python 3.14 enterprise standards for writing, reviewing and scaffolding code.

### 1.3.0 (2026-07-08)
#### Added
- A strict `mypy.ini` template (strict baseline, extra checks, targeted overrides) and a curated `ruff.toml` for Python 3.14 whose rule families are mapped to the skill's sections, with carve-outs for Pydantic and FastAPI.
- Tooling section now categorizes ruff rule families, including the pytest-style family, and specifies mandatory CI gates.
- Guidance on `Enum`/`StrEnum`/`IntEnum` and on typed models (`dataclass`, `NamedTuple`, `TypedDict`) in place of raw dicts and tuples.
- Sections on generators versus iterators, mutable default arguments, ordering, and PEP 654 exception groups.

Patch releases 1.4.1-1.4.2: added testcontainers-python for container-backed integration tests (pinned in the stack notes) and a rule to split folders on cohesion rather than file count.

### 1.0.0 (2026-05-22)
Covers modern typing, immutable dataclasses, Protocol-based dependency injection, asyncio discipline, pytest, and psycopg with SQL kept in `.sql` files.

## react-architect

React 19 standards with strict TypeScript and feature-based structure.

### 1.0.0 (2026-05-22)
Covers hooks-first composition, TanStack Query for server state, zustand for cross-tree client state, Suspense and error boundaries at async boundaries, Radix for accessibility, and isolated feature folders, with boilerplate in a recipes file.

Patch release 1.0.1: added a rule to split folders on cohesion, not file count.

## repo-tooling-architect

Repo-root developer tooling: editor config, version pinning, task runner, CI, pre-commit, env handling and dependency updates.

### 1.4.0 (2026-10-14)
#### Added
- A `TASKFILE.md` reference with the task naming grammar, a canonical verb and module vocabulary, and a standard task set for Go libraries.

#### Changed
- The Go-library template was brought in line with that set (default task, dependency check, split scan).

### 1.3.0 (2026-09-27)
#### Added
- Docker-push-under-act gotchas: a registry-host variable to bypass a tunnel in front of a same-machine registry, host networking for the buildx builder, and the `tee`/`tail` exit-code trap that makes a failure look like a pass.

### 1.2.0 (2026-09-27)
#### Added
- A CI section: GitHub Actions by default, with `act` to run the same workflow locally when hosted runners are unavailable, covering `workflow_dispatch`-only triggers, `.actrc` and secrets-example conventions, task wiring, and an event-file approach for tag-gated jobs.

### 1.1.0 (2026-07-07)
#### Added
- A `module:verb` task naming convention so `task --list` groups alphabetically by module, with un-prefixed names reserved for repo-wide entry points and guidance on moving to `includes:` sub-taskfiles as a module grows.

#### Changed
- Taskfile examples were rewritten to the new pattern (for example `proto:generate`, `test:unit`, `docker:up`).

### 1.0.0 (2026-05-22)
Covers `.editorconfig` and `.gitignore` layouts, Taskfile as the default runner, dotenv-based configuration with secrets kept in an external manager, and Renovate update ranges.

## rest-api-architect

Cross-language, framework-agnostic REST conventions for designing and auditing endpoints.

### 1.1.0 (2026-07-18)
#### Added
- OpenAPI 3.1 as an automatically validated single source of truth: `redocly lint` in CI for syntax and consistency, and Scalar as a suggested renderer for interactive docs; both are pinned in the stack notes.

Patch release 1.1.1: cites RFC 9457 (which obsoletes 7807) for problem-details errors, corrects the Idempotency-Key reference to an IETF draft, and adds RFC 9457 notes to the payload examples.

### 1.0.0 (2026-05-22)
Covers resource URLs, snake_case JSON, cursor pagination, Idempotency-Key replays and ETag-based optimistic concurrency, with authentication patterns split into a separate reference file.

## reveal-js-architect

Architecture guidance for reveal.js 6 presentations, including the React wrapper.

### 1.0.0 (2026-07-16)
Covers section and fragment structure, HTML versus Markdown authoring, theming, plugin configuration, `Reveal.initialize()` options, vanilla setup and the `@revealjs/react` extension.

Patch releases 1.0.1-1.0.2: forbids mutating `position`, `overflow` or `display` on core `<section>` elements, fixes bundle import paths for reveal.js 6 (no `dist/` prefix, conditional-export entry points), and adds fixed-height slide canvas, nested-`em` typography and viewport-colour rules to avoid letterboxing.

## ros2-architect

ROS 2 workspace standards across Jazzy, Kilted and Lyrical, for scaffolding and reviewing packages.

### 1.0.0 (2026-05-22)
Covers colcon with ament_cmake/ament_python layouts, lifecycle nodes, explicit QoS profiles, services, actions and parameters, Python launch files, and reproducible cross-platform environments with Pixi and RoboStack, with C++ and Python treated equally and multi-distro parameterized templates.

## rsk-guide

Operator's guide to the `rsk` CLI for installing, pinning and syncing skills in a project, and to the `rsk mcp` server.

### 0.4.0 (2026-10-02)
#### Added
- Documentation of the `rsk mcp` server: its four primitives and the profile, propose, install flow, with a note that approval is enforced by the client via the `requiresUserInteraction` meta rather than by the skill.
- A wider trigger so the skill also applies when starting work in a project or scaffolding, not only when rsk is mentioned.

#### Changed
- `pin --remove` replaces `unpin` (kept as a hidden alias), and `--include personal|official` replaces the separate `--personal` and `--official` flags (also kept as hidden aliases).

#### Breaking
- `rsk catalog --bundle` and the positional `rsk catalog <name>` form are removed; use `rsk catalog bundles` to list bundles and `rsk catalog bundle <name>` to list one bundle's skills.

### 0.3.0 (2026-05-22)
Covers the declarative `.rsk/` project manifest, the consolidated command surface (`new`, `install`, `pin`, `update`, `status`, `list`), bundle and global installs, and the official Anthropic skill cache.

Patch release 0.3.1: republished to verify the registry publish flow, with no content change.

## secrets-architect

Standards for wiring credentials into repos, containers, systemd units and CI from a single external store consumed by reference.

### 1.1.0 (2026-10-10)

#### Added
- Build-secret rule: a Compose `build.secrets` entry must be sourced from the environment, never from a `file:` path outside the build context, because of the BuildKit/Bake filesystem-read sandbox.
- Pitfalls for private-module tokens used in a Go build step and for variables the caller forgets to export.

#### Changed
- The world-readable (644) file exception is generalized from a specific case to any non-root container reader.

Patch release 1.1.1: renamed the example CI token to `ci-modules`.

### 1.0.0 (2026-10-10)

Introduces the skill: one secret store consumed by reference, a decision matrix for choosing the mechanism per consumer (Compose secrets, `*_FILE`, systemd credentials, bind mounts, CI), a by-kind-then-owner layout with 700/600 permissions, `_FILE` naming, a per-secret migration workflow, a live verification checklist and a list of common pitfalls.

## security-reviewer

Cross-language security review covering injection, authn/authz, secrets, insecure defaults and dependency vulnerabilities, reported by severity with file and line references and fixes.

### 1.0.0 (2026-05-22)

Initial release: an audit pipeline that combines static scanners (semgrep, gitleaks, trivy) with focused contextual review for injection flaws and token leakage, producing a Critical/High/Medium/Low report with concrete fixes.

## skill-builder

Scaffolds a new ralvaskills skill from the spec, interview-first, without guessing.

### 1.0.0 (2026-05-22)

Initial release: scaffolds a skill skeleton (with optional STACK and RECIPES side files), places it in the right folder and validates it against lean-description and token-budget rules. Single-line descriptions and recipe extraction are part of the scaffold.

## sql-architect

SQL standards for schema design, queries and migrations, with PostgreSQL as the primary engine and MySQL and SQLite differences noted.

### 1.0.0 (2026-05-22)

Initial release: UUID v7 primary keys, snake_case naming, soft delete, opaque JSON columns, forward-only migrations that are safe on large tables, parameter-bound queries kept in raw `.sql` files, and engine-specific behaviour split into separate reference files.

Patch release 1.0.1: index naming now follows each engine's idiom (PostgreSQL `_idx`/`_key` suffixes, with MySQL omitting the table name and SQLite keeping it).

## study-artifact-builder

Builds a single-file interactive HTML artifact for studying a paper's pipeline or dataset until it can be defended from memory.

### 1.1.0 (2026-10-20)

#### Added
- A fixed explanation form for each pipeline stage: why it exists, inputs and outputs, paper-versus-code steps, a real example, measured figures, non-obvious details, consequences and what was not verified.
- Guidance on probing a lexical or rule-based stage on real data (copy the construction exactly, print distributions, sweep arbitrary thresholds, record the environment) with a runnable probe skeleton.
- Dataset checks for derived reference sets (whole-word presence, frequency outliers, coverage, consumer behaviour) and a numeric audit procedure for results tables.

#### Changed
- References to the shared contracts repository were renamed accordingly.

### 1.0.0 (2026-10-15)

Initial release: an artifact workflow covering source pinning, paper-versus-code separation, per-component notes, ablation toggles, a detailed SVG diagram and headless verification, with separate guides for pipelines and datasets and a recipes file of skeletons.

## tdd

Test-driven development with a red-green-refactor loop, integration-style tests and test-first feature and bug work.

### 1.0.0 (2026-05-22)

Initial release as part of the shared catalogue, with supporting guides on deep modules, interface design, mocking, refactoring and writing good tests. Frontmatter gained a version and author, and the guidance was made language-agnostic.

## ui-ux-architect

UI/UX standards for building components, design systems and accessible interfaces.

### 1.0.0 (2026-05-22)

Initial release: a mandatory WCAG 2.2 AA accessibility floor, Radix headless primitives with Tailwind 4 and shadcn/ui, design-token theming, four required async view states (loading, error, empty, success), mobile-first layout and keyboard parity.

Patch release 1.0.1: error-format references now cite RFC 9457 instead of the obsolete RFC 7807.

## uru-scientific-paper-architect

Formats scientific manuscripts to the Revista Tecnocientifica URU submission rules.

### 1.1.0 (2026-07-28)

#### Added
- Rule requiring every acronym to be expanded on first use in each manuscript section, plus a matching item in the review checklist.

### 1.0.0 (2026-07-17)

Initial release: a manuscript spec covering cover page fields, typography, article structure by type, figure/table/equation numbering, bracketed-numeric citations, reference formats and an ethics checklist.

## uru-thesis-architect

Interview-first drafter for URU undergraduate thesis chapters, written one section at a time in Spanish toward the jury's evaluation rubric.

### 0.1.0 (2026-07-17)

Initial release: drafts sub-sections iteratively in the URU chapter order, marks missing user data with explicit `[falta: ...]` tokens instead of inventing content, and bundles a self-contained copy of the URU 2020 thesis norms.

## uru-thesis-defense-architect

Designs URU thesis-defense slide decks as a slide-by-slide spec.

### 1.0.0 (2026-07-16)

Initial release: a 10-field interview protocol, a fixed nine-section institutional slide structure, Spanish-only content, logo placement rules, per-section content weighting and a review checklist. Patch release 1.0.1 only renamed the bundled logo asset.

## uru-thesis-reviewer

Continuous-feedback reviewer for URU theses that emits ordered diff files for the author to apply manually.

### 0.2.0 (2026-07-22)

#### Added
- Cross-session issue tracking with stable issue IDs, a status axis separate from severity, and mandatory carry-forward of unresolved issues from the previous session's index.
- A required reversal note whenever a decision is reverted, and tutor comments must be paired with verbatim quotes.
- A standard handoff file at the end of each session, and an index template with an issue-tracking table and a file map.

### 0.1.0 (2026-05-22)

Initial release: reviews substance, structure, prose and citations against the URU norms, with checklists and templates moved into side files and the norms document trimmed of OCR noise.

Patch release 0.1.1: generated output files now use lowercase, accent-free kebab-case names.

## website-concept-architect

Walks through designing a website concept before any code is written.

### 1.0.0 (2026-05-22)

Initial release: a six-stage interview (purpose, audience, content, tone, visual, synthesis) that produces two or three distinct concept directions and an ASCII homepage wireframe, then hands off to ui-ux-architect for implementation.
