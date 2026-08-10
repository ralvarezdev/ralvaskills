# Stack Versions

Venue style files are released annually. This table records the template generation this skill was written against — it exists so `rsk status --stack` and a yearly manual review surface drift before a deadline does.

| Dependency | Pinned version | Purpose |
|---|---|---|
| neurips style | `neurips_2025.sty` | NeurIPS submission/final/preprint modes |
| icml style | `icml2025.sty` | ICML submission and `[accepted]` modes |
| iclr style | `iclr2025_conference.sty` | ICLR submission and `\iclrfinalcopy` |
| aaai style | `aaai25.sty` | AAAI; strictest package and font constraints |
| natbib | 8.31b | Citation package shipped by all four venue templates |
| latexdiff | CTAN 2026-01-02 | Marked-up revision diffs for rebuttals (CTAN publishes no version number) |

Base LaTeX toolchain (TeX Live, latexmk, biber, chktex) is inherited from
[latex-architect/STACK.md](../../languages/latex-architect/STACK.md) and not duplicated here.

## Notes

- **These style files change every year and the page limits change with them.** Treat this table as a staleness marker, not as authority — the current CFP is authoritative, per [SKILL.md §1](SKILL.md#1-the-one-rule-that-overrides-everything).
- **`natbib`, not `biblatex`** — this is the explicit carve-out from [latex-architect §8](../../languages/latex-architect/SKILL.md#8-bibliography). Venue templates ship `natbib` plus a matching `.bst`; converting them breaks the venue's citation rendering.
- All four venues assume **pdfLaTeX**. AAAI additionally requires Type 1 fonts and enforces embedding mechanically.
- Style files are **not vendored into this skill** — download them per submission from the official sources in [assets/VENUE_TEMPLATES.md](assets/VENUE_TEMPLATES.md).

_Last reviewed: 2026-07-30_
_Skill version at last review: 1.0.0_
