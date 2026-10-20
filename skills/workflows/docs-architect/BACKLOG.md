# Backlog

```
docs/backlog/
├── inbox.md          # raw capture, one line per thought, no format
├── <project>.md      # Doing / Next / Later / Someday
└── done/
    └── 2026-10.md    # closed items, one file per month
```

## Item format

```md
- [ ] Verify TON timers in the PLC agent  #plc-agent  P1
  plan: [plan](../plans/2026-10-12-plc-verification.md)
```

Every item has a project tag, a priority (P1/P2/P3) and, once it leaves the inbox, a link to the doc it came from (plan, design, review or change). An item without an origin link is allowed only in `Someday`.

## Sections per project file

| Section | Limit | Meaning |
|---|---|---|
| Doing | 1–2 | In progress now |
| Next | 3–5 | Ready to pick up; linked to an approved design |
| Later | none | Wanted, not scheduled |
| Someday | none | Maybe never |

The limits are what make the list a queue and not a pile. The linter warns when they are exceeded.

## Triage (weekly)

Take each inbox line and do exactly one of: delete it; append it to the project file with tag and priority; or open an `ideas/` doc when it needs thought. Leave the inbox empty.

## Pruning (monthly)

- Items in `Later` for 2–3 months without moving: delete, or demote to `Someday`. Do not archive them.
- Closed items move to `done/YYYY-MM.md` with the close date, keeping the origin link.

## Closing an item

1. Move the line to `done/YYYY-MM.md`, tick the box, add the date.
2. Set the originating doc to `status: done` once all of its items are closed.
3. Move that doc to `archive/` at the next monthly pass.

## Feedback loop

A review or a change that surfaces follow-up work creates new backlog items linking back to it. Never leave findings only inside the review.

## GitHub Issues

If the project tracks work in Issues, the backlog lives there. Keep `docs/` for the docs only and link the issue number from `related` or the item.
