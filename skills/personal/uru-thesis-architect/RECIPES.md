# URU Thesis Architect — Templates & Rubric

Reference material for drafting. Load on demand once the section is chosen (SKILL.md §4). Everything here serves the section-focused workflow — pick a chapter, use its input checklist and structure, draft, then self-check against the rubric.

## 1. Per-chapter input & structure templates

For each chapter: the **inputs to gather** before drafting (never invent them — §6 anti-fabrication) and the **required subsections** in URU order. Draft one subsection at a time.

### Cap. I — El Problema → `05-cap-i-problema.md`

**Inputs to gather:** el fenómeno/necesidad observado; evidencia o síntomas concretos; a quién afecta; qué pasa si no se resuelve; el objetivo general en una frase; los objetivos específicos; el argumento de valor; los límites (tema, espacio, tiempo).

**Subsections:**
- **Planteamiento del problema** — de lo general a lo particular; cierra con la pregunta/necesidad de investigación.
- **Objetivos** — `GENERAL` (una frase, verbo accionable medible) y `ESPECÍFICOS` (lista; cada uno un paso verificable del general).
- **Justificación e importancia** — valor teórico, práctico y metodológico; específico, no genérico.
- **Delimitación** — alcance temático, espacial y temporal; qué cubre y qué no.

### Cap. II — Marco Teórico → `06-cap-ii-marco-teorico.md`

**Inputs to gather:** antecedentes (autor, año, título, aporte directo a este trabajo); teorías/autores base; términos técnicos a definir; variables y su fuente.

**Subsections:**
- **Antecedentes** — cada uno conecta explícitamente con este problema (aporte, no resumen).
- **Bases teóricas** — profundidad a la medida de la tesis; ni resumen de manual ni monografía off-topic.
- **Definición de variables** — `VARIABLE | AUTOR | DEFINICIÓN`; cada variable respaldada por un autor.

### Cap. III — Marco Metodológico → `07-cap-iii-marco-metodologico.md`

**Inputs to gather:** tipo/nivel de investigación; diseño; población y muestra (o unidad de análisis) y cómo se seleccionó; técnicas e instrumentos y su validación; el procedimiento paso a paso.

**Subsections:** Tipo y nivel · Diseño · Población y muestra · Técnicas e instrumentos · Procedimiento. Cada elección **justificada**, no solo declarada; el tipo/diseño debe corresponder a los objetivos (un desajuste es `[blocking]` para el reviewer).

### Cap. IV — Resultados → `08-cap-iv-resultados.md`

**Inputs to gather:** por cada objetivo específico, qué se hizo, los datos/evidencia obtenidos, y las tablas/figuras.

**Subsections:** una por objetivo específico, en el mismo orden. Cada una presenta resultado + análisis, contextualizado contra el marco teórico. Claims proporcionales a la evidencia.

### Cap. V — Propuesta (opcional) → guarda como sección propia si aplica

**Inputs to gather:** la solución derivada de los resultados; recursos/tiempo/restricciones; cómo se mediría su adopción.

**Subsections:** descripción de la propuesta · factibilidad · plan de validación. Debe derivarse de los resultados, no estar pre-decidida.

### Conclusiones y Recomendaciones → `09-conclusiones.md` / `10-recomendaciones.md`

**Conclusiones:** una por objetivo específico, cada una sostenida por un resultado; sin material nuevo; sin sobre-alcance. **Recomendaciones:** accionables, dirigidas a audiencias identificables — no plateaux ("seguir investigando").

## 2. Write-to-the-rubric checklist

Run before returning any section. These are the [uru-thesis-reviewer](../uru-thesis-reviewer/SKILL.md) §6 dimensions as drafting targets — writing to pass them is the whole point.

- [ ] **Objetivo general** singular, medible, alcanzable dentro del alcance.
- [ ] **Objetivos específicos** con verbos accionables (`analizar`, `diseñar`, `evaluar`); descomponen el general en pasos verificables.
- [ ] **Justificación** específica y con evidencia — no "este tema es muy importante".
- [ ] **Antecedentes** conectan explícitamente con este problema; columna de aporte es aporte concreto, no resumen.
- [ ] **Variables** definidas con autor de respaldo.
- [ ] **Tipo/diseño metodológico** corresponde a los objetivos.
- [ ] **Instrumentos** con validación/confiabilidad descrita (o marcada `[falta: …]`).
- [ ] **Resultados** responden a cada objetivo específico; claims proporcionales a la evidencia.
- [ ] **Conclusiones**: una por objetivo específico; sin material nuevo; sin sobre-alcance.
- [ ] **Coherencia en cadena** intacta: `problema → objetivos → metodología → resultados → conclusiones` — la sección no rompe ningún eslabón ni contradice capítulos previos.
- [ ] **Idioma** 100% español en la prosa; tags/metadata en inglés.
- [ ] **Sin fabricación**: todo dato/cita proviene del autor o está marcado `[falta: …]`.
- [ ] **Formato URU** de citas y referencias (NORMAS §V–VI) donde la sección las incluya.

## 3. Section drafting & iteration protocol

The step-by-step loop (SKILL.md §8), expanded:

1. **Confirm the target** — one section or subsection, explicitly.
2. **Gather inputs** from §1 for that piece. Missing substance → ask, don't invent.
3. **Draft** using the subsection structure; apply Spanish + anti-fabrication rules.
4. **Self-check** against §2; fix silently before presenting.
5. **Present + iterate** — show the draft, mark any `[falta: …]`, take corrections, refine on the same section until the author is satisfied.
6. **Persist** — write/append to the section file; report the path.
7. **Only then advance** to the next subsection/section.

**Iterating a full chapter:** decompose the chapter into its §1 subsections and run steps 1–6 for each, in order. The chapter file fills progressively; at no point is the whole chapter generated in one turn. This keeps each piece small enough for the author to verify and own.

## 4. Output file layout

```
<output-base>/
  05-cap-i-problema.md
  06-cap-ii-marco-teorico.md
  07-cap-iii-marco-metodologico.md
  08-cap-iv-resultados.md
  09-conclusiones.md
  10-recomendaciones.md
  <optional: cap-v-propuesta.md>
```

Rules:

- **Kebab-case** — lowercase, dash-separated, no accents (filesystem safety). Same convention as the reviewer's session files, so a drafted section maps 1:1 to its review file.
- **Numbered in thesis order**, mirroring [uru-thesis-reviewer TEMPLATES § 5](../uru-thesis-reviewer/TEMPLATES.md#5-folder-layout-numbered-thesis-order).
- **One section per file**, filled incrementally across turns — never all at once.
- Preliminares (portada, resumen, abstract) and referencias are drafted only on request; they are not part of the core Cap. I–V drafting flow.
