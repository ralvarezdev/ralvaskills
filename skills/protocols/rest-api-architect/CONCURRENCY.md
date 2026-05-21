# Concurrency Control — `ETag` + `If-Match` Reference

Every editable resource exposes an `ETag` on read. Every `PUT` and `PATCH` requires `If-Match`. This file is the implementation reference; the rule itself is stated in [SKILL.md](SKILL.md) §9.

## Why it's required

Two clients fetch the same resource and both write back. Without preconditions, the second write silently overwrites the first.

```
Client A: GET /v1/orders/01J9...   →  {status: "pending"}
Client B: GET /v1/orders/01J9...   →  {status: "pending"}

Client A: PUT /v1/orders/01J9...   →  {status: "paid"}
Client B: PUT /v1/orders/01J9...   →  {status: "cancelled"}   ← A's change lost, A never notified
```

With `ETag` + `If-Match`, the second writer hits `412 Precondition Failed` and must re-fetch + retry.

## Wire protocol

```
GET /v1/orders/01J9...
200 OK
ETag: "v7-abc123"

{ ... }
```

```
PUT /v1/orders/01J9...
If-Match: "v7-abc123"
Content-Type: application/json
{ ... }

200 OK
ETag: "v8-def456"

{ ... }
```

Stale write:

```
PUT /v1/orders/01J9...
If-Match: "v7-abc123"

412 Precondition Failed
Content-Type: application/problem+json
{
  "type": "https://errors.myapp.io/concurrency/stale",
  "title": "Resource was modified by another request",
  "status": 412,
  "detail": "Re-fetch to get the current ETag and retry",
  "instance": "/v1/orders/01J9..."
}
```

## Where the ETag value comes from

Server-side, the ETag is derived from one of:

| Source | Pros | Cons |
|---|---|---|
| Monotonic version column (`version INT`) | Cheap, predictable, never collides | Requires schema column on every editable table |
| `updated_at` timestamp | No new column | Sub-millisecond collisions possible under high concurrency |
| Hash of canonical representation (SHA-256 of normalized JSON) | Truly content-based; survives no-op writes | Expensive on large resources |

**Default: monotonic version column.** Bump it in the same statement as the update (`SET ..., version = version + 1 WHERE id = ? AND version = ?`). The `WHERE version = ?` clause is itself the concurrency guard — `0` rows updated → `412`.

## Strong vs weak ETags

- **Strong** (`ETag: "v7-abc123"`) — bytes match exactly. Default.
- **Weak** (`ETag: W/"v7-abc123"`) — semantically equivalent (same after canonicalization). Use when you serve the same resource with different gzip compression / whitespace.

Most APIs use strong. Don't reach for weak unless you have a clear reason.

## `If-None-Match` on GET — free 304 caching

Clients can send the last ETag they have on a `GET` request:

```
GET /v1/orders/01J9...
If-None-Match: "v7-abc123"

304 Not Modified
ETag: "v7-abc123"
(no body)
```

If the server's current ETag matches, return `304` with no body. The client reuses its cached representation. Costs almost nothing server-side (a single lookup) and saves bandwidth proportional to the response size.

Always include the `ETag` header on the `304` so the client can keep using the same value.

## Concurrent edits in practice

When a `412` happens, the client must:

1. Re-`GET` the resource (new ETag, new state)
2. Re-apply its intended change against the new state
3. PUT/PATCH again with the new `If-Match`

If the change is irreconcilable with the new state (the resource transitioned to a state where the change no longer makes sense), surface a domain error — not another retry.

## Common mistakes

- **Forgetting `If-Match` on `PATCH`.** Easy to omit because PATCH "feels small." Enforce at the framework layer — reject the request if absent for any editable resource.
- **Generating the ETag *after* the response body is serialized.** Race: client gets ETag `X`, but the DB row was already at version `X+1` between read and serialization. Always compute ETag from the same query/transaction that produced the body.
- **Strong ETag on a JSON response with map field ordering that varies.** The bytes change, ETags mismatch, clients see "stale" on identical content. Use canonical serialization (sorted keys) or switch to a version-column derived ETag.
- **No `Vary: Accept-Encoding`.** A proxy caches the gzipped representation, returns it to a client that didn't send `Accept-Encoding: gzip`, ETag matches but bytes don't.
