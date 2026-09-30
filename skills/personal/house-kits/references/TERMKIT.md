# termkit

`github.com/ralvarezdev/termkit` — presentation primitives for Go CLIs: an adaptive light/dark theme, capability detection, tables, charts, typed parameter forms, an interactive session shell over a cobra tree, and output in table/json/yaml/csv. Reach for it before hand-writing table alignment, ANSI-aware truncation, a color theme, or a flags-to-form prompt.

## Packages

| Package | Owns |
|---|---|
| `termkit` | `Theme` (`TokyoNight`, `Catppuccin`, `ThemeFor`), `ColorEnabled`, `Capability`, ANSI-aware text helpers (`Width`, `Truncate`, `Wrap`, `StripANSI`), `Table[T]` + `WriteTable`, `Status`/`RegisterStatus`, `Capture`/`Panel`, `KeyMap`, `Progress` |
| `termkit/chart` | `Chart`, `BarChart`, `Spark`, `Plain` fallback; drawn in the active theme (its own package, so CLIs that do not chart never compile the chart libraries) |
| `termkit/output` | renders a command result as a terminal table (optionally with a chart) or json/yaml/csv |
| `termkit/form` | infers a field kind per cobra flag (bool, enum, date, numeric, file, text) and validates it; `huhform` renders the fields with charmbracelet/huh |
| `termkit/session` | the shared picker → form → captured run → result-paging shell over a cobra tree |
| `termkit/dates` | absolute and relative date parsing (`-1m`, `+2w`, `start of month`, `last friday`) |
| `termkit/markdown` | Markdown to ANSI in the active palette, for help and about text |
| `termkit/history`, `termkit/prefs`, `termkit/config` | optional file-backed command history, remembered theme and field values, and key/behavior settings loaded from viper |

## Recipe

```go
var projectTable = termkit.Table[Project]{
    Headers: []string{"ID", "Name"},
    Row:     func(p Project) []any { return []any{p.ID, p.Name} },
}

color := termkit.ColorEnabled(termkit.ColorAuto, noColor, forceColor)
rows := termkit.Rows(projects, projectTable)
termkit.WriteTable(os.Stdout, projectTable.Headers, rows, color, projectTable.Sort, termkit.TokyoNight)

cap := termkit.NewCapability(color)
fmt.Println(chart.Spark{Values: []float64{1, 3, 2, 5, 4}}.Render(30, cap))
```

## Rules

- **Color is decided once, by `ColorEnabled(mode, noColor, forceColor)`,** and machine-readable formats (`output` for `--output json|yaml|csv`) are never colored. The stdout-data / stderr-logs split and `NO_COLOR` are the CLI's job under [cli-tool-architect](../../../tooling/cli-tool-architect/SKILL.md).
- **Numeric and duration columns right-align and known status words are colorized** by `WriteTable`; register project-specific status words with `RegisterStatus` instead of coloring by hand.
- **One theme, passed as a value.** Charts, tables, forms and markdown all follow the capability's `Theme`; do not hardcode colors.
- **Validation failures go back to the form**: a `RunFunc` returns `termkit.FieldError` and the session keeps the entered values.
- **Use `Capture` for in-TUI rendering** instead of writing to stdout inside a session run.
- **Import `chart`, `huhform` and `session` only where used**; they carry the heavy dependencies.

## Consumers and stability

finance-platform's CLI, devtrack and rsk (ralvaskills). Pre-1.0 (`v0.x`): the API may change between minor versions. It is on Go 1.27.1, newer than the other kits ([../STACK.md](../STACK.md)). The module's `docs/` folder (`form-scope.md`, `session-scope.md`, `charts-scope.md`) holds the open work and lists per-CLI copies of this code that still need migrating.

## Not here

Command structure, flag/env/config precedence, exit codes and completions belong to [cli-tool-architect](../../../tooling/cli-tool-architect/SKILL.md); this kit renders and collects, it does not define a CLI's commands.
