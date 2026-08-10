# Venue Templates

Where to get the official style files, and how to approximate the classic NIPS 2017 look for documents that are not venue submissions.

## Why nothing is vendored here

Style files are **not copied into this skill**, deliberately:

- They are revised annually. A vendored copy is stale the moment the next author kit ships, and a stale `.sty` produces a paper that compiles cleanly and fails the venue's format check.
- Venues distribute them under their own terms as part of an author kit that also contains the checklist, the `.bst`, and a reference `main.tex`. Taking the `.sty` alone loses the rest.

Download per submission, from the venue, for the target year.

## Official sources

| Venue | Where |
|---|---|
| NeurIPS | Author kit linked from the current CFP at `neurips.cc` — provides `neurips_<year>.sty`, `neurips_<year>.tex`, and the paper checklist |
| ICML | Author instructions at `icml.cc` — `icml<year>.sty`, `icml<year>.bst`, example paper |
| ICLR | `github.com/ICLR/Master-Template` — per-year `iclr<year>_conference.sty` and `.bst` |
| AAAI | AAAI author kit from the current AAAI conference site — includes `aaai<yy>.sty`, `aaai<yy>.bst`, and the formatting instructions PDF, which must be read in full |

Verification before you start writing: compile the venue's **unmodified example paper** first. If that fails, the problem is your toolchain, not your document — solve it before there is any content to lose.

## Classic NIPS 2017 look

The single-column Times layout of the NIPS 2017 proceedings — the *Attention Is All You Need* look — is a common target for preprints, technical reports, and internal write-ups where no venue template applies.

**This is an approximation, not the real `nips_2017.sty`.** It reproduces the page geometry and font choice, not the exact title block, author formatting, or footer. Two consequences:

- For an **actual submission to any venue**, this is wrong by definition — use that venue's current style file.
- For a **preprint or report**, prefer a current venue template in `preprint` mode (NeurIPS supports `\usepackage[preprint]{neurips_<year>}`) — it gets you a nearly identical look, actively maintained, with a working title block.

Use the approximation only when neither applies:

```latex
\documentclass[10pt]{article}

\usepackage[letterpaper,
            textwidth=5.5in, textheight=9in,
            top=1in, headheight=12pt, headsep=25pt, footskip=30pt]{geometry}

\usepackage{times}                 % the 2017 look; newtxtext/newtxmath is the modern equivalent
\usepackage[T1]{fontenc}
\usepackage{microtype}
\usepackage{amsmath, amssymb, mathtools}
\usepackage{graphicx}
\usepackage{booktabs}
\usepackage[numbers]{natbib}
\usepackage{hyperref}
\usepackage[capitalise,noabbrev]{cleveref}
```

Note that `\usepackage{times}` is on the obsolete-font list in
[latex-architect §4](../../../languages/latex-architect/SKILL.md#4-preamble-discipline). It is used here only because it is what the original template used and what reproduces the look; `newtxtext` + `newtxmath` is the correct choice for a new document that merely wants Times.
