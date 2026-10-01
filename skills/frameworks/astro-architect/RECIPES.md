# Astro Recipes

Reference implementations extracted from [SKILL.md](SKILL.md). Load on demand
when you need a concrete skeleton; the rules and *why* stay in the skill body.

## Script processing

```astro
---
// Frontmatter runs on the server, at build time. Not shipped.
const title = 'Projects';
const count = 3;
---

<h1>{title}</h1>
<p>{count} projects</p>

<!-- Processed: bundled, deduped, hoisted, TypeScript-aware. Ships to browser. -->
<script>
  const rows = document.querySelectorAll<HTMLLIElement>('[data-project]');
  rows.forEach((row) => {
    row.addEventListener('click', () => {
      // ...
    });
  });
</script>
```

- **Processed `<script>`** is the default: TypeScript works, imports resolve,
  multiple identical tags across components are deduped into one.
- **A `<script>` inside a component is hoisted to the page head and rendered
  once**, no matter how many instances of the component exist. It cannot
  reference per-instance props or local state.
- **`is:inline` opts out completely** — no bundling, no TypeScript, no import.
  Use it for the pre-paint theme guard below and for third-party snippets you
  don't control.

```astro
<!-- Must be is:inline and in <head>: a bundled script may run after first paint. -->
<head>
  <script is:inline>
    (function () {
      var theme = localStorage.getItem('theme') || 'tokyo-night';
      var font = localStorage.getItem('font') || 'inter';
      document.documentElement.setAttribute('data-theme', theme);
      document.documentElement.setAttribute('data-font', font);
    })();
  </script>
</head>
```

- **`define:vars`** serializes server values into an inline script. The script
  is treated as `is:inline` and therefore cannot import anything.

```astro
---
const projectId = 'ralvaskills';
---
<button id="open">Open</button>
<script define:vars={{ projectId }}>
  document.getElementById('open')?.addEventListener('click', () => {
    fetch(`/api/projects/${projectId}`); // projectId is available here
  });
</script>
```

## Content collections

### Collection with schema + dynamic route

```ts
// src/content.config.ts
import { defineCollection } from 'astro:content';
import { glob } from 'astro/loaders';
import { z } from 'astro/zod';

const projects = defineCollection({
  loader: glob({ base: './src/content/projects', pattern: '**/*.md' }),
  schema: z.object({
    name: z.string(),
    cli: z.string().optional(),
    lang: z.string(),
    summary: z.string(),
    status: z.enum(['active', 'wip', 'archived']),
    category: z.enum(['tooling', 'robotics']),
    tags: z.array(z.string()).default([]),
    updated: z.coerce.date(),
    github: z.string().url(),
    demo: z.string().url().optional(),
    focus: z.string().optional(),
    pinned: z.boolean().default(false),
  }),
});

export const collections = { projects };
```

```yaml
# src/content/projects/ralvaskills.md
---
name: ralvaskills
cli: rsk
lang: Go
summary: >-
  Ever-growing collection of AI skills for OpenCode and Claude Code,
  enforcing strict clean architecture and professional standards.
status: active
category: tooling
tags: [go, cli, ai-tooling, skills]
updated: 2026-09-30
github: https://github.com/ralvarezdev/ralvaskills
demo: https://skills.ralvarez.dev
focus: Registry caching and version pinning across machines.
pinned: true
---

## The problem

Skills were scattered across machines with no versioning or sharing story.

## Architecture

...
```

```astro
---
// src/pages/projects/[slug].astro
import { getCollection, render } from 'astro:content';

export async function getStaticPaths() {
  const projects = await getCollection('projects');
  // Returns every path that should be generated. Static build fails if a
  // slug here doesn't correspond to a real entry.
  return projects.map((project) => ({
    params: { slug: project.id },
    props: { project },
  }));
}

const { project } = Astro.props;
const { Content } = await render(project);
---

<article>
  <h1>{project.data.name}</h1>
  <p>{project.data.summary}</p>
  <Content />
</article>
```

- **`getCollection` returns unsorted.** Always sort explicitly.
- **`render(entry)`** returns `{ Content }`. In Astro 5+ this replaced
  `entry.render()`.
- **`z.coerce.date()`** normalizes YAML date parsing, which otherwise varies by
  parser.
- **Import Zod from `astro/zod`** — a bare `zod` import can create duplicate
  instances and break schema resolution.

### Filtering, sorting, projecting

```astro
---
import { getCollection } from 'astro:content';

const all = await getCollection('projects');

// Filter in the query — returns a narrowed type and skips work.
const pinned = await getCollection('projects', ({ data }) => data.pinned);

// getCollection is unsorted; sort on every query that renders a list.
const byRecency = [...all].sort(
  (a, b) => b.data.updated.getTime() - a.data.updated.getTime(),
);

// Group once, reuse in the tree, the list panel, and the home page teaser.
const byCategory = all.reduce<Record<string, typeof all>>((acc, p) => {
  (acc[p.data.category] ??= []).push(p);
  return acc;
}, {});

// `undefined` keys are empty buckets — filter before rendering the separator.
const activeCategories = Object.entries(byCategory).filter(([, v]) => v.length > 0);
---

{activeCategories.map(([category, items]) => (
  <section>
    <h2>{category}</h2>
    <ul>{items.map((p) => <li>{p.data.name}</li>)}</ul>
  </section>
))}
```

### Remote data as a collection

```ts
// src/content.config.ts
const releases = defineCollection({
  loader: async () => {
    const res = await fetch('https://api.github.com/users/ralvarezdev/repos');
    const repos = await res.json();
    return repos.map((repo: any) => ({
      id: repo.name,
      name: repo.name,
      stars: repo.stargazers_count,
    }));
  },
  schema: z.object({ name: z.string(), stars: z.number() }),
});
```

- Remote loaders run **at build time**. They need a token in `import.meta.env`
  for higher rate limits, and they fail the build on a non-2xx response.
- Use one when the data is genuinely remote and unvetted. For a handful of
  hand-curated entries, a hand-written `glob` collection is better: no network
  dependency, no rate limit, no risk of leaking a private repo.
- **Never point a public loader at an authenticated endpoint.** Anything the
  build can read, anyone who reads the source can too.

## Hybrid rendering config

```js
// astro.config.mjs
// @ts-check
import { defineConfig } from 'astro/config';
import node from '@astrojs/node';

export default defineConfig({
  // Required for correct canonical URLs, sitemap, and RSS.
  site: 'https://example.com',
  // 'static' (default) or 'server'
  output: 'server',
  adapter: node({ mode: 'standalone' }),
  redirects: {
    '/blog': '/writing',
  },
  image: {
    remotePatterns: [{ protocol: 'https', hostname: 'images.example.com' }],
  },
});
```

Per-route opt-out back to static:

```astro
---
// src/pages/about.astro
export const prerender = true;
// ...
---
```

- **`site:` is not optional in practice.** Without it, canonical URLs, the
  sitemap, and RSS entries generate relative or wrong-absolute URLs.
- **`output: 'static'` + an adapter is contradictory.** If everything is static,
  drop the adapter and deploy to a CDN.

## Theme tokens via custom properties

```astro
---
// src/layouts/Base.astro
const { title } = Astro.props;
---
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>{title}</title>
    <script is:inline>
      (function () {
        var t = localStorage.getItem('theme') || 'tokyo-night';
        document.documentElement.setAttribute('data-theme', t);
      })();
    </script>
  </head>
  <body>
    <slot />
    <script is:inline>
      // No framework. The picker writes an attribute; CSS does the rest.
      document.querySelectorAll<HTMLElement>('[data-set-theme]').forEach((btn) => {
        btn.addEventListener('click', () => {
          var theme = btn.dataset.setTheme;
          if (!theme) return;
          document.documentElement.setAttribute('data-theme', theme);
          localStorage.setItem('theme', theme);
        });
      });
    </script>
  </body>
</html>
```

```css
/* src/styles/tokens.css — global, imported from the layout */
:root,
[data-theme='tokyo-night'] {
  --bg: #1a1b26;
  --bg-elev: #24283b;
  --fg: #c0caf5;
  --fg-muted: #a9b1d6;
  --fg-faint: #565f89;
  --accent: #7aa2f7;
}

[data-theme='nord'] {
  --bg: #2e3440;
  --bg-elev: #3b4252;
  --fg: #eceff4;
  --fg-muted: #d8dee9;
  --fg-faint: #4c566a;
  --accent: #88c0d0;
}
```

```astro
<style>
  /* Scoped by default: compiled and hashed to this component. */
  .card {
    background: var(--bg-elev);
    color: var(--fg);
  }
</style>

<style is:global>
  /* Escapes scoping — use for token definitions and resets. */
</style>
```

- **Tokens live in a `:root, [data-theme='x']` block** so the default applies
  without the attribute being set (SSR, JS disabled, first paint).
- **The picker is not an island.** Two custom properties and an attribute write
  is the whole implementation.
- **Contrast is a per-theme check.** A palette that passes WCAG AA against
  `--bg` may fail against a different theme's `--bg`. Verify each theme, not
  just the default.

## Actions

```ts
// src/actions/index.ts
import { defineAction } from 'astro:actions';
import { z } from 'astro/zod';

export const server = {
  contact: defineAction({
    accept: 'form',
    input: z.object({
      name: z.string().min(1),
      email: z.string().email(),
      message: z.string().min(10).max(2000),
    }),
    handler: async (input, context) => {
      // Re-verify auth here — the calling context is not a trust boundary.
      const session = await context.session?.get('user');
      if (!session) {
        return { ok: false, error: 'Not authenticated' } as const;
      }

      await deliverEmail(input); // server-only — never bundled
      return { ok: true, received: input.email } as const;
    },
  }),
};
```

```astro
---
// src/pages/contact.astro
import { actions } from 'astro:actions';

const result = Astro.getActionResult(actions.contact);
if (result && !result.error) {
  // POST / redirect / GET — prevents resubmission on refresh.
  return Astro.redirect('/thanks');
}
const errors = result?.error?.fields;
---
<form method="POST" action="?_action=contact">
  <label>Email <input type="email" name="email" required /></label>
  {errors?.email && <p role="alert">{errors.email}</p>}
  <button type="submit">Send</button>
</form>
```

- **Actions require on-demand rendering.** A fully static build cannot run them.
- **For a static site, prefer a plain `<form>` to a form service.** Adding a
  server runtime to collect one form is a bad trade.
- **Discriminated-union returns** (`{ ok: true } | { ok: false }`) so the page
  can branch in the template.

## Islands: choosing a client directive

```astro
---
import Counter from '../components/Counter.jsx';  // React island
import Chart from '../components/Chart.svelte';
---

<!-- Runs at page load. Critical path. Last resort. -->
<Counter client:load />

<!-- Loads when it scrolls into view. Right default for below-fold widgets. -->
<Chart client:visible />

<!-- Loads during requestIdleCallback. Right default for chrome that can
     arrive late: pickers, search boxes, theme toggles. -->
<SearchBox client:idle />

<!-- No server HTML at all; framework name required. Use when the component
     touches browser-only APIs during render. -->
<BrowserOnlyThing client:only="react" />
```

- **Every directive except `client:only` renders to static HTML first**, then
  hydrates. Prefer those — you get content without JS and avoid a layout shift.
- **`client:only` needs an explicit framework string** and ships no server HTML.
  It is the only correct use for components that crash on `window` during SSR.
- **Before adding any island**, check whether a processed `<script>` or a CSS
  `:checked`/`:target` does the job. On a content site, most don't need one.

## Endpoints

```ts
// src/pages/api/subscribe.json.ts
import type { APIRoute } from 'astro';

export const POST: APIRoute = async ({ request }) => {
  const body = await request.json();
  // Prefer Actions (SKILL.md §5). Reach for a bare endpoint only when a
  // non-Astro consumer must POST here — webhooks, external clients.
  return new Response(JSON.stringify({ ok: true }), {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  });
};
```

- Any `.ts` file in `src/pages/` exporting `GET`/`POST`/`PUT`/`DELETE` becomes
  an endpoint.
- Follow [rest-api-architect](../../protocols/rest-api-architect/SKILL.md) for
  status codes and error contracts if you use them.
- **Prefer Actions.** They give you Zod validation and typed client calls for
  free; a bare endpoint gives you neither.

## Dev server

```bash
pnpm astro dev --background     # start detached
pnpm astro dev status          # is it up, on what port
pnpm astro dev logs            # tail its output
pnpm astro dev stop            # stop it
pnpm check                     # astro check — TypeScript, run before commit
pnpm build                     # production build to dist/
pnpm preview                   # serve dist/ locally — what E2E should target
```