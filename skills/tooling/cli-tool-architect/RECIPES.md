# CLI Language Recipes

Implementation skeletons for the two canonical CLI stacks. See [SKILL.md](SKILL.md) for the language-agnostic conventions these implement.

## Go — cobra + pflag + viper

Canonical per [go-architect §11](../../languages/go-architect/SKILL.md#11-dependencies--logging).

```go
// cmd/root.go
var rootCmd = &cobra.Command{
    Use:   "tool",
    Short: "Brief description of the tool.",
    Long: `Longer description, displayed under tool --help.

Use 'tool <subcommand> --help' for command-specific help.`,
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

// cmd/list.go
var listCmd = &cobra.Command{
    Use:     "list",
    Short:   "List resources.",
    Example: "  tool list --output json | jq '.[]'",
    RunE:    runList,
}

func init() {
    rootCmd.AddCommand(listCmd)
    listCmd.Flags().StringP("output", "o", "text", "Output format: text, json, yaml")
    _ = viper.BindPFlag("output", listCmd.Flags().Lookup("output"))
}
```

- **Viper for config loading**: `viper.SetEnvPrefix("TOOL")`, `viper.SetConfigName("config")`, `viper.SetConfigType("toml")`, then `viper.AddConfigPath(xdg.ConfigHome())`. Wires the SKILL.md §3 precedence automatically.
- **`charmbracelet/log`** for prettier human output when slog's default text handler feels too sparse — keeps structure, adds color and alignment.
- **`charmbracelet/lipgloss`** for tables, boxes, padded layouts in `text` output mode.
- **Completions:** `cobra` generates them — register a `completion` subcommand following the cobra docs.

## Python — typer (+ rich)

Canonical per [python-architect §10](../../languages/python-architect/SKILL.md#10-tooling).

```python
import typer
from rich.console import Console

app = typer.Typer(help="Brief description of the tool.")
err = Console(stderr=True)

@app.command()
def list(
    output: str = typer.Option("text", "--output", "-o", help="text|json|yaml"),
):
    """List resources."""
    items = fetch_items()
    if output == "json":
        typer.echo(json.dumps(items))
    elif output == "yaml":
        typer.echo(yaml.dump(items))
    else:
        for item in items:
            typer.echo(f"{item.id}\t{item.name}")

if __name__ == "__main__":
    app()
```

- **`typer.echo`** for stdout; `Console(stderr=True)` (rich) for logs.
- **Config loading:** `tomllib` (stdlib, Python 3.11+) for reading; resolve XDG paths via the `platformdirs` package or hand-rolled.
- **`rich`** for tables, progress bars, color — auto-detects TTY, respects `NO_COLOR`, integrates with Python logging.
- **Completions:** `typer` generates them — `tool --install-completion`.
- **Distribution:** `uv tool install <package>` is the cleanest install path for Python CLIs in 2026; `pipx` is the legacy default.
