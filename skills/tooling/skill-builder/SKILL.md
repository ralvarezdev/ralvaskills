---
name: skill-builder
version: 1.0.0
description: Scaffold a new ralvaskills skill per SPECS.md — generates SKILL.md frontmatter and body skeleton, STACK.md template, folder placement, SPECS.md updates, validation checklist. Interview-first; never guesses. Use when user wants to author a new skill, mentions "new skill" or "skill template", or invokes /skill-builder.
---

# Skill Builder

Meta-skill that scaffolds new skills under `skills/<category>/<skill-name>/` following [SPECS.md](../../../docs/SPECS.md). Generates structure; hands off content authoring.

## 1. When to invoke

- User says "let's create a new skill", "scaffold a skill", "I want to add a `<topic>-architect`", or invokes `/skill-builder`.
- User wants to verify whether an existing folder structure matches the conventions.
- User wants to move / rename a skill and needs the SPECS updates handled correctly.

## 2. Source of truth

**SPECS.md is the canonical convention.** This skill cites and applies the rules; it does not redefine them. Sections in SPECS that this skill enforces:

- **Skill Authoring Guide** — SKILL.md format, STACK.md format, naming convention, description quality, supporting files (`UPPER_SNAKE_CASE.md`).
- **Repository & Folder Structure** — discovery rule (folder with `SKILL.md` is terminal; nested SKILL.md forbidden), category placement, personal-skills path-segment rule.
- **Which Skills Need a `STACK.md`** — heuristic table by parent folder; per-skill exceptions allowed.
- **Bundle & Skill Catalog** — bundle membership decisions.

If a question isn't answered here, **read SPECS.md before guessing.**

## 3. Interview-first protocol

**Never scaffold without an interview.** Ask these in order; wait for each answer:

1. **Topic + suffix:** what does the skill enforce? (e.g. "Kafka standards" → `-architect`; "code review for security" → `-reviewer`; "pattern catalog" → `-patterns`.) Confirm the resulting `<name>` matches the SPECS naming convention table.
2. **Category:** which parent folder? Use the existing folder structure in SPECS as the source of valid categories (`languages/`, `databases/`, `frameworks/`, `protocols/`, `encoding/`, `messaging/`, `tooling/`, `design/`, `ai-ml/`, `infra/`, `robotics/`, `quality/`, `frontend/`, `workflows/`, `personal/`). If the right category doesn't exist yet, ask whether to add it (SPECS update required).
3. **Version target:** if the category requires `STACK.md` (per the SPECS heuristic table), what versions does the skill target? (e.g. Go 1.26, Python 3.14, PostgreSQL 18.) For meta-skills with no library stack, confirm exemption.
4. **Bundle membership:** which existing bundle(s) should include this skill? Or is it standalone (no bundle)? Cross-reference the SPECS bundle catalog.
5. **Cross-skill ties:** which other skills should this one reference? (Language architect for idioms, sql-architect for data, rest-api-architect for HTTP conventions, etc.) These become `../<path>/SKILL.md` links in the body.
6. **Attribution:** is this skill adapted from an external source (e.g. `mattpocock/skills`)? If so, capture the attribution for the footer.
7. **Body shape preview:** propose 5–12 section headings based on the topic. Confirm with the user before writing the skeleton.
8. **Decisions list (if any):** when the skill encodes opinionated defaults (e.g. ORM vs raw SQL, auth pattern, file format), surface the decisions for the user *before* generating the body — same flow we used for sql-architect and others.

## 4. Naming + placement

Validate against the SPECS naming convention table:

| Suffix | Use for |
|---|---|
| `-architect` | Enforces structure/standards for a language, framework, protocol, infra component |
| `-arch` | Enforces an architectural pattern (`hexagonal-arch`) |
| `-reviewer` | Reviews code/contracts for issues |
| `-refactor` | Restructures existing code |
| `-planner` | Designs before implementation |
| `-patterns` | Catalog of patterns / principles for a domain |
| `-builder` | Scaffolds new artifacts of a given type |
| `-guide` | Documentation / usage guide for a specific tool |
| (none) | Workflow tools that don't fit above |

**If the topic doesn't fit any suffix**, propose a new suffix and update SPECS — don't shoehorn. The convention is meant to grow when new categories appear (we've added `-patterns`, `-builder`, `-guide` already).

Path follows category placement and the discovery rule: `skills/<category>/<name>/SKILL.md`. Never nest a `SKILL.md` inside another skill folder.

## 5. Frontmatter generation

Emit a **one-line** description in YAML scalar form (no `>` or `|`). The description sits in Claude's always-loaded skill index — every line costs tokens on every turn, forever.

```yaml
---
name: <skill-name>             # kebab-case; matches folder name
version: 1.0.0                 # always start at 1.0.0 unless it's a draft tracking a planned feature (e.g. 0.1.0)
description: <WHAT — short, comma-separated topics> — <version target if applicable>. Use when <trigger phrases>.
---
```

**Description quality rules (per [SPECS — Description Quality](../../../docs/SPECS.md#description-quality)):**

- WHAT first — topics the skill enforces, comma-separated, no marketing prose.
- WHEN last — trigger phrases the user might say or commands they might invoke (e.g. "Use when scaffolding or reviewing a FastAPI service" or "...invokes /fastapi-architect").
- For version-sensitive skills: mention the primary version target inline (e.g. "Go 1.26", "FastAPI 0.136 on Python 3.14").
- Drop "Enforces strict", "Comprehensive guide to", and similar filler.
- **Bad:** `description: Go best practices skill.` ← no when.
- **Bad (verbose, multi-line):** Anything using `>` or `|` and spanning more than one line.
- **Good:** `description: Go 1.26 architectural standards — memory-aligned structs, typed enums, interface design, goroutine safety, iterators, idiomatic errors. Use when writing, reviewing, or scaffolding Go code.`

## 6. Body skeleton

Generate placeholders, not content. **Body discipline (per [SPECS — SKILL.md Format](../../../docs/SPECS.md#skillmd-format)):**

- Rules first, code second. Each section explains the rule and the *why*; code is illustrative, not exhaustive.
- If a code block grows past ~15 lines or appears as a full reference implementation (Dockerfile, auth middleware, project tree, test scaffold), move it to `RECIPES.md` and link from the SKILL.md section with one line: *"Skeleton in [RECIPES.md](RECIPES.md)."*
- Don't restate patterns canonically explained elsewhere — link.
- Target SKILL.md body: under ~10 KB / ~250 lines unless the topic is genuinely dense (foundational language / cross-cutting protocol skills are allowed to run longer).

Standard section shape:

```markdown
# <Title>

<One-paragraph framing>. Targets **<version target if applicable>**.
Companion to [<related-skill>](../<path>/SKILL.md). Implementation skeletons in
[RECIPES.md](RECIPES.md). See [STACK.md](STACK.md) for pinned dependencies.

## 1. <Section name>

- **<Rule>** — <explanation>
- **<Rule>** — <explanation>

## 2. <Section name>

...

## N. Out of scope (if applicable)

- **<Concern>** → see [<other-skill>](../<path>/SKILL.md)

---

_Adapted from [<source>](<url>)._   <!-- only if attribution applies -->
```

Headings come from the body-shape preview in §3, step 7. Body content is filled in *after* the scaffold is approved, by the user or by Claude in a follow-up turn.

### When to scaffold a RECIPES.md

Create `RECIPES.md` at scaffold time when the skill will plausibly contain:

- Multiple full reference implementations (Dockerfile, middleware, repo layout, test scaffold).
- A canonical "here's how the full pattern looks in <language>" section.
- Multiple sibling examples (Go + Python, several frameworks).

Skip `RECIPES.md` when the skill is conceptual (catalog skills like `design-patterns`, methodology skills like `tdd`, glossary-only references). For those, code snippets stay inline because they're short and illustrate the concept, not provide a copy-pasteable skeleton.

If unsure, write the skill first; factor recipes out only when the body crosses ~10 KB or a single section's code blocks exceed half its prose.

## 7. STACK.md — when and how

**Check the SPECS "Which Skills Need a `STACK.md`" table** keyed on the parent folder. If yes (or "Depends" with version-sensitive content), generate:

```markdown
# Stack Versions

| Dependency | Pinned version | Purpose |
|---|---|---|
| <dep> | <version> | <one-line> |
| ... | ... | ... |

## Notes

- <one or two architectural-opinion lines that the skill enforces>
- <inheritance: if this skill inherits a base stack, link it>

_Last reviewed: <YYYY-MM-DD>_
_Skill version at last review: <version>_
```

**Fetching versions:** for Go libraries, query `https://proxy.golang.org/<module>/@latest`. For Python, `https://pypi.org/pypi/<package>/json`. For tools, use `gh api repos/<owner>/<name>/releases/latest`. Always pin actual versions, never `latest`.

**Exempt from STACK.md:**
- `workflows/` — language-agnostic process skills.
- `design/` — architectural patterns are version-agnostic.
- `quality/` — security/perf principles are stable (override only if the skill references framework-specific behavior).
- **Meta-skills** in `tooling/` that have no library stack (e.g. `skill-builder`, `rsk-guide`).

Mark the exemption explicitly in the roadmap notes when applicable.

## 8. Cross-skill linking

Use relative paths from the new skill to its peers:

```markdown
- [go-architect](../../languages/go-architect/SKILL.md) — Go idioms
- [sql-architect](../../databases/sql-architect/SKILL.md) — data access pattern
- [rest-api-architect](../../protocols/rest-api-architect/SKILL.md) — REST conventions
```

**Standard ties by category:**

| New skill category | Usually references |
|---|---|
| `frameworks/` (e.g. `fastapi-architect`) | matching language architect; `sql-architect`; `rest-api-architect` |
| `protocols/` (e.g. `rest-api-architect`) | usually framework-agnostic — minimal cross-references |
| `encoding/` (e.g. `protobuf-architect`) | language architect for code generation |
| `infra/` (e.g. `docker-architect`) | tooling defaults; sometimes language architects for per-language recipes |
| `quality/` (reviewer skills) | usually framework-agnostic; reference architect skills for "what they should already enforce" |
| `workflows/` | other workflow skills they chain with; pipeline references |

## 9. SPECS.md updates (atomic)

Every new skill requires four to six edits to SPECS.md, applied in one pass:

1. **Folder Structure tree** — add the new `<category>/<name>/` entry with status marker (`# ✅ exists (v1.0.0 — short note)`).
2. **Roadmap row** in the matching category section — `| <name> | ✅ | v1.0.0 — one-line scope summary |`.
3. **Global bundle row** (if the skill is in the global bundle) — `| <name> | <category> | <description> |`.
4. **Stack-specific bundle rows** (if relevant) — add to `gin`, `nethttp`, `fastapi`, `go-grpc`, etc.
5. **Naming convention table** (only if introducing a new suffix) — add the row.
6. **STACK.md needed table** (only if introducing a new category) — add the row.

**Verify after editing:** grep SPECS for the skill name; every mention should make sense in context. Check that no `📋 planned` row remains for a skill that now exists.

## 10. Attribution

If the skill is adapted from an external source (e.g. Matt Pocock, an Anthropic example, a community pattern), append the footer:

```markdown
---

_Adapted from [<source>](<url>)._
```

Adapt verbatim copies require strict attribution. Significantly evolved skills should still credit the origin idea — see the workflow skills (`grill-with-docs`, `tdd`, `improve-codebase-architecture`, etc.) for examples.

## 11. Validation checklist

Before declaring scaffold done, walk this checklist:

- [ ] Folder created at `skills/<category>/<name>/`
- [ ] `SKILL.md` present with valid YAML frontmatter (`name`, `version`, `description`)
- [ ] **Description is one line** in YAML scalar form (no `>` or `|`); has both WHAT and WHEN; version target named if applicable
- [ ] Name in frontmatter matches folder name
- [ ] Naming suffix valid per SPECS convention table
- [ ] No nested `SKILL.md` inside (single discovery point)
- [ ] All `.md` files in the folder use `UPPER_SNAKE_CASE` (except `SKILL.md` and `STACK.md` themselves)
- [ ] `STACK.md` present if category requires (per SPECS table) and not exempt; or exemption noted in roadmap
- [ ] `RECIPES.md` created if the skill will house full reference implementations (see §6); each long code block in SKILL.md replaced with a one-line pointer
- [ ] No code block in `SKILL.md` exceeds ~15 lines unless it's truly load-bearing for the rule it illustrates
- [ ] All cross-skill links resolve (relative paths point to real files)
- [ ] No emojis in `.md` files (per project convention)
- [ ] SPECS.md folder structure entry added with status marker
- [ ] SPECS.md roadmap row updated to ✅ with one-line summary
- [ ] SPECS.md bundle rows updated if skill is in any bundle
- [ ] Attribution footer added if skill is adapted from an external source
- [ ] No stale `📋 planned` rows remain for this skill in SPECS

If any item fails, fix before handing off.

## 12. Hand-off

The scaffold is ready. Content authoring is the next step:

- Each section heading needs its bullets / paragraphs / illustrative snippets.
- Each architectural opinion needs to be stated explicitly (default + alternative + when-to-switch).
- Each `STACK.md` row needs a real pinned version (fetch live from proxy.golang.org / pypi.org / GitHub releases — don't guess).
- **Long code recipes go in `RECIPES.md`**, not in `SKILL.md`. As content lands, audit any code block that grows past ~15 lines and factor it out.
- Target SKILL.md body: under ~10 KB / ~250 lines unless the topic is genuinely dense.

Surface the scaffold to the user, summarise what was generated and where, and ask: *"Want me to fill in the section content now, or would you prefer to author it yourself?"*
