# Stack Versions

| Dependency | Pinned version | Purpose |
|---|---|---|
| tailwindcss | 4.3 | Utility-first CSS, native CSS variable theme system, container queries |
| @radix-ui/react-* | 1.x (per primitive) | Headless accessible primitives (Dialog, Dropdown, Combobox, etc.) |
| shadcn/ui | recipe-based (no version) | Copy-paste component layer over Radix + Tailwind; lives in your repo |
| lucide-react | latest stable | Default icon set — open-source, broad, well-curated |
| class-variance-authority | latest | Variant-based component composition (`cva()`) |
| tailwind-merge | latest | Conflict-resolving className merge utility |
| @axe-core/react | latest | Accessibility violations surfaced in Vitest + Playwright |
| @testing-library/react | 16.3 | Component tests by role/label/text (see [react-architect](../../frameworks/react-architect/STACK.md)) |
| @playwright/test | 1.60 | End-to-end a11y + interaction tests |

## Notes

- **WCAG 2.2 AA is the mandatory floor** — `axe-core` enforces in CI; Lighthouse a11y ≥ 95 also required.
- **Radix primitives + Tailwind + shadcn/ui recipes** — components live in your repo, not as a dependency tree.
- **Design tokens in `@theme` (Tailwind 4)** with OKLCH colors, semantic naming (`--color-primary`, not `--color-blue-500`), dark mode via `prefers-color-scheme`.
- **`prefers-reduced-motion`** respected everywhere; Tailwind `motion-safe:` / `motion-reduce:` variants.
- **One icon set per project** (Lucide default). Never icon fonts.
- **Container queries (`@container`)** for components that adapt to their parent, not the viewport.
- **No MUI/Mantine/Chakra** — they ship more JS and constrain visual design; headless + Tailwind is the better trade in 2026.
- **Inherits the `react-architect` stack** for client patterns.

_Last reviewed: 2026-05-21_
_Skill version at last review: 1.0.0_
