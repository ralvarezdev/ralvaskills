---
name: rsk-guide
version: 0.1.0
description: >
  Quick reference for the `rsk` CLI — how to discover, install, update, and check
  ralvaskills bundles and individual skills. Use when the user mentions `rsk`,
  asks how to install/list/update skills, wants to set up ralvaskills, or wants
  to check what's installed and whether it's up to date. CLI is currently planned;
  commands below match the SPECS design.
---

# RSK Guide

> **Draft (v0.1.0).** The `rsk` CLI is in the roadmap. Verify against `rsk --help` once shipped. For creating *new* skills (vs managing existing ones), use `skill-builder`.

## Quick reference

### Discover

| Command | Purpose |
|---|---|
| `rsk list` | All known skills |
| `rsk list --bundle <name>` | Skills in a bundle |
| `rsk list --installed` | Currently installed |
| `rsk list --source local\|official` | Filter by source |
| `rsk list --personal` | Include personal skills |

### Install

```bash
rsk install <bundle...> [--global] [--for claude-code|opencode] [--skill <name>] [--version <v>] [--personal] [--dry-run]
```

- Bundle install goes to `./.claude/skills/` by default; `--global` writes to all configured tools' global dirs (`--for` scopes to one).
- `--version` works only for local skills/bundles; official skills are not co-versioned.
- Planned bundle skills are skipped with a warning and installed automatically when they ship.
- Personal skills require `--personal` (opt-in).

### Update

```bash
rsk update [bundle|skill] [--global] [--for <tool>] [--official] [--personal] [--dry-run]
```

`rsk update` alone = `git pull` on the repo + re-symlink everything. `--official` also re-fetches the Anthropic skills cache.

### Status

```bash
rsk status [--global] [--for <tool>] [--project] [--stack] [--refresh] [--personal]
```

`--stack` fetches latest versions live from `proxy.golang.org` and `pypi.org` (opt-in network call, 24h cache; `--refresh` bypasses). Without `--stack`, no network calls.

### Uninstall

```bash
rsk uninstall <bundle|skill> [--global] [--for <tool>] [--personal]
```

## Common workflows

**First-time setup**
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
| `go-rest` | local | Go REST services (Gin) |
| `go-cli` | local | Go command-line tools |
| `fastapi` | local | Python REST services (FastAPI) |
| `llm-app` | local | LLM apps / RAG pipelines |
| `ros2` | local | ROS2 robotics |

## Full reference

For the complete CLI specification — exhaustive flag list, `config.json` shape, discovery rules, edge-case behavior — see `docs/SPECS.md` in the [ralvaskills](https://github.com/ralvarezdev/ralvaskills) repo.
