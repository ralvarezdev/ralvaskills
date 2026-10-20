# Frontmatter Schemas

Two schemas. Frontmatter is mandatory on every doc in `docs/` except `docs/README.md`, `backlog/inbox.md`, `backlog/<project>.md`, `backlog/done/*.md` and `templates/*`.

## Standard docs (ideas, plans, designs, reviews, changes, guides, methodology)

```yaml
---
status: draft | active | done | obsolete
project: plc-agent        # project name, or "global"
date: 2026-10-20          # creation date; ISO 8601
related: [adr/0003-slug.md, plans/2026-10-12-topic.md]
---
```

- `status` lifecycle: `draft` → `active` → `done` → `obsolete`. Moving a doc to `archive/` accompanies `done` or `obsolete`.
- `date` is the day the doc was created, not last edited. `guides/` and `methodology/` keep it for provenance but their filenames are undated.
- `related` holds paths relative to `docs/`. Origins go here: a review that produced a backlog item is listed on the item and the item is listed on the review.
- `reviews/` and `changes/` are snapshots — normally written once with `status: done`.

## ADRs

```yaml
---
adr: 3
status: proposed | accepted | superseded | deprecated
date: 2026-10-20
project: global
supersedes: 1        # optional
superseded_by: 7     # filled when a later ADR replaces this one
---
```

- `adr` equals the number in the filename (`0003-slug.md` → `3`).
- `supersedes` / `superseded_by` must be symmetric. If ADR 7 has `supersedes: 1`, ADR 1 has `status: superseded` and `superseded_by: 7`.
- Only `status` and `superseded_by` change after acceptance.

## Why mandatory

An agent or a script cannot tell a live plan from a dead one by reading prose. `status` and `date` make triage, drafts-older-than-2-months checks and index generation mechanical.
