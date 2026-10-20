# Templates

One template per doc type. On scaffold, write each block into `docs/templates/<type>.md` (kebab-case filenames: `adr.md`, `idea.md`, `plan.md`, `design.md`, `review.md`, `change.md`, `guide.md`, `backlog-item.md`). Replace the placeholders; delete sections that add nothing.

## docs/README.md

```md
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
```

## adr.md

```md
---
adr: 0
status: proposed
date: YYYY-MM-DD
project: global
---

# <Decision title>

## Context
<Forces and constraints. Why a decision is needed now.>

## Options considered
- **<A>** — <trade-off>
- **<B>** — <trade-off>

## Decision
<What we chose and the deciding reason.>

## Consequences
<What gets easier, what gets harder, what to revisit.>
```

## idea.md

```md
---
status: draft
project: <project>
date: YYYY-MM-DD
related: []
---

# <Idea>

<The problem or opportunity, in a few lines. No commitment.>

## Open questions
- <question>
```

## plan.md

```md
---
status: draft
project: <project>
date: YYYY-MM-DD
related: []
---

# <Plan title>

## Scope
<In / out.>

## Phases
1. <phase> — <outcome>

## Exit criteria
- [ ] <observable condition>
```

## design.md

```md
---
status: draft
project: <project>
date: YYYY-MM-DD
related: []
---

# <Design title>

## Approach
<How it is built. Link decisions: [ADR 0002](../adr/0002-slug.md).>

## Components and interfaces
<Only what a reader needs to implement or review.>

## Risks
- <risk> — <mitigation>
```

## review.md

```md
---
status: done
project: <project>
date: YYYY-MM-DD
related: []
---

# <What was reviewed>

## Findings
- **<severity>** <finding> — <evidence>

## Follow-ups
- <backlog item, linked>
```

## change.md

```md
---
status: done
project: <project>
date: YYYY-MM-DD
related: []
---

# <What changed>

## Why
<Motivation.>

## Impact and migration
<Who is affected; steps to adopt.>
```

## guide.md

```md
---
status: active
project: <project>
date: YYYY-MM-DD
related: []
---

# <How to do X>

## When to use
## Steps
1. <step>
## Troubleshooting
```

## backlog-item.md

```md
- [ ] <Imperative, specific task>  #<project>  P2
  plan: [<title>](../plans/YYYY-MM-DD-slug.md)
```

## Agent pointer section

For `CLAUDE.md` / `AGENTS.md`. Keep it this short; do not copy the conventions in.

```md
## Documentation
- Structure and conventions: `docs/README.md`
- Architecture decisions: `docs/adr/` — read before structural changes; never edit the body of an accepted ADR
- Pending work: `docs/backlog/<project>.md`; loose ideas go to `docs/backlog/inbox.md`
- New docs start from `docs/templates/` and carry frontmatter
- On a relevant change: update `CHANGELOG.md` and, if warranted, add `docs/changes/`
```

If both files exist, make one the source (`AGENTS.md`) and have `CLAUDE.md` hold only Claude-specific lines plus `@AGENTS.md`.
