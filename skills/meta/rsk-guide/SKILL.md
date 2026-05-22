---
name: rsk-guide
version: 0.2.0
description: Operator's guide for the rsk CLI — manage ralvaskills via the .rsk/ project manifest (rsk project init, rsk skill add/pin) or via bundle installs, plus global installs and the official Anthropic cache. Use when the user mentions rsk, asks how to install/pin/update skills, wants to add a skill to a project, or wants to set up ralvaskills on a new machine. Companion to cli-tool-architect (CLI design) and skill-builder (authoring new skills).
---

# RSK Guide

> **Draft (v0.2.0).** The `rsk` CLI is in progress. Verify against `rsk --help` once shipped. For creating *new* skills (vs managing existing ones), use [`skill-builder`](../skill-builder/SKILL.md). For designing CLIs in general, see [`cli-tool-architect`](../../tooling/cli-tool-architect/SKILL.md).

## Always ask the user before taking action

Installing, pinning, updating, or uninstalling skills changes what tools and standards future agents will load in this and other projects. These are **architectural decisions**, not housekeeping. **Never run `rsk install`, `rsk uninstall`, `rsk update`, `rsk project init`, `rsk project remove`, or any `rsk skill <subcommand>` without explicit user confirmation.**

For every proposed action:

1. **State the action plainly** — which skills/bundles, target scope (`--global` / `--for` / project manifest), and what changes on disk (including `.rsk/CLAUDE.md` and `./CLAUDE.md` edits when pinning)
2. **Quote the exact command** you'd run
3. **Wait for the user to confirm**, redirect, or decline
4. **Use `--dry-run`** when available if the user wants to see the effect first

Read-only commands (`rsk list`, `rsk skill list`, `rsk status` without `--stack`) are safe to run without asking — they only display state. Anything that touches symlinks, the manifest, the project `CLAUDE.md`, the official cache, or makes network calls (`rsk status --stack`) needs an explicit go.

If the user gives a blanket "yes" up front (e.g. "set everything up for this project"), still pause and confirm before each *bundle* or *pin*. A slip from `go-grpc` to `go-rest`, installing globally when they meant locally, or pinning a skill they didn't want auto-loaded is hard to undo silently.

## Two workflows

Pick the one that matches the user's intent.

### A. Project manifest (`.rsk/`) — preferred for per-project skill sets

Mirrors `go.mod`: a declarative `rsk.mod` lists the skills this project depends on, `rsk.lock` records resolved versions, and pinned skills are auto-imported into `./CLAUDE.md` via `@.rsk/CLAUDE.md`. Lives in `<project>/.rsk/`.

```bash
rsk project init                      # creates .rsk/rsk.mod + .rsk/CLAUDE.md + appends import to ./CLAUDE.md
rsk skill add <name[@version]>        # adds to manifest, symlinks into .rsk/skills/, updates rsk.lock
rsk skill add <name> --pin            # also imports skill into .rsk/CLAUDE.md (auto-loaded in this project)
rsk skill pin <name>                  # pin an already-added skill
rsk skill unpin <name>                # remove the import (skill stays installed)
rsk skill list                        # show all manifest entries with pinned/installed marks
rsk skill upgrade <name>              # re-resolve + re-link to latest available version
rsk install                           # (no args) re-installs everything in rsk.mod (after clone, after edits)
rsk project remove                    # delete .rsk/ and remove @.rsk/CLAUDE.md import
```

**Pin vs add** — added skills are present on disk but not auto-loaded; pinned skills are imported in `.rsk/CLAUDE.md` and load every turn in that project. Pin only what truly needs auto-loading.

### B. Bundle install — fastest path for stack-standard setups

Install a curated bundle into the project (`.rsk/skills/`) or globally (`~/.claude/skills/`, `~/.config/opencode/skills/`). Does **not** create or update `rsk.mod` — use workflow A if you want a tracked manifest.

```bash
rsk install <bundle...>               # project-local; goes to .rsk/skills/
rsk install <bundle...> --global      # all configured global tool dirs
rsk install <bundle> --global --for claude-code   # one tool only
rsk install --skill <name>            # single skill, no bundle
rsk install --skill demo-script-architect --personal   # personal skill, explicit opt-in
rsk install <bundle> --dry-run        # preview without writing
```

> Project-local bundle installs land in `.rsk/skills/` (the same dir the manifest workflow uses). If a `.rsk/rsk.mod` already exists, prefer workflow A so the manifest stays in sync.

## Quick reference

### Discover (read-only — safe)

| Command | Purpose |
|---|---|
| `rsk list` | All known skills |
| `rsk list --bundle <name>` | Skills in a bundle |
| `rsk list --installed` | Currently installed |
| `rsk list --source local\|official` | Filter by source |
| `rsk list --personal` | Include personal skills |
| `rsk list -o json` | Machine-readable output |
| `rsk skill list` | Project manifest entries |
| `rsk status` | What's installed across project + global; bundle tags; pinned tags (no network) |
| `rsk status --global` / `--project` | Restrict scope |
| `rsk status --for <tool>` | Scope `--global` to one tool |

### Setup (state-changing — ask first)

```bash
rsk init                              # once per machine — writes ~/.config/rsk/config.json
rsk init --force                      # overwrite existing config
```

`rsk init` prompts for:
1. **Skill source**: local clone (path to your `ralvaskills` repo) **or** hosted registry (`skills.ralvarez.dev`)
2. **AI tools to support**: `claude-code`, `opencode`, or both
3. **Global skills dir** per selected tool (defaults: `~/.claude/skills/`, `~/.config/opencode/skills/`)
4. **Default scope** for bare `--global` (`all` configured tools, or a single tool)

### Update (state-changing — ask first)

```bash
rsk update [bundle|--skill <name>] [--global] [--for <tool>] [--official] [--personal] [--dry-run]
```

- Local-clone mode: `git pull` on the repo (symlinks update automatically). Add `--official` to also pull `anthropics/skills` cache.
- Registry mode: fetches latest index, re-downloads + re-links any skill whose installed version is behind.

### Uninstall (state-changing — ask first)

```bash
rsk uninstall <bundle|--skill <name>> [--global] [--for <tool>] [--personal] [--dry-run]
```

Project-local uninstall also cleans the matching entries from `rsk.mod` / `rsk.lock` / `.rsk/CLAUDE.md` if a manifest exists.

### Status with live fetch (network — ask first; **not yet implemented**)

```bash
rsk status --stack [--refresh]        # planned: fetch latest versions from proxy.golang.org / pypi.org
```

Per the spec this is opt-in with a 24h cache; the current build returns an explicit "not yet implemented" error. Don't promise drift checks until it ships.

## Common workflows

**First-time machine setup** (confirm each step with the user)
1. `rsk init` — pick local clone or registry, configure tools
2. `rsk install global --global` — universal skills, machine-wide

**Set up a new project with tracked manifest**
1. `cd <project>` then `rsk project init`
2. `rsk skill add <name>` per skill (use `--pin` for ones that should always auto-load)
3. Commit `.rsk/rsk.mod`, `.rsk/rsk.lock`, `.rsk/CLAUDE.md`, and the `@.rsk/CLAUDE.md` import line in `./CLAUDE.md`

**Set up a new project with a stack bundle** (no manifest needed)
1. `cd <project>` then `rsk install <stack-bundle>` — e.g. `rsk install go-grpc`

**Bring an existing manifest project up to date after `git clone`** — `rsk install` (no args)

**Add a single skill mid-project**
- Tracked: `rsk skill add <name>` (+ `--pin` if it should auto-load)
- Untracked: `rsk install --skill <name>`

**Update everything** — `rsk update` (local clone) or `rsk update --official` (also refresh Anthropic cache)

**Upgrade a single manifest skill** — `rsk skill upgrade <name>`

## Bundle quick map

| Bundle | Source | Use for |
|---|---|---|
| `global` | local | Every machine — universal workflow + meta skills |
| `docs` | official | Document creation (docx/xlsx/pdf/pptx/find-docs) |
| `design` | mixed | Frontend / UI projects |
| `go-grpc` | local | Go gRPC services |
| `gin` | local | Go REST services (Gin) |
| `nethttp` | local | Go REST services (stdlib net/http) |
| `go-cli` | local | Go command-line tools |
| `fastapi` | local | Python REST services (FastAPI) |
| `python-grpc` | local | Python gRPC services |
| `python-cli` | local | Python command-line tools |
| `llm-app` | local | LLM apps / RAG pipelines |
| `ros2` | local | ROS2 robotics |
| `event-driven` | local | Schema-first messaging (NATS/Kafka/RabbitMQ) |
| `observability` | local | Signals + dashboards/alerts |
| `code-review` | local | Reviewer triad: security, contracts, performance |

## What's in `.rsk/`

```
.rsk/
├── rsk.mod         # TOML: version, pinned[], skills{name = version-constraint}
├── rsk.lock        # TOML: resolved {name, version, source, path} per skill
├── skills/         # symlinks into the local repo or registry cache
│   └── <name>/     # → ralvaskills/skills/.../<name>/  or  registry cache
└── CLAUDE.md       # one `@skills/<name>/SKILL.md` line per pinned skill
```

`./CLAUDE.md` gets a single appended line: `@.rsk/CLAUDE.md`. Both `rsk project init` (creates it) and `rsk project remove` (removes it) are idempotent.

## Full reference

For the complete CLI specification — exhaustive flag list, `config.json` shape, catalog overrides, discovery rules, edge-case behavior — see `docs/SPECS.md` in the [ralvaskills](https://github.com/ralvarezdev/ralvaskills) repo.
