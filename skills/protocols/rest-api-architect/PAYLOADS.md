# REST Payload Examples

Canonical request/response shapes referenced by [SKILL.md](SKILL.md). Loaded on demand.

## 1. Cursor pagination

```
GET /v1/orders?limit=20&cursor=eyJpZCI6IjAxOTI...

200 OK
{
  "items": [ ... ],
  "next_cursor": "eyJpZCI6IjAxOTI...",   // null = end of stream
  "limit": 20
}
```

- **Cursor is opaque** — usually a base64-encoded `{last_id, last_sort_key}` tuple. Clients pass it back unmodified.
- **`limit` is a hint with a server-enforced max** (e.g. 100). Advertise the max in OpenAPI.
- **No total count by default.** Total counts on large tables are expensive — expose a separate `/count` endpoint only if a real consumer needs it.
- **Stable sort key is mandatory** — usually `(created_at DESC, id DESC)` so equal timestamps don't oscillate.

## 2. RFC 7807 problem details

Standard error shape (`application/problem+json`):

```json
{
  "type": "https://errors.myapp.io/order/already_cancelled",
  "title": "Order is already cancelled",
  "status": 409,
  "detail": "Order 01J9X... was cancelled at 2026-05-19T14:23:00Z",
  "instance": "/v1/orders/01J9X.../cancellation",
  "correlation_id": "req_01J9..."
}
```

- **`type` is a stable URL** identifying the error class — clients switch on it. Never change a type URL once published.
- **`title` is short, human-readable, fixed per type.** `detail` is the specific message for this instance.
- **`instance`** is the request URI that produced the error.
- **`correlation_id`** ties this response to server logs. Required on `5xx`.
- **One shape for every error.** Don't mix RFC 7807 with framework-default `{"detail": "..."}` shapes.

## 3. Validation errors (422)

```json
{
  "type": ".../validation",
  "title": "Validation failed",
  "status": 422,
  "errors": [
    { "field": "email", "code": "format", "message": "Must be a valid email" },
    { "field": "age",   "code": "min",    "message": "Must be >= 18" }
  ]
}
```

- `field` is dot-pathed for nested bodies (`shipping.address.zip`).
- `code` is a stable machine-friendly token (`required`, `min`, `max`, `format`, `unique`). Clients switch on this.
- `message` is human-readable text for direct display.

## 4. Rate-limit headers

```
RateLimit-Limit: 1000
RateLimit-Remaining: 873
RateLimit-Reset: 60
Retry-After: 60     // on 429 / 503 only
```

- `RateLimit-Reset` is seconds until the window resets (RFC 9331 draft).
- `Retry-After` may be seconds or an HTTP-date. Clients use this for backoff.
- Limit per *caller* (API key, user id), not per IP — IPs aren't reliable identity.
