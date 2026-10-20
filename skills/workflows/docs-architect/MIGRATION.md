# Migration

Bringing an existing `docs/` (and its backlogs) into the convention. **Propose first, move second.** Nothing is moved, renamed or deleted before the user approves the plan.

## 1. Inventory

List every doc and root-level planning file. For each, read the title, the first lines and `git log --follow` for its dates. Do not classify from the filename alone.

## 2. Classify

Assign each doc one type from [SKILL.md §2](SKILL.md#2-where-a-doc-goes) by its job:

| Signal | Type |
|---|---|
| Records one choice among alternatives | `adr/` |
| States scope, phases, "what we will do" | `plans/` |
| Describes how something is built | `designs/` |
| Findings from examining something | `reviews/` |
| Describes something already changed | `changes/` |
| Procedure someone follows repeatedly | `guides/` |
| Standing rule or convention | `methodology/` |
| Loose thought, no commitment | `ideas/` |
| Superseded, finished, or the code no longer matches | `archive/` |

Mixed docs (plan + design in one file): keep as the dominant type and note the split as a follow-up; do not split during migration.

## 3. Plan table

Present one table and wait:

| Current path | Type | New path | Frontmatter to add | Notes |
|---|---|---|---|---|

Rules for the new paths: kebab-case, lowercase, no accents; dated types get `YYYY-MM-DD-` from the first commit date; ADRs are numbered in chronological order. Fix inbound links in the same pass and list files with no determinable project or date as questions, not guesses.

## 4. Execute

Use `git mv` so history follows. Add frontmatter, rewrite relative links, then run `python3 scripts/docs_lint.py docs` until it is clean. Commit only when the user asks.

## 5. Backlog cleanup

Goal: leave only what is still pending.

For each item in each backlog, decide one of:

- **Resolved** — evidence found. Record the evidence: a commit hash, a file or symbol that now exists, a merged PR, a passing check. Move to `done/YYYY-MM.md` using the evidence date.
- **Pending** — keep, with project tag, priority and origin link.
- **Doubtful** — cannot be verified from the repo. Do not delete; list it for the user.
- **Stale** — in `Later` 2–3 months with no movement, or superseded by another item. Delete or demote to `Someday`.

A ticked box or a "done" note in the backlog is a claim, not evidence. Verify against the code and `git log` before moving anything. Deleting is irreversible outside git: work on a branch or leave it uncommitted so the diff can be reviewed.

Present a report per repo before editing:

| Item | Verdict | Evidence |
|---|---|---|

Then apply only the verdicts the user approved.

## 6. Many repos

Do one pilot repo end to end, adjust this procedure from what broke, then repeat. Libraries first if applications depend on their docs; keep each repo's change in its own commit.
