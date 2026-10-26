# ralvaskills-executor-port — agent guide

Not a standalone project: a `git worktree` of `ralvarezdev/ralvaskills`, checked
out on `feat/go-architect-external-executor` to rework the go-architect skill
around an external program behind a port. Everything here is the base repo's
catalogue and `rsk` CLI — read its guide at `../ralvaskills/AGENTS.md` and this
checkout's `docs/SPECS.md`; only the executor work is local to this branch.

## Commands

- Build: `task build` (→ `bin/rsk`); run without building: `task cli:run -- init`
- Test: `task test`; lint/format: `task lint`, `task fmt`, `task lint:doccheck`
- Registry preview: `task registry:check` (nothing is published)
- Generated token table: `task tokens`
- Dependency resolution from an empty cache: `task deps:check`

Tool versions come from `mise.toml` (Go 1.27.1, golangci-lint 2.14.0, deadcode).

## Layout

```
skills/languages/go-architect/   the skill under change: SKILL.md, STACK.md,
                                 LAYOUTS.md, RECIPES.md, assets/
skills/<domain>/<name>/          the rest of the catalogue, untouched here
cmd/{rsk,generate-registry,registry-remove,count-tokens}
internal/                        catalog, install, manifest, mcp, tool, source, ui
registry/index.json              published index; docs/ design docs
```

## Conventions

- The work is prose in a skill: a new rule goes in `SKILL.md`, its worked code in
  `RECIPES.md` under a matching heading, and the frontmatter `version` is bumped
  (currently 1.11.0) — CI cuts a release only for a changed skill with a new
  version.
- The uncommitted edits in this worktree (`SKILL.md`, `RECIPES.md`) are the
  branch's work in progress; check `git status` before staging, and never stage
  another checkout's files.
- Skills are the product; `internal/` and the Taskfile belong to the base repo and
  drive the `rsk` CLI for every project, so a change there is not local to this port.

## Boundaries

- Always: run `task lint` (and `task test` when Go code moved) before done.
- Ask first: editing `skills/**` beyond the go-architect work, and regenerating or
  syncing `registry/index.json` — pushing `main` publishes every changed skill.
- Never: push this branch's work to `main` or run `task registry:sync-index`
  without asking; put private project names or details into a bundled skill.
