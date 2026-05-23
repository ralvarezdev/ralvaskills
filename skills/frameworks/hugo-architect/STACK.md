# Stack Versions

| Dependency | Pinned version | Purpose |
|---|---|---|
| hugo | 0.161.1 | Static-site generator (Extended edition mandatory) |
| go | 1.26 | Required for Hugo Modules — see [go-architect](../../languages/go-architect/STACK.md) |
| dart-sass | bundled | SCSS compilation — ships inside Hugo Extended since 0.121 |
| postcss | 8.5 | Optional — Tailwind / autoprefixer chain invoked from Hugo Pipes |
| tailwindcss | 4.1 | Optional — production builds purge via `content` globbing |
| node | 22 LTS | Required only when PostCSS / Tailwind are in use |

## Notes

- **Hugo Extended edition is mandatory.** Standard edition cannot compile SCSS, cannot run image processing, and is missing several pipeline features. Pin the `-extended` build in CI.
- **Hugo Modules over git submodules** for themes and shared layouts — SemVer-versioned, resolved via Go's module proxy. Vendor with `hugo mod vendor` for reproducible builds.
- **TOML is the default for `hugo.toml` and front matter** — match across the repo to keep one parser in the author's head.
- **Hugo Pipes is the asset pipeline** — no Webpack, no Vite. PostCSS lives outside Hugo and is invoked from the pipe chain when needed.
- **Page Bundles (`index.md` + co-located assets)** are the default for any page with images; flat `.md` files are acceptable only for text-only posts.
- **Fingerprinting is mandatory in production.** `hugo --minify` plus `resources.Fingerprint` produces immutable URLs for long-cache headers.
- **Tailwind v4 changes the build model** — `@tailwindcss/postcss` plugin, no `tailwind.config.js` required; CSS-first configuration via `@theme` directive.

_Last reviewed: 2026-05-23_
_Skill version at last review: 1.0.0_
