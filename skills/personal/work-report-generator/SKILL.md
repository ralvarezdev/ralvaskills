---
name: work-report-generator
version: 1.0.0
description: Generate formal daily work reports from unstructured input. Asks output language once, never infers tasks or hours, keeps `reports/projects.md` + `reports/YYYY-MM-DD/{LOG,REPORT}.md`. Use on "reporte de trabajo", "daily report", or /work-report.
---

# Work Report Generator

Turn unstructured daily notes and commit messages into a formal `reports/YYYY-MM-DD/REPORT.md`. Strict no-inference: every task, project assignment, and per-project hour count comes from the user, asked explicitly. Report template, anti-patterns, and closing checklist in [RECIPES.md](RECIPES.md).

## 1. When to use

Daily or backfilled work reports written in a formal tone, in any language. Triggered by phrases like "generar reporte", "reporte de trabajo", "work report", "daily report", "reporte del día", or the `/work-report` slash command.

Skip this skill for: changelogs, release notes, retrospectives, CV/resume writing, stand-up summaries — those have different structures and audiences.

## 2. File layout

All artifacts live under `reports/` at the repository root. Each day gets its own folder.

```
reports/
├── projects.md            persistent catalog of general projects (one bullet per project, with short description)
└── YYYY-MM-DD/
    ├── LOG.md             per-day scratchpad: raw user input, commit dumps, tasks grouped by project, hours per project
    └── REPORT.md          formal report, generated/regenerated from LOG.md
```

- **`projects.md`** is the single source of truth for project names across all reports. Reuse names verbatim so reports stay linkable.
- **`LOG.md`** is the working scratchpad — append-only during a session. Paste raw input, record clarification answers, log per-project hours. It is the *input* to the formal report.
- **`REPORT.md`** is the *output*. Regenerate from `LOG.md` at the end of the session. Never edit it by hand mid-session.

Create `reports/`, `projects.md`, and the per-day folder lazily on first invocation. Do not pre-seed `projects.md` with example data.

## 3. Conversation flow (strict order)

Follow this order every time. Do not skip steps even if inputs are obviously inferable — inference is forbidden (§5).

### 3.1 Language (once per session)

Ask: **"¿En qué idioma quieres el reporte?"** (default question in Spanish; switch to English if the user opened the conversation in English).

- Store the answer for the rest of the session. Do not re-ask for subsequent reports in the same session.
- If the user changes language mid-session, honor the new choice for new reports only — do not retroactively rewrite past ones unless asked.

### 3.2 Date

Ask: **"¿Para qué fecha es el reporte?"** Accept `YYYY-MM-DD`, "today", "yesterday", or a relative reference. Convert to absolute `YYYY-MM-DD` and confirm before continuing.

### 3.3 Project catalog review

1. Read `reports/projects.md`. If missing or empty, state explicitly: *"El catálogo de proyectos está vacío. Vamos a construirlo conforme aparezcan en tus notas."*
2. If non-empty, print the current list and ask: **"¿Agregar, modificar o eliminar algún proyecto antes de empezar?"**
3. Apply changes to `projects.md` before moving on. Each entry is one line: `- <Project Name> — <one-line description>`.

### 3.4 Raw input collection

Ask the user to paste:

- Free-form comments about what they did (any order, any granularity).
- Optionally, commit subjects/SHAs (only as *input* — see §6).

Append everything verbatim to `LOG.md` under a `## Raw input` section.

### 3.5 Task extraction and project grouping

Walk through the raw input and split it into discrete tasks. For each candidate task, ask **one question**:

- **Project assignment:** *"¿A qué proyecto pertenece esta tarea?"* Show the current catalog. If the user names a new project, add it to `projects.md` and confirm the short description.

Do **not** ask hours per task. Hours are tracked at the project level (§3.6).

Record tasks under a `## Tasks by project` section in `LOG.md`, grouped by project, one bullet per task.

If an item is ambiguous (could be one task or several), **ask** — do not guess. If the user says "junta esos dos", merge into a single bullet.

### 3.6 Per-project hours

Once every task is grouped, walk through the projects with at least one task today. For each, ask:

- *"¿Cuánto tiempo total le dedicaste a `<Project Name>` hoy? (formato: `2h 30min`)"*

Record under a `## Hours` section in `LOG.md`.

Never estimate, round, or derive hours from the number/size of tasks. Ask.

### 3.7 Report generation

Only after every project with tasks today has a recorded hours figure, generate `REPORT.md` (§4).

### 3.8 Final review

Print the report path and a one-line summary of project totals + day total. Ask if the user wants edits before closing.

## 4. Report format

The report must be readable as raw text. **No tables, no HTML, no emojis, no bold/italic decorators, no code fences, no horizontal rules.** Only:

- Plain paragraphs.
- Numbered top-level sections: `1. Título`, `2. Título`, etc.
- Single-level bullets under each section with a `- ` prefix.
- One `Tiempo del proyecto` line per section; one `Tiempo total del día` line at the very end.

Full template, tone examples, hours-arithmetic rules in [RECIPES § 1–2](RECIPES.md#1-reportmd-template).

### Hours arithmetic

- `Tiempo del proyecto` is the value the user gave directly for that project today — do not derive from anything else.
- Sum project totals to produce `Tiempo total del día`.
- Normalize: `90min` → `1h 30min`; `2h 0min` → `2h`; `0h 45min` → `45min`.

## 5. Inference is forbidden

The user's hard rule: **never infer what was done**. Concretely:

- If raw input says "fixed OPC UA bug", do not write "fixed a critical OPC UA protocol violation" — ask what specifically was fixed, then use the user's words.
- If a commit subject says `feat(api): add /reports endpoint`, do not write the report based on the commit — ask the user what they did and how to frame it.
- If per-project hours are missing for any project with tasks today, do not estimate, round, or distribute time from the day total. Ask.
- If a project assignment is unclear for a task, do not pick the most recent project. Ask.
- If translation of a technical term is uncertain, follow §7 — do not invent a translation.

When in doubt, ask one specific question. Better to ask twice than to fabricate once.

## 6. Commits are input-only by default

Commits help the user remember what they did, but they do not appear in the report unless explicitly requested ("incluye los commits", "list the commits", "agrega los SHAs").

- Always paste commits into `LOG.md` under a `## Commits` subsection for traceability.
- Use commit subjects only as memory aids when asking clarification questions.
- If the user asks to include them, append a `Referencias` section at the very end of `REPORT.md`, listing `<short-sha> <subject>` one per line, no decoration.

## 7. Language and technical translation

The report is written in the language chosen in §3.1. Technical English terms follow four rules:

1. **Established translation exists** → use it, with the English term in parentheses on first use only (`throttling` → `limitación de tasa (throttling)`).
2. **No clean translation** → keep the English term as-is (`keep-alive`, `timeout`, `endpoint`).
3. **Acronyms and product names** → never translate (`OPC UA`, `SCADA`, `Raspberry Pi`).
4. **Function/symbol names from code** → preserve verbatim, no translation, no quoting.

Full examples per category in [RECIPES § 3](RECIPES.md#3-technical-translation--examples). When the output language is English, this section is a no-op.

## 8. Output deliverables per session

At the end of a session:

- `reports/projects.md` — updated with any new projects added during the session.
- `reports/<date>/LOG.md` — `## Raw input`, `## Commits` (if any), `## Tasks by project` (grouped), `## Hours` (per project).
- `reports/<date>/REPORT.md` — formal report regenerated from `LOG.md`, following §4 strictly.

If the user comes back the next day or week to amend a past report, re-read `LOG.md` for that date first and re-run the task extraction loop only on new input. Re-ask per-project hours only for projects whose task list changed.

## 9. Anti-patterns + closing checklist

Full anti-pattern table (12 common mistakes) in [RECIPES § 4](RECIPES.md#4-anti-patterns). Pre-close review checklist in [RECIPES § 5](RECIPES.md#5-closing-checklist).
