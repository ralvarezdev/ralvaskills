# Stack Versions

| Dependency | Pinned version | Purpose |
|---|---|---|
| Protobuf | per [protobuf-architect](../../encoding/protobuf-architect/STACK.md) | Event schema language; `buf breaking` in CI |
| buf CLI | 1.69 | Schema lint + breaking-change detection |
| NATS server (JetStream) | 2.14 | **Lightweight default**: single binary, low ops cost, persistence + K/V built in |
| Kafka | 3.x stable | **Log-based default**: replayable, partitioned, mature ecosystem |
| RabbitMQ | 4.3 | **Command-queue default**: flexible routing, work distribution, legacy fit |
| nats.go | 1.52 | NATS Go client |
| nats-py | 2.x | NATS Python client |
| confluent-kafka-python | 2.14 | Kafka Python client (production-grade) |
| segmentio/kafka-go | 0.4 | Kafka Go client (simpler than Sarama) |
| Debezium | latest stable | Postgres CDC for outbox publishing (when CDC pattern is used) |

## Notes

- **Schema format: Protobuf** is promoted. Same `.proto` + `buf breaking` discipline as gRPC. Optional CloudEvents envelope can wrap Protobuf payloads when consumers expect the CloudEvents contract.
- **Broker is per-system, not per-event.** Pick one broker for a system; rebalance only when the workload genuinely demands a different shape.
- **Outbox pattern is mandatory** for any write that emits an event — `BEGIN TX → INSERT business + INSERT outbox → COMMIT` then publish via a separate process (goroutine, sidecar, or Debezium CDC).
- **At-least-once delivery + idempotent consumers** = effective exactly-once. Don't chase protocol-level exactly-once.
- **Per-key partitioning** for ordering — events for the same aggregate land in the same partition.
- **DLQ + retry policy** for every consumer. DLQ depth is a monitored, alerted metric per [observability-architect](../../infra/observability-architect/SKILL.md).
- **Schema Registry (Confluent / BSR)** is the upgrade path for organizations publishing schemas to multiple consumer teams. Out of scope for solo / small-team use.

_Last reviewed: 2026-05-21_
_Skill version at last review: 1.0.0_
