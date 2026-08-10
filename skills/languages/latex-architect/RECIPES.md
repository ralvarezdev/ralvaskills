# LaTeX Architect — Reference Implementations

Skeletons referenced from [SKILL.md](SKILL.md). Load on demand.

## 1. Project tree & root file

```
paper/
├── main.tex              # documentclass + preamble + \input list only
├── preamble/
│   ├── packages.tex
│   └── macros.tex
├── sections/
│   ├── introduction.tex
│   ├── method.tex
│   ├── experiments.tex
│   └── conclusion.tex
├── figures/              # .pdf preferred, .png for photographs
├── refs.bib
├── .latexmkrc
├── .chktexrc
├── .latexindent.yaml
└── .gitignore            # build/ and all generated artifacts
```

`main.tex`:

```latex
\documentclass[11pt,a4paper]{article}

\input{preamble/packages}
\input{preamble/macros}

\title{Title}
\author{Author\thanks{Affiliation}}
\date{\today}

\begin{document}
\maketitle
\begin{abstract}
  \input{sections/abstract}
\end{abstract}

\input{sections/introduction}
\input{sections/method}
\input{sections/experiments}
\input{sections/conclusion}

\printbibliography

\appendix
\input{sections/appendix}
\end{document}
```

The root file contains no prose. Every content edit happens in `sections/`, so two people editing different sections never conflict.

## 2. `.latexmkrc` & `.gitignore`

`.latexmkrc` (biblatex + biber default):

```perl
$pdf_mode        = 1;      # pdflatex
$bibtex_use      = 2;      # run biber/bibtex and clean its output on -c
$biber           = 'biber %O %S';
$out_dir         = 'build';
$emulate_aux     = 1;      # keep aux files in $out_dir under older TeX Live
$pdflatex        = 'pdflatex -interaction=nonstopmode -halt-on-error -file-line-error %O %S';

# Treat these as fatal in CI: uncomment to fail the build on any warning.
# $warnings_as_errors = 1;

$clean_ext = 'bbl nav out snm synctex.gz run.xml bcf fls fdb_latexmk';
```

For `lualatex`, replace `$pdf_mode = 1` with `$pdf_mode = 4`. For `xelatex`, `$pdf_mode = 5`.

`.gitignore`:

```gitignore
build/
*.aux
*.bbl
*.bcf
*.blg
*.fdb_latexmk
*.fls
*.log
*.out
*.run.xml
*.synctex.gz
*.toc
*.lof
*.lot
```

Commit `.latexmkrc`, `.chktexrc`, `.latexindent.yaml`, `refs.bib`, and everything under `sections/` and `figures/`.

## 3. Preamble skeleton

`preamble/packages.tex` — order matters; see [SKILL.md §4](SKILL.md#4-preamble-discipline).

```latex
% --- encoding & fonts -------------------------------------------------
\usepackage[T1]{fontenc}
\usepackage{lmodern}
\usepackage{microtype}

% --- layout -----------------------------------------------------------
\usepackage[margin=2.5cm]{geometry}

% --- math -------------------------------------------------------------
\usepackage{amsmath, amssymb}
\usepackage{mathtools}          % must follow amsmath

% --- content ----------------------------------------------------------
\usepackage{graphicx}
\graphicspath{{figures/}}
\usepackage{booktabs}
\usepackage{siunitx}
\usepackage[font=small,labelfont=bf]{caption}
\usepackage{subcaption}         % never `subfigure`

% --- bibliography -----------------------------------------------------
\usepackage[backend=biber,style=numeric,sorting=nyt,maxbibnames=99]{biblatex}
\addbibresource{refs.bib}

% --- links (second-to-last) -------------------------------------------
\usepackage[hidelinks,breaklinks=true]{hyperref}

% --- cross-references (LAST, after hyperref) --------------------------
\usepackage[capitalise,noabbrev]{cleveref}
```

While drafting, add `\usepackage[l2tabu, orthodox]{nag}` at the very top and remove it before submission.

`preamble/macros.tex` — semantic only:

```latex
\newcommand{\model}{\textsc{Orion}}          % rename in one place
\newcommand{\dataset}{\textsc{WikiBench}}
\DeclareMathOperator*{\argmax}{arg\,max}
\DeclareMathOperator*{\argmin}{arg\,min}
\newcommand{\R}{\mathbb{R}}
```

## 4. Bibliography configuration

**biblatex + biber (default).** Loaded as shown in §3; emit with `\printbibliography` and cite with `\autocite{key}` / `\textcite{key}`.

A well-formed entry, with capitals brace-protected and the published venue rather than the preprint:

```bibtex
@inproceedings{vaswani2017attention,
  author    = {Vaswani, Ashish and Shazeer, Noam and Parmar, Niki and others},
  title     = {Attention Is All You Need},
  booktitle = {Advances in Neural Information Processing Systems 30 ({NIPS} 2017)},
  year      = {2017},
  pages     = {5998--6008},
}
```

**natbib carve-out** — for venue templates that ship `natbib` + a `.bst`. Do not convert these to biblatex.

```latex
\usepackage[numbers,sort&compress]{natbib}   % usually already in the venue .sty
\bibliographystyle{plainnat}                 % or the venue's .bst
\bibliography{refs}
```

Matching `.latexmkrc` line — bibtex instead of biber:

```perl
$bibtex_use = 2;
$biber      = '';      # unset; latexmk falls back to bibtex via the .aux
```

Citation commands differ: `\citep{x}` parenthetical, `\citet{x}` textual. `\autocite` does not exist under natbib.

## 5. Quality gates & CI

`.chktexrc` — suppress only what you have deliberately reviewed:

```
% Suppress: 8 (wrong dash), 46 (\rm etc.) are kept ON deliberately.
CmdLine { -n 3 }   % "You should put a space in front of parenthesis" — noisy
```

`.latexindent.yaml`:

```yaml
defaultIndent: "  "
onlyOneBackUp: 1
maximumIndentation: -1
removeTrailingWhitespace:
  beforeProcessing: 1
  afterProcessing: 1
modifyLineBreaks:
  oneSentencePerLine:
    manipulateSentences: 1
```

GitHub Actions build + lint. Uses the TeX Live container so the toolchain is pinned rather than installed ad hoc:

```yaml
name: latex
on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    container: texlive/texlive:latest
    steps:
      - uses: actions/checkout@v5
      - name: Lint
        run: chktex -q -l .chktexrc main.tex sections/*.tex
      - name: Build
        run: latexmk -pdf main.tex
      - uses: actions/upload-artifact@v4
        with:
          name: paper
          path: build/main.pdf
```

Pin the container to a dated tag (`texlive/texlive:TL2026-2026-04-01`) once the document is near submission — `latest` will roll to the next TeX Live and can change output.

## 6. Table and figure patterns

Table with `booktabs` + `siunitx` column alignment on the decimal point:

```latex
\begin{table}[t]
  \centering
  \caption{Accuracy on the held-out split.}
  \label{tab:results}
  \begin{tabular}{l S[table-format=2.1] S[table-format=1.2]}
    \toprule
    Model & {Accuracy (\si{\percent})} & {Latency (\si{\second})} \\
    \midrule
    Baseline & 81.4 & 1.20 \\
    \model   & 88.7 & 0.94 \\
    \bottomrule
  \end{tabular}
\end{table}
```

Note: `\caption` then `\label`, in that order. `S` columns need their headers in braces to opt out of numeric parsing.

Side-by-side figures with `subcaption`:

```latex
\begin{figure}[t]
  \centering
  \begin{subfigure}[b]{0.48\linewidth}
    \includegraphics[width=\linewidth]{loss}
    \caption{Training loss.}
    \label{fig:loss}
  \end{subfigure}
  \hfill
  \begin{subfigure}[b]{0.48\linewidth}
    \includegraphics[width=\linewidth]{acc}
    \caption{Validation accuracy.}
    \label{fig:acc}
  \end{subfigure}
  \caption{Training dynamics.}
  \label{fig:dynamics}
\end{figure}
```

Reference with `\cref{fig:loss,fig:acc}`, which renders as "Figures 1a and 1b" without you writing either word.
