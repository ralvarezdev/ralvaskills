# Stack Versions

This skill pins **specifications**, not libraries. Library choices live in the framework-architect skills that implement these conventions.

| Spec / Standard | Version | Purpose |
|---|---|---|
| HTTP/1.1 + HTTP/2 + HTTP/3 | RFC 9110 / 9112 / 9113 / 9114 | Transport and semantics |
| OpenAPI | 3.1 | API contract format |
| JSON | RFC 8259 | Body encoding |
| JSON Merge Patch | RFC 7396 | `PATCH` body format |
| Problem Details for HTTP APIs | RFC 7807 | Error response shape |
| Idempotency-Key Header Field | RFC 9457 (draft) | `Idempotency-Key` semantics |
| Conditional Requests (ETag, If-Match) | RFC 9110 §13 | Optimistic concurrency |
| Deprecation HTTP Header | RFC 9745 | Deprecation signaling |
| Sunset HTTP Header | RFC 8594 | Removal date signaling |
| RateLimit Header Fields | RFC 9331 (draft) | Rate-limit response headers |
| ISO 8601 / RFC 3339 | RFC 3339 | Timestamp encoding |
| UUID v7 | RFC 9562 | Resource ID format (sortable, distributed) |

## Notes

- **Conventions:**
  - JSON field casing: **`snake_case`**
  - Timestamps: **ISO 8601 with timezone**, string-encoded
  - Money: string-encoded decimals (`"99.99"`)
  - IDs: UUID v7 canonical hex with dashes
  - Versioning: **URL prefix** (`/v1/...`)
  - Pagination: **cursor**, not offset
  - Errors: **RFC 7807** problem-details
  - Idempotency: **`Idempotency-Key` mandatory** on POST/PATCH
  - Concurrency: **`ETag` + `If-Match` mandatory** on PUT/PATCH
- **No library pinning here.** For Python implementations see [fastapi-architect/STACK.md](../../frameworks/fastapi-architect/STACK.md); for Go see (planned) `gin-architect/STACK.md`.

_Last reviewed: 2026-05-20_
_Skill version at last review: 1.0.0_
