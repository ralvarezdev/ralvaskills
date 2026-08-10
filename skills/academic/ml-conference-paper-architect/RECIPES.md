# ML Conference Paper Architect — Reference Implementations

Skeletons referenced from [SKILL.md](SKILL.md). General LaTeX skeletons (project tree, `.latexmkrc`, CI) live in [latex-architect/RECIPES.md](../../languages/latex-architect/RECIPES.md) and are not repeated here.

## 1. Root file per venue

The structure is identical across venues; only the style-file line and the final-mode toggle change. Download the `.sty` from the official source — see [assets/VENUE_TEMPLATES.md](assets/VENUE_TEMPLATES.md).

**NeurIPS.** One option controls all three modes; the default (no option) is anonymous submission:

```latex
\documentclass{article}

\usepackage{neurips_2025}            % submission: anonymous
% \usepackage[final]{neurips_2025}   % camera-ready: de-anonymized
% \usepackage[preprint]{neurips_2025} % arXiv preprint

\usepackage[T1]{fontenc}
\usepackage{microtype}
\usepackage{amsmath, amssymb, mathtools}
\usepackage{graphicx}
\usepackage{booktabs}
\usepackage{hyperref}
\usepackage[capitalise,noabbrev]{cleveref}   % after hyperref

\title{Title}
\author{%
  Anonymous Author\\
  Anonymous Institution\\
  \texttt{anon@example.com}
}

\begin{document}
\maketitle
\input{sections/abstract}
\input{sections/introduction}
% ...
\bibliographystyle{plainnat}
\bibliography{refs}
\appendix
\input{sections/appendix}
\end{document}
```

**ICML.** Accepted mode is a package option, and the anonymity declaration is a separate command:

```latex
\usepackage{icml2025}              % submission
% \usepackage[accepted]{icml2025}  % camera-ready
\icmltitlerunning{Short Title}
```

**ICLR.** Final mode is a command, not an option — putting it in the wrong place silently does nothing:

```latex
\usepackage{iclr2025_conference, times}
% \iclrfinalcopy   % uncomment for camera-ready ONLY
```

**AAAI.** The most constrained: no `geometry`, no `fullpage`, no `authblk`, Type 1 fonts only, and a mandatory `\pdfinfo` block. Copy AAAI's own skeleton verbatim from the author kit rather than adapting one of the above.

## 2. Contributions list

Place at the end of the introduction. Each bullet is a claim the paper substantiates, with a pointer to where:

```latex
\paragraph{Contributions.}
\begin{itemize}
  \item We identify <specific gap> in existing <approach class> (\cref{sec:related}).
  \item We propose \model, which <one-sentence mechanism> (\cref{sec:method}).
  \item We show \model{} improves <metric> by <number> over <baseline> on
        <datasets>, and ablate each component (\cref{sec:experiments}).
\end{itemize}
```

Note `\model{}` with braces before a space — without them the macro swallows the following space.

## 3. Results table

Full `booktabs` + `siunitx` conventions in [latex-architect RECIPES §6](../../languages/latex-architect/RECIPES.md#6-table-and-figure-patterns). Venue-specific additions:

```latex
\begin{table}[t]
  \centering
  \caption{\model{} outperforms all baselines on both benchmarks.
           Mean $\pm$ std over 5 seeds; best in \textbf{bold}.}
  \label{tab:main}
  \begin{tabular}{lcc}
    \toprule
    Method & Bench-A & Bench-B \\
    \midrule
    Baseline~\citep{smith2023}  & $81.4 \pm 0.6$ & $74.2 \pm 1.1$ \\
    \model{} (ours)             & $\mathbf{88.7 \pm 0.4}$ & $\mathbf{80.1 \pm 0.9}$ \\
    \bottomrule
  \end{tabular}
\end{table}
```

Three things reviewers check on sight and that this encodes: variance reported, seed count stated, and a takeaway-led caption rather than "Results".

## 4. Pre-submission checklist

Mechanical checks, runnable from the repo root before upload:

```bash
# Page count of the main text (subtract appendix pages manually)
pdfinfo build/main.pdf | grep Pages

# Every font embedded — AAAI rejects on non-embedded fonts
pdffonts build/main.pdf | awk 'NR>2 && $NF=="no" {print "NOT EMBEDDED:", $1}'

# Author metadata leaked from the editor or OS
pdfinfo build/main.pdf | grep -Ei 'Author|Creator|Producer|Title'

# Unresolved references compile to "?" — catch them in the log, not the PDF
grep -Ei 'undefined (reference|citation)|LaTeX Warning: Reference' build/main.log

# Deanonymization slips in the source
grep -rniE 'our (previous|prior|earlier) work|github\.com/|acknowledg' sections/
```

Strip leaked metadata at build time by setting it explicitly rather than by post-processing:

```latex
\hypersetup{pdfauthor={}, pdftitle={}, pdfsubject={}, pdfkeywords={}}
```

Then the manual pass:

- [ ] Style file year matches the target venue's current year.
- [ ] Submission mode active (final/accepted toggle commented out).
- [ ] Figures legible at 100% print size and in grayscale.
- [ ] Every `\citep`/`\citet` renders correctly (no `[?]`).
- [ ] Checklist / impact / limitations sections present and consistent with the body.
- [ ] Clean checkout compiles with `latexmk` alone.

## 5. Reviewer response letter

For rebuttals and revisions, `latexdiff` produces a marked-up PDF showing exactly what changed between submission and revision — reviewers consistently rate this as reducing effort:

```bash
latexdiff --flatten submitted/main.tex revised/main.tex > diff.tex
latexmk -pdf diff.tex
```

`--flatten` is required when the document uses `\input`, otherwise only the root file is diffed.
