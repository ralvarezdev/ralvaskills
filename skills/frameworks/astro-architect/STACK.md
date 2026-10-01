# Stack Versions

| Dependency | Pinned version | Purpose |
|---|---|---|
| astro | 6.x latest (7.x when stable) | Framework |
| typescript | 5.x latest | Static typing — strict mode mandatory via `astro/tsconfigs/strict` |
| @astrojs/check | 0.9.x | `astro check` — editor + CLI type diagnostics |
| @astrojs/sitemap | 3.x latest | Sitemap generation; requires `site:` in config |
| @astrojs/rss | 4.x latest | RSS feed for content collections |
| @astrojs/mdx | 4.x latest | MDX support — only if content needs live components |
| @astrojs/node | 9.x latest | Node adapter — self-hosted on-demand deployment |
| @astrojs/image | — | **Not needed.** `astro:assets` is built in; the standalone package is deprecated |
| @biomejs/biome | 2.x latest | Lint + format |
| vitest | 3.x latest | Unit tests |
| @playwright/test | 1.x latest | E2E against `astro preview`, not `astro dev` |

## Framework islands

Add **only if an island is unavoidable** — see [SKILL.md §1](SKILL.md#1-the-rendering-model--the-one-rule-that-governs-everything-else).

| Dependency | Pinned version | Purpose |
|---|---|---|
| preact | 10.x latest | Lightest island runtime (~4 KB). Default choice over React for islands |
| @astrojs/preact | 4.x latest | Preact integration |

If the project already standardizes on React for islands, use
[react-architect](../react-architect/STACK.md) — otherwise prefer Preact.

## Notes

- **Zero JS by default is the invariant.** If the bundle is non-trivial on a
  static content page, an island was added that shouldn't have been.
- **Import Zod from `astro/zod`**, never bare `zod`. The bundled version is
  pinned to Astro's schema resolution; a mismatched instance breaks it.
- **Collections live in `src/content.config.ts`**, not the legacy
  `src/content/<name>/config.ts`. Migrate before adding new collections; don't
  run both for the same data.
- **`site:` in `astro.config.mjs` is effectively mandatory** — canonical URLs,
  sitemap, and RSS all depend on it.
- **Stay on `output: 'static'` unless a route genuinely needs a runtime.**
  Actions, Sessions, and Server Islands all require an adapter. Adding one adds
  a process to keep alive, cold starts, and a home for secrets.
- **Run `astro check` before every commit.** It is the only thing catching
  frontmatter-adjacent type errors at the boundary.

_Last reviewed: 2026-09-30_
_Skill version at last review: 1.0.0_