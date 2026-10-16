# contracts

`github.com/ralvarezdev/contracts` — the platform's protobuf contract module: the common message shapes every service and app speaks (cursor pagination, machine-readable error details, exact money, calendar dates) plus the APIs of the shared platform services (identity, notifier), kept in one reviewed place instead of drifting across repos. It is a Buf proto module and a Go module at once; generated Go (messages, gRPC and Connect stubs) is committed and imported directly, so there is no registry. Reach for it before re-declaring a `PageRequest`, an `ErrorDetail` or a `Money` in your own proto, and to get a platform service's client stubs without importing the service module.

## Packages

| Path | Owns |
|---|---|
| `proto/ralvarezdev/common/v1` | the shared value types, package `ralvarezdev.common.v1`, linted with buf `STANDARD`, `breaking` is `FILE` |
| `proto/ralvarezdev/<service>/v1` | a platform service's API (`identity`, `notifier`) |
| `gen/go/ralvarezdev/<name>/v1` | committed generated Go, import alias `commonv1`, `identityv1`, ...; `<name>v1connect` holds the Connect stubs (`task proto:generate`, remote plugins) |

| Message | Use |
|---|---|
| `PageRequest` (`page_size`, `cursor`) / `PageInfo` (`next_cursor`) | opaque cursor pagination for any list RPC; offset pagination is deliberately unsupported |
| `ErrorDetail` (`code`, `message`, `correlation_id`, `field_violations`) / `FieldViolation` (`field`, `description`) | the machine-readable failure body, shared across gRPC, Connect and REST (REST maps it onto RFC 9457 `application/problem+json`) |
| `Money` (`currency`, `amount`) | exact money: ISO 4217 code plus a decimal amount string, never a float |
| `Date` (`year`, `month`, `day`) | a calendar day with no time zone |

Instants and spans are not redefined: use `google.protobuf.Timestamp` and `google.protobuf.Duration`. The module is a leaf — its Go side imports only `google.golang.org/protobuf`.

## Recipe

```go
import commonv1 "github.com/ralvarezdev/contracts/gen/go/ralvarezdev/common/v1"

// server: read the page request, answer a page info
in := &commonv1.PageRequest{PageSize: req.GetPageSize(), Cursor: req.GetCursor()}
out := &commonv1.PageInfo{NextCursor: next} // "" on the last page

// one failure body across transports
detail := &commonv1.ErrorDetail{
    Code:            "identity.pool.invalid_name",
    Message:         "invalid pool name",
    CorrelationId:   requestID,
    FieldViolations: []*commonv1.FieldViolation{{Field: "pool.name", Description: "required"}},
}
```

Consume the messages from your own proto:

```proto
import "ralvarezdev/common/v1/pagination.proto";

message ListPoolsRequest  { ralvarezdev.common.v1.PageRequest page = 1; }
message ListPoolsResponse { repeated Pool pools = 1; ralvarezdev.common.v1.PageInfo page = 2; }
```

Regenerate with `task proto:generate` (Buf, remote plugins; needs network) and commit `gen/`. `task proto:check` fails when the committed `gen/` is stale; `task proto:breaking` runs `buf breaking`.

## Rules

- **Generated code is generated and committed.** Never hand-edit `gen/**`; edit `proto/` and run `task proto:generate`. CI gates drift with `task proto:check`.
- **Consume it as a Go module, not through a registry.** `go get github.com/ralvarezdev/contracts@<tag>` with `GOPRIVATE=github.com/ralvarezdev/*`; a consumer's own `.proto` never needs the source, because the generated packages are imported directly. Never point a consumer at an untagged version.
- **A contract lands here when a platform service owns it or two or more repos consume it.** App-private contracts stay with the app; do not add one app's shapes here.
- **Field numbers are reserved, never reused** (`buf breaking` `FILE`): a removal or renumber is a breaking change even when the wire format still decodes.
- **Cursors are opaque.** A client echoes `PageInfo.next_cursor` back and never parses or builds one; there is no offset variant, because offset drifts under concurrent writes.
- **`Money.amount` is a decimal string, not a float** — no exponent notation and no grouping separators; a leading sign is allowed. Binary floating point cannot represent decimal money exactly.
- **`ErrorDetail.code` is the contract**, a stable dotted `lower_snake` code owned by the producer; clients branch on it, never on the human-readable `message`.
- **Timestamps are well-known types.** `Date` means a calendar day only; an instant is `google.protobuf.Timestamp` and a span is `google.protobuf.Duration`.

## Not here

Generated Go clients or servers, the REST edge (cursors are issued by [RESTKIT](RESTKIT.md) `EncodeCursor`), domain messages, and the `Timestamp`/`Duration` well-known types. See [../STACK.md](../STACK.md) for the tag.
