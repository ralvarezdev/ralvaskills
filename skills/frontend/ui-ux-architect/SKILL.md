---
name: ui-ux-architect
version: 1.0.0
description: UI/UX standards — WCAG 2.2 AA, Radix + Tailwind 4 + shadcn/ui, design-token theming, mandatory loading/error/empty/success states on every async surface, mobile-first responsive, keyboard parity, contrast checked in CI. Use when designing UI components, building a design system, or auditing accessibility.
---

# UI / UX Architecture

Implementation-level standards for accessible, consistent, responsive frontends. Implements the *what* and *how*; framework-specific *where* lives in [react-architect](../../frameworks/react-architect/SKILL.md) and [nextjs-architect](../../frameworks/nextjs-architect/SKILL.md). See [STACK.md](STACK.md) for pinned dependencies.

## 1. Accessibility floor — WCAG 2.2 AA

Non-negotiable. Every shipped component meets WCAG 2.2 Level AA; AAA where reasonable.

- **Semantic HTML first.** A `<button>` is a button; a `<div onClick>` is a bug. `<nav>`, `<main>`, `<article>`, `<aside>`, `<header>`, `<footer>` mark out landmarks.
- **Every interactive element is keyboard-operable.** `Tab` to focus, `Enter`/`Space` to activate, `Escape` to dismiss, arrow keys inside composite widgets. No mouse-only behavior.
- **Visible focus indicator on every focusable element.** Don't `outline: none` without replacing it with something equally clear. Tailwind: `focus-visible:ring-2 focus-visible:ring-offset-2`.
- **Color contrast** ≥ 4.5:1 for normal text, ≥ 3:1 for large text and UI components. Enforced by `axe-core` in CI; rejected at build.
- **Form labels.** Every `<input>` has a `<label htmlFor>`; placeholder text is *not* a label.
- **ARIA only when semantic HTML can't say it.** `aria-label` for icon-only buttons, `aria-describedby` for inline help text, `role="alert"` for live announcements. Don't sprinkle `role="button"` — write `<button>`.
- **`prefers-reduced-motion`** respected — Tailwind: `motion-safe:` / `motion-reduce:` variants. Disable parallax, autoplay, large transitions for users who request it.

## 2. Component primitives — Radix + Tailwind + shadcn/ui

Don't reinvent dialogs, dropdowns, popovers, comboboxes. Modern accessibility is hard; let primitives do the keyboard/ARIA work.

- **[Radix UI](https://www.radix-ui.com/)** — headless, unstyled, accessible primitives. Focus management, keyboard nav, ARIA wiring all done.
- **Tailwind CSS 4** — utility-first styling on top of Radix. Variables-based theme system, design tokens via CSS custom properties.
- **shadcn/ui** — copy-paste component recipes (built on Radix + Tailwind). **Components live in your repo, not as a dep**; you own and adapt them.

```tsx
// components/ui/dialog.tsx — shadcn/ui-style Radix wrapper
"use client";
import * as DialogPrimitive from "@radix-ui/react-dialog";
import { cn } from "@/lib/utils";

export const Dialog = DialogPrimitive.Root;
export const DialogTrigger = DialogPrimitive.Trigger;

export function DialogContent({ className, children, ...props }: DialogPrimitive.DialogContentProps) {
  return (
    <DialogPrimitive.Portal>
      <DialogPrimitive.Overlay className="fixed inset-0 bg-black/50 data-[state=open]:animate-in" />
      <DialogPrimitive.Content
        className={cn(
          "fixed left-[50%] top-[50%] z-50 grid w-full max-w-lg translate-x-[-50%] translate-y-[-50%]",
          "gap-4 border bg-background p-6 shadow-lg",
          "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
          className,
        )}
        {...props}
      >
        {children}
      </DialogPrimitive.Content>
    </DialogPrimitive.Portal>
  );
}
```

- **Why not MUI / Mantine / Chakra?** They're great but opinionated about visual design; harder to evolve a custom design system; ship more JS. Radix headless + Tailwind gives full control with smaller bundles.
- **Never style by overriding library internals.** If Radix doesn't expose the part you need to style, file an issue / find another primitive.

## 3. Design tokens

Centralize every visual decision as a token. Components reference tokens; tokens evolve in one place.

```css
/* app/globals.css — Tailwind 4 with native CSS variables */
@theme {
  /* Colors — design system tokens, semantic naming */
  --color-background: oklch(1 0 0);
  --color-foreground: oklch(0.15 0 0);
  --color-primary: oklch(0.55 0.18 250);
  --color-primary-foreground: oklch(1 0 0);
  --color-muted: oklch(0.96 0 0);
  --color-muted-foreground: oklch(0.45 0 0);
  --color-destructive: oklch(0.6 0.21 25);
  --color-border: oklch(0.92 0 0);
  --color-ring: var(--color-primary);

  /* Radii */
  --radius-sm: 0.25rem;
  --radius-md: 0.375rem;
  --radius-lg: 0.5rem;

  /* Spacing scale — 4px base */
  --spacing-1: 0.25rem;
  --spacing-2: 0.5rem;
  --spacing-4: 1rem;
  --spacing-8: 2rem;
}

@media (prefers-color-scheme: dark) {
  @theme {
    --color-background: oklch(0.13 0 0);
    --color-foreground: oklch(0.97 0 0);
    /* ... dark mode overrides ... */
  }
}
```

- **Semantic over literal names.** `--color-primary` not `--color-blue-500`. Lets you change the palette without rewriting every component.
- **OKLCH for color.** Perceptually uniform, future-proof, supports wider gamuts. `oklch(L C H)` — lightness, chroma, hue.
- **Dark mode via CSS variables** — flip a few tokens, every component follows.
- **Spacing scale based on 4px** (or 8px). Don't introduce one-off values; reuse `--spacing-N`.
- **Type scale** centralized in tokens too (`--text-sm`, `--text-base`, etc.).

## 4. Responsive — mobile-first, container queries when shape matters

- **Mobile-first.** Default styles are mobile; `md:`, `lg:` add desktop overrides. Per Tailwind convention.
- **Container queries** (`@container`) when a component's layout depends on its container size, not the viewport. New in Tailwind 4 — use them for components that appear in narrow sidebars *and* wide pages.
- **Breakpoints** sparingly: `sm`, `md`, `lg`, `xl`, `2xl`. Don't invent new ones per project.
- **Touch targets** ≥ 44×44 px on mobile. WCAG 2.2 AA mandates 24×24 *minimum* but 44×44 is the modern usable size.
- **Fluid typography** — `clamp(1rem, 0.9rem + 0.5vw, 1.25rem)` scales smoothly without breakpoint jumps.
- **Test on real devices.** Browser DevTools mobile mode is a poor proxy for actual phones. At minimum: one iPhone, one Android.

## 5. The four required states — loading, error, empty, success

Every UI surface that touches data has four states. Build all four; don't pretend the only-success path is enough.

### Loading

- **Skeletons preferred over spinners** for content areas — they hint at the upcoming structure.
- **Spinners for "this button is doing something"** — short, focused.
- **`aria-busy="true"`** on the loading region so screen readers announce.
- **Suspense + skeleton** per [react-architect §6](../../frameworks/react-architect/SKILL.md#6-suspense--errorboundary).
- **No flashes of loading** for fetches < 300 ms (`setTimeout` the spinner display). User perceives the operation as instant.

### Error

- **State the problem in plain language.** "We couldn't load your orders." Not "Error 503".
- **Retry button** — most errors are transient.
- **Correlation ID visible** for support, per [observability-architect §5](../../infra/observability-architect/SKILL.md#5-correlation). "If this keeps happening, reference ID `req_01J9...`".
- **No stack traces in the UI.** Per [security-reviewer](../../quality/security-reviewer/SKILL.md) — log them server-side, show the ID.
- **Different shape per error class.** Network error gets retry; auth error redirects to login; validation error highlights the field.

### Empty

- **First-time empty differs from filtered empty.**
  - First-time: "You haven't created any orders yet. [Create your first.]"
  - Filtered: "No orders match the current filters. [Clear filters.]"
- **Action always present.** Empty without a way forward is a dead end.
- **Illustration optional.** A subtle SVG can soften an empty screen; an over-designed one screams.

### Success / data

- **Tabular data** uses `<table>` with `<th scope="col">`, `<caption>` for context.
- **Pagination indicator** (per [rest-api-architect §5](../../protocols/rest-api-architect/SKILL.md#5-pagination--cursor-not-offset) cursor pagination) — "Load more" button or infinite scroll with intersection observer.
- **Refresh / live update** signals when data is stale or refetching — small spinner near the title, not a full overlay.

## 6. Forms

- **Native HTML semantics.** `<form>`, `<label>`, `<input>` with the right types (`type="email"`, `type="tel"`, `type="number"`).
- **`autocomplete`** attributes on every relevant field. Browsers and password managers depend on them.
- **Inline validation** — show errors as the user finishes a field (`onBlur`), not on every keystroke (annoying). Submit-time validation is the final gate.
- **Error placement: below the input it concerns**, connected via `aria-describedby`. Screen readers announce automatically.
- **Submit button disabled while pending** — visual + `aria-disabled="true"` (not `disabled` attribute, which removes it from the tab order).
- **Inline success feedback** — "Saved" with a checkmark for 2s, then fade. Don't toast critical confirmations away.

## 7. Motion

- **`prefers-reduced-motion: reduce`** respected. Tailwind `motion-safe:transition-*`.
- **Functional motion only.**
  - Confirmation that an action happened (`scale-95` press feedback).
  - Spatial relationships (drawer slides in from the side).
  - State changes that wouldn't be perceptible otherwise (a modal appearing without animation is jarring).
- **No autoplaying video / animation** without user opt-in. Background videos break `prefers-reduced-motion`.
- **Easing**: `ease-out` for entrances (feels responsive), `ease-in` for exits, `ease-in-out` for transitions in place.
- **Duration** 100–300 ms for most UI animation; 0 ms when `prefers-reduced-motion: reduce`.

## 8. Icons

- **One icon set per project.** [Lucide](https://lucide.dev) (open-source, broad, well-curated) is the modern default.
- **`aria-hidden="true"` on decorative icons.** They're not announced.
- **`aria-label` on icon-only buttons.** Screen readers need *something* to read.
- **SVG sprites or per-icon components**, never icon fonts. Fonts cause FOIT and accessibility issues.
- **Don't mix icon sets.** A Lucide pen next to a Material Design save button looks unstable.

## 9. Internationalization (when in scope)

- **`<html lang="...">`** always set, even for English-only apps. Screen readers depend on it.
- **All copy through a translation function** if i18n is on the roadmap — `t("user.greeting")` even before there's a second language.
- **Directionality:** support RTL via `dir="rtl"` and CSS logical properties (`margin-inline-start` not `margin-left`). Tailwind 4's logical-property utilities cover this.
- **Date / number / currency formatting** via `Intl.DateTimeFormat`, `Intl.NumberFormat` — locale-aware, no string concatenation.

## 10. Testing accessibility

- **`axe-core`** integrated into Vitest tests + Playwright e2e. Zero violations of WCAG 2.2 AA before merge.
- **Manual keyboard pass** on every new component — tab through it, activate every control, no mouse.
- **Screen reader pass** for critical flows — NVDA or VoiceOver, walk through sign-up / checkout / settings change.
- **Color contrast checker** for every design token decision — `--color-primary` on `--color-background` ≥ 4.5:1.
- **Lighthouse a11y score** ≥ 95 in CI. Lighthouse misses things axe catches, and vice versa — run both.

## 11. Design system maturity ladder

Where a project sits on this ladder informs how much investment is justified:

| Level | What | Tooling |
|---|---|---|
| 0 — Ad hoc | Tailwind utilities everywhere, no shared components | Tailwind alone |
| 1 — Primitives | Shared `Button`, `Input`, `Card`, `Dialog` in `components/ui/` (shadcn-style) | + Radix + shadcn recipes |
| 2 — Tokens | Design tokens centralized, dark mode, semantic colors | + design tokens in `@theme` |
| 3 — Patterns | Composed `<UserCard>`, `<OrderList>` published as feature components | (no new tooling — convention) |
| 4 — Documented system | Storybook / docs published, visual regression tests | + Storybook + Chromatic |

- **Start at Level 1 for any non-throwaway project.** The cost of refactoring to it later is far higher than the cost of starting there.
- **Climb to Level 2 when more than one developer is touching styles.** Tokens prevent visual drift.
- **Level 4 earns its place when you have a real design system team or many product teams consuming the same components.**

## 12. Cross-skill ties

- [react-architect](../../frameworks/react-architect/SKILL.md) — client component patterns, hooks, state. This skill is what they render.
- [nextjs-architect](../../frameworks/nextjs-architect/SKILL.md) — server-component shells, server actions, image / font / metadata APIs.
- [rest-api-architect §7](../../protocols/rest-api-architect/SKILL.md#7-error-contracts--rfc-7807-problem-details) — the error contracts the UI consumes; correlation IDs surfaced in error states.
- [security-reviewer](../../quality/security-reviewer/SKILL.md) — UI surface for security headers (CSP), no stack traces in error UI, no PII in URLs.
- [observability-architect §5](../../infra/observability-architect/SKILL.md#5-correlation) — correlation IDs visible in the UI so support can match logs.
- [logic-cleaner](../../workflows/logic-cleaner/SKILL.md) / [code-design-refactor](../../workflows/code-design-refactor/SKILL.md) — same rules apply to component code as to any other code.
