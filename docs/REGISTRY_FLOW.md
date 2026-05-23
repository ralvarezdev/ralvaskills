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

`.github/workflows/publish-registry.yml` triggers when paths under `skills/**`, `cmd/generate-registry/**`, or `internal/**` change on `main`.

---

## 3. `generate-registry` runs

```
go run ./cmd/generate-registry \
  --skills-dir skills \
  --output-dir dist \
  --existing-index registry/index.json \
  --github-repo ralvarezdev/ralvaskills
```

The script:
- Walks `skills/`, reads each `SKILL.md` frontmatter (`version`, `description`) and flags any skill under `personal/` as personal
- Compares against `registry/index.json`
- Finds `go-architect` bumped from `1.0.0` → `1.1.0`
- Creates `dist/go-architect-v1.1.0.tar.gz` (entries rooted at `go-architect/`)
- Writes `dist/new-versions.json`: `[{name: "go-architect", version: "1.1.0", archive: "go-architect-v1.1.0.tar.gz"}]`
- Writes `dist/index.json` with the new version entry and `archive_url` pointing at GitHub Releases

---

## 4. CI publishes

A `Check for new versions` step counts entries in `dist/new-versions.json` and exports `count`. The three publish steps below are all gated on `count != '0'` — if no skill bumped its version, the job is a no-op:

```
GitHub Releases                           ← PRIMARY
  gh release create go-architect@v1.1.0 dist/go-architect-v1.1.0.tar.gz
  (already-existing tags are skipped, not failed)

R2 private bucket                         ← BACKUP
  aws s3 cp dist/go-architect-v1.1.0.tar.gz
            s3://ralvaskills/go-architect/v1.1.0.tar.gz

git commit registry/index.json [skip ci]  ← INDEX UPDATE
  cp dist/index.json registry/index.json && git push
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

Registry mode (no `repo_path` configured). `<cache>` below resolves to `<dirname(official_cache)>/registry/` — i.e. the registry cache sits next to the official cache configured at `rsk init`.

```
rsk install go-architect
 │
 ├── load config (~/.config/rsk/config.json)
 │     registry_url:   https://skills.ralvarez.dev
 │     official_cache: ~/.cache/rsk/official    (example)
 │     → registry cache: ~/.cache/rsk/registry/
 │
 ├── GET https://skills.ralvarez.dev/index.json
 │     → finds go-architect, latest: 1.1.0
 │     → archive_url: https://github.com/.../releases/download/
 │                    go-architect%40v1.1.0/go-architect-v1.1.0.tar.gz
 │
 ├── check cache: <cache>/go-architect/1.1.0/
 │     → not found, download needed
 │
 ├── GET archive_url
 │     → downloads go-architect-v1.1.0.tar.gz from GitHub Releases
 │
 ├── extract tarball → <cache>/go-architect/1.1.0/
 │     SKILL.md
 │     (plus any other files the skill ships)
 │
 ├── symlink ./.rsk/skills/go-architect  → <cache>/go-architect/1.1.0/
 │     (or the configured global dir(s) when --global is passed,
 │      e.g. ~/.claude/skills/, ~/.config/opencode/skills/)
 │
 └── update project manifest (project installs only)
       .rsk/rsk.mod   ← add  go-architect = "*"
       .rsk/rsk.lock  ← record name, version, source, path
```

---

## 7. User runs `rsk update`

Two modes, picked by config:

**Local-repo mode** (`repo_path` set): runs `git pull` in the local clone; symlinks pick up the new files automatically. `--official` also refreshes the `anthropics/skills` clone under `official_cache`.

**Registry mode** (`registry_url` set):

```
rsk update
 │
 ├── GET https://skills.ralvarez.dev/index.json
 │     → go-architect latest: 1.1.0
 │
 ├── scan symlinks in target dirs (project .rsk/skills/, or --global dirs)
 │     .rsk/skills/go-architect → <cache>/go-architect/1.0.0/   ← outdated
 │
 ├── download go-architect-v1.1.0.tar.gz
 │     extract → <cache>/go-architect/1.1.0/
 │
 └── re-link .rsk/skills/go-architect
       → <cache>/go-architect/1.1.0/
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
| `<dirname(official_cache)>/registry/<name>/<version>/` | Extracted skills | `rsk install` (registry mode) | Symlinks |
| `.rsk/rsk.mod`, `.rsk/rsk.lock` | Project skill manifest + lock | `rsk install` (project) | `rsk install`, `rsk status` |
