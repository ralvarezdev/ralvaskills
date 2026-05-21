# Idempotency — `Idempotency-Key` Reference

Every `POST` and `PATCH` requires an `Idempotency-Key` header. Without it the server returns `400 Bad Request`. This file is the implementation reference; the rule itself is stated in [SKILL.md](SKILL.md) §8.

## Request shape

```
POST /v1/payments
Idempotency-Key: 7f3a-9b2c-4d8e-...
Content-Type: application/json

{ "amount": "99.99", "currency": "USD", "source": "card_..." }
```

## Response semantics

| Scenario | Behavior |
|---|---|
| First request with this key | Execute, store response (status + body + headers), return it |
| Replay: same key + same caller + same method + same path + same body | Return the cached response unchanged; **do not re-execute** |
| Same key + same caller + same method + same path but **different body** | `409 Conflict` with `type: ".../idempotency_key_conflict"` |
| Key arrives after TTL has expired | Treated as first request; new execution |
| `GET`, `PUT`, `DELETE` | `Idempotency-Key` ignored — these methods are idempotent by HTTP semantics |

## Cache key

The cache lookup key is the tuple:

```
(caller_id, method, path, idempotency_key)
```

Including `caller_id` prevents key collisions between callers using the same opaque key (Stripe's pattern). Including `method` and `path` further reduces collision surface and makes the cache human-debuggable.

## TTL

Default **24 hours**. Long enough that:

- Mobile clients can retry after going offline and back online
- Backend retries from queues with hours of delay still match
- Operator-triggered manual retries (`curl ... --retry`) match

Set higher (e.g. 7 days) for endpoints whose effects are deeply external (payment captures, email sends, third-party API calls) — the cost of a duplicate execution justifies the storage.

## Storage

Server-side cache holds the full response (status, body, headers) keyed by the cache key. Backing storage can be:

- **Redis** with TTL keys — simplest, fast
- **Postgres** table with `expires_at` column and a background sweep — durable, queryable

The cache must be **strongly consistent across replicas** — a replay landing on a different replica must still see the cached response. Otherwise the guarantee is broken.

## Concurrency

Two concurrent requests with the same key:

1. First wins the "lock" (Redis `SET NX`, Postgres `INSERT ON CONFLICT DO NOTHING`).
2. Second request **blocks** until the first completes, then reads and returns the cached response.

Don't let both execute — that defeats the entire point.

## Client guidance (for SDKs)

- Generate one UUID v4 per **logical operation** (not per HTTP attempt). Reuse across retries.
- Discard the key once you've received a `2xx` or `409` from the server — neither will change with another attempt.
- On `5xx` or network failure, retry with the **same** key.

## What `Idempotency-Key` does NOT do

- It doesn't make a request idempotent at the database level — your handler is still responsible for not double-inserting if you somehow execute twice. The key prevents the framework from invoking your handler twice for the same retry, but a misconfigured server with no cache will still let duplicates through.
- It doesn't validate the body. A duplicate key with a different body is detected and rejected (`409`), but it's the client's bug — fix the client.
- It doesn't replace transactional design. Use it in concert with `INSERT ... ON CONFLICT` or `UNIQUE` constraints, not in place of them.
