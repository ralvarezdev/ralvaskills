# Registry Flow

End-to-end flow from editing a skill to a user installing it.

---

## 1. You update a skill

```
skills/languages/go-architect/SKILL.md
  version: 1.0.0  →  1.1.0
```

You commit and push to `main`.

---

## 2. CI detects the push

`.github/workflows/publish-registry.yml` triggers because `skills/**` changed.

---

## 3. `generate-registry` runs

```
go run ./cmd/generate-registry \
  --existing-index registry/index.json \
  --github-repo ralvarezdev/ralvaskills
```

The script:
- Walks `skills/`, reads every `SKILL.md` version
- Compares against `registry/index.json`
- Finds `go-architect` bumped from `1.0.0` → `1.1.0`
- Creates `dist/go-architect-v1.1.0.tar.gz`
- Writes `dist/new-versions.json`: `[{name: "go-architect", version: "1.1.0", archive: "go-architect-v1.1.0.tar.gz"}]`
- Writes `dist/index.json` with the new version entry and `archive_url` pointing at GitHub Releases

---

## 4. CI publishes

Three things happen in sequence:

```
GitHub Releases                           ← PRIMARY
  tag:   go-architect@v1.1.0
  asset: go-architect-v1.1.0.tar.gz

R2 private bucket                         ← BACKUP
  s3://ralvaskills/go-architect/v1.1.0.tar.gz

git commit registry/index.json [skip ci]  ← INDEX UPDATE
  pushed back to main
```

---

## 5. GitHub Pages picks up the index

`registry/` is the GitHub Pages source, so the commit from step 4 automatically publishes `registry/index.json` to:

```
https://skills.ralvarez.dev/index.json
```

Within ~30 seconds of the push. No deploy step needed.

---

## 6. User runs `rsk install go-architect`

```
rsk
 │
 ├── load config (~/.config/rsk/config.json)
 │     registry_url: https://skills.ralvarez.dev
 │
 ├── GET https://skills.ralvarez.dev/index.json
 │     → finds go-architect, latest: 1.1.0
 │     → archive_url: https://github.com/.../releases/download/
 │                    go-architect%40v1.1.0/go-architect-v1.1.0.tar.gz
 │
 ├── check cache: ~/.ralvaskills/cache/registry/go-architect/1.1.0/
 │     → not found, download needed
 │
 ├── GET archive_url
 │     → downloads go-architect-v1.1.0.tar.gz from GitHub Releases
 │
 ├── extract tarball → ~/.ralvaskills/cache/registry/go-architect/1.1.0/
 │     SKILL.md
 │     STACK.md
 │     RECIPES.md
 │
 └── symlink ./.claude/skills/go-architect
       → ~/.ralvaskills/cache/registry/go-architect/1.1.0/
```

---

## 7. User runs `rsk update`

```
rsk update
 │
 ├── GET https://skills.ralvarez.dev/index.json
 │     → go-architect latest: 1.1.0
 │
 ├── scan symlinks in target dirs
 │     .claude/skills/go-architect → .../1.0.0/   ← outdated
 │
 ├── download go-architect-v1.1.0.tar.gz
 │     extract → ~/.ralvaskills/cache/registry/go-architect/1.1.0/
 │
 └── re-link .claude/skills/go-architect
       → ~/.ralvaskills/cache/registry/go-architect/1.1.0/
```

---

## 8. Disaster recovery (if GitHub Releases are lost)

```
R2 bucket (private)
  s3://ralvaskills/go-architect/v1.1.0.tar.gz

→ download from R2
→ gh release create go-architect@v1.1.0 --attach go-architect-v1.1.0.tar.gz
→ index.json archive_urls unchanged — rsk users unaffected
```

---

## Data at rest

| Where | What | Who writes | Who reads |
|---|---|---|---|
| `skills/*/SKILL.md` | Source of truth | You | CI script |
| `registry/index.json` | Version catalog | CI (git commit) | GitHub Pages → `rsk` |
| GitHub Releases | Versioned tarballs | CI (`gh release create`) | `rsk` |
| R2 private | Tarball backup | CI (`aws s3 cp`) | Nobody (recovery only) |
| `~/.ralvaskills/cache/registry/` | Extracted skills | `rsk install` | Symlinks |
