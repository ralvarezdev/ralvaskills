---
name: ddd-architect
version: 1.0.0
description: Domain-Driven Design — strategic-first (bounded contexts, context mapping) with aggregates, value objects, domain events, repositories as the tactical toolkit. Builds on CONTEXT.md from grill-with-docs. Event sourcing out of scope. Use when shaping a domain model, splitting/merging contexts, or designing aggregate boundaries.
---

# DDD Architecture

Domain-Driven Design as practical structure, not ceremony. **Strategic decisions** (bounded contexts, integration patterns between them) carry far more weight than **tactical patterns** (entities, value objects, aggregates) — get the boundaries right and the tactics fall into place. Builds on the domain glossary captured by [grill-with-docs](../../workflows/grill-with-docs/SKILL.md). Pairs with [hexagonal-arch](../hexagonal-arch/SKILL.md) for code structure.

## 1. Strategic DDD — bounded contexts

A **bounded context** is the scope within which a term has one unambiguous meaning. `Order` in the Ordering context isn't the same `Order` in the Fulfillment context; they share an ID but mean different things.

- **One context = one CONTEXT.md** (per [CONTEXT_FORMAT.md](../../workflows/grill-with-docs/CONTEXT_FORMAT.md)).
- **Multiple contexts in one repo = CONTEXT-MAP.md** at the root + one `CONTEXT.md` per context folder. Pattern is enforced by `grill-with-docs`; this skill consumes the artifacts.
- **A context is not a microservice.** It's a *language* boundary. One service can hold several contexts; one context can span several services (rarely a good idea, but possible).
- **A context owns its language.** When the user says "user" in Billing, that's the Billing User — even if Identity has its own User. Don't share types across contexts to "save duplication"; the meaning is different.

## 2. Context mapping — the integration patterns

When two contexts must interact, name the relationship explicitly. The pattern dictates how they couple.

| Pattern | Use when | Coupling |
|---|---|---|
| **Shared Kernel** | Two contexts share a tiny model (`CustomerId`, `Money`) | Tight — every change is bilateral |
| **Customer / Supplier** | Upstream serves downstream; downstream has input on the contract | Medium — coordinated releases |
| **Conformist** | Downstream takes whatever upstream offers | Loose for upstream; downstream pays the cost |
| **Anticorruption Layer (ACL)** | Downstream wraps upstream's model in its own language | Loose — downstream is protected |
| **Open Host Service + Published Language** | Many downstreams consume; upstream publishes a stable schema | Loose for upstream (versioned) |
| **Separate Ways** | Two contexts deliberately don't talk | Zero — best when integration cost > benefit |
| **Partnership** | Two contexts succeed or fail together | Tightest — joint planning |

- **Default to ACL for any external system you don't own.** Legacy APIs, third-party services, vendor data formats — all wrapped in your own domain language. Same `INSERT ... ON CONFLICT` thinking as a DB boundary: never let the external model leak inward.
- **Record context maps in `CONTEXT-MAP.md`** (per grill-with-docs convention). Each relationship is a one-line entry: `Ordering → Fulfillment: events (OrderPlaced); Fulfillment uses ACL to convert to its own ShipmentRequest`.

## 3. Tactical DDD — building blocks

Tactical patterns matter, but they're load-bearing only when strategic decisions are already correct. Treat them as a toolkit, not a checklist.

- **Entity** — has identity (`id`) that persists across state changes. `User`, `Order`, `Shipment`. Equality by ID, not by attributes.
- **Value Object** — no identity, defined entirely by attributes. `Money(100, USD)`, `Address(street, city)`, `EmailAddress("a@b.com")`. **Immutable**. Equality by value. Reach for these aggressively — they catch bugs at construction and document invariants in the type system.
- **Aggregate** — a cluster of entities + value objects treated as one consistency boundary. Has a **root entity** that owns the lifecycle and enforces invariants for everything inside. External code references the aggregate root only.
- **Repository** — collection-like abstraction for retrieving aggregates by identity. One repository per aggregate root (`UserRepo`, `OrderRepo`); never one repository per entity inside an aggregate.
- **Domain Service** — when behavior doesn't belong to a single entity or value object (e.g. *"transfer money between two accounts"* — neither account owns the operation).
- **Application Service** — orchestrates use cases. Doesn't contain domain logic; coordinates aggregates, transactions, and side effects.

Make value objects whenever a primitive carries hidden meaning. `string email` invites every kind of validation drift; `EmailAddress` ensures it's been validated at construction.

## 4. Aggregate sizing — small + consistency boundaries

The most common DDD mistake is aggregates that are too big.

- **An aggregate is a transaction boundary.** Whatever's inside is updated atomically; whatever's outside is eventually consistent. Bigger aggregate = more contention, more transaction conflicts, more `ABORTED` retries.
- **Default: small.** Single entity + its value objects + a few directly-owned child entities. If you can't justify why an entity needs to be inside the aggregate, it doesn't.
- **Reference other aggregates by ID, not by object.** `Order` holds `customer_id`, not a `Customer`. Loading the related aggregate is a separate query.
- **Invariants drive the boundary.** "Order line total must equal sum of line items" → line items are inside the Order aggregate. "Order ships to one of customer's saved addresses" → no, address is referenced by ID; uniqueness is enforced at the Customer aggregate level.

## 5. Repositories

- **One per aggregate root.** `UserRepo`, `OrderRepo` — not `OrderLineRepo` (lines live inside Order).
- **Collection semantics:** `repo.find_by_id(id)`, `repo.add(aggregate)`, `repo.remove(aggregate)`. Don't expose query DSLs from the repo interface; specific queries live as named methods (`active_users_signed_up_after(date)`).
- **Return aggregates, not DTOs.** DTOs are an API-layer concern (see [fastapi-architect](../../frameworks/fastapi-architect/SKILL.md) / [gin-architect](../../frameworks/gin-architect/SKILL.md)).
- **Implementation goes through [sql-architect](../../databases/sql-architect/SKILL.md)** — `psycopg + .sql files via importlib.resources` (Python) or `sqlx + //go:embed` (Go). No ORM by default; the repo *is* the abstraction layer over SQL, not a wrapper around an ORM.

## 6. Domain events

Domain events express *something significant happened* in the domain — `OrderPlaced`, `ShipmentDispatched`, `PaymentCaptured`.

- **Past tense.** `OrderPlaced`, not `PlaceOrder`. Events describe what already happened.
- **Immutable.** Once published, never modified.
- **Minimal payload.** IDs + the few facts subscribers genuinely need. Don't dump the entire aggregate state.
- **Published by the aggregate root**, after the transaction commits. Don't publish inside the transaction — readers will see events for state that may roll back.
- **Cross-context integration** runs on domain events: Ordering emits `OrderPlaced`; Fulfillment subscribes and reacts. The transport is usually a message bus (see future `event-driven-architect`).
- **In-process eventing** is fine for same-context reactions; an event handler list registered with the aggregate works.

## 7. Domain services vs application services

| | Domain service | Application service |
|---|---|---|
| Lives in | Domain layer | Application layer |
| Contains | Domain logic that doesn't fit one aggregate | Use-case orchestration; no domain logic |
| Talks to | Aggregates, value objects, other domain services | Repositories, domain services, infrastructure |
| Example | `TransferService.transfer(from, to, amount)` enforces account invariants across two aggregates | `PlaceOrderUseCase.execute(input)` loads aggregates, calls domain methods, commits transaction, publishes events |

When in doubt, **start with a method on the aggregate**. Promote to a domain service only when the behavior genuinely spans aggregates. Avoid the trap of putting *all* logic in services and leaving aggregates anemic (the famous anti-pattern).

## 8. Anticorruption layers (ACL)

When integrating with a legacy system, third-party API, or external vendor model:

- **Wrap the external model immediately at the boundary.** Convert their `Customer` to your domain's `Customer` (or `Account`, or whatever your context calls it) before any other code sees the foreign type.
- **The ACL is just a translation function** in most cases — no need for a class hierarchy.
- **Validate aggressively** at the boundary. The external system isn't guaranteed to keep its contract; reject malformed data with a domain-meaningful error.
- **One ACL per external system**, lives in the infrastructure layer, called by application services.

## 9. When DDD is overkill

DDD is right when the domain is the hard part. It's wrong when the application is.

- **CRUD-only services** — login, sign-up, a settings page. Just write the CRUD. Adding aggregates and repositories is ceremony.
- **Throwaway prototypes** — DDD's value compounds over time; if the code is gone in two months, you don't need it.
- **Pure data pipelines** — ETL, batch jobs. The "domain" is the data shape, not behavior. Use plain functions and structs.
- **Tiny services with one bounded context and minimal logic** — `ddd-architect` is fine as guidance, but you may end up with one aggregate and two value objects. That's still DDD; you just won't write much about it.

If the domain has 3–5 entities with genuine invariants and policies, DDD pays. If the domain is "rows in a table the user edits," it doesn't.

## 10. Cross-skill ties

- **`grill-with-docs`** captures the language → CONTEXT.md → which this skill consumes. The glossary is the input; aggregates / value objects / events use those exact terms.
- **`hexagonal-arch`** structures the *code* (domain core + ports + adapters). DDD shapes the *meaning* inside the domain core. They compose: DDD answers "what's in the domain"; hexagonal answers "where does the domain live and what touches it."
- **`sql-architect`** is how repositories are implemented (psycopg / sqlx + raw SQL, no ORM).
- **`rest-api-architect`** / framework architects are the application-service entry points. DTOs translate to / from aggregates at the boundary.
- **`improve-codebase-architecture`** uses the same vocabulary (modules, interfaces, seams) and benefits from DDD because the domain language gives names to good seams.
