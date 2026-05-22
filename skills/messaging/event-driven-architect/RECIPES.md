# Event-Driven — Schemas & Patterns

Reference Protobuf and SQL patterns for the rules in [SKILL.md](SKILL.md). Loaded on demand.

## 1. Canonical event schema

```proto
// acme/shop/orders/v1/events.proto
syntax = "proto3";
package acme.shop.orders.v1;

import "google/protobuf/timestamp.proto";
import "buf/validate/validate.proto";

message OrderPlaced {
  // Envelope — every event has these four
  string event_id = 1 [(buf.validate.field).string.uuid = true];     // UUID v7 — sortable, unique
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

- **Standard envelope fields** on every event: `event_id` (UUID v7), `occurred_at`, `aggregate_id`, `schema_version`. Subscribers dedupe on `event_id`; replay tools sort by `occurred_at`.
- **Minimal payload.** IDs and the few facts subscribers need. Don't dump the full aggregate state — they can fetch via [sql-architect](../../databases/sql-architect/SKILL.md) repositories. Big payloads make schema evolution and replay expensive.
- **Field numbers reserved on delete** per [protobuf-architect §3](../../encoding/protobuf-architect/SKILL.md#3-field-numbering--reservation-discipline). Never reuse.
- **One file per resource's events** — `orders/v1/events.proto` holds every event the orders context emits.

## 2. Outbox table + publisher

### Schema

```sql
CREATE TABLE outbox (
    event_id      uuid PRIMARY KEY,
    topic         text NOT NULL,
    payload       bytea NOT NULL,
    occurred_at   timestamptz NOT NULL DEFAULT now(),
    published     boolean NOT NULL DEFAULT false,
    published_at  timestamptz
);

CREATE INDEX outbox_unpublished_idx ON outbox (occurred_at) WHERE NOT published;
```

### Atomic write (single transaction)

```
BEGIN;
  INSERT INTO orders (...);
  INSERT INTO outbox (event_id, topic, payload) VALUES ($1, $2, $3);
COMMIT;
```

### Publisher loop (goroutine / cron / CDC)

```
SELECT event_id, topic, payload FROM outbox WHERE NOT published ORDER BY occurred_at LIMIT 100;
-- for each row:
PUBLISH topic, payload;
UPDATE outbox SET published = true, published_at = now() WHERE event_id = $1;
```

- **Publisher is separate** — goroutine, sidecar, cron, or **CDC** (Debezium reading Postgres WAL). CDC is the most robust; goroutine is fine for small services.
- **At-least-once delivery** — same event can be republished if the publisher crashes between PUBLISH and UPDATE. Consumers must be idempotent (see §3).
- **Outbox table grows** — partition or purge rows older than 7–30 days. Don't keep forever.

## 3. Consumer idempotency

### Dedupe table (Postgres)

```sql
CREATE TABLE processed_events (
    event_id    uuid PRIMARY KEY,
    consumer    text NOT NULL,
    processed_at timestamptz NOT NULL DEFAULT now()
);
```

### Handler skeleton (Go)

```go
func (h *Handler) Handle(ctx context.Context, evt OrderPlaced) error {
    tx, _ := h.db.BeginTx(ctx, nil)
    defer tx.Rollback()

    var dup bool
    err := tx.QueryRowContext(ctx,
        `SELECT EXISTS(SELECT 1 FROM processed_events WHERE event_id = $1 AND consumer = $2)`,
        evt.EventId, "billing-service").Scan(&dup)
    if err != nil { return err }
    if dup { return nil } // already processed

    // ... business logic ...

    _, err = tx.ExecContext(ctx,
        `INSERT INTO processed_events (event_id, consumer) VALUES ($1, $2)`,
        evt.EventId, "billing-service")
    if err != nil { return err }
    return tx.Commit()
}
```

- **Dedupe by `event_id`.** Redis with TTL works too; choose based on storage cost vs query cost.
- **Idempotent side effects** also work: `INSERT … ON CONFLICT DO NOTHING`, `UPDATE … WHERE version = ?` (optimistic concurrency).
- **TTL on dedupe store** — events older than the broker's retention can't be replayed anyway.

## 4. DLQ topic + replay tooling

### Topic naming

- Original topic: `acme.shop.orders.v1.order_placed`
- DLQ topic: `acme.shop.orders.v1.order_placed.dlq`

### Retry policy

```
attempt 1 → fail → backoff 1s
attempt 2 → fail → backoff 5s
attempt 3 → fail → publish to DLQ; ack original
```

### Observability

```promql
<svc>_dlq_messages_total{topic="acme.shop.orders.v1.order_placed.dlq"} > 0
```

Alert on any non-zero value. A DLQ that quietly fills is a silent outage.

### Replay tool (CLI shape)

```
$ event-dlq inspect --topic order_placed.dlq --limit 10
$ event-dlq replay  --topic order_placed.dlq --event-id 01J9X...
$ event-dlq discard --topic order_placed.dlq --event-id 01J9X... --reason "schema mismatch (legacy)"
```

Both replay and discard are audited.

## 5. Topic / subject naming reference

```
<org>.<context>.<resource>.<version>.<event_name>

acme.shop.orders.v1.order_placed
acme.shop.orders.v1.order_cancelled
acme.shop.shipments.v1.shipment_dispatched
```

- Matches the Buf-style package path from [protobuf-architect](../../encoding/protobuf-architect/SKILL.md).
- **Version is part of the topic name**, not just the schema. New major version → new topic. Run them in parallel until consumers migrate.
- **Lowercase + dots** (NATS, Pub/Sub) or **lowercase + underscores** (Kafka topic names). Pick one per broker.

## 6. Choreography vs orchestration

| | Choreography | Orchestration |
|---|---|---|
| Coordination | Each service reacts to events from others | A central orchestrator sends commands |
| Coupling | Loose — services know events, not orchestrator | Tighter — services know the orchestrator |
| Visibility | Distributed; trace via correlation IDs | Centralized log of the saga state |
| Debugging | Harder — trace the event chain | Easier — one component knows the whole flow |
| Use when | 2–4 services, well-defined flow | 5+ services, complex branching, retries, compensation |

- Start with choreography — events flowing service-to-service, each subscriber reacts. Simpler.
- Promote to orchestration when the flow is genuinely complex — Temporal, Camunda, AWS Step Functions, or a custom orchestrator.
- Compensation actions for partial failures: `OrderCancelled` undoes `PaymentCaptured` via `RefundIssued`. Domain-level, not technical rollback.
- Correlation ID propagated through every event in a saga — ties the chain together in logs/traces.

## 7. Broker selection reference

| Broker | Strengths | Best for |
|---|---|---|
| **NATS / JetStream** | Lightweight (single binary), low latency, simple subject hierarchy, K/V store + persistence built-in | Modern event streams without Kafka's ops cost; service-to-service messaging |
| **Apache Kafka** | Persistent log, replayable, partitioned, mature ecosystem (Connect, Streams, Schema Registry), throughput | High-volume streams, analytics integration, replay-heavy workloads |
| **RabbitMQ** | Flexible routing (topic/direct/fanout/headers), TTL, DLQ built-in, established | Command queues, work distribution, integration with non-event-driven systems |
| **Cloud-managed equivalents** | (AWS SNS+SQS, GCP Pub/Sub, Azure Service Bus) | When ops cost matters more than feature parity |

- Lightweight default: NATS JetStream — single binary, conf-driven, easy to operate.
- Log-based default: Kafka — when replay, retention, and analytics consumers matter.
- Command queues / legacy fit: RabbitMQ — when integration with non-Kafka non-NATS systems forces the choice.
- Switching brokers mid-flight is expensive (every consumer needs a new client library + configs translate but the patterns don't).

## 8. Backpressure tuning per broker

| Broker | Knob | Typical starting value |
|---|---|---|
| Kafka | `max.poll.records` | 500 |
| Kafka | `max.partition.fetch.bytes` | 1 MB |
| RabbitMQ | `prefetch_count` | 50 |
| NATS JetStream | `MaxAckPending` | 1000 |

- **Lag-based autoscaling** — scale consumers when `kafka_consumergroup_lag` (or broker equivalent) grows.
- **Reject upstream** when persistently overloaded — return `503 Service Unavailable` with `Retry-After` per [rest-api-architect §3](../../protocols/rest-api-architect/SKILL.md#3-status-codes).
- **No unbounded in-memory queues.** Bound channel size + timeout in any handler that reads from one topic and writes to another.
