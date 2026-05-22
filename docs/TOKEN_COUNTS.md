# SKILL.md Token Estimates

> Auto-generated. **Do not edit by hand.** Run `task tokens` to refresh.
>
> Estimate: ~4 bytes/token (Claude tokenizer, English markdown). Actual range ±15%.

_Last updated: 2026-05-22 01:59 UTC · 39 skills_

## Load model

| What | When loaded | Estimated tokens |
|---|---|---:|
| All `description:` fields (skill index) | **Every turn** | ~2582 |
| All `SKILL.md` bodies | Only when skill is invoked | ~79446 |
| All side files (`STACK` + `RECIPES`) | On-demand only | ~33167 |
| Everything combined | Absolute maximum | ~112613 |

## Per skill

| # | Skill | Category | Body bytes | ~Body tkns | ~Desc tkns | ~Stack tkns | ~Recipes tkns |
|---|---|---|---:|---:|---:|---:|---:|
| 1 | `nextjs-architect` | frameworks | 12054 | ~3013 | ~80 | ~454 | ~1824 |
| 2 | `react-architect` | frameworks | 11859 | ~2964 | ~76 | ~464 | ~1425 |
| 3 | `event-driven-architect` | messaging | 11809 | ~2952 | ~62 | ~513 | ~2147 |
| 4 | `cli-tool-architect` | tooling | 11675 | ~2918 | ~69 | ~590 | ~1435 |
| 5 | `ros2-architect` | robotics | 11037 | ~2759 | ~79 | ~606 | ~1714 |
| 6 | `uru-thesis-reviewer` | personal | 10897 | ~2724 | ~73 | ~0 | ~0 |
| 7 | `observability-architect` | infra | 10528 | ~2632 | ~55 | ~545 | ~873 |
| 8 | `rest-api-architect` | protocols | 10373 | ~2593 | ~76 | ~427 | ~0 |
| 9 | `grpc-architect` | protocols | 10308 | ~2577 | ~79 | ~521 | ~832 |
| 10 | `repo-tooling-architect` | tooling | 10307 | ~2576 | ~68 | ~479 | ~0 |
| 11 | `api-contract-reviewer` | quality | 10184 | ~2546 | ~80 | ~0 | ~0 |
| 12 | `demo-presentation-architect` | personal | 9796 | ~2449 | ~76 | ~0 | ~1775 |
| 13 | `feature-planner` | workflows | 9419 | ~2354 | ~68 | ~0 | ~0 |
| 14 | `hexagonal-arch` | design | 9328 | ~2332 | ~76 | ~0 | ~767 |
| 15 | `ui-ux-architect` | frontend | 9283 | ~2320 | ~77 | ~466 | ~1387 |
| 16 | `code-design-refactor` | workflows | 9275 | ~2318 | ~71 | ~0 | ~0 |
| 17 | `sql-architect` | databases | 8890 | ~2222 | ~67 | ~295 | ~0 |
| 18 | `security-reviewer` | quality | 8765 | ~2191 | ~58 | ~0 | ~538 |
| 19 | `performance-reviewer` | quality | 8663 | ~2165 | ~62 | ~0 | ~611 |
| 20 | `work-report-generator` | personal | 8605 | ~2151 | ~61 | ~0 | ~1239 |
| 21 | `protobuf-architect` | encoding | 8077 | ~2019 | ~63 | ~474 | ~762 |
| 22 | `ddd-architect` | design | 7892 | ~1973 | ~58 | ~0 | ~0 |
| 23 | `nethttp-architect` | frameworks | 7817 | ~1954 | ~65 | ~397 | ~1508 |
| 24 | `improve-codebase-architecture` | workflows | 7722 | ~1930 | ~72 | ~0 | ~0 |
| 25 | `design-patterns` | workflows | 7587 | ~1896 | ~57 | ~0 | ~0 |
| 26 | `fastapi-architect` | frameworks | 7550 | ~1887 | ~67 | ~420 | ~1015 |
| 27 | `go-architect` | languages | 7332 | ~1833 | ~55 | ~385 | ~0 |
| 28 | `skill-builder` | tooling | 7303 | ~1825 | ~54 | ~0 | ~1696 |
| 29 | `gin-architect` | frameworks | 7151 | ~1787 | ~57 | ~413 | ~1145 |
| 30 | `docker-architect` | infra | 7090 | ~1772 | ~67 | ~333 | ~717 |
| 31 | `grafana-architect` | infra | 6922 | ~1730 | ~57 | ~407 | ~1149 |
| 32 | `python-architect` | languages | 6672 | ~1668 | ~59 | ~419 | ~0 |
| 33 | `demo-script-architect` | personal | 5181 | ~1295 | ~72 | ~0 | ~0 |
| 34 | `rsk-guide` | tooling | 4615 | ~1153 | ~60 | ~0 | ~0 |
| 35 | `tdd` | workflows | 4532 | ~1133 | ~51 | ~0 | ~0 |
| 36 | `grill-with-docs` | workflows | 3723 | ~930 | ~72 | ~0 | ~0 |
| 37 | `commit-author` | workflows | 2898 | ~724 | ~60 | ~0 | ~0 |
| 38 | `logic-cleaner` | workflows | 2772 | ~693 | ~70 | ~0 | ~0 |
| 39 | `caveman` | workflows | 1952 | ~488 | ~53 | ~0 | ~0 |

**Totals:** 317843 body bytes · ~79446 body tokens · ~2582 desc tokens · ~33167 side tokens

## Notes

- **Body tokens** cost only when the skill is invoked in a session.
- **Description tokens** cost on every single turn — keep descriptions short.
- `STACK.md` and `RECIPES.md` are never auto-loaded; they cost 0 per turn.
- For exact counts: run each file through [`tiktoken`](https://github.com/openai/tiktoken) with `cl100k_base`.
