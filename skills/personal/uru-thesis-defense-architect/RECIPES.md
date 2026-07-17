# URU Thesis Defense Architect — Templates & Checklists

Reference templates and the full structure table. Load on demand once the interview (SKILL.md §2) is complete.

## 1. Fixed structure table

Distribución de elementos y énfasis del documento, por sección. El número de lámina es un punto de referencia — cualquier sección puede extenderse a varias láminas cuando el contenido lo exige (ver SKILL.md §5).

### Lámina 1 — Portada (Apertura Institucional)

- **Distribución:** logo grande y centrado arriba; TÍTULO en tipografía pesada y mayor tamaño al centro; datos de autoría (Realizado por), tutoría (Tutor) y cierre cronológico (Lugar, Año) abajo.
- **Énfasis:** rigor formal e identificación clara de los actores de la investigación.

### Lámina 2 — El Problema

- **Distribución:** logo pequeño arriba a la izquierda; título al lado o centrado; el resto es lienzo libre.
- **Énfasis:** síntesis. Un esquema, un par de viñetas contundentes o un diagrama que explique el conflicto o necesidad que origina el proyecto — nunca un muro de texto.

### Lámina 3 — Objetivos del Trabajo

- **Distribución:** logo institucional reducido. Espacio fragmentado en dos categorías: GENERAL (bloque de texto destacado, arriba o a la izquierda) y ESPECÍFICOS (lista secuencial, abajo o a la derecha).
- **Énfasis:** claridad en la jerarquía de metas — separación inmediata entre el fin macro y los pasos técnicos para lograrlo.

### Lámina 4 — Justificación e Importancia / Delimitación

- **Distribución:** encabezado con logo compacto; espacio inferior dividido equivalentemente en dos mitades.
- **Énfasis:** doble propósito — argumentar el valor de la investigación (Justificación e Importancia) y delimitar el alcance (Delimitación) para que el jurado sepa qué cubre y qué no cubre el proyecto.

### Lámina 5 — Antecedentes

- **Distribución:** logo en esquina; formato de matriz/tabla `AUTOR, AÑO | TÍTULO | APORTE`.
- **Énfasis:** relación "Origen vs. Utilidad" — no basta con decir quién escribió algo antes; cada fila debe sintetizar el aporte directo (tecnológico, metodológico o algorítmico) que se extrae para el desarrollo propio.

### Lámina 6 — Definición de las Variables

- **Distribución:** logo en esquina; bloques conceptuales limpios `VARIABLE + AUTOR + DEFINICIÓN`.
- **Énfasis:** soporte teórico estricto — cada variable técnica o de ingeniería debe estar respaldada por un autor de referencia que valide el marco conceptual.

### Lámina 7 — Metodología

- **Distribución:** logo en esquina; cuadrícula o lista de cuatro componentes: Tipo, Diseño, Población-Muestra (o Unidad de Análisis), Técnicas e Instrumentos.
- **Énfasis:** rigor metodológico tradicional — naturaleza científica de la investigación y herramientas de recolección/procesamiento de información.

### Lámina 8 — Metodología de Desarrollo

- **Distribución:** logo en esquina; tabla o mapa de proceso cruzando FASES (eje principal / etapas del ciclo de desarrollo) × ACTIVIDADES (tareas específicas de cada etapa).
- **Énfasis:** ciclo de vida de la ingeniería — la hoja de ruta cronológica/metodológica seguida para construir la solución. Esta tabla es el ancla que Resultados debe referenciar.

### Láminas 9+ — Resultados (núcleo pesado)

- **Distribución:** logo en esquina; sección más libre, visual y dinámica. Puede llevar varias láminas.
- **Énfasis:** el hacer y la demostración de ingeniería. Dos directrices críticas:
  1. Explicar lo que se fue realizando en cada fase — correlación directa y ordenada con la Lámina 8.
  2. Extenderse cuanto sea necesario — es la única sección donde extender el espacio es la norma, no la excepción. Aquí van diagramas de bloques, arquitecturas, flujos de datos, capturas de interfaces, analíticas obtenidas.

### Lámina Final — Conclusiones y Recomendaciones (Cierre Técnico)

- **Distribución:** logo en esquina; bloques finales de texto analítico.
- **Énfasis:** "mostrar el producto" — más allá de la síntesis de hallazgos (Conclusiones) y mejoras futuras (Recomendaciones), el punto culminante es la validación práctica de la ingeniería, abriendo el espacio inmediato para la demostración real del software, hardware o sistema desarrollado.

## 2. Per-slide specification format

```markdown
## Lámina N — <Sección>

**Logo:** assets/uru-logo.png — grande, centrado | pequeño, superior-izquierda

**Contenido:**
- <campo 1 según la sección, p. ej. Título / Objetivo general / fila de tabla>
- <campo 2>
- ...

**Notas del orador:** <opcional, ≤40 palabras>
```

- Para láminas de tabla (Antecedentes, Variables, Metodología de Desarrollo), representar `**Contenido:**` como una tabla markdown real con las columnas fijas de esa sección.
- Para Objetivos, usar dos sub-bloques rotulados `GENERAL:` y `ESPECÍFICOS:` dentro de `**Contenido:**`.
- Para Resultados, prefijar cada lámina con la fase que cubre, p. ej. `## Lámina 9a — Resultados: Fase 1 — Análisis`.

## 3. Review checklist

Before handing off the deck spec:

- [ ] Todo el texto de contenido está en español (títulos, viñetas, tablas, notas)
- [ ] Lámina 1 tiene logo grande y centrado; toda lámina 2+ tiene logo pequeño en esquina superior izquierda
- [ ] Ninguna lámina tiene texto inventado — todo campo proviene de una respuesta del usuario
- [ ] Objetivos separa visualmente GENERAL de ESPECÍFICOS
- [ ] Antecedentes usa la tabla `AUTOR, AÑO | TÍTULO | APORTE` con aporte concreto, no resumen
- [ ] Variables usa la tabla `VARIABLE | AUTOR | DEFINICIÓN`
- [ ] Metodología de Desarrollo lista FASES × ACTIVIDADES
- [ ] Resultados abre con al menos una lámina por fase de Metodología de Desarrollo, en el mismo orden
- [ ] Ninguna sección fue forzada a una sola lámina si el contenido no cabía sin muro de texto
- [ ] Conclusiones y Recomendaciones cierra con la línea de transición a la demostración en vivo
- [ ] El Problema es síntesis (esquema/viñetas/diagrama), no un párrafo largo
