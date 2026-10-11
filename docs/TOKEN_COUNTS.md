# SKILL.md Token Estimates

> Auto-generated. **Do not edit by hand.** Run `task tokens` to refresh.
>
> Estimate: ~4 bytes/token for bodies, ~3 bytes/token for descriptions (Claude tokenizer). Actual range ±15%.

_Last updated: 2026-10-11 · 53 skills · 15 bundles_

## Load model

Description tokens hit **every turn** for installed skills. Body tokens are paid only when a skill is actually invoked. Side files are never auto-loaded.

| What | When loaded | Estimated tokens |
|---|---|---:|
| All `SKILL.md` bodies | Only when invoked | ~127746 |
| All side files (`STACK` + `RECIPES` + topic files) | On-demand only | ~100356 |
| All `description:` fields (every skill) | Every turn, if all installed | ~5839 |

## Session profiles

Budgets: **global ≤ 1500** desc tokens (paid by every project), **session ≤ 2200** desc tokens (global + one project bundle). Status: `ok` < 90% · `warn` ≥ 90% · `OVER` ≥ 100%.

**Global baseline** (`rsk install global --global`): 15 skills · ~1483 / 1500 desc tokens every turn (`warn`) · ~27535 body tokens if all invoked.

Per-project bundle additions on top of the global baseline:

| Bundle | Skills | + Desc tkns / turn | Session total | Budget | Status |
|---|---:|---:|---:|---:|---:|
| `docs` *(+5 external)* | 0 | +0 | ~1483 | 2200 | `ok` |
| `observability` | 2 | +151 | ~1634 | 2200 | `ok` |
| `event-driven` | 2 | +167 | ~1650 | 2200 | `ok` |
| `python-cli` | 3 | +261 | ~1744 | 2200 | `ok` |
| `go-cli` | 3 | +265 | ~1748 | 2200 | `ok` |
| `code-review` | 3 | +269 | ~1752 | 2200 | `ok` |
| `ros2` | 3 | +273 | ~1756 | 2200 | `ok` |
| `design` *(+1 external)* | 3 | +313 | ~1796 | 2200 | `ok` |
| `llm-app` *(+2 missing)* | 4 | +389 | ~1872 | 2200 | `ok` |
| `python-grpc` | 5 | +448 | ~1931 | 2200 | `ok` |
| `go-grpc` | 5 | +452 | ~1935 | 2200 | `ok` |
| `gin` | 5 | +472 | ~1955 | 2200 | `ok` |
| `fastapi` | 5 | +480 | ~1963 | 2200 | `ok` |
| `nethttp` | 5 | +484 | ~1967 | 2200 | `ok` |

## Personal / unbundled skills

These skills are not part of any bundle. They are installed individually with `rsk install <name> --include personal` and their desc tokens are only paid when explicitly installed.

| Skill | Category | ~Body tkns | ~Desc tkns | ~Side tkns |
|---|---|---:|---:|---:|
| `mcp-architect` | protocols | ~7584 | ~322 | ~9336 |
| `go-house-libs` | personal | ~1870 | ~202 | ~1003 |
| `go-library-builder` | personal | ~3915 | ~166 | ~3353 |
| `uru-scientific-paper-architect` | personal | ~2233 | ~141 | ~2021 |
| `uru-thesis-defense-architect` | personal | ~1798 | ~132 | ~1525 |
| `astro-architect` | frameworks | ~3252 | ~132 | ~3941 |
| `secrets-architect` | infra | ~2038 | ~132 | ~0 |
| `website-concept-architect` | frontend | ~3083 | ~122 | ~0 |
| `ml-conference-paper-architect` | academic | ~2617 | ~121 | ~1762 |
| `uru-thesis-reviewer` | personal | ~3358 | ~119 | ~7587 |
| `latex-architect` | languages | ~3099 | ~116 | ~2180 |
| `ci-cd-architect` | infra | ~2440 | ~114 | ~3539 |
| `uru-thesis-architect` | personal | ~1948 | ~114 | ~6629 |
| `pi-iteration-workflow` | personal | ~916 | ~108 | ~364 |
| `hugo-architect` | frameworks | ~2817 | ~107 | ~1407 |
| `demo-presentation-architect` | personal | ~2356 | ~101 | ~5548 |
| `demo-script-architect` | personal | ~1237 | ~96 | ~0 |
| `reveal-js-architect` | frameworks | ~2745 | ~93 | ~1203 |

## By category

| Category | Skills | ~Body tkns | ~Desc tkns | ~Side tkns |
|---|---:|---:|---:|---:|
| frameworks | 8 | ~20105 | ~890 | ~15621 |
| personal | 9 | ~19631 | ~1179 | ~28030 |
| protocols | 3 | ~12670 | ~530 | ~15337 |
| languages | 3 | ~12239 | ~278 | ~5781 |
| infra | 5 | ~10491 | ~486 | ~8490 |
| tooling | 2 | ~7077 | ~202 | ~2672 |
| quality | 3 | ~6718 | ~269 | ~1151 |
| refactoring | 4 | ~6557 | ~408 | ~6101 |
| frontend | 2 | ~5313 | ~225 | ~1853 |
| workflows | 4 | ~5179 | ~368 | ~2743 |
| meta | 3 | ~5121 | ~342 | ~1718 |
| design | 2 | ~4320 | ~179 | ~1519 |
| messaging | 1 | ~2878 | ~83 | ~2662 |
| robotics | 1 | ~2717 | ~105 | ~2321 |
| academic | 1 | ~2617 | ~121 | ~1762 |
| databases | 1 | ~2172 | ~90 | ~1359 |
| encoding | 1 | ~1941 | ~84 | ~1236 |

## By bundle

Catalog from `internal/config/catalog.toml`. `External` are `source = "official"` skills (e.g. `docx`, `xlsx`) that live outside this repo by design. `Missing` are `source = "local"` skills the bundle references that aren't in the local `skills/` tree.

| Bundle | Local | External | Missing | ~Body tkns (local) | ~Desc tkns (local) |
|---|---:|---:|---:|---:|---:|
| `global` | 15 | 0 | 0 | ~27535 | ~1483 |
| `docs` | 0 | 5 | 0 | ~0 | ~0 |
| `design` | 3 | 1 | 0 | ~8023 | ~313 |
| `go-grpc` | 5 | 0 | 0 | ~14146 | ~452 |
| `gin` | 5 | 0 | 0 | ~14090 | ~472 |
| `nethttp` | 5 | 0 | 0 | ~14253 | ~484 |
| `go-cli` | 3 | 0 | 0 | ~10388 | ~265 |
| `fastapi` | 5 | 0 | 0 | ~11726 | ~480 |
| `llm-app` | 4 | 0 | 2 | ~9370 | ~389 |
| `ros2` | 3 | 0 | 0 | ~7863 | ~273 |
| `python-grpc` | 5 | 0 | 0 | ~11742 | ~448 |
| `python-cli` | 3 | 0 | 0 | ~7984 | ~261 |
| `event-driven` | 2 | 0 | 0 | ~4819 | ~167 |
| `observability` | 2 | 0 | 0 | ~4235 | ~151 |
| `code-review` | 3 | 0 | 0 | ~6718 | ~269 |

## Per skill

| # | Skill | Category | Body bytes | ~Body tkns | ~Desc tkns | ~Stack tkns | ~Recipes tkns | ~Topic tkns |
|---|---|---|---:|---:|---:|---:|---:|---:|
| 1 | `mcp-architect` | protocols | 30232 | ~7584 | ~322 | ~1521 | ~7815 | ~0 |
| 2 | `go-architect` | languages | 22784 | ~5772 | ~83 | ~473 | ~2685 | ~0 |
| 3 | `repo-tooling-architect` | tooling | 15909 | ~4239 | ~109 | ~645 | ~0 | ~0 |
| 4 | `go-library-builder` | personal | 15657 | ~3915 | ~166 | ~628 | ~2725 | ~0 |
| 5 | `python-architect` | languages | 13242 | ~3368 | ~79 | ~443 | ~0 | ~0 |
| 6 | `uru-thesis-reviewer` | personal | 13430 | ~3358 | ~119 | ~0 | ~0 | ~7587 |
| 7 | `astro-architect` | frameworks | 12582 | ~3252 | ~132 | ~604 | ~3337 | ~0 |
| 8 | `latex-architect` | languages | 12396 | ~3099 | ~116 | ~339 | ~1841 | ~0 |
| 9 | `website-concept-architect` | frontend | 11524 | ~3083 | ~122 | ~0 | ~0 | ~0 |
| 10 | `rsk-guide` | meta | 11028 | ~2941 | ~199 | ~0 | ~0 | ~0 |
| 11 | `nextjs-architect` | frameworks | 11671 | ~2918 | ~108 | ~454 | ~1824 | ~0 |
| 12 | `event-driven-architect` | messaging | 11494 | ~2878 | ~83 | ~514 | ~2148 | ~0 |
| 13 | `react-architect` | frameworks | 11499 | ~2875 | ~102 | ~465 | ~1426 | ~0 |
| 14 | `cli-tool-architect` | tooling | 11334 | ~2838 | ~93 | ~591 | ~1436 | ~0 |
| 15 | `hugo-architect` | frameworks | 11266 | ~2817 | ~107 | ~434 | ~973 | ~0 |
| 16 | `reveal-js-architect` | frameworks | 10980 | ~2745 | ~93 | ~147 | ~1056 | ~0 |
| 17 | `ros2-architect` | robotics | 10867 | ~2717 | ~105 | ~607 | ~1714 | ~0 |
| 18 | `ml-conference-paper-architect` | academic | 10468 | ~2617 | ~121 | ~457 | ~1305 | ~0 |
| 19 | `rest-api-architect` | protocols | 10412 | ~2603 | ~102 | ~506 | ~0 | ~4141 |
| 20 | `observability-architect` | infra | 10238 | ~2577 | ~74 | ~546 | ~874 | ~0 |
| 21 | `api-contract-reviewer` | quality | 9798 | ~2511 | ~108 | ~0 | ~0 | ~0 |
| 22 | `grpc-architect` | protocols | 9932 | ~2483 | ~106 | ~521 | ~833 | ~0 |
| 23 | `ci-cd-architect` | infra | 9758 | ~2440 | ~114 | ~624 | ~2915 | ~0 |
| 24 | `hexagonal-arch` | design | 9537 | ~2419 | ~101 | ~0 | ~768 | ~0 |
| 25 | `feature-planner` | workflows | 9202 | ~2369 | ~114 | ~0 | ~0 | ~0 |
| 26 | `demo-presentation-architect` | personal | 9421 | ~2356 | ~101 | ~0 | ~1776 | ~3772 |
| 27 | `code-design-refactor` | refactoring | 8927 | ~2264 | ~96 | ~0 | ~0 | ~0 |
| 28 | `uru-scientific-paper-architect` | personal | 8930 | ~2233 | ~141 | ~0 | ~2021 | ~0 |
| 29 | `ui-ux-architect` | frontend | 8920 | ~2230 | ~103 | ~466 | ~1387 | ~0 |
| 30 | `sql-architect` | databases | 8686 | ~2172 | ~90 | ~295 | ~0 | ~1064 |
| 31 | `security-reviewer` | quality | 8474 | ~2119 | ~77 | ~0 | ~539 | ~0 |
| 32 | `performance-reviewer` | quality | 8350 | ~2088 | ~84 | ~0 | ~612 | ~0 |
| 33 | `secrets-architect` | infra | 7868 | ~2038 | ~132 | ~0 | ~0 | ~0 |
| 34 | `uru-thesis-architect` | personal | 7789 | ~1948 | ~114 | ~0 | ~1679 | ~4950 |
| 35 | `protobuf-architect` | encoding | 7763 | ~1941 | ~84 | ~474 | ~762 | ~0 |
| 36 | `nethttp-architect` | frameworks | 7494 | ~1928 | ~120 | ~398 | ~1508 | ~0 |
| 37 | `ddd-architect` | design | 7602 | ~1901 | ~78 | ~0 | ~0 | ~751 |
| 38 | `go-house-libs` | personal | 7477 | ~1870 | ~202 | ~1003 | ~0 | ~0 |
| 39 | `improve-codebase-architecture` | refactoring | 7437 | ~1860 | ~97 | ~0 | ~0 | ~3971 |
| 40 | `design-patterns` | refactoring | 7299 | ~1825 | ~122 | ~0 | ~0 | ~2130 |
| 41 | `fastapi-architect` | frameworks | 7220 | ~1805 | ~120 | ~420 | ~1016 | ~0 |
| 42 | `uru-thesis-defense-architect` | personal | 7191 | ~1798 | ~132 | ~0 | ~1525 | ~0 |
| 43 | `docker-architect` | infra | 7084 | ~1778 | ~89 | ~333 | ~718 | ~924 |
| 44 | `gin-architect` | frameworks | 6865 | ~1765 | ~108 | ~413 | ~1146 | ~0 |
| 45 | `skill-builder` | meta | 7030 | ~1758 | ~72 | ~0 | ~1718 | ~0 |
| 46 | `grafana-architect` | infra | 6632 | ~1658 | ~77 | ~407 | ~1149 | ~0 |
| 47 | `demo-script-architect` | personal | 4828 | ~1237 | ~96 | ~0 | ~0 | ~0 |
| 48 | `tdd` | workflows | 4279 | ~1119 | ~69 | ~0 | ~0 | ~1350 |
| 49 | `pi-iteration-workflow` | personal | 3662 | ~916 | ~108 | ~0 | ~364 | ~0 |
| 50 | `grill-with-docs` | workflows | 3376 | ~889 | ~104 | ~0 | ~0 | ~1393 |
| 51 | `commit-author` | workflows | 3208 | ~802 | ~81 | ~0 | ~0 | ~0 |
| 52 | `logic-cleaner` | refactoring | 2432 | ~608 | ~93 | ~0 | ~0 | ~0 |
| 53 | `caveman` | meta | 1687 | ~422 | ~71 | ~0 | ~0 | ~0 |

**Totals:** 505171 body bytes · ~127746 body tokens · ~5839 desc tokens · ~100356 side tokens

## Topic files

Side files in a skill directory other than `STACK.md` and `RECIPES.md` — typically topic-specific reference docs referenced by name from `SKILL.md` (e.g. `PAYLOADS.md`, `AUTH_PATTERNS.md`). These contribute to the `~Topic tkns` column above and are loaded only when the parent skill chooses to read them.

- `uru-thesis-reviewer` (~7587 tokens total):
  - `NORMAS_URU_2020.md` — ~4950 tokens
  - `TEMPLATES.md` — ~1830 tokens
  - `CHECKS.md` — ~807 tokens
- `uru-thesis-architect` (~4950 tokens total):
  - `NORMAS_URU_2020.md` — ~4950 tokens
- `rest-api-architect` (~4141 tokens total):
  - `CONCURRENCY.md` — ~1056 tokens
  - `IDEMPOTENCY.md` — ~869 tokens
  - `PAYLOADS.md` — ~787 tokens
  - `AUTH_PATTERNS.md` — ~767 tokens
  - `STATUS_CODES.md` — ~662 tokens
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
- `sql-architect` (~1064 tokens total):
  - `ENGINES.md` — ~563 tokens
  - `POSTGRES.md` — ~501 tokens
- `docker-architect` (~924 tokens total):
  - `CHECKLIST.md` — ~924 tokens
- `ddd-architect` (~751 tokens total):
  - `PATTERNS.md` — ~751 tokens

## Skills to consider trimming

Body > 2500 tokens — consider moving examples to `RECIPES.md` or topic files:

- `mcp-architect` (~7584 body tokens)
- `go-architect` (~5772 body tokens)
- `repo-tooling-architect` (~4239 body tokens)
- `go-library-builder` (~3915 body tokens)
- `python-architect` (~3368 body tokens)
- `uru-thesis-reviewer` (~3358 body tokens)
- `astro-architect` (~3252 body tokens)
- `latex-architect` (~3099 body tokens)
- `website-concept-architect` (~3083 body tokens)
- `rsk-guide` (~2941 body tokens)
- `nextjs-architect` (~2918 body tokens)
- `event-driven-architect` (~2878 body tokens)
- `react-architect` (~2875 body tokens)
- `cli-tool-architect` (~2838 body tokens)
- `hugo-architect` (~2817 body tokens)
- `reveal-js-architect` (~2745 body tokens)
- `ros2-architect` (~2717 body tokens)
- `ml-conference-paper-architect` (~2617 body tokens)
- `rest-api-architect` (~2603 body tokens)
- `observability-architect` (~2577 body tokens)
- `api-contract-reviewer` (~2511 body tokens)

Heaviest 10 descriptions — each desc token is paid every turn for any session that installs the skill:

- `mcp-architect` (~322 desc tokens)
- `go-house-libs` (~202 desc tokens)
- `rsk-guide` (~199 desc tokens)
- `go-library-builder` (~166 desc tokens)
- `uru-scientific-paper-architect` (~141 desc tokens)
- `astro-architect` (~132 desc tokens)
- `uru-thesis-defense-architect` (~132 desc tokens)
- `secrets-architect` (~132 desc tokens)
- `design-patterns` (~122 desc tokens)
- `website-concept-architect` (~122 desc tokens)

## Notes

- **Body tokens** cost only when the skill is invoked in a session.
- **Description tokens** cost every turn for any session that has the skill installed. The cost depends on which bundles are installed, not on the corpus total.
- Side files (`STACK.md`, `RECIPES.md`, topic files) are never auto-loaded; they cost 0 per turn.
- For exact counts: run each file through [`tiktoken`](https://github.com/openai/tiktoken) with `cl100k_base`.
