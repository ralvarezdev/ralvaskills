# Codebase Hygiene Reviewer — Reference

Commands and the report template. Loaded on demand.

## 1. Tool matrix

| Goal | Go | TypeScript / React |
|---|---|---|
| Health gate (run first) | `go build ./... && go vet ./...` | `tsc --noEmit` |
| Unreachable code | `deadcode -test ./...` | `knip --include files` |
| Unused exports | `deadcode -test ./...` (exported funcs no root reaches) | `knip --include exports` |
| Duplicate exports | | `knip --include duplicates` |
| Unused deps | `go mod tidy -diff` | `knip --include dependencies` |
| Unexported unused | `golangci-lint run --enable-only=unused ./...` | `tsc --noEmit` with `noUnusedLocals` |
| Structural clones (one module) | `golangci-lint run --enable-only=dupl ./...` | `jscpd` (see § 5) |
| Clones across modules / repos, non-code files | `jscpd` (see § 5) | `jscpd` (see § 5) |
| Size / complexity | `funlen`, `gocognit`, `maintidx`, revive `file-length-limit` | ESLint `max-lines`, `complexity` |
| Files tracked | `git ls-files '*.go'` | `git ls-files '*.ts' '*.tsx'` |

Exclude generated paths in the tool config, not by deleting hits from the report afterwards.

## 2. Go: library with several consumers

`deadcode` roots on `main`, so a library has no roots of its own. Measure usage from each consumer:

```sh
# run inside each consumer module
deadcode -test -filter='<library-module-path>' ./...
```

`-filter` restricts the report to packages matching the regexp, so the output reads "library symbols this consumer never reaches". A symbol is a deletion candidate only when **every** consumer reports it unreachable.

Package coverage (who imports what):

```sh
go list -deps -test ./... | grep '<library-module-path>' | sort -u
```

Packages in the library that appear in no consumer's list are candidates for removal or for being made internal.

## 3. TypeScript: generated and ignored paths

```jsonc
// knip.json
{
  "ignore": ["**/*.gen.ts", "**/schema.d.ts", "dist/**"],
  "ignoreDependencies": []
}
```

Route files, `page.tsx`, `layout.tsx` and similar framework-convention entries are roots; add them under `entry` rather than treating them as unused.

## 4. Byte-identical generated files

```sh
git ls-files '*.sql.go' | xargs md5sum | sort | uniq -w32 -D
```

Identical hashes across modules mean duplicated generation (for example the same sqlc models in several services). Report as duplication; the fix is a shared module or one generator input, never a manual edit.

## 5. Cross-module duplication with jscpd

`jscpd` (respects `.gitignore` by default) is a token-based copy/paste detector that treats Go, TypeScript, SQL, proto, YAML and Dockerfiles alike. Use it where `dupl` cannot reach: clones between modules or services, and duplication in non-code files. Inside a single Go module prefer `dupl`, which already runs in the lint gate.

```sh
npx jscpd --min-tokens 70 --min-lines 8 \
  --ignore '**/*.pb.go,**/*.sql.go,**/*.gen.ts,**/schema.d.ts,**/dist/**,**/vendor/**' \
  --reporters console,json --output "${TMPDIR:-/tmp}/jscpd" <path-to-workspace>
```

Raise `--min-tokens` if the output is mostly boilerplate (error handling, struct tags). Every clone pair still needs a human check: repeated shapes that are intentionally separate (per-service config, adapters) are not findings. Report results as duplication with both locations, never as a delete instruction.

## 6. Report template

```markdown
# Hygiene report — <repo> — <date>

Scope: <tracked files, excluded paths, tests as roots yes/no, consumers audited>

| Repo | Files | LOC | Dead | Unused exports | Dup | Split candidates |
|---|---|---|---|---|---|---|

## Dead code
| Symbol | Location | Evidence | Confidence |
|---|---|---|---|

## Unused exports
| Symbol | Location | Consumers searched | Verdict (unexport / deprecate / keep) |
|---|---|---|---|

## Duplication
## Split candidates
| Target | Size evidence | Named seam | Verdict (split / monitor) |
|---|---|---|---|

## Quick wins
## Do not touch
<hit + reason it is a false positive>

## Priority
- Now:
- Later:
- Won't do:
```

Output file name is kebab-case: `hygiene-report-<timestamp>.md`.
