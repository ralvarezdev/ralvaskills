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

## Per-project review checklist

- [ ] How many binaries / deployables? Separate release cadence or teams?
- [ ] How many real contexts (own model and language)? Which are entangled?
- [ ] Which features have real ports with 2+ adapters?
- [ ] Is there a god package or an import cycle today? Where?
- [ ] Does anything need to be extracted as a module or service soon?
- [ ] Is the dependency direction enforced by a test, or only by convention?
- [ ] Which layout above is the lightest that fits? What pain would trigger moving down?
- [ ] Where is the decision recorded, and when is it re-reviewed?
