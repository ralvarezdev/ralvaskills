# Stack Versions

| Dependency | Pinned version | Purpose |
|---|---|---|
| react | 19.2 | UI library |
| react-dom | 19.2 | DOM renderer |
| typescript | 6.0 | Static typing — **strict mode mandatory** |
| @tanstack/react-query | 5.100 | Server state (queries, mutations, cache) |
| zustand | 5.0 | Client global state when justified |
| zod | 3.x latest | Runtime validation at API + form boundaries |
| react-hook-form | 7.x latest | Form state for non-trivial forms |
| @vitejs/plugin-react | 6.0 | Vite bundler integration |
| vitest | 4.1 | Unit + component test runner |
| @testing-library/react | 16.3 | Component testing — query by role/label/text |
| @playwright/test | 1.60 | End-to-end browser tests |
| axe-core | 4.11 | Accessibility violations (a11y) — see [ui-ux-architect](../../frontend/ui-ux-architect/STACK.md) |
| @biomejs/biome | 2.4 | Lint + format (replaces ESLint + Prettier) |

## Notes

- **TypeScript is non-negotiable.** `tsconfig.json` enables `strict`, `noUncheckedIndexedAccess`, `noImplicitOverride`, `noFallthroughCasesInSwitch`, `exactOptionalPropertyTypes`.
- **`@tanstack/react-query` for server state** — never store fetched data in `useState` / `zustand`.
- **`zustand` only when Context + local state don't fit.** Most apps don't need it.
- **`@biomejs/biome`** replaces both ESLint and Prettier in 2026 — one tool, one config. Per [repo-tooling-architect](../../tooling/repo-tooling-architect/SKILL.md).
- **No CSS-in-JS with runtime cost.** Tailwind (covered in [ui-ux-architect](../../frontend/ui-ux-architect/STACK.md)) or CSS Modules.
- **Vitest** for component + unit; **Playwright** for browser e2e. No Jest in new code.
- **For Next.js apps**, react + react-dom are pinned by Next; this STACK still applies for client-side patterns.

_Last reviewed: 2026-05-21_
_Skill version at last review: 1.0.0_
