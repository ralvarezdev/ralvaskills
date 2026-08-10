# Stack Versions

| Dependency | Pinned version | Purpose |
|---|---|---|
| texlive | 2026 | Distribution; default toolchain for local and CI builds |
| latexmk | 4.88 | Sole build entry point; handles rerun convergence |
| biber | 2.21 | Bibliography backend for biblatex |
| biblatex | 3.21 | Bibliography package (pairs with biber 2.21) |
| tectonic | 0.17.0 | Hermetic alternative engine for reproducible CI builds |
| latexindent | 4.0.2 | Deterministic source formatting (pre-commit) |
| chktex | 1.7.10 | Static linter for mechanical LaTeX errors (CI) |

## Notes

- **biblatex and biber versions are coupled** — biber 2.21 requires biblatex 3.21. A mismatch fails with a version-check error, not a subtle bug. Upgrade both together, and pin both in CI.
- Engine default is `pdflatex` (TeX Live 2026); `lualatex` is the switch target for system fonts or native Unicode. Conference style files frequently assume pdfLaTeX — see [ml-conference-paper-architect](../../academic/ml-conference-paper-architect/SKILL.md).
- Tectonic is listed as an alternative, not the default: package coverage is still incomplete and some venue `.sty` files fail on it.
- TeX Live ships a new release each spring; the `texlive/texlive:latest` container tracks it. Pin to a dated tag near submission.

_Last reviewed: 2026-07-30_
_Skill version at last review: 1.0.0_
