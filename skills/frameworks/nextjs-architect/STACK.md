# Stack Versions

| Dependency | Pinned version | Purpose |
|---|---|---|
| next | 16.2 | Framework (App Router only) |
| react | 19.2 | UI library (pinned by Next) — see [react-architect](../react-architect/STACK.md) |
| react-dom | 19.2 | DOM renderer |
| typescript | 6.0 | Static typing — strict mode mandatory |
| zod | 3.x latest | Server action validation + boundary types |
| @vercel/otel | latest | OpenTelemetry instrumentation when deploying on Vercel; generic OTel SDK otherwise — see [observability-architect](../../infra/observability-architect/STACK.md) |
| next-auth | 5.x (Auth.js) | Session-based auth for first-party apps |
| @biomejs/biome | 2.4 | Lint + format |
| vitest | 4.1 | Unit + component tests |
| @playwright/test | 1.60 | End-to-end tests against `next start` |

## Notes

- **App Router only.** Pages Router is legacy; no new code there.
- **Server components default; `"use client"` opts out.** Hoist the directive to the smallest leaf component that needs it.
- **Server actions are the canonical mutation mechanism.** API routes earn their place for webhooks, external clients, streaming.
- **Hybrid data access** (per the project's choice): direct DB from RSC for reads, API routes for writes — but `react-architect` and project conventions can override per project.
- **Runtime: node default; edge** only when latency justifies it (auth checks, redirects, A/B routing).
- **`next/image` and `next/font` are mandatory** — never `<img>` or `@import` for fonts.
- **Standalone output** (`output: "standalone"`) for Docker deploys per [docker-architect](../../infra/docker-architect/SKILL.md).
- **Inherits the `react-architect` stack** for client-side patterns; this file pins only what's Next-specific.

_Last reviewed: 2026-05-21_
_Skill version at last review: 1.0.0_
