# SKILL.md Token Estimates

> Auto-generated. **Do not edit by hand.** Run `task tokens` to refresh.
>
> Estimate: ~4 bytes/token for bodies, ~3 bytes/token for descriptions (Claude tokenizer). Actual range ±15%.

_Last updated: 2026-05-23 · 41 skills · 15 bundles_

## Load model

Description tokens hit **every turn** for installed skills. Body tokens are paid only when a skill is actually invoked. Side files are never auto-loaded.

| What | When loaded | Estimated tokens |
|---|---|---:|
| All `SKILL.md` bodies | Only when invoked | ~84141 |
| All side files (`STACK` + `RECIPES` + topic files) | On-demand only | ~62141 |
| All `description:` fields (every skill) | Every turn, if all installed | ~3933 |

## Session profiles

Budgets: **global ≤ 1500** desc tokens (paid by every project), **session ≤ 2200** desc tokens (global + one project bundle). Status: `ok` < 90% · `warn` ≥ 90% · `OVER` ≥ 100%.

**Global baseline** (`rsk install global --global`): 15 skills · ~1412 / 1500 desc tokens every turn (`warn`) · ~25284 body tokens if all invoked.

Per-project bundle additions on top of the global baseline:

| Bundle | Skills | + Desc tkns / turn | Session total | Budget | Status |
|---|---:|---:|---:|---:|---:|
| `docs` *(+5 external)* | 0 | +0 | ~1412 | 2200 | `ok` |
| `observability` | 2 | +151 | ~1563 | 2200 | `ok` |
| `event-driven` | 2 | +167 | ~1579 | 2200 | `ok` |
| `go-cli` | 3 | +256 | ~1668 | 2200 | `ok` |
| `python-cli` | 3 | +261 | ~1673 | 2200 | `ok` |
| `code-review` | 3 | +269 | ~1681 | 2200 | `ok` |
| `ros2` | 3 | +273 | ~1685 | 2200 | `ok` |
| `design` *(+1 external)* | 3 | +313 | ~1725 | 2200 | `ok` |
| `llm-app` *(+2 missing)* | 4 | +389 | ~1801 | 2200 | `ok` |
| `go-grpc` | 5 | +443 | ~1855 | 2200 | `ok` |
| `python-grpc` | 5 | +448 | ~1860 | 2200 | `ok` |
| `gin` | 5 | +463 | ~1875 | 2200 | `ok` |
| `nethttp` | 5 | +475 | ~1887 | 2200 | `ok` |
| `fastapi` | 5 | +480 | ~1892 | 2200 | `ok` |

## Personal / unbundled skills

These skills are not part of any bundle. They are installed individually with `rsk install <name> --personal` and their desc tokens are only paid when explicitly installed.

| Skill | Category | ~Body tkns | ~Desc tkns | ~Side tkns |
|---|---|---:|---:|---:|
| `website-concept-architect` | frontend | ~3083 | ~122 | ~0 |
| `ci-cd-architect` | infra | ~2440 | ~114 | ~3539 |
| `demo-presentation-architect` | personal | ~2356 | ~101 | ~5548 |
| `uru-thesis-reviewer` | personal | ~2636 | ~97 | ~7058 |
| `demo-script-architect` | personal | ~1237 | ~96 | ~0 |
| `work-report-generator` | personal | ~2104 | ~82 | ~1240 |

## By category

| Category | Skills | ~Body tkns | ~Desc tkns | ~Side tkns |
|---|---:|---:|---:|---:|
| frameworks | 5 | ~11291 | ~558 | ~9070 |
| infra | 4 | ~8373 | ~354 | ~7566 |
| personal | 4 | ~8333 | ~376 | ~13846 |
| quality | 3 | ~6718 | ~269 | ~1151 |
| refactoring | 4 | ~6557 | ~408 | ~6101 |
| tooling | 2 | ~5519 | ~184 | ~2506 |
| frontend | 2 | ~5313 | ~225 | ~1853 |
| workflows | 4 | ~5027 | ~368 | ~2743 |
| protocols | 2 | ~4985 | ~208 | ~5771 |
| meta | 3 | ~4722 | ~289 | ~1718 |
| design | 2 | ~4178 | ~179 | ~1519 |
| languages | 2 | ~3448 | ~153 | ~804 |
| messaging | 1 | ~2878 | ~83 | ~2662 |
| robotics | 1 | ~2717 | ~105 | ~2321 |
| databases | 1 | ~2141 | ~90 | ~1274 |
| encoding | 1 | ~1941 | ~84 | ~1236 |

## By bundle

Catalog from `internal/config/catalog.toml`. `External` are `source = "official"` skills (e.g. `docx`, `xlsx`) that live outside this repo by design. `Missing` are `source = "local"` skills the bundle references that aren't in the local `skills/` tree.

| Bundle | Local | External | Missing | ~Body tkns (local) | ~Desc tkns (local) |
|---|---:|---:|---:|---:|---:|
| `global` | 15 | 0 | 0 | ~25284 | ~1412 |
| `docs` | 0 | 5 | 0 | ~0 | ~0 |
| `design` | 3 | 1 | 0 | ~8023 | ~313 |
| `go-grpc` | 5 | 0 | 0 | ~10059 | ~443 |
| `gin` | 5 | 0 | 0 | ~9902 | ~463 |
| `nethttp` | 5 | 0 | 0 | ~10065 | ~475 |
| `go-cli` | 3 | 0 | 0 | ~6332 | ~256 |
| `fastapi` | 5 | 0 | 0 | ~9798 | ~480 |
| `llm-app` | 4 | 0 | 2 | ~7432 | ~389 |
| `ros2` | 3 | 0 | 0 | ~6067 | ~273 |
| `python-grpc` | 5 | 0 | 0 | ~9915 | ~448 |
| `python-cli` | 3 | 0 | 0 | ~6188 | ~261 |
| `event-driven` | 2 | 0 | 0 | ~4819 | ~167 |
| `observability` | 2 | 0 | 0 | ~4235 | ~151 |
| `code-review` | 3 | 0 | 0 | ~6718 | ~269 |

## Per skill

| # | Skill | Category | Body bytes | ~Body tkns | ~Desc tkns | ~Stack tkns | ~Recipes tkns | ~Topic tkns |
|---|---|---|---:|---:|---:|---:|---:|---:|
| 1 | `website-concept-architect` | frontend | 11524 | ~3083 | ~122 | ~0 | ~0 | ~0 |
| 2 | `nextjs-architect` | frameworks | 11671 | ~2918 | ~108 | ~454 | ~1824 | ~0 |
| 3 | `event-driven-architect` | messaging | 11494 | ~2878 | ~83 | ~514 | ~2148 | ~0 |
| 4 | `react-architect` | frameworks | 11499 | ~2875 | ~102 | ~465 | ~1426 | ~0 |
| 5 | `cli-tool-architect` | tooling | 11334 | ~2838 | ~93 | ~591 | ~1436 | ~0 |
| 6 | `ros2-architect` | robotics | 10867 | ~2717 | ~105 | ~607 | ~1714 | ~0 |
| 7 | `repo-tooling-architect` | tooling | 9968 | ~2681 | ~91 | ~479 | ~0 | ~0 |
| 8 | `uru-thesis-reviewer` | personal | 10542 | ~2636 | ~97 | ~0 | ~0 | ~7058 |
| 9 | `observability-architect` | infra | 10238 | ~2577 | ~74 | ~546 | ~874 | ~0 |
| 10 | `rsk-guide` | meta | 9451 | ~2542 | ~146 | ~0 | ~0 | ~0 |
| 11 | `api-contract-reviewer` | quality | 9798 | ~2511 | ~108 | ~0 | ~0 | ~0 |
| 12 | `rest-api-architect` | protocols | 10005 | ~2502 | ~102 | ~427 | ~0 | ~3990 |
| 13 | `grpc-architect` | protocols | 9932 | ~2483 | ~106 | ~521 | ~833 | ~0 |
| 14 | `ci-cd-architect` | infra | 9758 | ~2440 | ~114 | ~624 | ~2915 | ~0 |
| 15 | `feature-planner` | workflows | 9202 | ~2369 | ~114 | ~0 | ~0 | ~0 |
| 16 | `demo-presentation-architect` | personal | 9421 | ~2356 | ~101 | ~0 | ~1776 | ~3772 |
| 17 | `hexagonal-arch` | design | 8972 | ~2277 | ~101 | ~0 | ~768 | ~0 |
| 18 | `code-design-refactor` | refactoring | 8927 | ~2264 | ~96 | ~0 | ~0 | ~0 |
| 19 | `ui-ux-architect` | frontend | 8920 | ~2230 | ~103 | ~466 | ~1387 | ~0 |
| 20 | `sql-architect` | databases | 8564 | ~2141 | ~90 | ~295 | ~0 | ~979 |
| 21 | `security-reviewer` | quality | 8474 | ~2119 | ~77 | ~0 | ~539 | ~0 |
| 22 | `work-report-generator` | personal | 8295 | ~2104 | ~82 | ~0 | ~1240 | ~0 |
| 23 | `performance-reviewer` | quality | 8350 | ~2088 | ~84 | ~0 | ~612 | ~0 |
| 24 | `protobuf-architect` | encoding | 7763 | ~1941 | ~84 | ~474 | ~762 | ~0 |
| 25 | `nethttp-architect` | frameworks | 7494 | ~1928 | ~120 | ~398 | ~1508 | ~0 |
| 26 | `ddd-architect` | design | 7602 | ~1901 | ~78 | ~0 | ~0 | ~751 |
| 27 | `improve-codebase-architecture` | refactoring | 7437 | ~1860 | ~97 | ~0 | ~0 | ~3971 |
| 28 | `design-patterns` | refactoring | 7299 | ~1825 | ~122 | ~0 | ~0 | ~2130 |
| 29 | `fastapi-architect` | frameworks | 7220 | ~1805 | ~120 | ~420 | ~1016 | ~0 |
| 30 | `go-architect` | languages | 7055 | ~1796 | ~74 | ~385 | ~0 | ~0 |
| 31 | `gin-architect` | frameworks | 6865 | ~1765 | ~108 | ~413 | ~1146 | ~0 |
| 32 | `skill-builder` | meta | 7030 | ~1758 | ~72 | ~0 | ~1718 | ~0 |
| 33 | `docker-architect` | infra | 6762 | ~1698 | ~89 | ~333 | ~718 | ~0 |
| 34 | `grafana-architect` | infra | 6632 | ~1658 | ~77 | ~407 | ~1149 | ~0 |
| 35 | `python-architect` | languages | 6376 | ~1652 | ~79 | ~419 | ~0 | ~0 |
| 36 | `demo-script-architect` | personal | 4828 | ~1237 | ~96 | ~0 | ~0 | ~0 |
| 37 | `tdd` | workflows | 4279 | ~1119 | ~69 | ~0 | ~0 | ~1350 |
| 38 | `grill-with-docs` | workflows | 3376 | ~889 | ~104 | ~0 | ~0 | ~1393 |
| 39 | `commit-author` | workflows | 2599 | ~650 | ~81 | ~0 | ~0 | ~0 |
| 40 | `logic-cleaner` | refactoring | 2432 | ~608 | ~93 | ~0 | ~0 | ~0 |
| 41 | `caveman` | meta | 1687 | ~422 | ~71 | ~0 | ~0 | ~0 |

**Totals:** 331942 body bytes · ~84141 body tokens · ~3933 desc tokens · ~62141 side tokens

## Topic files

Side files in a skill directory other than `STACK.md` and `RECIPES.md` — typically topic-specific reference docs referenced by name from `SKILL.md` (e.g. `PAYLOADS.md`, `AUTH_PATTERNS.md`). These contribute to the `~Topic tkns` column above and are loaded only when the parent skill chooses to read them.

- `uru-thesis-reviewer` (~7058 tokens total):
  - `NORMAS_URU_2020.md` — ~4950 tokens
  - `TEMPLATES.md` — ~1301 tokens
  - `CHECKS.md` — ~807 tokens
- `rest-api-architect` (~3990 tokens total):
  - `CONCURRENCY.md` — ~1056 tokens
  - `IDEMPOTENCY.md` — ~869 tokens
  - `AUTH_PATTERNS.md` — ~767 tokens
  - `STATUS_CODES.md` — ~662 tokens
  - `PAYLOADS.md` — ~636 tokens
- `improve-codebase-architecture` (~3971 tokens total):
  - `HTML_REPORT.md` — ~1698 tokens
  - `LANGUAGE.md` — ~951 tokens
  - `INTERFACE_DESIGN.md` — ~681 tokens
  - `DEEPENING.md` — ~641 tokens
- `demo-presentation-architect` (~3772 tokens total):
  - `LAYOUTS.md` — ~3772 tokens
- `design-patterns` (~2130 tokens total):
  - `PATTERNS.md` — ~2130 tokens
- `grill-with-docs` (~1393 tokens total):
  - `ADR_FORMAT.md` — ~703 tokens
  - `CONTEXT_FORMAT.md` — ~690 tokens
- `tdd` (~1350 tokens total):
  - `TESTS.md` — ~410 tokens
  - `MOCKING.md` — ~370 tokens
  - `DEEP_MODULES.md` — ~310 tokens
  - `INTERFACE_DESIGN.md` — ~163 tokens
  - `REFACTORING.md` — ~97 tokens
- `sql-architect` (~979 tokens total):
  - `POSTGRES.md` — ~501 tokens
  - `ENGINES.md` — ~478 tokens
- `ddd-architect` (~751 tokens total):
  - `PATTERNS.md` — ~751 tokens

## Skills to consider trimming

Body > 2500 tokens — consider moving examples to `RECIPES.md` or topic files:

- `website-concept-architect` (~3083 body tokens)
- `nextjs-architect` (~2918 body tokens)
- `event-driven-architect` (~2878 body tokens)
- `react-architect` (~2875 body tokens)
- `cli-tool-architect` (~2838 body tokens)
- `ros2-architect` (~2717 body tokens)
- `repo-tooling-architect` (~2681 body tokens)
- `uru-thesis-reviewer` (~2636 body tokens)
- `observability-architect` (~2577 body tokens)
- `rsk-guide` (~2542 body tokens)
- `api-contract-reviewer` (~2511 body tokens)
- `rest-api-architect` (~2502 body tokens)

Heaviest 10 descriptions — each desc token is paid every turn for any session that installs the skill:

- `rsk-guide` (~146 desc tokens)
- `website-concept-architect` (~122 desc tokens)
- `design-patterns` (~122 desc tokens)
- `fastapi-architect` (~120 desc tokens)
- `nethttp-architect` (~120 desc tokens)
- `feature-planner` (~114 desc tokens)
- `ci-cd-architect` (~114 desc tokens)
- `gin-architect` (~108 desc tokens)
- `nextjs-architect` (~108 desc tokens)
- `api-contract-reviewer` (~108 desc tokens)

## Notes

- **Body tokens** cost only when the skill is invoked in a session.
- **Description tokens** cost every turn for any session that has the skill installed. The cost depends on which bundles are installed, not on the corpus total.
- Side files (`STACK.md`, `RECIPES.md`, topic files) are never auto-loaded; they cost 0 per turn.
- For exact counts: run each file through [`tiktoken`](https://github.com/openai/tiktoken) with `cl100k_base`.
