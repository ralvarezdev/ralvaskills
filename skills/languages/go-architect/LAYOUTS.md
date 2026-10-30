# Repo layouts for Go projects

There is no single correct layout — only better fits for a project's shape. This file is a
catalogue to **review per project** and to revisit when the project changes. [SKILL.md](SKILL.md)
§10, §16 and §17 hold the rules; this file holds the options, trade-offs and a review checklist.

## How to use this file

1. Fill in the review checklist at the bottom for the project.
2. Pick the **lightest** layout that fits. Move down the list only when a concrete pain appears
   (god package, import cycle, a context that must be extracted, a team boundary).
3. Record the choice and its reason in the repo (an ADR or the repo's `AGENTS.md`).
4. Re-review when contexts, binaries or teams change.

## Selector

| Project shape | Layout |
|---|---|
| Library or SDK | 1. Library |
| One binary, few concepts, CRUD-ish | 2. Flat by concept |
| Several features, one team, one deployable | 3. Feature-first |
| Domain-rich features, real ports, 2+ adapters | 4. Feature-first, escalated |
| Many contexts in one repo | 5. Grouped at the root |
| Several deployables, separate teams or release cadence | 6. Multi-module workspace (§16) |
| Tightly entangled contexts, one team, one binary | 7. Layer-first |

Module count, process types, generated code, tests and ownership are orthogonal to this choice — see the sections after the layouts.

## Layouts

### 1. Library
Root package is the public API; `internal/` for everything else; no layer folders. Examples,
`cmd/` helpers and generated code live beside, never inside, the API package.

### 2. Flat by concept
```
cmd/<binary>/main.go
internal/<concept>/      one package per concept, file-prefixed inside
pkg/                     only if something is genuinely reusable outside the repo
```
Cost: none until a package becomes a god package; that is the signal to move to 3.

### 3. Feature-first
```
<feature>/
    *.go        model + rules + use cases + the interfaces it consumes, flat, concept-named
    <impl>/     implementations named by technology or variant (postgres, nats, docker)
    internal/   private helpers
cmd/<binary>/   main.go only
internal/<binary>/   composition root (avoids overloading `app`)
```
Interfaces are declared next to the consumer; fakes live in `fake/`, `testing/` or `_test.go`.
Cost: model and use cases share a package, so purity is by convention and by test, not by compiler.

### 4. Feature-first, escalated
```
<feature>/
    core/                 model, rules, ports the rules declare — no framework imports
    <use-case layer>/     operations | usecase | application — named for what it holds
    adapters/<tech>/      everything touching a technology
```
Use for domain-heavy features only (real invariants, 2+ adapters per port, a model that must be
compile-time-isolated). Skip any layer a feature does not need. Avoid stutter (`order/core/core`).

### 5. Grouped at the root
```
platform/<ctx>/     group named for the area, not a generic `contexts/`
billing/<ctx>/
pkg/  contracts/  cmd/  internal/
```
Introduce groups past ~7 contexts so the root stays at a handful of folders; cap groups at ~5–7.
Keep reserved names (`pkg`, `contracts`, `cmd`, `internal`) distinct from group names. With
fewer contexts, put them straight at the root. `**/` path patterns make the rank rule
indifferent to grouping.

### 6. Multi-module workspace
`go.work` with one module per context or per group; heavy or shared adapters can be their own
module (`<ctx>-<tech>`) so `go.mod` — not a test — forbids the import. Extract a group once it
passes ~8–10 contexts or gets its own team. See SKILL.md §16.

### 7. Layer-first
```
core/<ctx>/  services/<ctx>/  adapters/<ctx>/  contracts/  pkg/  cmd/  internal/
```
Makes one rank rule trivial to enforce for tightly interdependent contexts owned by one team.
Costs: a context lives in three trees and cannot be lifted out as a module or deployable;
shared layers attract catch-all packages; it cuts against §10. If contexts are not really
separate units, redraw the contexts before choosing this.

## Module count and versioning

One `go.mod` is the default. Add modules only for a concrete reason:

| Reason to split | Cost to weigh |
|---|---|
| Independent release cadence or semver (a published library inside a service repo) | Tag per module (`<path>/v1.2.0`), `replace` directives for local dev, changelogs per module |
| Different dependency sets that must not leak (a heavy adapter, a CLI vs a server) | More `go.mod`/`go.sum` to keep in step; Renovate/Dependabot noise |
| Team boundary where `go.mod` should forbid imports | CI must build and test each module; tooling that assumes one module (some linters, coverage) needs `go.work` or loops |
| Extractable context or group | Cross-module refactors become multi-step |

A single module with many `cmd/*` binaries and an enforced rank rule covers most repos. Do not split only because the tree is big. Major-version directories (`v2/`) and versioned `api/` packages belong to libraries and wire contracts, not to internal layout.

## Process types

The layouts above describe the model. The process shape on top differs:

- **HTTP/gRPC API:** `cmd/<api>/main.go` parses flags and env, `internal/<api>` wires the graph and server lifecycle, transports sit beside the use cases (`<feature>/http`, `<feature>/grpc`) or in `<feature>/<transport>/`.
- **CLI:** `cmd/<cli>/main.go` calls `internal/cli`; one file per subcommand; commands call feature APIs and never hold business rules.
- **Worker / consumer:** handlers per message type live in the feature (`<feature>/consumer`), the subscription and retry wiring lives in the composition root. Keep the outbox and idempotency logic with the feature that owns the event.
- **Several processes, one codebase:** one `cmd/<binary>` and one `internal/<binary>` each; features are shared, wiring is not.

## Generated code and SQL

- **Where:** next to the consumer that owns the schema — `<feature>/<impl>/db` for sqlc output and `//go:embed` SQL, `contracts/<name>/gen` for proto/OpenAPI. A central `gen/` is fine when output is shared by many features.
- **Committed or not:** commit generated code when consumers must build without the generator; otherwise generate in CI. Pick one rule per repo and write it down.
- **Tooling:** exclude generated paths from the rank test, linters and coverage by pattern (`**/gen/**`, `*.pb.go`, `*.sql.go`), and mark files `// Code generated ... DO NOT EDIT.`
- **Migrations:** per-service next to the schema owner, one numbered sequence per database, not per feature.

## Cross-context communication

- **Integration events and shared contract types** live in the publishing context's contract package (or `contracts/` when wire-only). Consumers depend on the contract, never on the publisher's model.
- **Shared kernel** stays tiny and deliberate: value objects only, no behaviour that one context would want to change alone.
- **Rule:** a context never imports another context's `core/`, feature root or adapters. If two contexts need the same type, either the kernel owns it or each defines its own and a translator sits at the boundary.

## Test layout

- Unit tests sit beside the code (`_test.go`); `testdata/` beside the package that reads it.
- Integration and end-to-end tests that need Docker or a network live in the package they exercise behind a build tag (`//go:build integration`), or in `<feature>/integration` when they span implementations.
- Shared helpers go in `internal/testutil` (or a `<feature>/fake` package when they are specific), never in `pkg/`.
- The architecture test lives in one place (`internal/archtest`), may import anything to inspect it, and is itself exempt from the rank rule. Test code may import across layers; production code may not.

## Ownership and CI scoping

Layout follows team boundaries. Align top-level directories with `CODEOWNERS` entries and with path-filtered CI jobs, so a change to one group runs one group's checks. If a directory has no single owner, it is usually either a shared library that needs one, or two contexts that were never separate.

## Non-Go parts of the repo

- **Polyglot monorepo:** put Go under one subtree (`backend/`, `go/`) with `go.mod` or `go.work` at its root, so Go tooling and module paths are unambiguous.
- **Ops and tooling:** `deploy/`, `scripts/`, `tools/` (build-time Go tools, pinned via `tools.go` or the `tool` directive), `docs/`. Keep them out of `cmd/` and `internal/`.
- **The root** holds `go.mod`/`go.work`, the task runner file, `README`, `AGENTS.md`, CI config and `CODEOWNERS` — little else.

### App repo with a SPA — the recommended shape

Most app repos pair a Go module with a JavaScript/TypeScript single-page app. Default to the
shape below when that is the case: one Go module, one SPA, wire contracts beside them.

```
repo/
├── backend/                 the Go module: go.mod lives here, nothing above it
│   ├── go.mod  go.sum  sqlc.yaml  .air.toml
│   ├── Dockerfile[.…]
│   ├── cmd/<binary>/        main.go only: flags/env, then internal/<binary>
│   ├── internal/<binary>/   composition root per binary, plus private helpers
│   ├── <ctx>/               bounded contexts, flat; postgres/, transport/<proto> as siblings
│   ├── kernel/              tiny shared kernel (value objects only)
│   ├── pkg/                 only code another repo could import
│   └── db/                  migrations
├── frontend/                the SPA; its build output is copied into the Go embed dir
├── contracts/               source of truth: openapi/, proto/, schema/
├── config/  bin/  scripts/  docs/  deploy/  mockups/
└── AGENTS.md  README.md  CHANGELOG.md  Taskfile.yml  mise.toml
```

Rules, in the order they bite:

1. **One module, rooted at the subtree that holds only Go.** `backend/` when the repo has another
   toolchain, the repo root when it is Go only. Add a second module only for one of the reasons in
   *Module count and versioning*; a big tree is not one.
2. **The module path names the directory that holds `go.mod`**
   (`github.com/<owner>/<repo>/backend`, not `github.com/<owner>/<repo>` or a bare `myapp`). A path
   that points above its own `go.mod` makes `go install`/`go get` resolve to a directory with no
   module, and every import path lies about where the file is.
3. **Each binary owns a composition root; `main.go` holds wiring only.** The root is named for
   the binary (`internal/<binary>`) or for what it is (`internal/bootstrap`, `internal/api`) — the
   same list as the cross-cutting rule below. With one binary a single shared root is enough;
   with two or more, give each binary its own so a change to one lifecycle cannot disturb another.
4. **Contexts stay flat** at the module root, implementations as sibling subpackages
   (`postgres/`, `transport/mcp/`, `remote/grpc/`). Group under `ledger/`, `billing/` only past
   ~7 contexts — see layout 5.
5. **Operational directories exist once, at the repo root:** `config/`, `bin/`, `scripts/`,
   `docs/`, `deploy/`. A `backend/scripts/` next to a root `scripts/` is two homes for one concern,
   and `backend/config/` splits configuration by language instead of by scope.
6. **Contracts are the source of truth; generated code lives inside the module; embedded assets
   live under `internal/`.** Do not mirror `contracts/openapi/` as a second `openapi/` at the
   module root: put the merged artefact and its `embed.go` in `internal/<name>/` so the two are not
   mistaken for each other.
7. **A frontend build is copied, never reached with `..`.** `//go:embed` cannot escape its
   package directory, so a sibling `frontend/` needs an explicit copy step
   (`cp -r frontend/dist internal/static/dist`) in the task runner, CI and the Dockerfile. That
   step is deliberate; do not "simplify" it away, and do not put a Go package inside `frontend/`
   to avoid it.
8. **An adapter never imports another adapter's SDK.** A REST client used by a CLI or a worker must
   not depend on the MCP SDK (or any other transport's) for a helper: the caller owns token
   resolution and passes it in. Otherwise every binary in the repo links every transport.

#### Naming the frontend directory

There is no ecosystem winner: among ~30 active Go repos with a SPA, `web/`, `ui/` and `frontend/`
are used by about a third each, and the rest name the product (`site/`, `console/`, `portal/`,
`dashboard/`, `client/`). Pick by the rule that makes the tree read as one thing:

| Directory | Reads as | Use when |
|---|---|---|
| `frontend/` | the other toolchain | the Go module is under `backend/` — the pair is the point |
| `web/` | the web face of the product | the module is at the repo root and the SPA is a compiled asset |
| `ui/` | the UI of this tool | the SPA is embedded with `//go:embed` (desktop-ish or single-binary tools) |
| `site/` `console/` `portal/` `dashboard/` | a product surface with its own name | the product has several, or the name is part of the vocabulary |
| `web_src/` | sources only, build at the root | the build system must own the root (no `package.json` inside) |

`frontend/` without a `backend/` reads as a leftover; `backend/` + `web/` reads as a rename in
progress. Keep the pair consistent, and rename the *pair*, not one half.

## Smells that mean the layout needs review

- Packages named `utils`, `common`, `helpers`, `models`, `types`, `shared`.
- One global `domain/` or `models/` holding every context.
- `pkg/` used as a dumping ground for anything "reusable" with a single consumer.
- Directories nested deeper than about 3–4 levels, or one package per type.
- A package with hundreds of files and no internal structure (god package), or two packages that always change together.
- Import cycles worked around with interfaces in a third package instead of moving the code.
- Rank-rule exceptions that only grow.
- `app` (or any layer name) meaning two different things in two places.

## Cross-cutting rules (all layouts)

- **Enforce direction with a test.** Rank the layers (`pkg`/`contracts` < feature root or `core`
  < use-case layer < implementations/`adapters` < `internal/<binary>` < `cmd`); each may import
  its own rank or lower. Use `**/` patterns, add an import-level check where a package mixes
  layers, and record exceptions in the test so they stay visible and shrink. Across contexts,
  import only another context's public API or contract package.
- **No infrastructure in the model.** Clients, workers, queue consumers and SQL do not live in
  `domain/`, `core/` or a feature root.
- **Wire schemas get their own home** (`contracts/`), apart from `pkg/` libraries and any domain.
- **Define "context" tightly.** A real consistency boundary with its own model and language, not
  one folder per noun. Merge tiny ones into a neighbour.
- **Name the composition root for what it is** (`internal/<binary>`, `bootstrap`), not `app`,
  when `app` already means a layer elsewhere.

## What large Go codebases tend to do

Component- or subsystem-first with flat, concept-named packages — not a fixed per-context layer
triplet. The component root holds model, logic and interfaces; implementations are sibling
subpackages by variant; subpackages are split by concern only when self-contained; big packages
are organized by file prefix before folders; `internal/` holds unstable helpers; shared types or
APIs get their own package or module; `cmd/<binary>` stays thin; import direction is checked
mechanically (import-restriction files, `depguard`, custom `go list` tests). Staying flat forever
ends in a god package — that is the signal to extract a concern or escalate the feature, not a
reason to adopt layers up front.

## Migrating an existing tree

1. Record the rank rule in the architecture test first, with no files moved.
2. Extract wire contracts and generic libraries (`contracts/`, `pkg/`).
3. Rename layers mechanically (one stage, one guard).
4. Split domain from adapters context by context, last, each behind its own guard test, and
   outside windows where other sessions are editing the same files.
5. Normalize the module boundary last: module root, module path, and the operational directories
   (`config/`, `bin/`, `scripts/`) at the repo root. It touches CI, Docker files and every doc, so
   do it in one window with no functional change, and land no other refactor in it.

## Per-project review checklist

- [ ] How many binaries / deployables? Separate release cadence or teams?
- [ ] How many real contexts (own model and language)? Which are entangled?
- [ ] Which features have real ports with 2+ adapters?
- [ ] Is there a god package or an import cycle today? Where?
- [ ] Does anything need to be extracted as a module or service soon?
- [ ] Is the dependency direction enforced by a test, or only by convention?
- [ ] Which layout above is the lightest that fits? What pain would trigger moving down?
- [ ] Does each module split have a concrete reason (release, deps, team, extraction)?
- [ ] Which process types exist (API, CLI, worker), and is wiring per binary?
- [ ] Where do generated code, SQL and migrations live, and are they committed?
- [ ] Is the module rooted at the subtree that holds only Go, and does the module path name that
      directory?
- [ ] Does each binary have a wiring-only `main.go` and its own composition root (`internal/<binary>` or a named one)?
- [ ] Do `config/`, `bin/`, `scripts/`, `docs/` and `deploy/` exist once, at the repo root?
- [ ] Does the frontend build get copied into the embed directory, and is that step documented?
- [ ] Does any adapter import another transport's SDK?
- [ ] How do contexts communicate; where do integration events and contracts live?
- [ ] Are directories aligned with `CODEOWNERS` and path-filtered CI?
- [ ] Which smells above are present today?
- [ ] Where is the decision recorded, and when is it re-reviewed?
