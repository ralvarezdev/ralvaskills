# DDD — Reference Tables

Strategic and tactical pattern tables referenced by [SKILL.md](SKILL.md). Loaded on demand.

## 1. Context mapping patterns

| Pattern | Use when | Coupling |
|---|---|---|
| **Shared Kernel** | Two contexts share a tiny model (`CustomerId`, `Money`) | Tight — every change is bilateral |
| **Customer / Supplier** | Upstream serves downstream; downstream has input on the contract | Medium — coordinated releases |
| **Conformist** | Downstream takes whatever upstream offers | Loose for upstream; downstream pays the cost |
| **Anticorruption Layer (ACL)** | Downstream wraps upstream's model in its own language | Loose — downstream is protected |
| **Open Host Service + Published Language** | Many downstreams consume; upstream publishes a stable schema | Loose for upstream (versioned) |
| **Separate Ways** | Two contexts deliberately don't talk | Zero — best when integration cost > benefit |
| **Partnership** | Two contexts succeed or fail together | Tightest — joint planning |

- **Default to ACL** for any external system you don't own — legacy APIs, third-party services, vendor data formats. Never let the external model leak inward.
- **Record relationships in `CONTEXT-MAP.md`** (per grill-with-docs convention). One-line entries: `Ordering → Fulfillment: events (OrderPlaced); Fulfillment uses ACL to convert to its own ShipmentRequest`.

## 2. Domain service vs application service

| | Domain service | Application service |
|---|---|---|
| Lives in | Domain layer | Application layer |
| Contains | Domain logic that doesn't fit one aggregate | Use-case orchestration; no domain logic |
| Talks to | Aggregates, value objects, other domain services | Repositories, domain services, infrastructure |
| Example | `TransferService.transfer(from, to, amount)` enforces account invariants across two aggregates | `PlaceOrderUseCase.execute(input)` loads aggregates, calls domain methods, commits transaction, publishes events |

- **Start with a method on the aggregate.** Promote to a domain service only when the behavior genuinely spans aggregates.
- Avoid the **anemic-domain** anti-pattern — putting *all* logic in services and leaving aggregates with no behavior.

## 3. Tactical building blocks

| Block | Identity | Equality | Typical use |
|---|---|---|---|
| **Entity** | Has `id` | By ID | `User`, `Order`, `Shipment` — survives state changes |
| **Value Object** | None | By attributes | `Money(100, USD)`, `Address(...)` — immutable |
| **Aggregate** | Root entity ID | Through root | Transaction boundary; root owns invariants |
| **Repository** | One per aggregate root | — | Collection semantics over persistence |
| **Domain Service** | Stateless | — | Behavior spanning multiple aggregates |
| **Application Service** | Stateless | — | Use-case orchestration; no domain logic |

Reach for **value objects aggressively**: `string email` invites validation drift; `EmailAddress` ensures it's validated at construction.
