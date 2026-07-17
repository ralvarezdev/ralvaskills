# URU Thesis Reviewer — File Templates

Boilerplate for the per-session output files. Used by [SKILL.md](SKILL.md) §4 (output structure) and §5 (file templates). Loaded on demand.

## 1. `01-index.md`

```markdown
# Revisión URU — YYYY-MM-DD

**Source:** `<absolute path to .docx>`
**Scope:** <full thesis | Capítulo II | grammar only | ...>
**Reviewer pass:** <1st | follow-up Nº | ...>

## Resumen ejecutivo

<2-3 sentence overall impression. Strongest aspect. Weakest aspect. Where to start.>

## Conteo por severidad

| Sección | blocking | suggest | nit | Total |
|---|---:|---:|---:|---:|
| Resumen | 0 | 2 | 1 | 3 |
| Cap. I | 3 | 5 | 2 | 10 |
| ... | | | | |
| **Total** | **X** | **Y** | **Z** | **N** |

## Mapa de archivos

- `02-resumen.md` — 3 observaciones
- `03-abstract.md` — sin observaciones
- `04-introduccion.md` — 7 observaciones (1 blocking)
- ...

## Temas transversales

<Recurring issues to fix everywhere, not section-specific.
 e.g. "Uso inconsistente de primera persona en Cap. II y IV — la norma exige tercera persona.">

## Recomendaciones de orden

1. Resolver primero los `[blocking]` (norm violations, factual errors).
2. Cap. III necesita reestructuración antes de revisar prosa.
3. Bibliografía: 4 fuentes débiles identificadas, ver `11-referencias.md`.
```

## 2. `nn-section.md` — one per reviewed section

```markdown
# <Section name as it appears in thesis>

**Ubicación:** página N–M, §X.Y
**Observaciones:** total N (blocking: A, suggest: B, nit: C)

---

## [blocking] <one-line summary> — §X.Y, párrafo N

**Norma:** `NORMAS §V. CITAS > Cita textual de larga extensión`

> Original (cita textual del documento):
> "<paste exact original sentence/paragraph>"

- "El autor menciona que la productividad aumento un 30%, según un estudio reciente."
+ "Según Ramírez (2021, p. 47), la productividad aumentó un 30% en el período evaluado."

**Por qué:** Falta autor y año en la cita indirecta (norma URU exige `Apellido, año` para toda referencia parafraseada). Además, "aumento" sin tilde es error ortográfico (RAE — verbo en pretérito).

---

## [suggest] <one-line summary> — §X.Y, párrafo N

> Original:
> "..."

- "..."
+ "..."

**Por qué:** <reasoning>

---

## [nit] <one-line summary>

- "..."
+ "..."

**Por qué:** <reasoning>
```

## 3. `11-referencias.md` — bibliography section

```markdown
# Referencias bibliográficas

**Total fuentes:** N
**Citadas en texto:** N    **Solo en lista:** N    **Citadas pero ausentes:** N

## [blocking] Fuentes citadas pero no listadas

- "Pérez (2019)" — citada en Cap. II §2.3, ausente de la lista.
- ...

## [blocking] Fuentes en la lista pero no citadas

- García, M. (2015). ...
- ...

## [suggest] Fuentes débiles

### García, M. (2015). Manual de gestión.

**Problema:** Editorial no académica, sin revisión por pares, contenido divulgativo para tema técnico central de la tesis.

**Alternativa sugerida:**
+ Hernández, R., Fernández, C. y Baptista, P. (2014). *Metodología de la investigación* (6ª ed.). México: McGraw-Hill.
  — Referencia estándar en metodología, citada en >50,000 trabajos académicos.

### <next weak source>
...

## [suggest] Formato de referencia

- García M, 2015, Manual de gestión, McGraw Hill
+ García, M. (2015). *Manual de gestión*. México: McGraw-Hill.

**Por qué:** Norma URU exige `Apellido, Inicial. (Año). Título en cursiva. Ciudad: Editorial.` con sangría francesa.
```

## 4. Substantive-feedback diff shape

When the issue is substantive (not a textual fix), the `+` side is a **prompt** for the author, not rewritten text. Example:

```
- "El presente estudio busca conocer la situación del mantenimiento en la empresa."
+ "<reformular como objetivo medible — e.g.: 'Analizar las prácticas actuales de mantenimiento
   preventivo en la planta X durante el período 2024–2025, identificando desviaciones respecto
   al estándar ISO 55000.' Confirmar alcance temporal y referencia normativa con el tutor.>"

**Por qué:** "Conocer la situación" no es un objetivo verificable (NORMAS §III: lenguaje preciso;
metodológicamente, no permite definir cuándo se ha cumplido). Requiere verbo accionable + objeto
delimitado + criterio de éxito.
```

The author makes the substantive call — Claude doesn't write the thesis.

## 5. Folder layout (numbered, thesis order)

```
<output-base>/YYYY-MM-DD/
  01-index.md
  02-resumen.md
  03-abstract.md
  04-introduccion.md
  05-cap-i-problema.md
  06-cap-ii-marco-teorico.md
  07-cap-iii-marco-metodologico.md
  08-cap-iv-resultados.md
  09-conclusiones.md
  10-recomendaciones.md
  11-referencias.md
  12-anexos.md           ← only if anexos have issues
```

Rules:

- Numbering reflects **thesis order**, not severity.
- Skip a section file entirely if it has zero feedback — note "sin observaciones" in `01-index.md` instead of creating an empty file.
- Small sections with few items may be merged with their neighbor (`02-resumen-y-abstract.md`) — note merges in the index.
- `referencias` always second-to-last (or last if no anexos).
- File names: `nn-section-name.md`, kebab-case (lowercase, dash-separated), no accents (filesystem safety).
