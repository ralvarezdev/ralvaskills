---
name: grpc-architect
version: 1.0.0
description: >
  Enforces strict vanilla-gRPC standards: service definitions in `.proto` files,
  `status.Error` with standard gRPC codes, domain-error → code mapping, mandatory
  interceptor chain (auth, logging, recovery, validation, metrics), client-side
  deadlines, context propagation, reflection on in dev / off in prod, and
  in-process testing via `bufconn`. Language-agnostic at the protocol level;
  Go-specific implementation examples. Use when designing or reviewing a gRPC
  service, writing handlers, configuring interceptors, debugging streaming
  patterns, or choosing gRPC over REST.
---

# gRPC Architecture

Vanilla gRPC over HTTP/2 — backend-to-backend services. Pair with [protobuf-architect](../../encoding/protobuf-architect/SKILL.md) for schema design. Go-specific implementation examples reference [go-architect](../../languages/go-architect/SKILL.md); other languages follow the same protocol-level conventions with idiomatic substitutions. See [STACK.md](STACK.md) for pinned dependencies.

## 1. Service definition

One service per file, named after the resource. Methods are verb-noun, request and response are always typed messages — never raw primitives or `google.protobuf.Empty` as input.

```proto
// acme/shop/users/v1/users.proto
service UserService {
  rpc GetUser(GetUserRequest) returns (User);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);   // cursor-paginated
  rpc UpdateUser(UpdateUserRequest) returns (User);
  rpc DeleteUser(DeleteUserRequest) returns (google.protobuf.Empty);
  rpc WatchUsers(WatchUsersRequest) returns (stream UserEvent);  // server-stream
}
```

- **`<Verb><Noun>Request` / `<Verb><Noun>Response`** naming for every method's I/O message. Even when the response is a single resource, prefer `CreateUserResponse { User user = 1; }` over returning `User` directly — leaves room to add fields without bumping the major version.
- **Pagination via cursor**, mirroring [rest-api-architect §5](rest-api-architect/SKILL.md#5-pagination--cursor-not-offset). `ListUsersRequest { string cursor = 1; int32 limit = 2; }` → `ListUsersResponse { repeated User users = 1; string next_cursor = 2; }`.
- **`google.protobuf.Empty`** only as a response type for "fire and forget" actions with no useful return. Never as input.

## 2. Error handling — `status.Error` with codes

Use gRPC standard codes, return errors via `status.Error(code, msg)`. Map domain errors to codes in one central place.

### Standard codes — the ones architects actually use

| Code | Use for |
|---|---|
| `OK` | success |
| `INVALID_ARGUMENT` | request fails schema or business validation |
| `FAILED_PRECONDITION` | request valid but system state forbids it (e.g. delete a non-empty resource) |
| `OUT_OF_RANGE` | numeric / range-specific violation distinct from `INVALID_ARGUMENT` |
| `UNAUTHENTICATED` | missing or invalid credentials |
| `PERMISSION_DENIED` | authenticated but not authorized |
| `NOT_FOUND` | resource doesn't exist |
| `ALREADY_EXISTS` | unique-constraint or idempotency-key collision (with a different body) |
| `ABORTED` | concurrency conflict (ETag-equivalent) — client should re-fetch and retry |
| `RESOURCE_EXHAUSTED` | rate limit; per-tenant quota |
| `DEADLINE_EXCEEDED` | the request didn't complete in time — set by the runtime |
| `UNAVAILABLE` | transient — load-balancer drained, restart in progress; client retries |
| `INTERNAL` | unexpected server-side failure — bug or external dependency error |
| `UNIMPLEMENTED` | method exists in proto but server doesn't handle it (use during rollout) |

**Don't reach for `INTERNAL` as a default.** Map every known domain error to a specific code.

### Domain → code mapping

```go
// internal/errors/grpc.go
func toGRPC(err error) error {
    switch {
    case errors.Is(err, users.ErrNotFound):
        return status.Error(codes.NotFound, err.Error())
    case errors.Is(err, users.ErrEmailExists):
        return status.Error(codes.AlreadyExists, err.Error())
    case errors.Is(err, auth.ErrInvalidToken):
        return status.Error(codes.Unauthenticated, "invalid token")
    case errors.Is(err, auth.ErrForbidden):
        return status.Error(codes.PermissionDenied, "")
    case errors.Is(err, db.ErrStale):
        return status.Error(codes.Aborted, "resource was modified")
    default:
        slog.Error("unexpected error", "err", err)
        return status.Error(codes.Internal, "internal error")  // never leak details
    }
}
```

- **`status.WithDetails`** attaches structured detail (`google.rpc.ErrorInfo`, `google.rpc.BadRequest`) when clients need machine-readable error context — equivalent to REST's RFC 7807 (see [rest-api-architect §7](../../protocols/rest-api-architect/SKILL.md#7-error-contracts--rfc-7807-problem-details)). Always include a correlation id.
- **Never leak stack traces or DB errors** to clients. Log server-side; return a generic `INTERNAL` with the correlation id.

## 3. Request / response shape

- **Always typed messages.** Don't define a method as `rpc Ping(StringValue) returns (StringValue)` — wrap in `PingRequest` / `PingResponse`.
- **Validation at the boundary** via `protovalidate` (see [protobuf-architect §5](../../encoding/protobuf-architect/SKILL.md#5-validation--protovalidate-cel)). Enforced server-side via an interceptor (§4).
- **No business logic in generated handler files.** Generated handlers are thin shims that call into service-layer code (same discipline as REST routers per [fastapi-architect](../../frameworks/fastapi-architect/SKILL.md) / [gin-architect](../../frameworks/gin-architect/SKILL.md)).

## 4. Interceptors — mandatory chain

Interceptors are gRPC's middleware. **Order matters** — outermost first.

```go
import (
    "google.golang.org/grpc"
    "buf.build/go/protovalidate"
    protovalidate_grpc "buf.build/go/protovalidate/grpc"
)

server := grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        recovery.UnaryServerInterceptor(),         // 1. panic → INTERNAL
        requestid.UnaryServerInterceptor(),        // 2. assign correlation id
        slogger.UnaryServerInterceptor(),          // 3. structured access log
        auth.UnaryServerInterceptor(authSvc),      // 4. authenticate (sets user in ctx)
        protovalidate_grpc.UnaryServerInterceptor( // 5. validate request per protovalidate
            mustNewValidator(),
        ),
        metrics.UnaryServerInterceptor(),          // 6. instrument latency/codes
    ),
    grpc.ChainStreamInterceptor(/* analogous stream interceptors */),
)
```

- **Recovery first** — catches panics anywhere downstream and converts to `INTERNAL` with correlation id (never a stack trace).
- **Auth before validation** — no point validating an unauthenticated request's body. Per-method authorization (scopes/roles) happens inside the handler or via a small `WithAuthFunc` interceptor.
- **Validation is centralized via `protovalidate`** — don't hand-write validation in every handler. The interceptor calls `validator.Validate(req)` and returns `INVALID_ARGUMENT` with `google.rpc.BadRequest` details on failure.
- **Same interceptor chain for streaming RPCs** via `ChainStreamInterceptor`. Streaming validation requires handling per-message in client/bidi streams.

## 5. Streaming patterns

gRPC supports four call types. Pick the simplest one that meets the requirement.

| Pattern | Use for | Pitfalls |
|---|---|---|
| **Unary** | Default — request/response | None — start here |
| **Server-stream** | Server emits N responses to one request (event feeds, paginated downloads that don't fit one response, log tail) | Connection state outlives the request; resume tokens needed for restarts |
| **Client-stream** | Client uploads N messages, server returns one summary (large uploads, batch ingest) | Backpressure from server requires careful flow control |
| **Bidi-stream** | Genuinely interactive (chat, collaborative editing, control protocols) | Connection lifecycle complexity; reconnect / resume logic; deadlines |

- **Don't reach for streaming "to save round-trips"** — unary with proper pagination is usually fine and orders of magnitude simpler.
- **Server-streams need resume tokens.** Pass `start_after_id` or a cursor in the request so a disconnected client can resume from a known point.
- **Bidi-streams need a clear protocol** — define exactly which side sends what and when. Sketch the message flow in the `.proto` comments; future-you will thank you.
- **Set generous deadlines** on streams — but **always** set them. An unbounded stream is a leak.

## 6. Deadlines & context propagation

Every gRPC call has a deadline. Clients set; servers respect; downstream calls inherit the remaining time.

- **Clients always set a deadline:**
  ```go
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  resp, err := client.GetUser(ctx, &pb.GetUserRequest{Id: id})
  ```
- **Server respects** — check `ctx.Err()` periodically in long-running handlers; abort work the moment the deadline fires.
- **Propagate context to all downstream calls** — DB queries, HTTP calls, other gRPC calls. Deadlines and cancellation flow automatically.
- **Default deadlines per call type:** unary 5–30s; server-stream often much longer (minutes / hours) but always bounded.
- **Server-side deadline guard:** wrap the entire handler in `context.WithTimeout` slightly less than the client deadline to leave headroom for response serialization.

## 7. Metadata vs message fields

| Use metadata for | Use message fields for |
|---|---|
| Auth tokens (`authorization: Bearer ...`) | Business data |
| Request IDs / correlation IDs (read by middleware) | Anything the handler reads as part of business logic |
| Tracing context (`traceparent`) | — |
| Rate-limit hints (`x-tenant-id` for routing) | — |

- **Metadata is HTTP/2 headers under the hood.** Don't send large payloads here.
- **Keys are case-insensitive ASCII;** values are strings (binary metadata uses the `-bin` suffix).
- **Standardize one correlation-id header** (e.g. `x-request-id`) — interceptor reads it on entry, injects into context, logs against it.

## 8. Reflection

```go
import "google.golang.org/grpc/reflection"

if cfg.Env != "prod" {
    reflection.Register(server)
}
```

- **Reflection enables `grpcurl`** and IDE plugins to introspect the service without the `.proto` file.
- **Off in production.** Leaks the entire service surface; useful only when debugging.
- **Health service** (`grpc.health.v1`) is always on — load balancers and orchestrators need it.

## 9. Testing — `bufconn` for in-process

The `google.golang.org/grpc/test/bufconn` package gives an in-memory listener — full server + client without a real socket. Faster than `net.Pipe`, simpler than spinning up a test server on a port.

```go
func TestGetUser(t *testing.T) {
    lis := bufconn.Listen(1024 * 1024)
    srv := grpc.NewServer(/* interceptors */)
    pb.RegisterUserServiceServer(srv, newTestHandler(t))
    go func() { _ = srv.Serve(lis) }()
    t.Cleanup(srv.Stop)

    conn, err := grpc.NewClient("passthrough://bufnet",
        grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) {
            return lis.Dial()
        }),
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    require.NoError(t, err)
    t.Cleanup(func() { _ = conn.Close() })

    client := pb.NewUserServiceClient(conn)
    resp, err := client.GetUser(t.Context(), &pb.GetUserRequest{Id: "test"})
    require.NoError(t, err)
    require.Equal(t, "test", resp.Id)
}
```

- **`bufconn` for unit + integration tests**; spin a real server on a random port only when you specifically need the full network path (TLS, HTTP/2 frame behavior, etc.).
- **Table-driven tests** per [go-architect §9](../../languages/go-architect/SKILL.md#9-testing) — one row per (input, expected code, expected error type).
- **`grpcurl`** is the manual-testing tool. Pin it in your task runner / mise config.

## 10. When to pick gRPC over REST

gRPC's wins are real but specific:

- **Backend-to-backend** — gRPC's binary framing and HTTP/2 multiplexing beat JSON-over-HTTP/1.1 in throughput and tail latency at scale.
- **Strong contracts** — `.proto` is the canonical schema; clients in any language are generated. No OpenAPI drift.
- **Streaming** — first-class server-stream / client-stream / bidi.
- **Compact wire format** — binary; smaller than JSON for the same payload.

REST is the better default when:

- **Browser callers** — vanilla gRPC isn't browser-callable without a gateway. Connect-RPC fixes this; consider it (or a separate REST facade) if browser is in scope.
- **Public APIs** — external consumers expect REST; `curl` works without tooling; OpenAPI is the universal documentation format.
- **Cache-friendly reads** — HTTP caching (ETag, Cache-Control) is built-in; gRPC has no equivalent.
- **Small scope** — for one CRUD service, a Gin/FastAPI REST API is shorter to build.

The two coexist: gRPC for east-west backend traffic, REST for north-south client-facing endpoints. If you need both for the same service, **Connect-RPC** lets one set of `.proto` files serve gRPC, gRPC-Web, and Connect (browser-friendly) — worth considering as the upgrade path rather than running gRPC + a separate REST service.

`grpc-gateway` (REST/JSON facade in front of a gRPC service from extra proto annotations) is an alternative, but adds spec complexity, deployment surface, and error-translation discipline. Reach for it only when you must serve both protocols from the same binary and Connect-RPC isn't a fit.
