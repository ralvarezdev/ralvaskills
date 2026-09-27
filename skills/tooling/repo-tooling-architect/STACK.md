# Stack Versions

| Dependency | Pinned version | Purpose |
|---|---|---|
| mise | 2026.5 | **Default** tool / runtime version manager (`mise.toml`) |
| proto | 0.57 | **Alternative** tool version manager (`.prototools`) — strict scope, cleaner separation |
| Task | 3.51 | **Default** task runner (`Taskfile.yml`) + dotenv loader |
| just | 1.51 | **Alternative** task runner (`justfile`) — Makefile-style brevity |
| pre-commit | 4.6 | Pre-commit hook framework — minimal hook set only |
| gitleaks | 8.21 | Secret detection (called by pre-commit) |
| Renovate | 43.x | **Default** dependency updater (`renovate.json`) |
| Dependabot | (GitHub-native) | Acceptable alternative for tiny single-language repos |
| act | 0.2.89 | Local GitHub Actions runner — free alternative when Actions billing is a blocker |

## Notes

- **Tool version manager:** `mise` is the default for broader adoption, `winget`/`brew`/`scoop` packaging, and asdf-legacy plugin compatibility. `proto` is the alternative when strict separation of concerns matters (Task owns env + scripts; tool manager only pins binaries).
- **Task runner:** `Task` is the default for native cross-platform support (built-in POSIX sh on Windows), incremental builds via `sources:`/`generates:`, and `--watch` mode. `just` is the alternative when Makefile-style recipe brevity is preferred.
- **Env loading:** delegated to `Task`'s `dotenv:` for non-secrets; external secret manager (1Password CLI / doppler / cloud) for real secrets. **No `direnv`** in either default or alternative stack.
- **Pre-commit hooks: minimal only.** Trailing whitespace, EOF fixer, large-files-check, merge-conflict, YAML / JSON validation, gitleaks. Language linters stay in editor (on save) and CI.
- **Both `mise` and `proto` are committed:** the pinning file is the source of truth. CI uses the same tool, the same versions.
- **`.editorconfig` and `.gitignore`** are always included; no version to pin.
- **`act` is opt-in, not default:** GitHub Actions stays the default CI target. `act` only enters the stack when Actions billing specifically becomes a blocker — see §8.
- **A Docker-push job under `act`, against a registry on the same machine, needs two extra fixes** beyond the base act setup: a `REGISTRY_HOST` variable to bypass any tunnel in front of the registry for local runs, and `driver-opts: network=host` on `docker/setup-buildx-action` — see §8's numbered gotchas. Confirmed by an actual end-to-end local publish, not written speculatively.

_Last reviewed: 2026-09-27_
_Skill version at last review: 1.3.0_
