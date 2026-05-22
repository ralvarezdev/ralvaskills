# UI/UX — Reference Implementations

Component recipes and token templates referenced by [SKILL.md](SKILL.md). Loaded on demand.

## 1. Radix + shadcn dialog wrapper

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

- Components live in your repo, not as a dep — you own and adapt them.
- Never style by overriding library internals. If Radix doesn't expose the part you need, file an issue / find another primitive.

## 2. Tailwind 4 `@theme` design tokens

```css
/* app/globals.css — Tailwind 4 with native CSS variables */
@theme {
  /* Colors — semantic naming, OKLCH for perceptually uniform palettes */
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

- **Semantic over literal names.** `--color-primary` not `--color-blue-500`.
- **OKLCH** — perceptually uniform, future-proof, wider gamut. `oklch(L C H)`.
- **Spacing on a 4 px base.** Don't introduce one-off values.
- **Type scale** centralized too (`--text-sm`, `--text-base`, etc.).

## 3. Required UI states — shape contract

Every async surface needs **loading / error / empty / success**. Concrete shapes:

### Loading
- Skeletons for content areas; spinners for button-level "working".
- `aria-busy="true"` on the loading region.
- Suppress loading flash for fetches < 300 ms.

### Error
- Plain language: "We couldn't load your orders." Not "Error 503".
- Retry button (most errors are transient).
- Correlation ID visible for support: "If this keeps happening, reference `req_01J9…`".
- No stack traces in the UI — log server-side, show the ID.
- Differentiate: network → retry, auth → redirect to login, validation → highlight field.

### Empty
- First-time empty: "You haven't created any orders yet. [Create your first.]"
- Filtered empty: "No orders match the current filters. [Clear filters.]"
- Action always present.
- Illustration optional; subtle.

### Success / data
- Tabular data uses `<table>` with `<th scope="col">`, `<caption>`.
- Pagination matches [rest-api-architect §5](../../protocols/rest-api-architect/SKILL.md#5-pagination--cursor-not-offset) cursors — "Load more" or intersection-observer infinite scroll.
- Refetch indicator near the title; no full-screen overlay.

## 4. Design system maturity ladder

| Level | What | Tooling |
|---|---|---|
| 0 — Ad hoc | Tailwind utilities everywhere, no shared components | Tailwind alone |
| 1 — Primitives | Shared `Button`, `Input`, `Card`, `Dialog` in `components/ui/` (shadcn-style) | + Radix + shadcn recipes |
| 2 — Tokens | Design tokens centralized, dark mode, semantic colors | + design tokens in `@theme` |
| 3 — Patterns | Composed `<UserCard>`, `<OrderList>` published as feature components | (no new tooling — convention) |
| 4 — Documented system | Storybook / docs published, visual regression tests | + Storybook + Chromatic |

- **Start at Level 1** for any non-throwaway project — refactoring up later costs more.
- **Climb to Level 2** when >1 developer is touching styles — tokens prevent visual drift.
- **Level 4** when you have a design-system team or many product teams consuming the same components.

## 5. Accessibility testing setup

```ts
// vitest.setup.ts — axe in every render
import { expect } from "vitest";
import { toHaveNoViolations } from "jest-axe";
expect.extend({ toHaveNoViolations });
```

```ts
// any.test.tsx
import { render } from "@testing-library/react";
import { axe } from "jest-axe";

it("has no a11y violations", async () => {
  const { container } = render(<MyComponent />);
  expect(await axe(container)).toHaveNoViolations();
});
```

Playwright e2e:

```ts
import AxeBuilder from "@axe-core/playwright";
test("home page has no a11y violations", async ({ page }) => {
  await page.goto("/");
  const results = await new AxeBuilder({ page }).analyze();
  expect(results.violations).toEqual([]);
});
```

Lighthouse a11y score ≥ 95 in CI; Lighthouse and axe miss different things, so run both.
