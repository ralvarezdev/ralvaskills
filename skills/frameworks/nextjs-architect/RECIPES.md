# Next.js Recipes

Reference implementations extracted from [SKILL.md](SKILL.md). Load on demand when you need a concrete skeleton; the rules and *why* stay in the skill body.

## App Router folder layout

```
app/
├── layout.tsx                        # root layout (html/body, Providers)
├── page.tsx                          # /
├── loading.tsx                       # default loading UI for /
├── error.tsx                         # error UI for /
├── not-found.tsx                     # 404 UI
├── (auth)/                           # route group — doesn't add to URL
│   ├── login/page.tsx                # /login
│   └── signup/page.tsx               # /signup
├── dashboard/
│   ├── layout.tsx                    # nested layout for /dashboard/*
│   ├── page.tsx                      # /dashboard
│   ├── @analytics/page.tsx           # parallel route slot
│   └── users/
│       ├── page.tsx                  # /dashboard/users
│       └── [id]/
│           ├── page.tsx              # /dashboard/users/[id]
│           └── edit/page.tsx         # /dashboard/users/[id]/edit
└── api/
    └── webhooks/
        └── stripe/route.ts           # POST /api/webhooks/stripe
```

- `page.tsx` = route UI; `layout.tsx` wraps it. Both default to server components.
- `loading.tsx` wraps the route in Suspense automatically.
- `error.tsx` is a client component that wraps the route in an ErrorBoundary.
- Route groups `(name)` organize without adding URL segments.
- `@slot` is a parallel route — independent loading/error states, rendered into named slots in the parent layout.

## Edge runtime declaration

```tsx
// app/api/quick-check/route.ts
export const runtime = "edge";

export async function GET(req: Request) {
  return Response.json({ ok: true });
}
```

- Default is `node` — full Node API, npm packages with native modules, deepest compatibility.
- `runtime = "edge"` for low-latency, geographic-edge routes (auth verification, redirects, feature flags, A/B routing).
- Edge restrictions: no Node-only modules (`fs`, `child_process`), no native bindings, smaller bundle limit (~1 MB).
- Standard Postgres drivers don't work on edge; use HTTP-based drivers (Neon serverless, Cloudflare D1, Turso libsql).

## Metadata via `generateMetadata`

```tsx
// app/dashboard/users/[id]/page.tsx
import type { Metadata } from "next";

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { id } = await params;
  const user = await getUser(id);
  return {
    title: user ? `${user.name} — Users` : "User not found",
    description: user?.bio,
    openGraph: { images: user?.avatarUrl ? [user.avatarUrl] : [] },
  };
}
```

- Per-route `generateMetadata` — async, runs on the server, has access to params.
- Root `metadata` export in `app/layout.tsx` for defaults (title template, OG defaults).
- Don't hand-write `<head>` — Next manages it for SEO + social embed correctness.

## Cache layers reference

| Cache | Default | Override |
|---|---|---|
| **`fetch()` data cache** | Off by default in Next 16 | `fetch(url, { cache: "force-cache", next: { revalidate: 60 } })` |
| **Full route cache** | Static routes cached at build; dynamic routes per-request | `export const dynamic = "force-static"` / `"force-dynamic"` |
| **Router cache (client)** | Caches RSC payloads for 30 s | `router.refresh()` to bust |
| **`unstable_cache`** | Wrap arbitrary data functions with TTL + tag-based invalidation | Use tags for `revalidateTag()` |

- Don't cache user-specific data globally. Use `cookies()` / `headers()` to opt out, or scope the cache to the user.
- `revalidateTag` after mutations so server-rendered data refreshes when the underlying data changes.
- Profile before tuning. Most apps don't need cache surgery; trust the defaults until something's measurably slow.

## Server / client boundary

Server component fetches data and passes it to a client leaf for interactivity. The `"use client"` directive sits on the leaf, not the page.

```tsx
// app/dashboard/users/page.tsx — server component (default)
import { db } from "@/lib/db";
import { UserListClient } from "./UserListClient";

export default async function UsersPage() {
  const users = await db.query.users.findMany();   // runs on server, no JS shipped for this
  return <UserListClient users={users} />;
}
```

```tsx
// app/dashboard/users/UserListClient.tsx — client component (opt-in)
"use client";
import { useState } from "react";

type Props = { users: User[] };
export function UserListClient({ users }: Props) {
  const [filter, setFilter] = useState("");
  /* ... interactive UI ... */
}
```

## Server action with Zod validation + form binding

Server action validates input with Zod, re-checks auth inside the action, and returns a discriminated union the form consumes via `useActionState`.

```tsx
// app/dashboard/users/actions.ts
"use server";
import { revalidatePath } from "next/cache";
import { z } from "zod";
import { getCurrentUser } from "@/lib/auth";
import { db } from "@/lib/db";

const Schema = z.object({
  name: z.string().min(1).max(100),
  email: z.string().email(),
});

export async function createUser(formData: FormData) {
  const user = await getCurrentUser();
  if (!user) throw new Error("Unauthorized");

  const parsed = Schema.safeParse(Object.fromEntries(formData));
  if (!parsed.success) {
    return { ok: false as const, fieldErrors: parsed.error.flatten().fieldErrors };
  }

  await db.insert(users).values(parsed.data);
  revalidatePath("/dashboard/users");
  return { ok: true as const };
}
```

```tsx
// app/dashboard/users/NewUserForm.tsx
"use client";
import { useActionState } from "react";
import { createUser } from "./actions";

export function NewUserForm() {
  const [state, formAction, pending] = useActionState(createUser, null);
  return (
    <form action={formAction}>
      <input name="name" required />
      <input name="email" type="email" required />
      <button disabled={pending}>Create</button>
      {state?.fieldErrors && /* render errors */}
    </form>
  );
}
```

## Read directly from a server component

Server component fetches and renders. `notFound()` triggers the closest `not-found.tsx`.

```tsx
// app/dashboard/users/[id]/page.tsx
import { notFound } from "next/navigation";
import { getUser } from "@/lib/users";

export default async function UserPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const user = await getUser(id);
  if (!user) notFound();
  return <UserDetail user={user} />;
}
```

## Per-region streaming with Suspense

Each `<Suspense>` boundary streams independently. Pair with skeletons sized to the eventual content to avoid layout shift.

```tsx
// app/dashboard/page.tsx
import { Suspense } from "react";

export default function Dashboard() {
  return (
    <div>
      <h1>Dashboard</h1>
      <Suspense fallback={<AnalyticsSkeleton />}>
        <Analytics />            {/* slow async server component */}
      </Suspense>
      <Suspense fallback={<RecentSkeleton />}>
        <Recent />               {/* slow async server component */}
      </Suspense>
    </div>
  );
}
```
