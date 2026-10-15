# Stack Versions

This skill pins **protocol-level** versions (Connect / gRPC / gRPC-Web wire behavior) and **Go-specific** runtime libraries. Connect's `connect-go` is the reference implementation and the only one most projects need; other languages (TypeScript `@connectrpc/connect`, Python, Kotlin) follow the same protocol-level conventions with their own runtimes — pin those in the consuming project's manifest.

| Dependency | Pinned version | Purpose |
|---|---|---|
| Connect protocol | Connect + gRPC + gRPC-Web over HTTP/1.1 or 2 | Spec / wire formats served by one handler |
| connectrpc.com/connect (Go) | 1.19 | Handler, generic client, interceptors, error model — see [go-architect](../../languages/go-architect/STACK.md) |
| buf.build/connectrpc/go | Buf remote plugin | Connect handler generation driven by `buf.gen.yaml` |
| protoc-gen-connect-go | 1.19 | Standalone generator (same plugin, invoked without Buf) |
| connectrpc.com/grpcreflect | 1.3 | Server reflection for `grpcurl` / IDE introspection — dev only |
| connectrpc.com/grpchealth | 1.4 | `grpc.health.v1.Health` service — always on |
| connectrpc.com/validate | 0.4 | protovalidate interceptor for request validation |
| connectrpc.com/cors | 0.1 | CORS middleware for direct browser callers |
| connectrpc.com/otelconnect | 0.8 | OpenTelemetry instrumentation interceptor |
| buf | 1.71 | Lint, breaking, and code-generation driver |
| grpcurl | 1.9 | Ad-hoc client + reflection debugging (server must expose reflection) |

## Notes

- **Connect is not gRPC-only.** The generated handler speaks Connect's own POST/JSON+proto protocol, gRPC, and gRPC-Web simultaneously. Choose the protocol per caller at request time; no separate build or gateway.
- **The gRPC protocol is one of three, not the whole.** A native `*grpc.Server` adapter from the same `.proto` is kept as deliberate insurance and for pure-gRPC peers — see [grpc-architect](../grpc-architect/SKILL.md) and [GRPCKIT](../../personal/go-house-libs/references/GRPCKIT.md).
- **Browser streaming caveats.** Connect and gRPC carry server-streams to browsers; client- and bidi-streams require HTTP/2 and are not supported by gRPC-Web in the browser. Design browser-facing APIs as unary + server-stream.
- **Validation via protovalidate.** `connectrpc.com/validate` wraps the same `buf.build/go/protovalidate` rules that [protobuf-architect §5](../../encoding/protobuf-architect/SKILL.md#5-validation--protovalidate-cel) defines. Don't hand-write per-handler validation.
- **Reflection and health are separate modules** with their own release cadence; pin them independently of the core `connectrpc.com/connect` module.
- **Per-language runtimes** (TypeScript, Python, Kotlin) follow the same protocol conventions; pin those in the consuming project's `package.json` / `pyproject.toml` / `build.gradle`.
- **Cross-check [go-architect STACK](../../languages/go-architect/STACK.md)** before bumping the base Go toolchain or `google.golang.org/protobuf`; the Connect runtime tracks protobuf-go.

_Last reviewed: 2026-10-15_
_Skill version at last review: 1.0.0_
