# CI/CD — GitHub Actions Recipes

Concrete YAML for the principles in [SKILL.md](SKILL.md). Each section is anchored to a SKILL.md reference; jump in via the body links, not by reading top-to-bottom.

Pinned action / tool versions live in [STACK.md](STACK.md). Action SHAs in examples below are illustrative; treat them as placeholders to refresh via Renovate.

## Table of Contents

1. [Workflow file layout](#1-workflow-file-layout)
2. [Concurrency and triggers](#2-concurrency-and-triggers)
3. [Composite action — language setup](#3-composite-action--language-setup)
4. [Action pinning + Renovate](#4-action-pinning--renovate)
5. [OIDC cloud auth](#5-oidc-cloud-auth)
6. [Language caching](#6-language-caching)
7. [Matrix strategy](#7-matrix-strategy)
8. [Reusable workflow — CI gate](#8-reusable-workflow--ci-gate)
9. [Build & push image](#9-build--push-image)
10. [release-please](#10-release-please)
11. [Deploy with environment approval](#11-deploy-with-environment-approval)

---

## 1. Workflow file layout

```
.github/
├── workflows/
│   ├── ci.yaml              # on: pull_request, push to main; lint + test + build
│   ├── release.yaml         # on: push tag v*; build + sign + publish artifacts
│   ├── deploy.yaml          # on: workflow_dispatch or release; promote artifact
│   ├── scheduled-scan.yaml  # on: schedule (cron); trivy + gitleaks + dep-review
│   └── release-please.yaml  # on: push to main; manage version PR + tag
├── actions/                 # composite actions, one folder per action
│   └── setup-go-cached/
│       └── action.yaml
└── dependabot.yaml          # or renovate.json at repo root
```

One concern per workflow. Keep job names stable — branch-protection rules reference them by name.

---

## 2. Concurrency and triggers

Cancel stale PR runs; queue main and release runs.

```yaml
name: ci
on:
  pull_request:
    paths-ignore: ['docs/**', '**.md']
  push:
    branches: [main]
    paths-ignore: ['docs/**', '**.md']

concurrency:
  group: ci-${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: ${{ github.event_name == 'pull_request' }}

permissions:
  contents: read
```

```yaml
name: scheduled-scan
on:
  schedule:
    - cron: '17 4 * * 1'   # Monday 04:17 UTC — off-peak, off-hour
  workflow_dispatch:        # always allow manual rerun

concurrency:
  group: scheduled-scan
  cancel-in-progress: false
```

---

## 3. Composite action — language setup

Factor out common setup steps when ≥3 jobs share them.

```yaml
# .github/actions/setup-go-cached/action.yaml
name: Setup Go with cache
description: Installs Go, restores module + build caches keyed on go.sum.
inputs:
  go-version-file:
    description: Path to go.mod
    default: go.mod
runs:
  using: composite
  steps:
    - uses: actions/setup-go@v5
      with:
        go-version-file: ${{ inputs.go-version-file }}
        cache: true
    - name: Verify modules
      shell: bash
      run: go mod verify
```

Consume it:

```yaml
jobs:
  test:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v5
      - uses: ./.github/actions/setup-go-cached
      - run: go test ./...
```

---

## 4. Action pinning + Renovate

**SHA pin third-party actions:**

```yaml
- uses: aquasecurity/trivy-action@b6643a29fecd7f34b3597bc6acb0a98b03d33ff8 # v0.28.0
- uses: gitleaks/gitleaks-action@ff98106e4c7b2bc287b24eaf42907196329070c7 # v2.3.6
```

**Renovate config** that bumps SHAs while preserving the `# vX.Y.Z` comment:

```json
{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  "extends": ["config:recommended", "helpers:pinGitHubActionDigests"],
  "packageRules": [
    {
      "matchManagers": ["github-actions"],
      "pinDigests": true,
      "groupName": "github-actions"
    }
  ]
}
```

**Minimal workflow-level permissions:**

```yaml
permissions:
  contents: read           # default-deny everything else
jobs:
  publish:
    permissions:
      contents: write      # for creating releases
      packages: write      # for pushing images
      id-token: write      # for OIDC
```

---

## 5. OIDC cloud auth

**AWS — no static keys in repo secrets:**

```yaml
permissions:
  id-token: write
  contents: read

jobs:
  deploy:
    runs-on: ubuntu-24.04
    steps:
      - uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: arn:aws:iam::123456789012:role/gha-deploy
          aws-region: us-east-1
      - run: aws sts get-caller-identity
```

AWS trust policy on the role (`gha-deploy`):

```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": { "Federated": "arn:aws:iam::123456789012:oidc-provider/token.actions.githubusercontent.com" },
    "Action": "sts:AssumeRoleWithWebIdentity",
    "Condition": {
      "StringEquals": {
        "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
      },
      "StringLike": {
        "token.actions.githubusercontent.com:sub": "repo:ralvarezdev/ralvaskills:ref:refs/heads/main"
      }
    }
  }]
}
```

Tighten the `sub` claim — `repo:owner/repo:ref:refs/heads/main` is broader than `repo:owner/repo:environment:prod`. Use the narrowest claim the workflow needs.

**GCP — workload identity federation:**

```yaml
- uses: google-github-actions/auth@v2
  with:
    workload_identity_provider: projects/123/locations/global/workloadIdentityPools/gh/providers/gha
    service_account: gha-deploy@my-project.iam.gserviceaccount.com
```

---

## 6. Language caching

**Go** — `setup-go` handles modules + build cache automatically:

```yaml
- uses: actions/setup-go@v5
  with:
    go-version-file: go.mod
    cache: true        # caches ~/go/pkg/mod and the build cache
```

**Python (uv)** — `setup-python` with `cache: 'pip'` works for pip; uv has native cache support:

```yaml
- uses: actions/setup-python@v5
  with:
    python-version-file: pyproject.toml
- uses: astral-sh/setup-uv@v3
  with:
    enable-cache: true
    cache-dependency-glob: 'uv.lock'
- run: uv sync --frozen
```

**Node** — `setup-node` with `cache: 'pnpm'` (or `'npm'`, `'yarn'`):

```yaml
- uses: pnpm/action-setup@v4
  with: { version: 9 }
- uses: actions/setup-node@v4
  with:
    node-version-file: .nvmrc
    cache: pnpm
- run: pnpm install --frozen-lockfile
```

**Docker layer cache** — GitHub Actions cache (`type=gha`) for ephemeral, registry cache for production:

```yaml
- uses: docker/build-push-action@v6
  with:
    cache-from: type=registry,ref=ghcr.io/${{ github.repository }}:buildcache
    cache-to:   type=registry,ref=ghcr.io/${{ github.repository }}:buildcache,mode=max
```

---

## 7. Matrix strategy

```yaml
jobs:
  test:
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-24.04, macos-14, windows-2022]
        go: ['1.25', '1.26']
        exclude:
          - os: windows-2022
            go: '1.25'           # narrow the matrix; not every cell earns it
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v5
      - uses: actions/setup-go@v5
        with: { go-version: ${{ matrix.go }}, cache: true }
      - run: go test ./...
```

Use `fail-fast: false` for compatibility surveys; default `true` saves runner minutes when any failure blocks the merge.

---

## 8. Reusable workflow — CI gate

Define once, call from many repos.

```yaml
# .github/workflows/reusable-go-ci.yaml
name: Reusable Go CI
on:
  workflow_call:
    inputs:
      go-version-file:
        type: string
        default: go.mod
permissions:
  contents: read
jobs:
  lint-and-test:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v5
      - uses: actions/setup-go@v5
        with: { go-version-file: ${{ inputs.go-version-file }}, cache: true }
      - uses: golangci/golangci-lint-action@v6
        with: { version: v2.0 }
      - run: go test -race -coverprofile=cover.out ./...
```

Consume from another repo:

```yaml
jobs:
  ci:
    uses: ralvarezdev/ci-workflows/.github/workflows/reusable-go-ci.yaml@<sha>
```

---

## 9. Build & push image

Multi-arch, signed, with attestations. Image-level rules (Dockerfile, USER, etc.) live in [docker-architect](../docker-architect/SKILL.md).

```yaml
name: release
on:
  push:
    tags: ['v*']

permissions:
  contents: read
  packages: write
  id-token: write       # for Cosign OIDC signing
  attestations: write

jobs:
  publish:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v5

      - uses: docker/setup-qemu-action@v3
      - uses: docker/setup-buildx-action@v3

      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - id: meta
        uses: docker/metadata-action@v5
        with:
          images: ghcr.io/${{ github.repository }}
          tags: |
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha,prefix=,format=short

      - id: build
        uses: docker/build-push-action@v6
        with:
          context: .
          platforms: linux/amd64,linux/arm64
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          provenance: true
          sbom: true
          cache-from: type=registry,ref=ghcr.io/${{ github.repository }}:buildcache
          cache-to:   type=registry,ref=ghcr.io/${{ github.repository }}:buildcache,mode=max

      - uses: sigstore/cosign-installer@v3
      - name: Sign image
        run: cosign sign --yes ghcr.io/${{ github.repository }}@${{ steps.build.outputs.digest }}
```

---

## 10. release-please

**Manifest-driven config** (`release-please-config.json`):

```json
{
  "release-type": "go",
  "packages": {
    ".": {
      "package-name": "ralvaskills",
      "include-component-in-tag": false
    }
  },
  "changelog-sections": [
    { "type": "feat", "section": "Features" },
    { "type": "fix",  "section": "Bug Fixes" },
    { "type": "perf", "section": "Performance" },
    { "type": "deps", "section": "Dependencies" }
  ]
}
```

**Workflow:**

```yaml
name: release-please
on:
  push:
    branches: [main]

permissions:
  contents: write
  pull-requests: write

jobs:
  release-please:
    runs-on: ubuntu-24.04
    steps:
      - uses: googleapis/release-please-action@v4
        with:
          config-file: release-please-config.json
          manifest-file: .release-please-manifest.json
```

When release-please merges the version-bump PR it creates a `v*` tag, which triggers `release.yaml` (§9).

---

## 11. Deploy with environment approval

```yaml
name: deploy
on:
  workflow_dispatch:
    inputs:
      image_tag:
        description: Image tag or digest to deploy
        required: true
  release:
    types: [published]

concurrency:
  group: deploy-prod         # one prod deploy at a time
  cancel-in-progress: false

permissions:
  id-token: write
  contents: read

jobs:
  deploy-prod:
    runs-on: ubuntu-24.04
    environment:
      name: production       # required reviewers + secrets scoped here
      url: https://app.example.com
    steps:
      - uses: actions/checkout@v5
      - uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ vars.PROD_DEPLOY_ROLE }}
          aws-region: us-east-1
      - name: Promote image
        run: ./scripts/promote.sh "${{ inputs.image_tag || github.event.release.tag_name }}"
```

The `environment: production` block enforces the manual reviewer gate and scopes any secrets/vars to prod. The `concurrency` group prevents overlapping prod deploys.
