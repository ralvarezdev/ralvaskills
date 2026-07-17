---
name: reveal-js-architect
version: 1.0.1
description: reveal.js 6.0.1 presentation architecture — section/fragment structure, HTML vs Markdown-mode authoring, theming, plugin config, Reveal.initialize() options, vanilla setup, React extension via @revealjs/react 0.2.1. Use when building, structuring, or reviewing a reveal.js deck.
---

# Reveal.js Architect

Targets **reveal.js 6.0.1** (vanilla) with an extension for **@revealjs/react 0.2.1**. Produces actual reveal.js source (HTML/JS deck scaffold or Markdown-mode slides) — never a slide-spec document; for that, see [demo-presentation-architect](../../personal/demo-presentation-architect/SKILL.md) (unrelated output format, not a dependency). Full project skeletons in [RECIPES.md](RECIPES.md); pinned versions in [STACK.md](STACK.md).

## 1. Scope & output

- **Output is real reveal.js source** — `index.html` + `<section>` markup, or Markdown-mode slides, plus any `main.js` config. Never a slide-spec `.md` document.
- **Vanilla is the default target.** Reach for the React extension (§9) only when the deck already lives inside a React app or needs component-driven slide content (data-fetched slides, shared design-system components).
- **This skill does not write slide copy for the user.** It structures decks; content (titles, bullets) comes from the user or from a spec produced by [demo-presentation-architect](../../personal/demo-presentation-architect/SKILL.md).

## 2. Interview-first protocol

**Never scaffold a deck without finishing the interview.** Ask in order; wait for each answer:

1. **Topic + slide count/duration** — what's the deck about, how long is the talk.
2. **Vanilla or React?** — plain HTML/JS project, or embedded in an existing React app via `@revealjs/react`.
3. **Authoring mode** — HTML `<section>` markup, or Markdown-mode (`data-markdown`)? See §4 for the trade-off.
4. **Theme** — a built-in theme (§5), or a custom CSS override.
5. **Plugins needed** — syntax highlighting, speaker notes, math, zoom, search (§6).
6. **Delivery** — static HTML file opened directly, or a bundler-based project (npm + Vite)?

## 3. Deck structure conventions

- **Horizontal `<section>` = new topic; nested vertical `<section>` = a sub-point of the same topic.** Don't nest vertical slides more than one level deep — audiences lose the "back" direction.
- **One idea per slide**, same discipline as [demo-presentation-architect §3](../../personal/demo-presentation-architect/SKILL.md) — split rather than cram.
- **Fragments (`class="fragment"`) reveal progressively within a slide** — use for build-up bullets or step-by-step diagrams, not as a substitute for splitting a slide.
- **Speaker notes** go in `<aside class="notes">` inside each section — visible only in speaker view (`S` key), never on the audience-facing screen.
- **`data-transition` overrides per-slide** when a specific slide needs a different transition than the deck default — use sparingly, inconsistent transitions read as sloppy.
- **Never set `position`, `overflow`, or `display` directly on `<section>`.** reveal.js owns those properties on slide elements to drive its stacking and transition system (e.g. `position: absolute` on non-present slides, toggling `display` via the `.present` class). Custom CSS that overrides them breaks navigation silently — slides go blank on advance with no console error. Style content *inside* the section (a wrapper `<div>`), never the section itself.

## 4. Content authoring mode

- **HTML sections** — full control over markup, required when a slide embeds custom components, tables with rowspan/colspan, or fine-grained fragment ordering.
- **Markdown-mode (`data-markdown`)** — faster to author, easier to diff/review, preferred for text-heavy decks (bullets, code blocks, quotes). Separate slides within one block with `---`, vertical slides with `--`.
- **Don't mix modes within one deck** unless one section genuinely needs HTML control the rest doesn't — pick a default per deck and stay consistent.
- **External Markdown files** (`data-markdown="slides.md"`) keep content in version control separate from the shell; prefer this for long text-heavy decks over inline `<script type="text/template">`.

## 5. Theming & assets

- **Start from a built-in theme** (`black`, `white`, `league`, `sky`, `night`, `serif`, `simple`, `solarized`, `moon`, `beige`, `blood`, `dracula`) — never build from scratch unless brand requirements demand it.
- **Custom overrides go in a separate stylesheet loaded after the theme**, not by editing the theme file in place — keeps upgrades painless.
- **Images and media live alongside the deck**, referenced with relative paths; large video/audio should be lazy-loaded via `data-src` and `data-lazy-load` conventions to keep initial load fast.

## 6. Plugins

Standard plugin set, loaded and registered in `Reveal.initialize({ plugins: [...] })`:

- **`RevealHighlight`** — syntax-highlighted code blocks; pair with `<pre><code data-line-numbers>` for line-focus during a talk.
- **`RevealNotes`** — enables the speaker-notes view (`S` key); always include if `<aside class="notes">` is used anywhere.
- **`RevealMarkdown`** — required whenever any slide uses `data-markdown`.
- **`RevealMath`** (KaTeX flavor preferred over MathJax for load speed) — only when the deck has actual formulas.
- **`RevealZoom`** — Alt/Option-click zoom into any element; low cost, safe default to include.
- **`RevealSearch`** — only for long reference decks the audience will navigate themselves, not for live-talk decks.

Don't register a plugin the deck doesn't use — each one adds a script tag and init cost.

## 7. Config conventions

- **Sane defaults to keep**: `hash: true` (deep-linkable slides), `controls: true`, `progress: true`, `center: true` unless the deck's layouts are top-aligned by design.
- **`transition`** — `slide` is the safe default; reserve `fade`/`zoom`/`convex` for decks where the transition itself carries meaning (rare).
- **`width`/`height`** — leave at defaults (1280×720) unless the deck targets a specific screen ratio; reveal.js auto-scales via its `viewDistance` and CSS scaling, don't hardcode pixel layouts inside slides.
- **`slideNumber`** — enable for reference/training decks the audience revisits; skip for live pitch decks where it's visual clutter.

## 8. Vanilla setup

CDN link for a single-file deck, or an npm + Vite project for anything version-controlled and multi-file. Full project skeleton (folder tree, `package.json`, `main.js` with plugin registration) in [RECIPES § Vanilla project skeleton](RECIPES.md#vanilla-project-skeleton).

## 9. React extension (@revealjs/react)

- **Use when** the deck lives inside an existing React app, or slides need live component state (charts fed by app data, shared design-system components) rather than static markup.
- **`<Reveal>` wraps `<Slide>` children** — each `<Slide>` maps to a `<section>`; fragments become a `<Fragment>` component, not the `class="fragment"` attribute.
- **Reveal's imperative API (`Reveal.initialize`, plugin registration) is wrapped by the library's own hooks/props** — don't reach into the vanilla API directly except for edge cases the wrapper doesn't expose.
- **Don't reach for this extension by default** — see [react-architect](../react-architect/SKILL.md) for general React conventions; this section only covers the reveal.js-specific integration surface. Full component skeleton in [RECIPES § React project skeleton](RECIPES.md#react-project-skeleton).

## 10. Anti-patterns & review checklist

- Wall-of-text slides (no fragment/split discipline) — same failure mode as [demo-presentation-architect](../../personal/demo-presentation-architect/SKILL.md) flags for spec decks.
- Mixing HTML and Markdown-mode within one deck without a reason.
- Registering plugins that aren't used anywhere in the deck.
- Speaker notes that duplicate on-slide text instead of adding presenter-only context.
- Hardcoded pixel layouts fighting reveal.js's auto-scaling instead of using its centering/sizing config.
- Custom CSS setting `position`, `overflow`, or `display` on `<section>` — silently breaks the stacking/transition system (blank slides on advance, no console error).
- React extension reached for out of habit when a static vanilla deck would do.

Before handing off: every slide has a clear single idea, plugins list matches what's actually used, transitions are consistent unless a specific slide has a stated reason to diverge, and speaker notes (if any) are present only where they add value.
