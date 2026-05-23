# Work Report Generator — Templates & Checks

Report template, anti-patterns, translation examples, and closing checklist for [SKILL.md](SKILL.md). Loaded on demand.

## 1. `REPORT.md` template

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

## 3. Technical translation — examples

1. **Established translation exists** → use the translated term, with the English term in parentheses on first use only.
   - `throttling` → `limitación de tasa (throttling)`
   - `fine-tuning` → `ajuste fino (fine-tuning)`
   - `handshake` → `handshake` (no widely accepted Spanish equivalent — keep English)
2. **No clean translation** → keep the English term as-is.
   - `keep-alive`, `timeout`, `bridge`, `pool`, `endpoint`, `dashboard`, `commit`, `runtime`.
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
| Editing `REPORT.md` directly to add a task | Append to `LOG.md`, then regenerate |
| Describing the source artifact (`Cerró el ticket X`, `Mergeó el commit Y`, `Completó el PR Z`) | Describe the result of the work (`Resolvió el timeout en el worker de despacho`) |
| Generating `REPORT.md` without the explicit "¿algo más?" close in §3.7 | Wait for affirmative close from the user before generating |
| Translating `LOG.md` content to the chosen output language | LOG stays verbatim in the source language; only `REPORT.md` is translated |
| Reformatting / normalizing the raw input pasted into `LOG.md` | Append verbatim; preserve formatting, indentation, line breaks |
| Batching multiple dates in a single session | One date per invocation; repeat §3.1–§3.9 per date for backfills |
| Pasting only commit subjects into `LOG.md` | Paste full commit messages (subject + body); same for issues/tickets |
| Section header with decoration: `1. Detector de Anomalías Online (4h 30min)` | Section header is the project name verbatim: `1. Detector de Anomalías Online` |
| Merging two raw items into one task because they look related | Merge only on explicit user instruction ("junta esos dos") |
| Splitting one raw item into multiple tasks on your own | Ask when ambiguous; split only on explicit user instruction |

## 5. Closing checklist

- [ ] Language matches user's session choice (REPORT in chosen language; LOG verbatim in source language)
- [ ] User explicitly confirmed close in §3.7 before `REPORT.md` was generated
- [ ] Single date scope — no content from other dates leaked into this report
- [ ] Every task has a project from `projects.md`
- [ ] Every project with tasks today has a `Tiempo del proyecto` figure in `Xh Ymin` format
- [ ] No per-task time appears anywhere in `REPORT.md`
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
- [ ] `reports/<date>/LOG.md` contains raw input (verbatim, with full commit/issue bodies), tasks grouped by project, and per-project hours
