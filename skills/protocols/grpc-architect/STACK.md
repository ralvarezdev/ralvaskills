# Stack Versions

This skill pins **protocol-level** versions (gRPC spec, codes) and **Go-specific** runtime libraries. Other languages (Python `grpcio`, Java, Rust, etc.) follow the same protocol-level conventions with their own runtimes; pin those in the consuming project's manifest.

| Dependency | Pinned version | Purpose |
|---|---|---|
| gRPC protocol | HTTP/2 + Protobuf wire format | Spec |
| google.golang.org/grpc | 1.81 | Go runtime — see [go-architect](../../languages/go-architect/STACK.md) |
| protoc-gen-go-grpc | 1.6.2 | Go service stub generator (referenced from `buf.gen.yaml`) |
| buf.build/go/protovalidate (Go) | 1.2 | Server-side request validation via interceptor |
| google.golang.org/grpc/cmd/protoc-gen-go-grpc | 1.6.2 | gRPC code generation plugin |
| grpcurl | 1.9 | Command-line client for ad-hoc requests + reflection debugging |
| grpc-health-protocol | v1 | `grpc.health.v1.Health` service — always on |

## Notes

- **Vanilla gRPC, not Connect-RPC.** Browser callers and gRPC-Web are out of scope here. If browser support becomes a requirement, [Connect-RPC](https://connectrpc.com) is the documented upgrade path (same `.proto` files, additional generator).
- **Protocol design comes from [protobuf-architect](../../encoding/protobuf-architect/SKILL.md).** This skill targets the runtime behavior of services that already follow proto3 + Buf conventions.
- **Validation library:** `protovalidate-go` server-side, via a unary/stream interceptor. Don't hand-write validation in handlers; the field constraints live in `.proto` files (see [protobuf-architect §5](../../encoding/protobuf-architect/SKILL.md#5-validation--protovalidate-cel)).
- **`grpc-gateway` is out of scope** by default. Connect-RPC is the cleaner answer when REST clients also need to talk to the same service.
- **Per-language runtimes** (Python `grpcio`, Java `grpc-java`, Rust `tonic`) follow the same protocol conventions; pin those in the consuming project's `pyproject.toml` / `pom.xml` / `Cargo.toml`.

_Last reviewed: 2026-05-21_
_Skill version at last review: 1.0.0_
