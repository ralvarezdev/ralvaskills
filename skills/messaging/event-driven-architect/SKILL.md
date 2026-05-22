---
name: event-driven-architect
version: 1.0.0
description: Event-driven architecture — event/command/message taxonomy, Protobuf schemas + `buf breaking`, topic naming, idempotency, mandatory outbox pattern, partitioning, DLQs, schema evolution. Broker-agnostic; NATS lightweight, Kafka log-based, RabbitMQ command queues. Use when designing event flows or auditing eventual-consistency boundaries.
---

# Event-Driven Architecture

Patterns for services that communicate via asynchronous events. **Broker-agnostic** at the conceptual level; defaults inline per use case (NATS for lightweight, Kafka for log-based replay, RabbitMQ for command queues). **Protobuf schemas** for events ([protobuf-architect](../../encoding/protobuf-architect/SKILL.md)) so contracts get the same `buf breaking` discipline as gRPC. **Outbox pattern is mandatory** for any write that emits an event. See [STACK.md](STACK.md) for pinned brokers and client libraries.

## 1. Event vs message vs command — the three shapes

| Shape | Direction | Semantics | Example |
|---|---|---|---|
| **Event** | Past-tense, broadcast | "Something happened" — fact about the past, anyone can listen | `OrderPlaced`, `PaymentCaptured` |
| **Command** | Imperative, point-to-point | "Do this" — request to one specific handler | `CancelOrder`, `SendEmail` |
| **Message** | Generic envelope | Container for either — used when the distinction doesn't matter | (mostly an implementation detail) |

- **Events are immutable past-tense facts.** `OrderPlaced` was placed; nothing changes that. Subscribers react however they want.
- **Commands have one intended handler.** Multiple handlers reacting to a command is almost always wrong — it's an event in disguise. Rename it.
- **Naming:** events are `<Noun><PastVerb>` (`OrderPlaced`, `ShipmentDispatched`); commands are `<Verb><Noun>` (`PlaceOrder`, `SendShipment`).
- **Choose the shape per use case**, not per technology. Both Kafka and RabbitMQ can carry either; the discipline is in the schema and contract.

## 2. Schema — Protobuf

Per [protobuf-architect](../../encoding/protobuf-architect/SKILL.md): events are `.proto` messages, code-generated, validated by `protovalidate`, and protected from breaking changes by `buf breaking` in CI.

```proto
// acme/shop/orders/v1/events.proto
syntax = "proto3";
package acme.shop.orders.v1;

import "google/protobuf/timestamp.proto";
import "buf/validate/validate.proto";

message OrderPlaced {
  string event_id = 1 [(buf.validate.field).string.uuid = true];     // UUID v7 for ordering
  google.protobuf.Timestamp occurred_at = 2 [(buf.validate.field).required = true];
  string aggregate_id = 3 [(buf.validate.field).string.uuid = true]; // the order ID
  int32 schema_version = 4 [(buf.validate.field).int32.gte = 1];

  // Payload — minimal facts, IDs over embedded objects
  string customer_id = 10 [(buf.validate.field).string.uuid = true];
  string currency = 11 [(buf.validate.field).string.len = 3];
  string total_amount = 12;  // decimal as string per rest-api-architect §12
  repeated string item_ids = 13;
}
```

- **Standard envelope fields** on every event: `event_id` (UUID v7 — sortable, unique), `occurred_at`, `aggregate_id`, `schema_version`. Subscribers dedupe on `event_id`; replay tools sort by `occurred_at`.
- **Minimal payload.** IDs and the few facts subscribers need. Don't dump the full aggregate state — subscribers can fetch what they need via [sql-architect](../../databases/sql-architect/SKILL.md) repositories. Big payloads make schema evolution painful and replay expensive.
- **Field numbers reserved on delete** per [protobuf-architect §3](../../encoding/protobuf-architect/SKILL.md#3-field-numbering--reservation-discipline). Never reuse.
- **One file per resource's events** — `orders/v1/events.proto` holds every event the orders context emits.

## 3. Topic / subject naming

Hierarchical, snake_case, versioned. Pick a convention and enforce it.

```
acme.shop.orders.v1.order_placed
acme.shop.orders.v1.order_cancelled
acme.shop.shipments.v1.shipment_dispatched
```

- **`<org>.<context>.<resource>.<version>.<event_name>`** matches the Buf-style package path from protobuf-architect.
- **Version is part of the topic name**, not just the schema. New major version → new topic. Run them in parallel until consumers migrate.
- **Lowercase + dots** (NATS, Pub/Sub) or **lowercase + underscores** (Kafka topic names). Pick one for your broker and stick to it.
- **Document the catalog** somewhere queryable — Confluent Schema Registry, BSR, or a simple `events.md` in the repo. Producers and consumers need to agree on what exists.

## 4. Outbox pattern — mandatory for "DB write + event emit"

The dual-write problem: an HTTP handler writes a row and publishes an event. If the DB commits but the broker rejects, the event is lost — silent inconsistency. If the broker accepts but the DB rolls back, subscribers process a phantom event.

**Outbox fixes this with a single transactional write.**

```
1. HTTP handler  →  BEGIN TX
2.                  INSERT INTO orders (...)
3.                  INSERT INTO outbox (event_id, topic, payload, created_at)
4.                  COMMIT
5. Outbox publisher (separate goroutine / cron / CDC reader)
                 →  SELECT FROM outbox WHERE NOT published ORDER BY created_at
6.                  PUBLISH to broker
7.                  UPDATE outbox SET published = true
```

- **Outbox is a regular table** in the same DB as the aggregate. The write is atomic with the business write.
- **Publisher is separate** — a goroutine, a sidecar, a cron, or **CDC** (Debezium reading Postgres WAL). CDC is the most robust; goroutine is fine for small services.
- **At-least-once delivery** — same event can be republished if the publisher crashes between PUBLISH and UPDATE. Consumers must be idempotent (see §6).
- **Outbox table grows** — partition or purge published rows older than the retention window (7–30 days). Don't keep it forever.
- **Why mandatory:** there is no working pattern that avoids both the lost-event and phantom-event failure modes without the outbox. Anything else (publish-before-commit, publish-after-commit) is broken under failure.

## 5. Ordering and partitioning

Event ordering is the single hardest part of event-driven systems. **Order is per-key, not global.**

- **Order is preserved within a partition / subject**, but not across partitions. Kafka partitions by message key; NATS via subject hierarchy; RabbitMQ via consistent-hash exchanges.
- **Partition key is the aggregate ID.** All events for `order_id=abc` land on the same partition, processed in order by one consumer. Different orders process in parallel.
- **Globally-ordered events are a smell.** If you "need" global order, you actually need a single consumer (and you've lost scaling), or you're modeling the domain wrong.
- **Consumers process one partition at a time per instance.** Concurrent processing within a partition breaks ordering. Most client libs handle this; verify your config.

## 6. Idempotency — consumer must dedupe

Brokers deliver at least once. Consumers see the same event more than once under network failure, restart, or rebalance.

- **Dedupe by `event_id`.** Every consumer keeps a small store (Redis with TTL, or a `processed_events` table) of recently-seen IDs. Reject duplicates.
- **Idempotent side effects** — designing the handler so a duplicate is harmless. `INSERT ... ON CONFLICT DO NOTHING`, `UPDATE ... WHERE version = ?` (with optimistic concurrency).
- **TTL on dedupe store** — events older than the broker's retention window can't be replayed anyway; the dedupe state can roll off.
- **Exactly-once illusion**: a system that combines an idempotent consumer with at-least-once delivery is "effectively exactly-once" from the business perspective. Don't chase true exactly-once at the protocol level — it's much more expensive than just making consumers idempotent.

## 7. Dead-letter queues (DLQs)

Some events can't be processed — schema mismatch, downstream service down for too long, business invariant violation. Don't let them block the partition.

- **Every consumer has a DLQ.** A topic / subject named `<original>.dlq` receives messages the consumer gave up on.
- **Retry policy first, DLQ second.** Try N times with exponential backoff; if still failing, DLQ. Typical: 3 attempts, then DLQ.
- **DLQs are monitored.** Per [observability-architect](../../infra/observability-architect/SKILL.md): a Prometheus counter `<svc>_dlq_messages_total` with an alert on any non-zero value. A DLQ that quietly fills up is a silent outage.
- **DLQ tooling** — operator scripts to inspect, replay, or discard DLQ messages. Replay sends back to the original topic; discard logs and drops. Both audited.

## 8. Backpressure

When the consumer can't keep up with the producer, the system needs to slow down — gracefully.

- **Prefetch / consumer concurrency limits.** Don't let one consumer instance buffer 10,000 in-flight messages. Tune `max.poll.records` (Kafka), `prefetch_count` (RabbitMQ), `MaxAckPending` (NATS JetStream).
- **Lag-based autoscaling.** Watch consumer-group lag as a metric; scale out the consumer pool when lag grows. Kafka's `kafka_consumergroup_lag` is standard.
- **Reject upstream** when persistent overload — return `503 Service Unavailable` with `Retry-After` per [rest-api-architect §3](../../protocols/rest-api-architect/SKILL.md#3-status-codes). Better than building a backlog you can't drain.
- **No unbounded queues in memory.** A handler that reads from one topic and writes to another with an internal `chan` or `asyncio.Queue` needs bounded size + a timeout.

## 9. Schema evolution

Per [protobuf-architect §4](../../encoding/protobuf-architect/SKILL.md#4-versioning) — additive changes stay in the version; breaking changes go to a new `vN`.

- **Additive:** new optional field, new enum value, new event type on a new topic. Safe.
- **Breaking:** field removal, type change, semantic change. Bump to `<topic>.v2`; both `v1` and `v2` run side-by-side until consumers migrate.
- **`buf breaking` in CI** catches accidental breakage in the proto files. Per [protobuf-architect §8](../../encoding/protobuf-architect/SKILL.md#8-breaking-change-detection--buf-breaking-in-ci).
- **Consumer compatibility tests** — for each known consumer version, replay a sample event and assert it deserializes cleanly. Catches things `buf breaking` can't (semantic-level breakage that's wire-compatible).

## 10. Saga / orchestration / choreography

For multi-step workflows that span services. Two opposing patterns:

| | Choreography | Orchestration |
|---|---|---|
| Coordination | Each service reacts to events from others | A central orchestrator sends commands |
| Coupling | Loose — services know events, not orchestrator | Tighter — services know the orchestrator |
| Visibility | Distributed; trace via correlation IDs | Centralized log of the saga state |
| Debugging | Harder — must trace the event chain | Easier — one component knows the whole flow |
| Use when | 2–4 services, well-defined flow | 5+ services, complex branching, retries, compensation |

- **Start with choreography** — events flowing service-to-service, each subscriber reacts. Simpler.
- **Promote to orchestration** when the flow is genuinely complex — Temporal, Camunda, AWS Step Functions, or a custom orchestrator service.
- **Compensation actions** for partial failures: `OrderCancelled` undoes `PaymentCaptured` via `RefundIssued`. Domain-level compensations, not technical rollback.
- **Correlation ID** propagated through every event in a saga — ties the chain together in logs/traces ([observability-architect §5](../../infra/observability-architect/SKILL.md#5-correlation)).

## 11. Broker selection — when each fits

The pattern in §1–10 works on any modern broker. Where the rubber meets the road:

| Broker | Strengths | Best for |
|---|---|---|
| **NATS / JetStream** | Lightweight (single binary), low latency, simple subject hierarchy, K/V store + persistence built-in | Modern event streams without Kafka's ops cost; service-to-service messaging in microservices |
| **Apache Kafka** | Persistent log, replayable, partitioned, mature ecosystem (Connect, Streams, Schema Registry), throughput | High-volume event streams, analytics integration, replay-heavy workloads |
| **RabbitMQ** | Flexible routing (topic/direct/fanout/headers), TTL, DLQ built-in, established | Command queues, work distribution, integration with non-event-driven systems |
| **Cloud-managed equivalents** | (AWS SNS+SQS, GCP Pub/Sub, Azure Service Bus) | When ops cost matters more than feature parity |

- **Lightweight default: NATS JetStream** — single binary, conf-driven, easy to operate, fits all the patterns in this skill.
- **Log-based default: Kafka** — when replay, retention, and analytics consumers matter.
- **Command queues / legacy fit: RabbitMQ** — when integration with non-Kafka non-NATS systems pushes the choice.

Pick once per system; switching brokers mid-flight is expensive (every consumer needs new client library + the patterns translate but not the configs).

## 12. Cross-skill ties

- [protobuf-architect](../../encoding/protobuf-architect/SKILL.md) — event schemas + `buf breaking` discipline.
- [grpc-architect](../../protocols/grpc-architect/SKILL.md) — synchronous counterpart; when a call should be RPC vs event.
- [sql-architect](../../databases/sql-architect/SKILL.md) — outbox table lives in your domain DB; same transactional discipline.
- [ddd-architect §6](../../design/ddd-architect/SKILL.md#6-domain-events) — domain events as the natural source of integration events.
- [observability-architect](../../infra/observability-architect/SKILL.md) — consumer lag, DLQ depth, processing latency are first-class metrics. Correlation IDs propagated through every event.
- [grafana-architect](../../infra/grafana-architect/SKILL.md) — per-consumer-group lag dashboards; DLQ alerts.
- [rest-api-architect §8](../../protocols/rest-api-architect/SKILL.md#8-idempotency--idempotency-key-mandatory) — the `Idempotency-Key` discipline at the HTTP layer is the same idea consumers need internally.
