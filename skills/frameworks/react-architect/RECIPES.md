# React Recipes

Reference scaffolds extracted from [SKILL.md](SKILL.md). Load on demand when you need a concrete template; the rules and *why* stay in the skill body.

## Feature-based project structure

Mirrors [fastapi-architect](../fastapi-architect/SKILL.md#1-project-structure--feature-based) and [gin-architect](../gin-architect/SKILL.md#1-project-structure--feature-based) — same shape, different language.

```
src/
├── App.tsx
├── main.tsx                      # ReactDOM render / Vite entry
├── config/                       # env loading via zod parse
├── lib/                          # cross-feature utilities (api client, query client setup)
├── components/                   # cross-feature primitives (Button, Modal, Form)
├── hooks/                        # cross-feature hooks (useDebounce, useMediaQuery)
├── features/
│   ├── users/
│   │   ├── api.ts                # TanStack Query keys + fetchers
│   │   ├── components/
│   │   │   ├── UserList.tsx
│   │   │   ├── UserForm.tsx
│   │   │   └── UserCard.tsx
│   │   ├── hooks/
│   │   │   └── useUsers.ts
│   │   ├── schemas.ts            # Zod request/response schemas
│   │   ├── types.ts              # User, UserCreate, UserUpdate types
│   │   └── index.ts              # public exports
│   └── orders/
│       └── ...
└── routes/                       # router-aware route components (App Router pages live in nextjs-architect)
```

- **A feature folder owns its UI, hooks, types, schemas, and API queries.** A component from `features/users/components/` is only imported within `features/users/`. Cross-feature reuse moves to `src/components/`.
- **`schemas.ts` per feature** — Zod schemas for every request body and response shape. Parse at the API boundary; throw on parse failure.
- **`index.ts` exports the public surface** of each feature. Importing from `features/users/components/UserList` is cheating; import from `features/users` and re-export deliberately.

## Built-in hooks reference

| Hook | Use for |
|---|---|
| `useState` | Component-local state |
| `useReducer` | Component-local state with non-trivial transitions or many setters |
| `useContext` | Cross-tree state that's read often, written rarely (auth, theme, locale) |
| `useEffect` | **Synchronization** with external systems (subscribe/unsubscribe to a non-React store, set up a non-React UI listener). **Not** for fetching — TanStack Query handles that. |
| `useLayoutEffect` | DOM measurement before paint (rare; usually wrong) |
| `useId` | Stable unique IDs for `htmlFor`/`aria-labelledby` |
| `useTransition` | Mark non-urgent updates so urgent ones don't block |
| `useDeferredValue` | Same idea, value-side |
| `use` (React 19) | Unwrap a promise inside a component (with Suspense) — preferred over manual loading state where it fits |

## `tsconfig.json` — strict baseline

Every flag earns its place. `noUncheckedIndexedAccess` and `exactOptionalPropertyTypes` together close most of TypeScript's safety gaps.

```jsonc
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitOverride": true,
    "noFallthroughCasesInSwitch": true,
    "exactOptionalPropertyTypes": true,
    "verbatimModuleSyntax": true,
    "module": "preserve",
    "moduleResolution": "bundler",
    "jsx": "react-jsx",
    "target": "ES2024"
  }
}
```

## Component skeleton — `Props` type + loading/error guard

```tsx
// features/users/components/UserCard.tsx
import { useUser } from "../hooks/useUser";

type UserCardProps = {
  userId: string;
  onEdit?: () => void;
};

export function UserCard({ userId, onEdit }: UserCardProps) {
  const { data, isPending, isError, error } = useUser(userId);

  if (isPending) return <UserCardSkeleton />;
  if (isError) return <UserCardError error={error} retry={refetch} />;

  return (
    <article aria-label={`User ${data.name}`}>
      <h3>{data.name}</h3>
      {/* ... */}
      {onEdit && <button onClick={onEdit}>Edit</button>}
    </article>
  );
}
```

## Compound vs. configuration component shape

Configuration shape (fragile — every new feature is a new prop):

```tsx
<DataTable
  columns={[...]}
  rows={[...]}
  showHeader
  sortable
  filterable
  selectable
  paginated
  ...
/>
```

Composition shape (extensible — slots compose):

```tsx
<DataTable rows={...}>
  <DataTable.Header>...</DataTable.Header>
  <DataTable.Body>...</DataTable.Body>
  <DataTable.Pagination />
</DataTable>
```

## Zustand store — focused, single-purpose

A multi-step checkout flow modeled as one store. Selectors at the call site let consumers subscribe to just the slice they read.

```ts
import { create } from "zustand";

interface CheckoutStore {
  step: 1 | 2 | 3;
  shipping: ShippingInfo | null;
  payment: PaymentInfo | null;
  next: () => void;
  setShipping: (s: ShippingInfo) => void;
  reset: () => void;
}

export const useCheckoutStore = create<CheckoutStore>((set) => ({
  step: 1,
  shipping: null,
  payment: null,
  next: () => set((s) => ({ step: Math.min(s.step + 1, 3) as 1|2|3 })),
  setShipping: (s) => set({ shipping: s }),
  reset: () => set({ step: 1, shipping: null, payment: null }),
}));
```

## Suspense + ErrorBoundary at a route

Loading and error both handled declaratively. TanStack Query's `useSuspenseQuery` plugs into the same boundaries.

```tsx
function UsersRoute() {
  return (
    <ErrorBoundary fallback={<RouteError />}>
      <Suspense fallback={<RouteSkeleton />}>
        <UserList />
      </Suspense>
    </ErrorBoundary>
  );
}
```
