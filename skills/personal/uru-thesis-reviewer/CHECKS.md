# URU Thesis Reviewer — Detailed Check Catalog

Per-dimension checklists referenced by [SKILL.md](SKILL.md) §6. Substantive checks (§6.1–6.2 in SKILL.md) stay in the body because they drive the workflow; this file holds the formal/linguistic dimensions (§6.3–6.8 of the original spec) that get pulled when reviewing prose.

## 1. Formal aspects (URU §IV)

- **Font** — Arial 12 pt body. Only checkable from `.docx` directly; flag as "verify manually" if reviewing `.md`.
- **Margins** — 3 cm; 4 cm top on capitular pages.
- **Line spacing** — 1.5 body; 1.0 for citas largas, tablas, refs; 3.0 around RESUMEN/ABSTRACT.
- **Page numbering** — top-right, Arial; no number on preliminary or capitular pages.
- **Title formatting** — sentence case, bold, left-aligned, no terminal period, max 4 numbering levels.
- **Table/figure** — tabla → caption above; figura → caption below; numbered consecutively; "Continuación" on multi-page.

## 2. Redacción (URU §III)

- **Third person.** Flag every `yo/nosotros/mi/nuestro` — propose `el autor / los autores` or impersonal.
- **Verb tense.** Past for TEG, future for anteproyecto. Flag mixed tenses within the same section.
- **Acronyms.** First use per chapter must spell out, then `(SIGLAS)`. Flag undefined siglas.
- **No abbreviations in body** (allowed in citas, tablas, figuras).
- **Avoid uppercase abuse** in titles.
- **Style consistency** — viñetas style, citation form, terminology.

## 3. Grammar & orthography (RAE)

- Accents (tildes), punctuation, agreement (concordancia), preposition use.
- Anglicisms — flag and propose Spanish equivalents unless the term is a domain standard.
- Run-on sentences (>40 words) — propose splits.
- Passive abuse — propose active alternatives only when they don't violate third-person rule.

## 4. Semantics & clarity

- Ambiguous referents ("esto", "lo anterior" without clear antecedent).
- Undefined technical terms used before introduction.
- Logical leaps — claim → evidence chain breaks.
- Repetitive ideas across paragraphs.
- Section openings without bridge from previous section.

## 5. Citations (URU §V)

- Every claim that isn't original work has author + year.
- Format matches case (1 author / 2 / 3–5 / 6+ / group / no author / no date).
- Textual citations < 40 words: in-line, quoted, with `(Author, year, p.X)`.
- Textual citations > 40 words: block, 1 cm sangría both sides, single-spaced, no quotes.
- Cross-check: every in-text citation appears in `Referencias` and vice versa.

## 6. Bibliography (URU §VI)

- Hanging indent (sangría francesa).
- Alphabetical by first author last name.
- Same author multiple works → chronological, oldest first.
- Format per source type (book / journal / thesis / electronic / legal) per URU Table 4.
- **Source quality** — for each source, ask:
  - Peer-reviewed or institutional?
  - Recent enough for the claim (< 10 years for tech, < 5 for fast-moving fields)?
  - Primary source, not a textbook summary, when the claim is specific?
  - Verifiable (DOI, URL that resolves, ISBN)?
- For weak sources, attempt to find a replacement via `find-docs` or `WebSearch`. Cite specifically (full reference) so the author can verify.
