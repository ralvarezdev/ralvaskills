# ralvaskills (`rsk`) — agent guide

The skill registry and the `rsk` CLI. `skills/` holds the published catalogue (grouped by
domain); `cmd/rsk` is the CLI that installs, pins and syncs skills into a project; `internal/mcp`
is the `rsk mcp` server; `registry/` is the generated index served at
`https://skills.ralvarez.dev/index.json`.

> **Publishing is load-bearing.** Pushing `main` runs the registry workflow and **auto-publishes
> every changed skill**. Do not edit `skills/**` unless the change is meant to ship, and never put
> the user's private project names or details into a bundled skill. Planning docs go in `docs/`,
> never `skills/`.

## Commands

- `task build` — build the CLI; `task cli:build`, `task cli:run -- <args>`, `task cli:lint`,
  `task cli:fmt`, `task cli:tag` (release)
- Gates: `task lint`, `task fmt`, `task lint:oneliners`, `task lint:separators`,
  `task lint:doccheck`, `task deadcode`, `task deps:check`
- Registry: `task registry:generate`, `registry:sync-index`, `registry:check`, `registry:clean`,
  `registry:remove[*]`, `registry:lint`, `registry:fmt`
- Tokens: `task tokens`, `task tokens:lint`, `task tokens:fmt`

Run `task lint` (and the registry/token checks when those areas change) before declaring done.

## Layout

- `cmd/rsk/` — the CLI: subcommands, flags, output. Subcommands: `init`, `new`, `destroy`,
  `install`, `uninstall`, `update`, `pin`, `tools`, `catalog`, `list`, `status`, `completion`,
  `mcp`.
- `internal/` — `catalog` (indexing), `config` (machine config `~/.config/rsk/config.json`,
  bundles), `cmdx` (cobra helpers), `fsx` (atomic writes), `install` (source selection,
  constraint solving), `manifest` (`rsk.mod`/`rsk.lock`), `mcp` (project detection, tool
  registration, resources), `skill`, `source` (registry/local/official), `tool` (the Claude /
  OpenCode / MCP adapters), `ui` (termkit), `update`.
- `cmd/{generate-registry,registry-remove,count-tokens}` — registry build helpers.
- `skills/<domain>/<name>/SKILL.md` — the catalogue (domains: academic, databases, design,
  encoding, frameworks, frontend, infra, languages, messaging, meta, protocols, quality,
  refactoring, robotics, tooling, workflows). One skill per directory; the directory name matches
  the frontmatter `name`.
- `registry/index.json` — the published index; `docs/` — design docs (e.g. `MCP_SERVER_PLAN.md`,
  `REGISTRY_FLOW.md`, `SPECS.md`).

## Conventions (non-obvious only)

- **Skill frontmatter:** `name` (≤64, lowercase/hyphen, must match the directory), `description`
  (≤1024; says *what* and *when*), optional `license`, `compatibility`, `metadata`,
  `allowed-tools`. Only the description loads at start; the body loads on activation, so keep it
  under ~500 lines and use `scripts/`/`references/`/`assets/` for bulk.
- **Publishing:** bump the skill's version in its frontmatter, push to `main`; CI diffs against
  `registry/index.json`, builds a tarball per new/changed skill, cuts a
  `<skill>@v<version>` release, backs it up to R2 and commits the updated index
  (`docs/REGISTRY_FLOW.md`).
- **`.rsk/` per project:** `rsk.mod` (version, `tools` scope, skills, pinned) and `rsk.lock`
  (resolved versions + cache paths). `rsk new`/`pin` regenerate the adapter files; a second run is
  idempotent. The Claude adapter writes `.rsk/CLAUDE.md` (skill imports) and appends
  `@.rsk/CLAUDE.md` to the project's `CLAUDE.md`; the OpenCode adapter writes `opencode.json`'s
  `instructions` array. **rsk does not generate or manage AGENTS.md today** — that is a P6 design
  (a delimited, generated skills block in AGENTS.md or `.rsk/AGENTS.md`), see the rsk↔AGENTS.md
  design in the finance-platform repo's `docs/2026-10-12-ralvaskills-rsk-agents-plan.md` §7.
- **Skill sources:** `registry` (published), `local`, `official` (Anthropic). `source` resolves
  and caches under `~/.config/rsk/cache/`.
- Keep the CLI's data-on-stdout / logs-on-stderr contract and `--output json` where present.

## Agent surfaces

- **`rsk mcp`** — stdio server (launched and killed per client session; registered in
  `.claude/settings.json` under `mcp`). Planned tools: `search_skills` (read-only),
  `install_skills` (destructive; `_meta["anthropic/requiresUserInteraction"] = true`),
  `project_profile` (read-only), plus the `rsk://catalog` resource. Read-only tools carry
  `readOnlyHint`; `install_skills` carries `destructiveHint`. In progress (design approved
  2026-10-02); see `docs/MCP_SERVER_PLAN.md`.
- **`rsk` CLI** — the human-facing surface; `rsk <subcommand> --help`. Skills pinned for this repo
  itself (via `rsk`): cli-tool-architect, go-architect, logic-cleaner, repo-tooling-architect,
  skill-builder.
- **Registry** — `https://skills.ralvarez.dev/index.json` (GitHub Pages); consumers `rsk update`
  then `rsk install <name>`.

## Boundaries

- **Always:** bump a skill's version to publish it; keep a bundled skill generic; run `task lint`
  and the registry checks before done.
- **Ask first:** editing anything under `skills/**` (it publishes on push); regenerating
  `registry/index.json`; changing adapter config-writing (`internal/tool/*`) — it touches every
  consumer project.
- **Never:** write the user's private project names or details into `skills/**`; put planning docs
  in `skills/` (use `docs/`); commit or push unless asked; commit secrets.

## Reference

- `README.md` — what rsk is and how to install it
- `docs/MCP_SERVER_PLAN.md` — the `rsk mcp` design
- `docs/REGISTRY_FLOW.md` and `docs/REGISTRY_SETUP.md` — publishing pipeline
- `docs/SPECS.md` — skill/catalogue specification
- The AGENTS.md ↔ rsk design: `2026-10-12-ralvaskills-rsk-agents-plan.md` §7 (kept in the
  `finance-platform` repo's `docs/`)
