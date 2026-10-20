# Documentation

Index and conventions for this folder. One doc type per folder; the folder and the
frontmatter `status` tell you the state.

| Folder | Holds |
|---|---|
| adr/ | Architecture decisions, numbered, immutable once accepted |
| backlog/ | inbox.md, one file per project, done/ per month |
| ideas/ plans/ designs/ | The path from thought to build |
| reviews/ changes/ | Dated records; never rewritten |
| guides/ methodology/ | Living procedures and working rules |
| templates/ archive/ | Doc templates; finished or obsolete docs |

Rules: frontmatter on every doc; relative links; drafts older than 2 months are promoted
or deleted; moving a file changes its state.
Routine: capture daily, triage weekly, prune monthly.
Check: `python3 <skill>/scripts/docs_lint.py docs`

## Index

## designs/

| Date | Doc | Status |
|---|---|---|
| 2026-05-20 | [ralvaskills — Technical Specification](designs/2026-05-20-technical-specification.md) | active |

## guides/

| Date | Doc | Status |
|---|---|---|
| 2026-05-22 | [Registry Flow](guides/registry-flow.md) | active |
| 2026-05-22 | [Registry Setup](guides/registry-setup.md) | active |

## plans/

| Date | Doc | Status |
|---|---|---|
| 2026-10-02 | [MCP Server Plan — `rsk mcp`](plans/2026-10-02-mcp-server-plan.md) | active |

