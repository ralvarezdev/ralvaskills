# React Recipes

Reference scaffolds extracted from [SKILL.md](SKILL.md). Load on demand when you need a concrete template; the rules and *why* stay in the skill body.

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
