# HTTP Status Codes — Reference Table

The status codes architects need to use correctly. Load this reference when designing a new endpoint, reviewing error responses, or debugging client-server status confusion.

| Code | Meaning | Use for |
|---|---|---|
| `200` | OK | Successful GET, PUT, PATCH, DELETE with response body |
| `201` | Created | Successful POST that created a resource. Include `Location` header pointing to the new resource |
| `202` | Accepted | Async work queued; response body includes how to check status |
| `204` | No Content | Successful DELETE with no body, or PATCH returning nothing |
| `400` | Bad Request | Malformed request (bad JSON, missing required field) — schema problems |
| `401` | Unauthorized | Missing or invalid credentials. **Not "you're not allowed"** — that's 403 |
| `403` | Forbidden | Credentials valid but caller can't do this action |
| `404` | Not Found | Resource doesn't exist (or shouldn't be visible to this caller) |
| `409` | Conflict | Request can't be applied to current resource state (e.g. duplicate UNIQUE, idempotency-key collision with different body) |
| `410` | Gone | Resource existed but is permanently removed (vs 404 "never existed") |
| `412` | Precondition Failed | `If-Match` ETag mismatch — concurrent edit detected (see CONCURRENCY.md) |
| `415` | Unsupported Media Type | Wrong `Content-Type` header |
| `422` | Unprocessable Entity | Syntactically valid request but semantically invalid (validation) |
| `429` | Too Many Requests | Rate limit hit. Include `Retry-After` header |
| `500` | Internal Server Error | Something genuinely unexpected; include a correlation id, never a stack trace |
| `503` | Service Unavailable | Temporary unavailability (maintenance, dependency down). Include `Retry-After` |

## Hard rules

- **Never `200 OK` for errors.** Returning `{"success": false, "error": ...}` with a 200 status is wrong and breaks every HTTP-aware tool.
- **401 vs 403 strict.** 401 = "who are you?" (missing/invalid creds). 403 = "I know who you are, and no" (creds valid, action forbidden). Mixing them confuses clients debugging access issues.
- **404 vs 410.** 404 = "I have no record of this resource (or you can't see it)". 410 = "this resource definitely existed and is gone forever". Use 410 when a delete should be observable to caches and clients (e.g. takedown).
- **422 vs 400.** 400 = the request didn't parse at all. 422 = it parsed but its content is invalid (business rules, validation).
- **All `5xx` include a `correlation_id`** the user can quote to support — the body always has it; logs always tag it. Never expose stack traces.
