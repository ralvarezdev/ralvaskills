---
name: protobuf-architect
version: 1.0.0
description: >
  Enforces strict Protocol Buffers standards: proto3 syntax, Buf-style package
  naming (`<org>.<product>.<resource>.<version>`), `protovalidate` CEL-based
  validation, `buf` toolchain for lint / breaking-change detection / code
  generation, field-number reservation discipline, and well-known-type
  conventions. Language-agnostic at the schema level; code generation per
  target language. Use when designing or reviewing `.proto` files, evolving
  an existing schema, configuring `buf.gen.yaml` / `buf.yaml`, or auditing a
  proto module for breaking changes.
---

# Protocol Buffers Architecture

Language-agnostic schema standards. Proto files target **proto3** + `buf` toolchain. Validation uses `protovalidate` (CEL). Code generation is per-language and out of scope here — see the language architect skill for the generated-code idioms (e.g. [go-architect](../../languages/go-architect/SKILL.md) when generating for Go). See [STACK.md](STACK.md) for pinned tool versions.

## 1. File layout & package naming

Buf-style hierarchical packages — `<org>.<product>.<resource>.<version>`. Mirrors the on-disk path. One service or coherent message group per file.

```
proto/
└── acme/
    └── shop/
        └── users/
            └── v1/
                ├── users.proto        # service + request/response messages
                ├── user.proto         # the User resource message + enums
                └── events.proto       # domain events emitted by this resource
```

```proto
// acme/shop/users/v1/users.proto
syntax = "proto3";

package acme.shop.users.v1;

option go_package = "github.com/acme/shop/gen/go/acme/shop/users/v1;usersv1";
option java_package = "com.acme.shop.users.v1";
option csharp_namespace = "Acme.Shop.Users.V1";

import "acme/shop/users/v1/user.proto";
import "google/protobuf/timestamp.proto";
import "buf/validate/validate.proto";
```

- **Package = directory path.** `acme/shop/users/v1/` → `package acme.shop.users.v1;`. Enforced by `buf lint`.
- **Per-language `option *_package`** declarations live in every file. Set them up-front; don't refactor later (it's a breaking change).
- **One `service`** definition per file; supporting messages can live alongside or in sibling files.

## 2. Message design

- **All fields are `optional` semantically in proto3.** Use the explicit `optional` keyword (proto3.15+) when "field absent" must be distinguishable from "field present with zero value" — common for PATCH-style partial updates.
- **Field names: `snake_case`.** Enforced by `buf lint`. The language plugin converts to the target idiom (Go `CamelCase`, Python `snake_case`, etc.).
- **Message names: `PascalCase` singular.** `User`, not `Users` (the collection is `repeated User`).
- **Enum names: `PascalCase`.** Enum *value* names: `UPPER_SNAKE_CASE` **prefixed with the enum name** to keep them unambiguous across imports:
  ```proto
  enum OrderStatus {
    ORDER_STATUS_UNSPECIFIED = 0;
    ORDER_STATUS_PENDING = 1;
    ORDER_STATUS_PAID = 2;
    ORDER_STATUS_CANCELLED = 3;
  }
  ```
  - **`*_UNSPECIFIED = 0` is mandatory** — proto3 default is 0; making the zero value an explicit "unspecified" sentinel prevents accidental defaults from being mistaken for real states.
- **No primitive wrappers (`google.protobuf.StringValue` etc.) at API boundaries.** Use `optional` instead — it's the modern equivalent and doesn't require importing `wrappers.proto`.
- **Avoid `map<K, V>` when ordering or evolution matters.** `repeated KeyValue` with explicit `key`/`value` fields gives you ordering, validation, and the ability to add metadata per entry later.

## 3. Field numbering & reservation discipline

- **Field numbers 1–15 use 1 byte on the wire; 16–2047 use 2 bytes.** Reserve 1–15 for fields read on every request (IDs, status, frequently-accessed metadata).
- **Never reuse a field number.** When deleting a field, `reserved` it forever:
  ```proto
  message User {
    reserved 4, 7 to 9;
    reserved "deprecated_email", "legacy_role";

    string id = 1;
    string email = 2;
    ...
  }
  ```
- **Never change a field's type.** Changing `int32 age = 5;` to `string age = 5;` is a breaking change at the wire level. Delete + reserve + introduce a new field with a new number instead.
- **`buf breaking` enforces both rules** in CI (see §8).

## 4. Versioning

- **Version is part of the package path** (`acme.shop.users.v1`), never appended to message names (`UserV1` is wrong).
- **A new major version = a new `vN` package + a new directory.** `v1/` and `v2/` coexist; clients migrate at their pace.
- **Additive changes stay in the same version.** Adding a new optional field, a new enum value, or a new service method does not break clients — no version bump needed.
- **Breaking changes always go to a new major version.** Field removal, type change, semantic change, required-field tightening.
- **Deprecate before remove.** Add `[deprecated = true]` on the field; communicate via `Deprecation` headers when wrapped in REST; remove only in the next major version after migration is complete.

## 5. Validation — `protovalidate` (CEL)

Buf's `protovalidate` is the modern, declarative replacement for the older `protoc-gen-validate` (PGV). Constraints are CEL expressions attached as field options; validation runs at runtime via per-language `protovalidate` libraries.

```proto
import "buf/validate/validate.proto";

message CreateUserRequest {
  string email = 1 [(buf.validate.field).string = {
    email: true,
    max_len: 254
  }];
  string password = 2 [(buf.validate.field).string = {
    min_len: 12,
    max_len: 1024
  }];
  int32 age = 3 [(buf.validate.field).int32.gte = 0];
  repeated string tags = 4 [(buf.validate.field).repeated = {
    max_items: 32,
    items: { string: { min_len: 1, max_len: 64 } }
  }];
}
```

- **Validate at the boundary** — server-side, in the gRPC interceptor (see [grpc-architect §4](../../protocols/grpc-architect/SKILL.md#4-interceptors)). Don't rely on clients to validate.
- **CEL allows cross-field rules:**
  ```proto
  option (buf.validate.message).cel = {
    id: "password_confirmation_matches",
    message: "passwords must match",
    expression: "this.password == this.password_confirmation"
  };
  ```
- **Skip the old `protoc-gen-validate` (PGV).** It's deprecated; protovalidate's CEL is more expressive and the runtime is maintained.

## 6. Code generation — `buf generate` + `buf.gen.yaml`

One config file at the repo root drives all generators. Versions of the plugins are pinned in `buf.gen.yaml` itself; no `protoc` invocation by hand.

```yaml
# buf.gen.yaml
version: v2
managed:
  enabled: true
  override:
    - file_option: go_package_prefix
      value: github.com/acme/shop/gen/go
plugins:
  - remote: buf.build/protocolbuffers/go:v1.36.11
    out: gen/go
    opt: paths=source_relative
  - remote: buf.build/grpc/go:v1.6.2
    out: gen/go
    opt:
      - paths=source_relative
      - require_unimplemented_servers=false
  - remote: buf.build/bufbuild/validate-go:v1.2.0
    out: gen/go
    opt: paths=source_relative
```

```bash
buf generate                     # regenerate all targets
```

- **Generated code is committed** in monorepos and in any repo whose consumers don't run `buf` themselves. Commit it under `gen/<lang>/...` clearly separated from hand-written source.
- **Generated code is never edited** by hand. If you need to add helpers, write them in sibling hand-written files.
- **Generator plugin versions are pinned** in `buf.gen.yaml` — never `:latest`.

## 7. Linting — `buf lint`

`buf lint` runs the `STANDARD` rule set out of the box (proto3 conventions, naming, file structure). Add to CI.

```yaml
# buf.yaml
version: v2
modules:
  - path: proto
lint:
  use:
    - STANDARD
breaking:
  use:
    - FILE
```

Common lint catches:

- File package doesn't match directory path
- Field name isn't `snake_case`
- Enum value not prefixed with enum name
- Missing `*_UNSPECIFIED = 0`
- Missing per-language `option *_package`

## 8. Breaking-change detection — `buf breaking` in CI

```bash
buf breaking --against '.git#branch=main,subdir=proto'
```

`FILE`-level rules (the default) treat the file as the unit of compatibility — appropriate when generated code is consumed file-by-file. Stricter `PACKAGE` and `WIRE` modes exist; pick `FILE` unless you have a reason to differ.

- **Run on every PR** — block merges that introduce breaking changes without an intentional new `vN` package.
- **The check compares against the main branch state.** Workflows that want to compare against a tagged release can pass `--against '.git#tag=v1.5.0,subdir=proto'`.

## 9. Well-known types

Use the Google-provided well-known types instead of inventing equivalents.

| Use case | Type |
|---|---|
| Point-in-time timestamp | `google.protobuf.Timestamp` (RFC 3339, nanosecond precision) |
| Duration | `google.protobuf.Duration` |
| Wall-clock date (no time) | `google.type.Date` (from `googleapis`) |
| Money | `google.type.Money` (from `googleapis`) — currency code + amount |
| Unstructured JSON | `google.protobuf.Struct` (only for genuinely schemaless payloads) |
| Empty response | `google.protobuf.Empty` |
| Optional field marker | proto3 `optional` keyword (no wrapper needed) |
| Field mask for partial updates | `google.protobuf.FieldMask` |

- **Always import via the canonical path.** Don't copy these into your own packages.
- **`Timestamp` over `int64 unix_ts`** for any human-relevant time. Wire size is similar; human readability wins.
- **`Struct` is a hint of bad design.** Schemaless data in a typed schema is a smell — model it relationally if you can. Same opinion as [sql-architect §10](../../databases/sql-architect/SKILL.md) on JSONB.

## 10. Common types / shared messages

When two services need the same domain concept (`Money`, `Address`, `CustomerId`), put it in a shared sub-package:

```
proto/
└── acme/
    └── shop/
        ├── types/                    # shared types
        │   └── v1/
        │       ├── money.proto
        │       └── address.proto
        ├── users/v1/
        └── orders/v1/
```

- **`types/v1/`** is its own versioned package. It evolves on the same major-version contract as anything else.
- **Avoid premature sharing.** Two services with similar-looking `Address` messages might genuinely have different semantics; force the sharing only when the model is provably the same.
- **Never put services in the shared package.** Only messages.
