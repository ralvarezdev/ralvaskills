# Registry Setup

This document walks through the one-time setup for the skill registry.

## Architecture

```
skills.ralvarez.dev/index.json   ← GitHub Pages  (registry/index.json in repo)
GitHub Releases assets           ← primary tarball download source for rsk
Cloudflare R2 (private)          ← cold backup of tarballs, never read by users
```

`rsk` downloads tarballs from GitHub Releases URLs embedded in `index.json`.
R2 is write-only from CI — it exists so you can recover or migrate without GitHub.

---

## 1. Enable GitHub Pages

1. Go to **GitHub → `ralvarezdev/ralvaskills` → Settings → Pages**
2. **Source:** `GitHub Actions`
3. (No branch/folder picker — the workflow uploads `registry/` as the site root.)

The `Publish Registry` workflow runs on pushes that touch `skills/**`, `cmd/generate-registry/**`, or `internal/**`. On every run it generates the index, then uploads `registry/` via `actions/upload-pages-artifact` + `actions/deploy-pages`. After the first successful run, `https://skills.ralvarez.dev/index.json` serves the current index.

> **Why GitHub Actions, not "Deploy from a branch"?** The branch-folder source only allows `/` or `/docs` as the publish folder — `/registry` cannot be selected. Deploying via Actions side-steps that restriction and keeps `registry/index.json` where it already lives.

> **Custom domain:** Add `skills.ralvarez.dev` in the Pages settings. Since `ralvarez.dev` is on Cloudflare, add a CNAME record pointing `skills` at `<username>.github.io`. GitHub provisions TLS automatically.

---

## 2. Create the R2 Bucket (private backup)

1. Go to **Cloudflare dashboard → R2 Object Storage → Create bucket**
2. Bucket name: `ralvaskills`
3. Location: **Automatic**
4. Click **Create bucket** — leave it private (no public access, no custom domain)

---

## 3. Create an R2 API Token

1. Go to **Cloudflare dashboard → R2 → Manage R2 API Tokens → Create API Token**
2. Settings:
   - **Token name:** `ralvaskills-ci`
   - **Permissions:** `Object Read & Write`
   - **Bucket:** Specify → `ralvaskills`
   - **TTL:** No expiry
3. Click **Create API Token** and copy:
   - **Access Key ID**
   - **Secret Access Key**

Also copy your **Cloudflare Account ID** from the dashboard sidebar.

---

## 4. Add GitHub Secrets

Go to **GitHub → `ralvarezdev/ralvaskills` → Settings → Secrets and variables → Actions** and add:

| Secret name | Value |
|---|---|
| `CLOUDFLARE_ACCOUNT_ID` | Your Cloudflare Account ID |
| `R2_ACCESS_KEY_ID` | Access Key ID from step 3 |
| `R2_SECRET_ACCESS_KEY` | Secret Access Key from step 3 |

`GITHUB_TOKEN` is provided automatically — no setup needed.

---

## 5. Trigger the First Publish

The workflow triggers on pushes to `main` that touch any of:

- `skills/**`
- `cmd/generate-registry/**`
- `internal/**`

There is no `workflow_dispatch` trigger, so the "Run workflow" button won't appear in the Actions UI. To trigger the first publish, push a small change under one of those paths (e.g. bump a skill's `version:` or touch `skills/.keep`).

The first run will:
1. Read `registry/index.json` from the repo (starts empty)
2. Create tarballs for every skill at its current version
3. Count entries in `dist/new-versions.json` — if zero, the publish steps below are skipped
4. Create one GitHub Release per skill (e.g. `go-architect@v1.0.0`); already-existing tags are skipped, not failed
5. Back up each tarball to the private R2 bucket
6. Commit an updated `registry/index.json` back to `main` with `[skip ci]`
7. Deploy `registry/` to GitHub Pages — this step runs on every workflow execution, not just when new versions are published

Subsequent runs only publish skills whose `version:` field changed, but the Pages deploy always refreshes from the current `registry/`.

---

## 6. Verify

```bash
# Index is served from GitHub Pages
curl https://skills.ralvarez.dev/index.json | jq '.skills | keys'

# A tarball is reachable from GitHub Releases
curl -IL https://github.com/ralvarezdev/ralvaskills/releases/download/go-architect%40v1.0.0/go-architect-v1.0.0.tar.gz
```

---

## 7. Recovery from R2

If GitHub Releases are ever lost, every tarball is in the private R2 bucket at:
```
s3://ralvaskills/<skill>/v<version>.tar.gz
```

To restore: download from R2, re-upload as GitHub Release assets, the `index.json` URLs stay unchanged.

---

## Summary

| Resource | Location |
|---|---|
| Index | `registry/index.json` in repo → `https://skills.ralvarez.dev/index.json` |
| Tarballs (primary) | GitHub Releases assets |
| Tarballs (backup) | Private R2 bucket `ralvaskills` |
| CI workflow | `.github/workflows/publish-registry.yml` |
| Local task | `task registry:generate` |
