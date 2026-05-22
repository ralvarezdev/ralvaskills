---
name: rsk-guide
version: 0.1.1
description: Operator's guide for the rsk CLI — discover, install, update, check ralvaskills bundles and skills. Use when the user mentions rsk, asks how to install/list/update skills, or wants to set up ralvaskills. Companion to cli-tool-architect (which defines how to design any CLI) and skill-builder (which authors new skills). CLI is in progress; commands track SPECS.
---

# RSK Guide

> **Draft (v0.1.1).** The `rsk` CLI is in progress. Verify against `rsk --help` once shipped. For creating *new* skills (vs managing existing ones), use [`skill-builder`](../skill-builder/SKILL.md). For designing CLIs in general, see [`cli-tool-architect`](../../tooling/cli-tool-architect/SKILL.md).

## Always ask the user before taking action

Installing, updating, or uninstalling skills changes what tools and standards future agents will load in this and other projects. These are **architectural decisions**, not housekeeping. **Never run `rsk install`, `rsk update`, `rsk uninstall`, or any state-changing command without explicit user confirmation.**

For every proposed action:

1. **State the action plainly** — which skills/bundles, target scope (`--global` / `--for` / project-local), and what changes on disk
2. **Quote the exact command** you'd run
3. **Wait for the user to confirm**, redirect, or decline
4. **Use `--dry-run`** when available if the user wants to see the effect first

Read-only commands (`rsk list`, `rsk status` without `--stack`) are safe to run without asking — they only display state. But anything that touches symlinks, fetches official caches, or makes network calls (`rsk status --stack`) needs an explicit go.

If the user gives a blanket "yes" up front (e.g. "set everything up for this project"), still pause and confirm before each *bundle* — a slip from `go-grpc` to `go-rest` or installing globally when they meant locally is hard to undo silently.

## Quick reference

### Discover (read-only — safe)

| Command | Purpose |
|---|---|
| `rsk list` | All known skills |
| `rsk list --bundle <name>` | Skills in a bundle |
| `rsk list --installed` | Currently installed |
| `rsk list --source local\|official` | Filter by source |
| `rsk list --personal` | Include personal skills |
| `rsk status` | What's installed; version drift (no network) |

### Install (state-changing — ask first)

```bash
rsk install <bundle...> [--global] [--for claude-code|opencode] [--skill <name>] [--version <v>] [--personal] [--dry-run]
```

- Bundle install goes to `./.claude/skills/` by default; `--global` writes to all configured tools' global dirs (`--for` scopes to one).
- `--version` works only for local skills/bundles; official skills are not co-versioned.
- Planned bundle skills are skipped with a warning and installed automatically when they ship.
- Personal skills require `--personal` (opt-in).

### Update (state-changing — ask first)

```bash
rsk update [bundle|skill] [--global] [--for <tool>] [--official] [--personal] [--dry-run]
```

`rsk update` alone = `git pull` on the repo + re-symlink everything. `--official` also re-fetches the Anthropic skills cache.

### Status with live fetch (network — ask first)

```bash
rsk status [--global] [--for <tool>] [--project] [--stack] [--refresh] [--personal]
```

`--stack` fetches latest versions live from `proxy.golang.org` and `pypi.org` (opt-in network call, 24h cache; `--refresh` bypasses). Without `--stack`, no network calls.

### Uninstall (state-changing — ask first)

```bash
rsk uninstall <bundle|skill> [--global] [--for <tool>] [--personal]
```

## Common workflows

**First-time setup** (confirm each step with the user)
1. `rsk init` (once per machine — writes `~/.config/rsk/config.json`)
2. `rsk install global --global` — universal skills, machine-wide
3. In a project: `rsk install <stack-bundle>` — e.g. `rsk install go-grpc`

**Add a skill mid-project:** `rsk install --skill <name>`

**Check for outdated stacks:** `rsk status --stack`

**Switch to a pinned version (local skills only):** `rsk install <bundle> --version v1.1.0`

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

## Full reference

For the complete CLI specification — exhaustive flag list, `config.json` shape, discovery rules, edge-case behavior — see `docs/SPECS.md` in the [ralvaskills](https://github.com/ralvarezdev/ralvaskills) repo.
