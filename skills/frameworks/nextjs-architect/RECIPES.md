# Next.js Recipes

Reference implementations extracted from [SKILL.md](SKILL.md). Load on demand when you need a concrete skeleton; the rules and *why* stay in the skill body.

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
