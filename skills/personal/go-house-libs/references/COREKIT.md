# corekit

`github.com/ralvarezdev/corekit` — the transport-neutral error vocabulary the house Go kits share, plus the domain-agnostic conversions in `xconv`. A domain classifies its failures once (`domain.ErrNotFound -> KindNotFound`); each transport renders the same typed error through its own table: an HTTP status for a REST problem, a `codes.Code` in [GRPCKIT](GRPCKIT.md), a `connect.Code` in [CONNECTRPCKIT](CONNECTRPCKIT.md), a safe message for an MCP tool error. The root is stdlib-only and imports no HTTP, gRPC, Connect or MCP package. Current tag v0.3.0. Reach for it before hand-writing a per-transport error switch or a project-local error code.

## Packages

| Package | Owns |
|---|---|
| `corekit` | `Kind` + `Class`, `ClassOf`, the neutral `Issue`/`Severity`, the typed `Error` with `errors.Is` sentinels, `New`/`Invalid`, `KindOf`/`IssuesOf`, and the `Classifier` (`Rule`, `IssueCarrier`, `Wrap`) |
| `corekit/xconv` | domain-agnostic conversions: `MapSlice`, `Deref`, optional string/bool pointers, date/decimal parsing. The only subpackage with a dependency (`shopspring/decimal`) |

The `Kind` set (zero value `KindUnknown`, so an unclassified failure is safe by default), with the class each maps to:

| Kind | Class: `Name`, HTTP status, default message |
|---|---|
| `KindNotFound` | `not_found`, 404, "not found" |
| `KindInvalid` | `invalid`, 422, "validation failed" |
| `KindConflict` | `conflict`, 409, "already exists" |
| `KindForbidden` | `forbidden`, 403, "forbidden" |
| `KindUnauthenticated` | `unauthenticated`, 401, "unauthenticated" |
| `KindUnavailable` | `unavailable`, 503, "a dependency is unavailable" (`Retryable`) |
| `KindPrecondition` | `precondition_failed`, 412, "the resource changed since it was read" |
| `KindUnknown` | `unknown`, 500, "internal error" |

`ClassOf(k)` returns the `Class` and never panics — an unknown or out-of-range Kind gives the `KindUnknown` class. `Kind.HTTPStatus()`, `DefaultMessage()`, `Retryable()` and `String()` read it back.

## Recipe

```go
// The domain classifies its sentinels once, at the service boundary.
classify := corekit.NewClassifier(
    corekit.Rule{Is: movement.ErrNotFound, Kind: corekit.KindNotFound},
    corekit.Rule{Is: movement.ErrVersionConflict, Kind: corekit.KindConflict},
)

// Handlers return *corekit.Error; the transport maps it to a code.
err := classify.Wrap(store.Update(ctx, m))
if err != nil {
    return nil, err
}

// Read the classification back anywhere; both work through wrapping.
if kind, ok := corekit.KindOf(err); ok && kind == corekit.KindNotFound { /* ... */ }
if typed, ok := errors.AsType[*corekit.Error](err); ok { _ = typed.Issues }
```

A validation failure carries neutral `Issue`s (`Pointer` is an RFC 6901 JSON Pointer), and a domain type exposing `ValidationIssues() []corekit.Issue` is promoted by `Wrap`:

```go
err := corekit.Invalid([]corekit.Issue{ // nil when no error-severity issue
    {Pointer: "/amount", Code: "required", Message: "amount is required", Severity: corekit.SeverityError},
})
```

The Kind → code table lives with each transport, never here:

```go
grpckit.KindCode(kind)        // codes.Code
connectrpckit.KindCode(kind)  // connect.Code
corekit.ClassOf(kind).HTTPStatus // 404, 422, 409, ...

// install the transport's mapper so handlers can just return *corekit.Error
grpckit.NewServer(grpckit.ServerConfig{MapError: grpckit.KindErrorMapper(logger)})
connectrpckit.NewServer(connectrpckit.ServerConfig{MapError: connectrpckit.KindErrorMapper(logger)})
```

## Rules

- **Classify once, in the domain.** `New(k, err)` or `Classifier.Wrap` at the service boundary; a transport never inspects domain sentinels.
- **`KindUnknown` never goes on the wire.** It is the safe zero value; a mapper logs the real error and answers an internal failure (grpckit/connectrpckit `KindErrorMapper` do exactly this).
- **`errors.Is`/`errors.As` work through wrapping.** `Error.Is` matches a sentinel of the same Kind (`errors.Is(err, corekit.ErrNotFound)`) whatever cause it carries; `errors.AsType[*corekit.Error]` reaches `Kind` and `Issues`.
- **`KindConflict` is not `KindPrecondition`.** 409 is a duplicate/unique clash; 412 / `FailedPrecondition` is a failed optimistic-concurrency precondition (the resource changed since it was read).
- **`Issue` is the one validation shape.** `Pointer` is an RFC 6901 JSON Pointer, `SeverityError` (the zero value) invalidates and `SeverityWarning` does not; transports send only error-severity violations. `HasErrors` is the check.
- **`Retryable` marks `KindUnavailable` only**, the failures a read-only caller may retry.
- **`Wrap` returns unclassified errors unchanged**, so a caller keeps logging them as unexpected; only a matched `Rule`, an `IssueCarrier`, or an already-typed error becomes a `*corekit.Error`.
- **The root stays stdlib-only.** Anything with a dependency goes in a subpackage (`xconv` is the only one).

## Not here

The transport tables and their `KindErrorMapper` (grpckit's `codes.Code`, connectrpckit's `connect.Code`), HTTP problem details ([RESTKIT](RESTKIT.md) `problem`, [GINKIT](GINKIT.md)), MCP tool errors ([MCPKIT](MCPKIT.md)), and any domain sentinel (the domain owns those). corekit owns the vocabulary the transports key on. See [../STACK.md](../STACK.md) for the tag.
