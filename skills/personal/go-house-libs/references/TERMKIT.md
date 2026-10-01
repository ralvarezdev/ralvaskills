# termkit

`github.com/ralvarezdev/termkit` — presentation primitives for Go CLIs: an adaptive light/dark theme, capability detection, tables, charts, typed parameter forms, an interactive session shell over a cobra tree, and output in table/json/yaml/csv. Reach for it before hand-writing table alignment, ANSI-aware truncation, a color theme, or a flags-to-form prompt.

## Packages

| Package | Owns |
|---|---|
| `termkit` | `Theme` (`TokyoNight`, `Catppuccin`, `ThemeFor`), `ColorEnabled`, `Capability`, ANSI-aware text helpers (`Width`, `Truncate`, `Wrap`, `StripANSI`), `Table[T]` + `WriteTable`, `Status`/`RegisterStatus`, `Capture`/`Panel`, `KeyMap`, `Progress` |
| `termkit/chart` | `Chart` (line; `Fit`: `FitStretch` default, `FitNatural`), `BarChart`, `Spark`, `Plain` fallback; drawn in the active theme (its own package, so CLIs that do not chart never compile the chart libraries) |
| `termkit/output` | renders a command result as a terminal table (optionally with a chart) or json/yaml/csv; `Options.Layout` / `SetLayout` frame tables (`TableBoxed`, `TableBorderless`, `TableHeaderless`), `Table[T].Title` and `SetCaptureTitle` label captured tables, `ChartWith(title, series, fit)` picks a chart fit |
| `termkit/cli` | cobra plumbing for the shared `--output/-o` flag: `Printer` (from `OutputConfig`) adds the flag, validates it, completes it, and resolves the `output.Formatter`; `WithFormatter`/`FormatterFromContext` hand a session's formatter to an in-process command |
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

## Recipe: the shared `--output` flag

```go
printer := cli.NewPrinter(cli.OutputConfig{
    Persistent:    true,                     // flag inherited by every subcommand
    DefaultFormat: output.FormatTable,       // shown as the flag default in --help
    Default: func(cmd *cobra.Command) output.OutputFormat { return cfgFormat(cmd) }, // e.g. from config
    Options: func(cmd *cobra.Command) output.Options { return output.Options{Layout: termkit.TableBorderless} },
})
printer.AddFlag(rootCmd, output.FormatCSV)   // base is table|json|yaml; extras per command

// in RunE: one render path for table, json, yaml, csv, and a session capture
return printer.Print(cmd, projects, func(f *output.Formatter) {
    _ = output.Render(f, projects, projectOutTable)
})
```

## Rules

- **Color is decided once, by `ColorEnabled(mode, noColor, forceColor)`,** and machine-readable formats (`output` for `--output json|yaml|csv`) are never colored. The stdout-data / stderr-logs split and `NO_COLOR` are the CLI's job under [cli-tool-architect](../../../tooling/cli-tool-architect/SKILL.md).
- **Numeric and duration columns right-align and known status words are colorized** by `WriteTable`; register project-specific status words with `RegisterStatus` instead of coloring by hand.
- **One theme, passed as a value.** Charts, tables, forms and markdown all follow the capability's `Theme`; do not hardcode colors.
- **Validation failures go back to the form**: a `RunFunc` returns `termkit.FieldError` and the session keeps the entered values.
- **Use `Capture` for in-TUI rendering** instead of writing to stdout inside a session run.
- **`--output` is validated before the command runs** (after any existing `PreRunE`/`PersistentPreRunE`, so a persistent flag's `Default` can read loaded config); `text` stays accepted as the legacy spelling of `table`. The flag is excluded from generated session forms unless `ExcludeForm` is false, and `Default` is not consulted when `--output` is given.
- **Under a session capture, `-o json` is recorded as the result's raw viewport** instead of being written, so the interactive view can show it; a table's `Layout` is ignored under a capture (the session renders it).
- **Line charts stretch by default**: a series shorter than the plot is interpolated across the full width and a longer one is downsampled to show its whole history. Set `Chart.Fit = chart.FitNatural` when X is a real scale (one point per column).
- **The session picker lists runnable commands that have subcommands** (e.g. `catalog` and `catalog bundles` both appear); only pure groups (no `Run`/`RunE`) are collapsed.
- **Import `chart`, `huhform` and `session` only where used**; they carry the heavy dependencies.

## Consumers and stability

Used by several CLIs. Pre-1.0 (`v0.x`, current v0.64.1): the API may change between minor versions. It is on Go 1.27.1, newer than the other kits ([../STACK.md](../STACK.md)). The module's `docs/` folder (`form-scope.md`, `session-scope.md`, `charts-scope.md`) holds the open work and lists per-CLI copies of this code that still need migrating.

## Not here

Command structure, flag/env/config precedence, exit codes and completions belong to [cli-tool-architect](../../../tooling/cli-tool-architect/SKILL.md); this kit renders and collects, it does not define a CLI's commands.
