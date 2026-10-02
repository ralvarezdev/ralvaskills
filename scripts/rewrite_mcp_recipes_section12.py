"""Reescribe la seccion 12 de RECIPES.md en mcp-architect.

El archivo commiteado tiene la seccion duplicada (dos copias identicas de
"## 12. Schema-driven tool contracts (Go)"). Este script deja exactamente una,
con las correcciones de v2.1.1:

- --only-models pasa a ser condicional (quita validacion, no solo tipos)
- el $ref con description hermano NO necesita allOf
- se documenta el costo real del flag segun quien deserialice

Uso: python scripts/rewrite_mcp_recipes_section12.py
"""

from pathlib import Path

RECIPES = (
    Path(__file__).resolve().parent.parent
    / "skills"
    / "protocols"
    / "mcp-architect"
    / "RECIPES.md"
)

HEADING = "## 12. Schema-driven tool contracts (Go)"
NEXT_HEADING = "## 4. Resource + template + subscription"

SECTION = '''## 12. Schema-driven tool contracts (Go)

The Go SDK derives `inputSchema` from the `In` struct by reflection and the structured-output
schema from `Out`. That is the right default for a handful of tools. Past ~10, hand-written
structs drift from each other and from the API they wrap, and nothing catches it. This is the
pipeline that fixes that: JSON Schema as the source, Go types generated, tool bodies
hand-written.

Worked example: `finance-platform`, 42 tools — arguments at `158fc32`, outputs at `eadd08c`,
shared enums at `7b7f21d`.

### Decide first: authored or derived

**Author the schemas when the tool surface is not the API surface.** An MCP tool is a curated view
for a model, not an endpoint. In finance, `profile_id` does not exist in the client types at all —
it lives in the URL path and the tool resolves it from the caller's session. `list_movements`
omits the API's `sort` and `future`. A tool that mirrors its endpoint exactly has nothing to
curate.

**Derive them when the tool is a thin pass-through** — same fields, same names, nothing added or
dropped. Then generation is right and you should not hand-maintain a parallel spec.

**Do not share MCP enums with API enums.** An MCP enum is the set a tool *accepts*; an API enum is
the domain's. `create_movement` takes 2 of the 9 movement types the API knows. One schema listing
nine values of which six error is worse than useless — the model reads it as available. Author
both and let them differ.

The failure mode to watch: reaching for the API's enum to avoid duplication, then shipping a tool
whose schema advertises nine values it cannot use.

### Layout

Keep schemas at the repo root beside the OpenAPI tree, with generated Go under the module:

```
openapi/contexts/*.yaml    → backend/internal/api/handler/*.gen.go
json/mcpserver/*/*.json    → backend/internal/mcpserver/schemas/*.gen.go
```

Putting `json/` inside the module inverts the repo's own convention — no other schema tree lives
there.

### Arguments

`github.com/atombender/go-jsonschema` generates the Go struct. These flags are load-bearing:

```
go run github.com/atombender/go-jsonschema@latest \\
  -p movement -o list.gen.go \\
  --disable-omitzero \\                  # drop the omitzero tag
  --capitalization ID \\                 # profile_id → ProfileID, not ProfileId
  --struct-name-from-title \\            # title → struct name, not the filename
  --tags json                           # drop yaml/mapstructure
```

Add `--only-models` only when nothing deserializes these types outside the SDK — see below.

- **`--struct-name-from-title` is not optional.** Without it the type is named after the *file*:
  `list.json` gives `ListJson`, `list_arguments.json` gives `ListArgumentsJson`. The schema's
  `title` is what names the type, so every schema needs one.
- The generated file says `DO NOT EDIT`. Commit it (same as the OpenAPI's generated types) so a
  plain `go build` works without the generator, and add a check task that regenerates and diffs.

#### `--only-models`: decide by who deserializes the struct

The default emits a custom `UnmarshalJSON` per enum. It is **not** redundant — it is the only
validation these types get when unmarshaled outside the SDK. It enforces `required`, `minimum`,
and the enum set:

```go
func (j *Req) UnmarshalJSON(value []byte) error {
    if _, ok := raw["name"]; raw != nil && !ok {
        return fmt.Errorf("field name in Req: required")
    }
    ...
    if plain.N != nil && 5 > *plain.N {
        return fmt.Errorf("field %s: must be >= %v", "n", 5)
    }
```

`--only-models` removes all of it: a `n: 3` against `minimum: 5` then deserializes cleanly. The
choice is not "duplicate types or clean types":

| Who unmarshals the struct | Flag | Why |
|---|---|---|
| Only `mcp.AddTool` | `--only-models` | The SDK already validates the input against the `inputSchema` it derived from this same schema. The rest is dead code. |
| Also a CLI, a test, a config loader | omit it | There it is the only validation, and dropping it silently accepts bad data. |

Ask who consumes the types before choosing — it is not knowable from the tool code.

The flag also removes a duplicate type. Without it, a shared enum referenced from a consumer file
yields a *third* type named after that consumer's filename (`usea.json` → `UseADialect`) carrying
the same values as the shared one. If you cannot use `--only-models`, give each enum its own file
so the derived name cannot collide.

### Outputs

Do **not** generate Go for outputs. Declare the schema and hand it to the SDK:

```go
mcp.AddTool(server, &mcp.Tool{
    Name:         "get_movement",
    OutputSchema: outputs["MovementOutputSchema"],
}, handler)
```

The SDK validates the result against it before it reaches the wire (`mcp/server.go`, "validating
tool output"), so a wrong shape fails at the call instead of confusing the model.

**Describe the wire, not the Go type.** Three cases where they differ, each of which had to be
fixed in the reference implementation:

- **Money as string.** `Decimal` was `type Decimal = string` — a deliberate alias so no float
  rounding enters. A `"type": "number"` schema makes the server reject its own output.
- **The re-wrapped envelope.** `list_movements` returns `movements`, not the API envelope's `data`,
  because the HTTP client re-wraps after decoding.
- **`omitempty` fields are optional.** The paging meta serializes with `omitempty`, so a first
  empty page omits every counter. Declaring them `required` rejects the server's own empty
  response.

Read the actual type and its JSON tags; do not assume the Go field name is the wire key
(`per_page` is `perPage`).

### Three things that will bite you

**1. The SDK resolves `OutputSchema` with no loader.** A `$ref` to another document panics at
`AddTool` — this is startup, not a validation error:

```go
panic: AddTool "list_movements": output schema: loading ./x.json:
  cannot resolve remote schemas: no loader passed to Schema.Resolve
```

Inline every `$ref` at generation time; the resolver then only ever sees one self-contained
document.

**2. `go-jsonschema` does not follow `$ref` at all.** A referenced property silently becomes
`interface{}` — the type and enum gone, with no error:

```go
Type interface{} `json:"type"`   // was {"type":"string","enum":["income","expense"]}
```

Same fix, different reason. Both paths should share one inliner so there is a single definition
of "resolved".

**3. `go-jsonschema` does not descend into `allOf`.** Relevant only when you use `allOf` to
*override* a shared definition's constraint — flatten the single-element wrapper before
generating, or the property degrades to `interface{}` like any unresolved ref.

Note what is **not** a trap: a `$ref` with a sibling `description` needs no wrapper. In draft
2020-12 `go-jsonschema` reads it directly and keeps the comment on the generated field.

### Sharing

Put reusable definitions in one file and reference them:

```
json/mcpserver/shared/enums.json
  $defs: movement_type_input, movement_type_output, currency, decimal, date

json/mcpserver/movement/create_arguments.json
  "type": { "$ref": "../shared/enums.json#/$defs/movement_type_input" }
```

Commit both the sources (with refs) and the resolved copies — the sources so a reader sees the
sharing, the resolved ones because that is what the generators read. Resolve relative refs against
the containing file's directory per RFC 3986, not the schema root.

Every `$defs` entry needs a `description`: it is inlined into every schema that references it, so
it is the only text the model reads about those values.

On nullability: `anyOf: [{$ref}, {type: null}]` degrades to `interface{}` like any other
indirection. Keep optional properties a plain `$ref` and let `null` live in the definition when the
field is genuinely nullable — then cover it with a test, because an explicit `null` against an enum
that does not list `null` is exactly what your optional fields will hit.

### Verify

Test both directions, and prove the rejection tests are not vacuous — mutate the schema and watch
the test fail:

```go
func TestOutputSchemaRejectsNumericAmount(t *testing.T) {
    // amount: "42.50" passes; amount: 42.5 must be refused
    // flipping the schema's amount to "number" makes this test fail
}
```

Also assert the schema is an object with `additionalProperties: false`, that every property has a
description, and that a cross-file `$ref` was inlined rather than left dangling. A schema linter
over the source tree catches malformed JSON before generation does — go-jsonschema is lenient and
will happily emit Go from a broken schema.

Expect existing tests to need conforming fixtures. A fake response missing a field the schema
requires is exactly the bug the schema exists to catch.

---

'''


def main() -> int:
    lines = RECIPES.read_text(encoding="utf-8").split("\n")

    # localizar cada copia de la seccion
    starts = [i for i, l in enumerate(lines) if l == HEADING]
    if not starts:
        print("seccion 12 no encontrada")
        return 1
    if len(starts) > 1:
        print(f"seccion 12 duplicada x{len(starts)}; se conserva la primera")

    first = starts[0]
    # el fin de la seccion es el proximo heading de nivel ## que no sea ella misma
    end = len(lines)
    for i in range(first + 1, len(lines)):
        if lines[i].startswith("## ") and lines[i] != HEADING:
            end = i
            break

    # absorber el separador '---' y la linea en blanco previos al siguiente heading
    while end > first and lines[end - 1].strip() in ("", "---"):
        end -= 1

    out = lines[:first] + SECTION.split("\n") + lines[end:]
    RECIPES.write_text("\n".join(out), encoding="utf-8")

    check = RECIPES.read_text(encoding="utf-8").split("\n")
    n = sum(1 for l in check if l == HEADING)
    print(f"seccion 12 reescrita: {len(check)} lineas, {n} copia(s)")
    return 0 if n == 1 else 1


if __name__ == "__main__":
    raise SystemExit(main())