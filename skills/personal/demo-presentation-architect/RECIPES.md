# Demo Presentation — Templates & Reference

Slide spec template, ordering tables, anti-patterns, and the closing checklist for [SKILL.md](SKILL.md). Layout catalog is separate in [LAYOUTS.md](LAYOUTS.md). Loaded on demand.

## 1. Per-slide specification format

Every slide in the output `.md` follows the same block shape. Field names inside `### Content` come from the chosen layout's required/optional fields in [LAYOUTS § 2](LAYOUTS.md).

```markdown
## Slide N — <headline in deck language>

**Layout:** <layout-id from LAYOUTS §2>
**Purpose:** <one line — why this slide exists in the flow>

### Content
<exact text content, structured per the chosen layout. Use the field names
defined for that layout in LAYOUTS.md (eyebrow, title, hub, nodes[], cards[],
visual.caption, etc.). Write final copy — never placeholders.>

### Visual hierarchy
1. <element the eye should land on first>
2. <element second>
3. <element third (optional)>

### Speaker notes (optional)
<one or two sentences — reminders, not a script>

### Transition
<one sentence connecting this slide's idea to the next>
```

The `Transition` line is what makes the deck feel like a story rather than a pile of slides. Mandatory on every slide except the final one.

## 2. Narrative arcs

Pick the arc that matches the goal — it drives the order of body slides:

| Arc | Sequence | When to pick |
|---|---|---|
| **Problem → Solution** | context → pain → mechanism → evidence → ask | Sales, internal proposals, pitches |
| **System tour** | overview → component A → component B → … → recap | Architecture walkthroughs, demos |
| **Before / After** | baseline → change → result → next | Migrations, refactors, redesigns |
| **Discovery** | question → method → finding → implication | Research presentations, post-mortems |
| **Tutorial** | goal → step 1 → step 2 → … → outcome | How-tos, onboarding |

Mixing arcs mid-deck breaks the audience's mental model — pick one and stay with it.

## 3. Title takeaway phrasing — examples

| Topic phrasing (weak) | Takeaway phrasing (strong) |
|---|---|
| "Onboarding times" | "Cuts onboarding from <em>3 days to 4 hours.</em>" |
| "Architecture" | "One assistant. <em>Four kinds of sources</em> underneath." |
| "Our team" | "A compact team focused on <em>AI for operations.</em>" |
| "Migration results" | "<em>Zero downtime</em> on the cutover." |

The `<em>` accent always carries the takeaway word.

## 4. Ordering items within a layout

Different layouts have different ordering principles. Apply the one that matches:

| Layout family | Order by |
|---|---|
| `title-bullets` | Importance (most important first) — unless bullets describe a sequence, then chronological |
| `card-grid` / `profile-grid` | Salience left-to-right, top-to-bottom; the `featured` card anchored visually (center or top) |
| `feature-comparison` | Primary/flagship on the left, secondary on the right |
| `diagram-hub-spoke` | Nodes numbered clockwise from top-left |
| `recap-numbered` | Same order as the slides being recapped |
| `stat-row` | Largest-impact stat first (or chronological if it's a timeline) |
| `mode-comparison` | "Before" / "offline" on the left; "after" / "online" on the right |
| `showcase-split` | Visual on the left, text on the right (Western reading flow) |

## 5. Body word budget per layout

| Layout | Body word budget |
|---|---|
| `cover`, `section-divider`, `quote` | ≤40 words total |
| `title-paragraph`, `closing-cta` | ≤80 words total |
| `title-bullets`, `card-grid`, `profile-grid` | ≤120 words across all bullets / cards |
| `diagram-hub-spoke`, `showcase-split`, `mode-comparison` | ≤140 words across all node / card bodies |
| `recap-numbered`, `stat-row` | ≤80 words (these are summaries — brevity is the point) |

Over budget → move detail to speaker notes or split the slide.

## 6. Anti-patterns

| Anti-pattern | Wrong | Right |
|---|---|---|
| Placeholder text | "Bullet 1, bullet 2, bullet 3" | Final copy: "Cuts onboarding time from 3 days to 4 hours" |
| Multiple ideas per slide | One slide covering problem + solution + metrics | Three slides, one per idea, with transitions |
| Invented layout | `Layout: cool-split-thing` | Pick from [LAYOUTS § 2](LAYOUTS.md) or propose an addition first |
| No title accent | Plain title with no `<em>` | Exactly one `<em>` phrase per content-slide title |
| Wall of text | 8 dense bullets, 30 words each | ≤5 bullets, ≤12 words each; move detail to speaker notes |
| Mixed languages | Title in English, body in Spanish | Whole deck in the language confirmed in the interview |
| No transition | Slides feel disconnected | Every non-final slide has a `Transition` line |
| Decorative-only image | Stock photo with no caption / alt | Image earns its place; include `alt:` and a caption tying it to the idea |
| Two `featured` cards | Two profile cards or two offer cards marked `featured` | At most one `featured` per grid — defeats the purpose otherwise |

## 7. Review checklist

Before handing the `.md` to the user, verify:

- [ ] Language confirmed in the interview is used consistently across every slide (titles, eyebrows, bullets, footers, alt text)
- [ ] Every content slide title has exactly one `<em>` accent phrase ([LAYOUTS § 4](LAYOUTS.md))
- [ ] Every slide cites a `Layout:` from [LAYOUTS § 2](LAYOUTS.md) — no invented names
- [ ] Every non-cover slide has an `eyebrow` and a `slide-footer` (`section_label · brand_mark · page`)
- [ ] Page numbering excludes the cover (`01 / TT`, …)
- [ ] Every slide has explicit `Visual hierarchy` ordering
- [ ] Every non-final slide has a `Transition` line
- [ ] Slide count matches the duration / budget agreed in the interview
- [ ] Required slides from the user's constraints (title, agenda, Q&A, contact) are present
- [ ] Card-count caps respected: `card-grid` / `profile-grid` / `recap-numbered` / hub-spoke nodes ≤6; `stat-row` ≤5; `pill_tags` ≤4 per cluster
- [ ] At most one `featured` card per grid
- [ ] No placeholder text, no lorem ipsum, no `[TODO]`
- [ ] Image fields have `alt:` text and a caption
- [ ] Every title carries the **takeaway**, not just the topic label
- [ ] Body word budget per layout respected (§5)
- [ ] Narrative arc from §2 is consistent across the deck (no mid-deck arc switch)
- [ ] No concept appears on more than two slides

## 8. Output deliverables

1. **Deck outline** — one line per slide: `N. <layout-id> — <headline>`. Produced after the interview, approved before expansion.
2. **Full slide spec `.md`** — every slide in the per-slide format from §1, using field names from [LAYOUTS.md](LAYOUTS.md).
3. **Layout usage audit** — count of each layout used; flag overuse of any single layout (>40% of slides) or unused-but-expected layouts (e.g. no `cover`, no `closing-cta`).
4. **Transition map** — list of every `Transition` line in order, so the user can read the deck as a narrative.
5. **Open questions** — anything the interview didn't resolve and that the user must answer before the deck is final.
