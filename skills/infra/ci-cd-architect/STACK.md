# Stack Versions

Lean stack — this skill is principles-first; the rows below track the third-party actions and tools referenced in [RECIPES.md](RECIPES.md). `rsk status --stack` uses this to surface drift when an action major bump changes behavior.

| Dependency | Pinned version | Purpose |
|---|---|---|
| github-actions runner | ubuntu-24.04 | Default Linux runner image referenced in recipes |
| actions/checkout | v5 | Source checkout (first-party) |
| actions/setup-go | v5 | Go toolchain + module/build cache |
| actions/setup-python | v5 | Python toolchain + pip cache |
| actions/setup-node | v4 | Node toolchain + npm/pnpm/yarn cache |
| astral-sh/setup-uv | v3 | uv install + cache for Python projects |
| docker/setup-buildx-action | v3 | BuildKit builder for multi-arch / advanced cache |
| docker/setup-qemu-action | v3 | QEMU for cross-arch emulation (arm64 on amd64 runners) |
| docker/build-push-action | v6 | Build + push + SBOM + provenance attestations |
| docker/login-action | v3 | Registry login (ghcr.io, Docker Hub, etc.) |
| docker/metadata-action | v5 | Tag + label generation from refs |
| aws-actions/configure-aws-credentials | v4 | OIDC role assumption — AWS |
| google-github-actions/auth | v2 | OIDC workload identity — GCP |
| googleapis/release-please-action | v4 | Manifest-driven release PRs + tags |
| aquasecurity/trivy-action | 0.28 | Image + filesystem vulnerability scanning |
| gitleaks/gitleaks-action | v2 | Secret scanning |
| sigstore/cosign-installer | v3 | Image signing with OIDC (no key material) |
| golangci/golangci-lint-action | v6 | Go lint aggregator (v2 of the lint binary) |
| gh CLI | 2.62 | Manual workflow dispatch + release scripting |

## Notes

- **Principles-first skill.** SKILL.md is platform-agnostic; STACK.md tracks only the actions/tools cited in RECIPES.md so `rsk status --stack` can surface drift.
- **First-party actions (`actions/*`, `docker/*`)** pinned by major tag (`@v5`) in recipes — a common compromise. SHA-pin if your threat model warrants it.
- **Third-party actions** SHOULD be SHA-pinned in real workflows; the recipe examples show major tags for readability with a note pointing to §4 for the Renovate-managed SHA pattern.
- **Cosign + OIDC** is the default signing path — no long-lived keys checked in.
- Image base / runtime versions belong to [docker-architect/STACK.md](../docker-architect/STACK.md); not duplicated here.

_Last reviewed: 2026-05-22_
_Skill version at last review: 1.0.0_
