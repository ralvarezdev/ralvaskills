# SKILL.md Token Estimates

> Auto-generated. **Do not edit by hand.** Run `task tokens` to refresh.
>
> Estimate: ~4 bytes/token (Claude tokenizer, English markdown). Actual range ±15%.

_Last updated: 2026-05-22 00:25 UTC · 36 skills_

## Load model

| What | When loaded | Estimated tokens |
|---|---|---:|
| All `description:` fields (skill index) | **Every turn** | ~2589 |
| All `SKILL.md` bodies | Only when skill is invoked | ~83053 |
| All side files (`STACK` + `RECIPES`) | On-demand only | ~18586 |
| Everything combined | Absolute maximum | ~101639 |

## Per skill

| # | Skill | Category | Body bytes | ~Body tkns | ~Desc tkns | ~Stack tkns | ~Recipes tkns |
|---|---|---|---:|---:|---:|---:|---:|
| 1 | `event-driven-architect` | messaging | 14295 | ~3573 | ~85 | ~513 | ~0 |
| 2 | `skill-builder` | tooling | 14171 | ~3542 | ~80 | ~0 | ~0 |
| 3 | `nextjs-architect` | frameworks | 13848 | ~3462 | ~80 | ~454 | ~878 |
| 4 | `ui-ux-architect` | frontend | 13700 | ~3425 | ~77 | ~466 | ~0 |
| 5 | `design-patterns` | workflows | 13612 | ~3403 | ~86 | ~0 | ~0 |
| 6 | `observability-architect` | infra | 13503 | ~3375 | ~84 | ~545 | ~0 |
| 7 | `react-architect` | frameworks | 13485 | ~3371 | ~76 | ~464 | ~720 |
| 8 | `ros2-architect` | robotics | 13240 | ~3310 | ~79 | ~606 | ~936 |
| 9 | `cli-tool-architect` | tooling | 12651 | ~3162 | ~69 | ~590 | ~698 |
| 10 | `performance-reviewer` | quality | 11870 | ~2967 | ~79 | ~0 | ~0 |
| 11 | `rest-api-architect` | protocols | 11708 | ~2927 | ~76 | ~427 | ~0 |
| 12 | `grafana-architect` | infra | 11467 | ~2866 | ~86 | ~407 | ~0 |
| 13 | `security-reviewer` | quality | 11156 | ~2789 | ~98 | ~0 | ~0 |
| 14 | `ddd-architect` | design | 10761 | ~2690 | ~82 | ~0 | ~0 |
| 15 | `grpc-architect` | protocols | 10308 | ~2577 | ~79 | ~521 | ~832 |
| 16 | `repo-tooling-architect` | tooling | 10307 | ~2576 | ~68 | ~479 | ~0 |
| 17 | `api-contract-reviewer` | quality | 10184 | ~2546 | ~80 | ~0 | ~0 |
| 18 | `feature-planner` | workflows | 9419 | ~2354 | ~68 | ~0 | ~0 |
| 19 | `hexagonal-arch` | design | 9328 | ~2332 | ~76 | ~0 | ~767 |
| 20 | `code-design-refactor` | workflows | 9275 | ~2318 | ~71 | ~0 | ~0 |
| 21 | `sql-architect` | databases | 8890 | ~2222 | ~67 | ~295 | ~0 |
| 22 | `protobuf-architect` | encoding | 8077 | ~2019 | ~63 | ~474 | ~762 |
| 23 | `nethttp-architect` | frameworks | 7817 | ~1954 | ~65 | ~397 | ~1508 |
| 24 | `improve-codebase-architecture` | workflows | 7722 | ~1930 | ~72 | ~0 | ~0 |
| 25 | `fastapi-architect` | frameworks | 7550 | ~1887 | ~67 | ~420 | ~1015 |
| 26 | `go-architect` | languages | 7332 | ~1833 | ~55 | ~385 | ~0 |
| 27 | `gin-architect` | frameworks | 7151 | ~1787 | ~57 | ~413 | ~1145 |
| 28 | `docker-architect` | infra | 7090 | ~1772 | ~67 | ~333 | ~717 |
| 29 | `python-architect` | languages | 6672 | ~1668 | ~59 | ~419 | ~0 |
| 30 | `demo-script-architect` | personal | 5181 | ~1295 | ~72 | ~0 | ~0 |
| 31 | `rsk-guide` | tooling | 4615 | ~1153 | ~60 | ~0 | ~0 |
| 32 | `tdd` | workflows | 4532 | ~1133 | ~51 | ~0 | ~0 |
| 33 | `grill-with-docs` | workflows | 3723 | ~930 | ~72 | ~0 | ~0 |
| 34 | `commit-author` | workflows | 2898 | ~724 | ~60 | ~0 | ~0 |
| 35 | `logic-cleaner` | workflows | 2772 | ~693 | ~70 | ~0 | ~0 |
| 36 | `caveman` | workflows | 1952 | ~488 | ~53 | ~0 | ~0 |

**Totals:** 332262 body bytes · ~83053 body tokens · ~2589 desc tokens · ~18586 side tokens

## Notes

- **Body tokens** cost only when the skill is invoked in a session.
- **Description tokens** cost on every single turn — keep descriptions short.
- `STACK.md` and `RECIPES.md` are never auto-loaded; they cost 0 per turn.
- For exact counts: run each file through [`tiktoken`](https://github.com/openai/tiktoken) with `cl100k_base`.
