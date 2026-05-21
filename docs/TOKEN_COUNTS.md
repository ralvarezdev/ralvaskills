# SKILL.md Token Estimates

> Auto-generated. **Do not edit by hand.** Run `go run ./scripts/count-tokens` to refresh.
>
> Estimate: ~4 chars/token (Claude tokenizer, English markdown). Actual range ±15%.

_Last updated: 2026-05-21 18:11 UTC · 26 skills_

## Load model

| What | When loaded | Estimated tokens |
|---|---|---:|
| All `description:` fields (skill index) | **Every turn** | ~1765 |
| All `SKILL.md` bodies | Only when skill is invoked | ~51369 |
| Everything combined | If all skills active at once | ~51369 |

## Per skill

| # | Skill | Category | Body bytes | ~Body tokens | ~Desc tokens |
|---|---|---|---:|---:|---:|
| 1 | `skill-builder` | tooling | 14171 | ~3542 | ~80 |
| 2 | `design-patterns` | workflows | 13612 | ~3403 | ~86 |
| 3 | `cli-tool-architect` | tooling | 12651 | ~3162 | ~69 |
| 4 | `rest-api-architect` | protocols | 11708 | ~2927 | ~76 |
| 5 | `ddd-architect` | design | 10761 | ~2690 | ~82 |
| 6 | `grpc-architect` | protocols | 10308 | ~2577 | ~79 |
| 7 | `repo-tooling-architect` | tooling | 10307 | ~2576 | ~68 |
| 8 | `feature-planner` | workflows | 9419 | ~2354 | ~68 |
| 9 | `hexagonal-arch` | design | 9328 | ~2332 | ~76 |
| 10 | `code-design-refactor` | workflows | 9275 | ~2318 | ~71 |
| 11 | `sql-architect` | databases | 8890 | ~2222 | ~67 |
| 12 | `protobuf-architect` | encoding | 8077 | ~2019 | ~63 |
| 13 | `nethttp-architect` | frameworks | 7817 | ~1954 | ~65 |
| 14 | `improve-codebase-architecture` | workflows | 7722 | ~1930 | ~72 |
| 15 | `fastapi-architect` | frameworks | 7550 | ~1887 | ~67 |
| 16 | `go-architect` | languages | 7332 | ~1833 | ~55 |
| 17 | `gin-architect` | frameworks | 7151 | ~1787 | ~57 |
| 18 | `docker-architect` | infra | 7090 | ~1772 | ~67 |
| 19 | `python-architect` | languages | 6672 | ~1668 | ~59 |
| 20 | `demo-script-architect` | personal | 5181 | ~1295 | ~72 |
| 21 | `rsk-guide` | tooling | 4615 | ~1153 | ~60 |
| 22 | `tdd` | workflows | 4532 | ~1133 | ~51 |
| 23 | `grill-with-docs` | workflows | 3723 | ~930 | ~72 |
| 24 | `commit-author` | workflows | 2898 | ~724 | ~60 |
| 25 | `logic-cleaner` | workflows | 2772 | ~693 | ~70 |
| 26 | `caveman` | workflows | 1952 | ~488 | ~53 |

**Totals:** 205514 bytes · ~51369 body tokens · ~1765 description tokens (always-loaded)

## Notes

- **Body tokens** cost only when the skill is invoked in a session.
- **Description tokens** cost on every single turn — keep descriptions short.
- `RECIPES.md` and other side files are never auto-loaded; they cost 0 per turn.
- For exact counts: run each file through [`tiktoken`](https://github.com/openai/tiktoken) with `cl100k_base`.
