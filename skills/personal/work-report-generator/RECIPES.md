# Work Report Generator — Templates & Checks

Report template, anti-patterns, translation examples, and closing checklist for [SKILL.md](SKILL.md). Loaded on demand.

## 1. `report.md` template

```
Reporte de Actividades — YYYY-MM-DD

1. <Project Name>
   - <Human-readable description of the first task done for this project today, formal tone, full sentence, technical terms translated per §7.>
   - <Description of the next task.>
   - <Description of the next task.>

   Tiempo del proyecto: Xh Ymin.

2. <Next Project>
   - <Task description.>
   - <Task description.>

   Tiempo del proyecto: Xh Ymin.

Tiempo total del día: Xh Ymin.
```

Rules:

- One entry per project worked that day; all tasks grouped as a bulleted list.
- Each task bullet is a **single human-readable description** — no title prefix, no `Title: description` shape, no nested sub-bullets.
- Tasks may be re-ordered within a project for natural reading flow; content stays faithful to user words.
- Hours recorded **once per project**, on the `Tiempo del proyecto` line. No per-task time.
- Section numbering follows the order in `projects.md`, filtered to projects that have tasks on this date.
- Translate the labels (`Tiempo del proyecto`, `Tiempo total del día`, `Reporte de Actividades`) to the chosen output language.
- **No** "Engineer / Project / Header" preamble. Go straight to `1. <Project>`.
- **No** intro paragraph at the top of each project section.
- **No** commit hashes unless the user explicitly asked for them in this session.

## 2. Tone — formal, impersonal

- ✓ "Se refactorizó el módulo de despacho de publicaciones para asegurar..."
- ✓ "The publish dispatch module was refactored to ensure..."
- ✗ "Estuve trabajando en arreglar el bug de..."
- ✗ "I fixed the bug in..."

Don't embellish. Adjectives implying impact, scope, or difficulty ("critical", "major", "comprehensive") only if the user used them.

**Level of detail — no exact failure/lint/pass-rate counts.** Don't enumerate precise quantities of failures, errors, lint findings, or test/benchmark pass rates (e.g. "74 lint violations", "380/412 diagnostics (92%)", "21/21 tests passing", "48/80 pass rate"). Summarize qualitatively instead ("se remediaron la mayoría de las violaciones de lint", "se validó exitosamente el harness de pruebas"), unless the user explicitly asks for the precise figures. This does not apply to numbers describing the scope of the actual deliverable (files touched, tasks completed, corpus/dataset size) — those stay as given.

- ✗ "se ejecutó una limpieza de análisis estático, remediando 380 de 412 diagnósticos iniciales (92%)"
- ✓ "se ejecutó una limpieza de análisis estático, remediando la gran mayoría de los diagnósticos iniciales"

**Verb precision.** Don't default to one verb for every deliverable — match the verb to what actually happened.

- `generó`/`generaron` — for produced artifacts: documentation, diagrams, scaffolds, schema/migration files.
- `desarrolló`/`desarrollaron` — for actively built/coded things: tools, adapters, CI pipelines, backend services, features with real logic.
- `entregó`/`entregaron` — only if the work was actually delivered or shared with someone. Finishing something is not the same as delivering it — most day-to-day work bullets should use `generó`/`desarrolló`/`implementó`/etc., not `entregó`.

**Costs and pricing — ask, don't assume.** If raw input or commit content touches costs, pricing, or dollar figures, ask the user whether to include that in the report before writing the bullet. Don't default to including it (it's rarely relevant to a work-activity report) and don't default to silently omitting it either — some users may want it kept. When in doubt, ask once and remember the answer for the rest of the session.

**No personal-tooling references.** Don't mention the user's own Claude Code skills, prompts, or internal AI tooling as if they were part of the technical work being reported (e.g. "se habilitaron linters del skill go-architect"). Describe the underlying practice or standard instead ("se habilitaron linters siguiendo las buenas prácticas de Go").

## 3. Technical translation — examples

1. **Established translation exists** → use the translated term, with the English term in parentheses on first use only.
   - `throttling` → `limitación de tasa (throttling)`
   - `fine-tuning` → `ajuste fino (fine-tuning)`
   - `handshake` → `handshake` (no widely accepted Spanish equivalent — keep English)
2. **No clean translation** → keep the English term as-is.
   - `keep-alive`, `timeout`, `bridge`, `pool`, `endpoint`, `dashboard`, `commit`, `runtime`.
   - Watch for anglicisms with a clean, natural Spanish equivalent that's easy to default to in English out of habit — these should be translated, not kept: `tracking` → `seguimiento`, `hyperlinks` → `enlaces`, `de-duplicaron`/`de-duplicated` → `se eliminó la duplicación de` / `se depuraron duplicados` (avoid the hyphenated calque).
   - Some terms have a "technically correct" Spanish translation that's still a stiff calque in narrative prose — prefer natural phrasing over the dictionary-accurate cognate. `git submodule deinit` → don't write `se desinicializó el submódulo` (accurate but stilted); write `se eliminó por completo el submódulo, dando de baja su registro en Git`.
3. **Acronyms and product names** → never translate.
   - `OPC UA`, `SCADA`, `HMI`, `VTScada`, `YOLOv11`, `Raspberry Pi`, `Hailo-8L`.
4. **Function/symbol names from code** → preserve verbatim, no translation, no quoting.
   - `forward_publish_response`, `RequestHandle`, `BadUserAccessDenied`.

When the output language is English, this is a no-op — write everything in natural English.

## 4. Anti-patterns

| Wrong | Right |
|---|---|
| Inferring a task from a commit subject | Ask the user what they did, then write |
| Estimating hours because the user "probably worked all morning" | Ask for the exact figure in `Xh Ymin` |
| Asking hours per task | Hours are per project only — ask once per project |
| Distributing a known day-total across projects yourself | Ask the per-project figure even when the day total is known |
| Writing a `Title: description` shape per task bullet | Each bullet is a single human-readable description |
| Writing an intro paragraph at the top of a project section | Go straight to the bulleted task list |
| Adding "Ingeniero: …" / "Proyecto: …" header | Start with `1. <Project>` directly |
| Writing the report in English when the user asked for Spanish | Re-read the language answer before generating |
| Translating `OPC UA` to `OPC UA (Comunicaciones Unificadas de Plataforma Abierta)` | Leave acronyms untouched |
| Including commit SHAs in the body when not asked | Use commits only as clarification prompts |
| Bolding section titles or using `##` headers in the formal report | Use `1.`, `2.`, … plain text headers |
| Editing `report.md` directly to add a task | Append to `raw.md`, then regenerate |
| Describing the source artifact (`Cerró el ticket X`, `Mergeó el commit Y`, `Completó el PR Z`) | Describe the result of the work (`Resolvió el timeout en el worker de despacho`) |
| Generating `report.md` without the explicit "¿algo más?" close in §3.7 | Wait for affirmative close from the user before generating |
| Translating `raw.md` content to the chosen output language | It stays verbatim in the source language; only `report.md` is translated |
| Reformatting / normalizing the raw input pasted into `raw.md` | Append verbatim; preserve formatting, indentation, line breaks |
| Batching multiple dates in a single session | One date per invocation; repeat §3.1–§3.9 per date for backfills |
| Pasting only commit subjects into `commit.md` | Paste full commit messages (subject + body); same for issues/tickets |
| Pasting commit dumps into `raw.md`'s raw input | Commits live in `commit.md`, separate from manual raw input |
| Section header with decoration: `1. Detector de Anomalías Online (4h 30min)` | Section header is the project name verbatim: `1. Detector de Anomalías Online` |
| Enumerating exact failure/lint/test-pass-rate counts (`74 violations`, `21/21 tests`, `48/80 pass rate`) | Summarize qualitatively; keep deliverable-scope numbers (files touched, tasks completed) as-is |
| Merging two raw items into one task because they look related | Merge only on explicit user instruction ("junta esos dos") |
| Splitting one raw item into multiple tasks on your own | Ask when ambiguous; split only on explicit user instruction |
| Expanding ambiguous shorthand from raw input based on the most plausible guess (e.g. "la Pi" silently expanded to "Raspberry Pi") | Ask what it refers to — the same shorthand can mean different things across sessions (hardware, an API, a typo, a codename); never hardcode one interpretation |
| Mentioning costs, pricing, or dollar figures because the underlying commit/raw input had them | Ask the user first whether cost content should be included in the report |
| Describing the user's own Claude Code skill/tooling as part of the technical work ("siguiendo el skill go-architect") | Describe the underlying practice/standard instead ("siguiendo las buenas prácticas de Go") |
| Defaulting every deliverable bullet to "se entregó" | Match the verb to what happened: "generó" for artifacts, "desarrolló" for built things, "entregó" only if actually shared |

## 5. Closing checklist

- [ ] Language matches user's session choice (REPORT in chosen language; LOG verbatim in source language)
- [ ] User explicitly confirmed close in §3.7 before `report.md` was generated
- [ ] Single date scope — no content from other dates leaked into this report
- [ ] Every task has a project from `projects.md`
- [ ] Every project with tasks today has a `Tiempo del proyecto` figure in `Xh Ymin` format
- [ ] No per-task time appears anywhere in `report.md`
- [ ] Day total equals the sum of project totals
- [ ] No header preamble (Engineer/Project) above section 1
- [ ] No project-level intro paragraph above the task list
- [ ] No task bullet uses a `Title: description` shape
- [ ] No task bullet references the source artifact (commit / ticket / issue / PR) as its subject
- [ ] Section headers contain the project name verbatim — no parenthetical decoration
- [ ] No tables, no emojis, no HTML, no `##` decorators, no nested bullets in the body
- [ ] Technical terms follow the translation rules
- [ ] No content not traceable to user input or clarification answers
- [ ] `projects.md` reflects any new projects introduced this session
- [ ] `reports/<date>/raw.md` contains raw input (verbatim, manual notes only), tasks grouped by project, and per-project hours
- [ ] `reports/<date>/commit.md`, if present, holds the day's commit dump (full bodies, grouped by repo) — not duplicated into `raw.md`
- [ ] No costs/pricing content unless the user was explicitly asked and opted in
- [ ] No reference to the user's own Claude Code skills/tooling in the technical narrative
- [ ] Verbs match what happened ("generó" for artifacts, "desarrolló" for built things, "entregó" only if actually shared)
