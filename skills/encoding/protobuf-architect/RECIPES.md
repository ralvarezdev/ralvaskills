# Protocol Buffers Recipes

Reference snippets for the conventions in [SKILL.md](SKILL.md).

## File layout & package declaration

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

## Enum with mandatory UNSPECIFIED

```proto
enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  ORDER_STATUS_PENDING = 1;
  ORDER_STATUS_PAID = 2;
  ORDER_STATUS_CANCELLED = 3;
}
```

## Field-number reservation

```proto
message User {
  reserved 4, 7 to 9;
  reserved "deprecated_email", "legacy_role";

  string id = 1;
  string email = 2;
}
```

## Validation — `protovalidate` (CEL)

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

Cross-field rules via CEL:

```proto
option (buf.validate.message).cel = {
  id: "password_confirmation_matches",
  message: "passwords must match",
  expression: "this.password == this.password_confirmation"
};
```

## Code generation — `buf.gen.yaml`

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
buf generate
```

## Lint + breaking-change config — `buf.yaml`

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

```bash
buf breaking --against '.git#branch=main,subdir=proto'
```

## Shared types sub-package

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
