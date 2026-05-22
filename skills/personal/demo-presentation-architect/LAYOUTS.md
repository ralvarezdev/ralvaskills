# Demo Presentation Layouts

Generic slide-deck layout catalog. Derived from the bundled reference deck at `references/example-deck.html` (or whatever HTML the user supplies as a visual exemplar). Layouts here are language- and topic-agnostic — they describe *shape and field set*, not styling.

## How to read this catalog

- §1 defines **primitive blocks** that compose into layouts.
- §2 enumerates the **layout catalog**. The skill must pick a layout from this list per slide; new layouts are added here before being used.
- §3 is the **decision flow** for picking a layout from intent.
- §4 spells out the **design conventions** that span the whole deck (eyebrow style, accent rule, footer pattern). These apply to every layout unless explicitly overridden.

When a slide does not fit any layout cleanly, propose a new entry to §2 and wait for the user to approve before using it.

## 1. Primitive blocks

These compose into layouts. Most layouts use 3–6 of them.

### 1.1 `eyebrow`
Small context label above the title. Usually 2–5 words, sentence-case or uppercase per the deck's convention. Names the section the slide belongs to, *not* the slide itself.

Example: `eyebrow: "The team"` on a slide titled "A compact, full-stack team focused on AI for operations."

### 1.2 `title-with-accent`
The slide headline. ≤10 words. Exactly **one** accent phrase per title, marked by `<em>...</em>`. The accent carries the takeaway word — what the audience should remember if they only read the title.

Example: `title: "Engineering in PDF, <em>queryable in plain English.</em>"`

### 1.3 `subtitle`
A single sentence beneath the title. Used on cover and on dense slides where the title alone is too compressed. ≤25 words.

### 1.4 `slide-footer`
Three-part footer at the bottom of every non-cover slide: `section_label · brand_mark · page`. The `page` uses the format `NN / TT` where NN is the current slide number (excluding cover) and TT is the total non-cover slide count.

Example: `Documentation · PlantTalk · 05 / 07`.

### 1.5 `card`
Repeating unit inside grids. Fields:
- `tag` — small label at top (e.g. "Flagship product", "Mode 01 · Offline")
- `icon` or `media` — small visual identifier
- `heading` — card name (≤5 words)
- `body` — 1–2 short sentences
- `pill_tags` — optional inline keyword tags (≤4)
- `featured` — boolean; marks one card in a grid as visually prominent

### 1.6 `callout-band`
Full-width emphasized statement at slide bottom. Icon + 1 sentence with one bolded phrase. Used to deliver the slide's takeaway when the body is dense.

Example: `[icon] **Zero new infrastructure.** PlantTalk reads the APIs of what you already have running.`

### 1.7 `pill-tags`
Inline keyword tags (≤4 per cluster). Use to summarize scope, supported items, or qualifiers. Each pill is 1–3 words.

### 1.8 `mini-visual`
Embedded mini chart, mock, or diagram inside a card. Caption mandatory. Used in `mode-comparison` and `showcase-split` layouts. Specify chart type (line / bar / sparkline / mock) and the data being shown.

### 1.9 `numbered-step`
A card variant with a large number (`01`, `02`, …) as the dominant visual. Used in `recap-numbered`.

### 1.10 `person-card`
Card variant for `profile-grid`. Fields:
- `avatar` — image OR initials fallback (when no photo)
- `role_tag` — e.g. "Demo lead", "Speaker", "Backend"
- `name` — full name
- `role` — formal title (1 line)
- `bio` — 1–2 sentences
- `featured` — marks the most relevant person for this audience (e.g. the demo lead)

### 1.11 `handoff-band`
Horizontal strip used at slide bottom on the final or transition slide: avatar + label (`Handing over to`) + name + role + arrow text.

## 2. Layout catalog

Each layout lists: `purpose`, `when to use`, `when NOT to use`, `required fields`, `optional fields`, `structural shape`, and an `anti-pattern` to avoid.

### 2.1 `cover`
**Purpose:** open the deck. The first impression.
**When to use:** slide 1 only.
**When NOT to use:** anywhere else — this layout has no page number and a unique meta footer.
**Required fields:** `brand_mark`, `hero_title` (with accent), `subtitle`, `meta_line` (event + speaker/date)
**Optional fields:** `tagline`, `background_media`
**Shape:** centered brand mark above hero title, subtitle paragraph below, meta footer at the bottom corners.
**Anti-pattern:** including a generic page number footer. Cover slides do not participate in the numbering sequence.

### 2.2 `section-divider`
**Purpose:** mark the boundary between major sections of the deck.
**When to use:** decks with ≥3 distinct sections.
**When NOT to use:** short decks (≤8 slides) — section dividers waste slide budget.
**Required fields:** `section_number`, `section_title`
**Optional fields:** `subtitle`, `roman_numeral`
**Shape:** large centered section title, optional subtitle, no body.
**Anti-pattern:** putting body content here. If it has content, it's a regular content slide, not a divider.

### 2.3 `title-bullets`
**Purpose:** baseline content slide.
**When to use:** when the slide's content reduces cleanly to 3–6 parallel points.
**When NOT to use:** when bullets would read like a wall of text — split into multiple slides or pick `card-grid`.
**Required fields:** `eyebrow`, `title-with-accent`, `bullets` (3–6, parallel structure)
**Optional fields:** `lede` (1-sentence lede before bullets), `callout-band`
**Shape:** eyebrow → title → optional lede → bullet list → optional callout.
**Anti-pattern:** ≥7 bullets, or bullets longer than 12 words each.

### 2.4 `title-paragraph`
**Purpose:** narrative content slide — no list.
**When to use:** when the idea is a single argument, not enumerable.
**When NOT to use:** when content has natural parallel structure (use `title-bullets`).
**Required fields:** `eyebrow`, `title-with-accent`, `body` (1–2 short paragraphs)
**Optional fields:** `callout-band`, `pull-quote-inline`
**Anti-pattern:** prose longer than ~80 words. Trim or split.

### 2.5 `card-grid`
**Purpose:** present N parallel items with the same shape.
**When to use:** 3–6 items that share the same fields (icon, heading, body).
**When NOT to use:** items that aren't visually parallel; one item that's far more important (use `feature-comparison`).
**Required fields:** `eyebrow`, `title-with-accent`, `cards[]` (3–6 of `card` primitive)
**Optional fields:** `lede`, `callout-band` below the grid
**Shape:** title at top, uniform card grid below. Card count drives grid layout: 3 → 1 row, 4 → 2×2 or 1×4, 6 → 2×3.
**Anti-pattern:** mixing two card sizes in the same grid. Use `feature-comparison` or `card-grid + small-row` instead.

### 2.6 `profile-grid`
**Purpose:** introduce a team, panel, or set of stakeholders.
**When to use:** 3–6 people. Each gets a `person-card`.
**When NOT to use:** ≥7 people (too dense — split or summarize); or for showing a single person (use a sidebar instead).
**Required fields:** `eyebrow`, `title-with-accent`, `people[]` (3–6 `person-card`s)
**Optional fields:** mark **one** card as `featured` to draw the eye (e.g. the demo lead).
**Anti-pattern:** more than one `featured` person. Defeats the purpose.

### 2.7 `feature-comparison`
**Purpose:** position exactly two offerings side by side. Flagship vs. companion; product vs. service; current vs. proposed.
**When to use:** when the slide's job is to contrast or anchor a primary + secondary.
**When NOT to use:** ≥3 items (use `card-grid`); items not comparable along the same axes.
**Required fields:** `eyebrow`, `title-with-accent`, `cards[]` of length 2 (one with `featured: true`)
**Optional fields:** `common_row` — a thin row of 3–5 shared traits below the two cards
**Shape:** two large cards spanning the slide width; optional thin commonalities row underneath.
**Anti-pattern:** comparing along inconsistent axes — both cards must list the same kinds of fields in the same order.

### 2.8 `diagram-hub-spoke`
**Purpose:** show that one thing integrates with several others. The hub-and-spoke mental model.
**When to use:** when the system has a clear central element (assistant, platform, broker) with N peripheral sources/consumers.
**When NOT to use:** linear pipelines (use a `pipeline` layout instead — add to §2 if needed); fully meshed systems.
**Required fields:** `eyebrow`, `title-with-accent`, `hub` (one card with `mark`, `tag`, `heading`, `body`), `nodes[]` (3–6 cards with `position`, `number`, `icon`, `heading`, `body`)
**Optional fields:** animated connectors (visual hint only — the spec just notes "animated: true")
**Shape:** central hub, peripheral nodes around it, connectors implied by position.
**Anti-pattern:** more than 6 nodes — the diagram becomes a hairball.

### 2.9 `showcase-split`
**Purpose:** anchor an abstract idea with a concrete visual.
**When to use:** when there is a real artifact to show (document mock, screenshot, chart) and a parallel bullet list interpreting it.
**When NOT to use:** when no concrete visual exists; or when the visual is decorative (don't fake it).
**Required fields:** `eyebrow`, `title-with-accent`, `visual` (`media` + `caption` + `citation`), `lede`, `bullets`
**Optional fields:** `citation_target` — what section/page/source the visual cites
**Shape:** two columns. Left ≈55%: the visual block. Right ≈45%: lede paragraph + bullet list.
**Anti-pattern:** visual on the right and text on the left — Western reading flow expects the artifact-anchor on the left.

### 2.10 `mode-comparison`
**Purpose:** show the same concept in two modes (offline/online, before/after, mode A/B).
**When to use:** exactly two modes, each with its own mini-visual.
**When NOT to use:** ≥3 modes (use `card-grid` with `mini-visual` per card).
**Required fields:** `eyebrow`, `title-with-accent`, `intro` (left column: `lede` + `bullets`), `modes[]` of length 2 (each: `tag`, `heading`, `body`, `mini-visual`)
**Optional fields:** `callout-band` at the bottom of the intro column
**Shape:** two columns. Left ≈45%: lede + bullets + optional callout. Right ≈55%: two stacked mode cards.
**Anti-pattern:** modes that share no comparison axis — they should be two configurations of the same thing.

### 2.11 `recap-numbered`
**Purpose:** summarize what was covered or what will be covered next.
**When to use:** closing recap, agenda preview, or end-of-section roundup.
**When NOT to use:** when you've already shown the same recap on a prior slide.
**Required fields:** `eyebrow`, `title-with-accent`, `items[]` (3–6 `numbered-step` cards)
**Optional fields:** `handoff-band` at slide bottom (when this is a handoff/closing slide)
**Anti-pattern:** items that don't map 1-to-1 with the actual content the audience saw or will see.

### 2.12 `closing-cta`
**Purpose:** the deck's final ask or next step.
**When to use:** the very last slide.
**When NOT to use:** earlier — calls to action mid-deck dilute the closing impact.
**Required fields:** `title-with-accent`, `cta_text` (one sentence)
**Optional fields:** `contact_block` (name + role + email/url), `qr_code`, `handoff-band`
**Anti-pattern:** multiple CTAs. Pick one.

### 2.13 `quote`
**Purpose:** anchor a slide on a pull-quote — customer, expert, internal voice.
**When to use:** when the quote *is* the slide. Sparingly — once or twice per deck max.
**When NOT to use:** as a decorative aside; quotes deserve their own slide or none.
**Required fields:** `eyebrow`, `quote_text` (≤25 words), `attribution` (name + role)
**Optional fields:** `context` (1 sentence above or below the quote)
**Anti-pattern:** an unattributed or paraphrased quote. If you can't cite the source verbatim, don't use this layout.

### 2.14 `stat-row`
**Purpose:** lead with a few headline numbers.
**When to use:** when the slide's takeaway is quantitative — funding, scale, performance, growth.
**When NOT to use:** when numbers need context (use `title-bullets` with inline numbers instead).
**Required fields:** `eyebrow`, `title-with-accent`, `stats[]` (3–5 entries, each: `number`, `label`, optional `context`)
**Optional fields:** `source_line` at the bottom citing where the numbers come from
**Anti-pattern:** numbers without a source line. Every statistic must be traceable.

## 3. Picking a layout — decision flow

Walk the questions in order; stop at the first match.

1. Is this the first slide? → `cover`
2. Is this the last slide? → `closing-cta` (or `recap-numbered` + `handoff-band` if there is a live demo or speaker change next)
3. Are we entering a new major section? → `section-divider`
4. Is the slide built around a single quote? → `quote`
5. Is the takeaway a set of numbers? → `stat-row`
6. Is there exactly one concrete visual to anchor on? → `showcase-split`
7. Is the slide showing two modes/states of the same thing? → `mode-comparison`
8. Is the slide showing one central thing with N peripheral things? → `diagram-hub-spoke`
9. Is the slide comparing exactly two offerings? → `feature-comparison`
10. Is the slide introducing a team/panel? → `profile-grid`
11. Is the slide presenting N parallel items with a shared shape? → `card-grid`
12. Does the content reduce to 3–6 parallel bullets? → `title-bullets`
13. Else → `title-paragraph`

## 4. Design conventions (deck-wide)

These apply to every layout unless the user explicitly overrides. The reference HTML demonstrates them, but they're meant as defaults that any deck can adopt.

- **Accent rule.** Every content slide title has exactly one `<em>` accent phrase. The accent is the takeaway word — what stays in the audience's memory.
- **Eyebrow rule.** Every non-cover, non-divider slide has an `eyebrow`. Short, names the section.
- **Footer rule.** Every non-cover slide has the 3-part `slide-footer`. Page numbering excludes the cover.
- **Single-takeaway rule.** A slide has one job. If the spec needs two unrelated titles or two unrelated bullets blocks, split into two slides.
- **Card count caps.** `card-grid` ≤6, `profile-grid` ≤6, `recap-numbered` ≤6, `diagram-hub-spoke` ≤6 nodes, `stat-row` ≤5. Beyond the cap, split or summarize.
- **Pill cap.** ≤4 pills per cluster.
- **Title length.** ≤10 words.
- **Bullet length.** ≤12 words per bullet, parallel structure across bullets.
- **Language consistency.** Once §2 of `SKILL.md` confirms a deck language, every text field across every slide uses it. No mixed-language titles or footers.
- **Image fields.** Every image field includes `alt:` text describing what's shown — never decorative-only.

## 5. Adding a new layout

If a slide's intent doesn't fit any layout in §2:

1. Stop. Don't invent a layout name in the slide spec.
2. Propose a new entry to §2 with: id, purpose, when-to-use, when-NOT-to-use, required/optional fields, shape, anti-pattern.
3. Wait for the user to approve.
4. Add the layout to §2 *and* to the decision flow in §3.
5. Then write the slide using the new layout.

This keeps the catalog tight and prevents the deck spec from drifting into ad-hoc shapes.
