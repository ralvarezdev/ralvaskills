# Stack Versions

This skill pins the **language-agnostic** proto toolchain. Per-language code-generation plugins are pinned in your repo's `buf.gen.yaml`; the versions below are sane current defaults.

| Dependency | Pinned version | Purpose |
|---|---|---|
| proto syntax | proto3 | Schema language |
| buf CLI | 1.69 | Lint, breaking-change detection, code generation, format |
| `buf.yaml` schema | v2 | Module config format |
| `buf.gen.yaml` schema | v2 | Generation config format |
| protovalidate spec | 1.2 | CEL-based validation constraints (`buf/validate/validate.proto`) |
| protovalidate-go (Go runtime) | 1.2 | Runtime validation for Go-generated code |
| protoc-gen-go (plugin) | 1.36.11 | Go message generator (referenced from `buf.gen.yaml`) |
| protoc-gen-go-grpc (plugin) | 1.6.2 | Go gRPC service generator |
| google.golang.org/protobuf | 1.36 | Go runtime — see [go-architect](../../languages/go-architect/STACK.md) |

## Notes

- **No raw `protoc` invocations.** All generation goes through `buf generate` and `buf.gen.yaml`. The `protoc` binary is not pinned because `buf` orchestrates plugins directly.
- **Plugin versions** for code generation are pinned in `buf.gen.yaml` per-repo. The versions above are sane defaults; specific projects may pin different versions to match their language runtime.
- **`protovalidate` replaces the deprecated `protoc-gen-validate` (PGV).** CEL expressions are more expressive and the runtime is actively maintained.
- **Buf Schema Registry (BSR)** is the upgrade path when you publish schemas to external consumers. Skip until you genuinely need centralized schema distribution.
- **Per-language runtime libraries** (e.g. `protovalidate-go`, `protovalidate-python`, `protovalidate-java`) are pinned in the language-architect STACK or the consuming project's manifest.

_Last reviewed: 2026-05-21_
_Skill version at last review: 1.0.0_
