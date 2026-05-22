# CLI Language Recipes

Implementation skeletons for the two canonical CLI stacks. See [SKILL.md](SKILL.md) for the language-agnostic conventions these implement.

## Exit-code reference

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Generic / unspecified failure |
| `2` | Misuse — bad flag, missing required arg, invalid subcommand |
| `64`–`78` | sysexits.h categories — useful for shell scripts to dispatch on |
| `126` | Command found but not executable (rare for user CLIs) |
| `127` | Command not found |
| `130` | Killed by SIGINT (Ctrl-C) — runtime handles this |
| custom | Domain-specific (`tool deploy` could return `10` for "deploy refused: dirty tree", `11` for "deploy refused: missing approval") |

- Document every non-standard exit code in `--help` or a dedicated `tool help exit-codes` page.
- Don't reuse codes across categories within one tool — once `10` means "dirty tree", it can't later mean something else.
- Misuse vs failure: parsing errors are 2; the tool worked but the operation failed is 1 or a custom non-zero.

## Output discipline — stdout/stderr separation

```bash
tool list > items.json       # writes data only to items.json
tool list 2> tool.log         # writes logs only to tool.log
tool list | jq '.[] | .name'  # pipes data into jq; logs still visible on the terminal
```

- Data → stdout. The thing the user wants — JSON, the filename created, the resource ID, the table.
- Logs, progress, errors → stderr.
- Errors that prevent producing data exit non-zero.

## Output format flag examples

```
tool list                          # tabular text
tool list --output json            # machine-parseable
tool list --output yaml            # machine-parseable, human-skimmable
tool get user 01J9... -o json      # short form
```

- `-o` short form is standard (`kubectl`, `gh`, `oc`).
- JSON output is stable and documented — clients depend on it. Schema changes are breaking.
- Don't pretty-print JSON by default — emit JSONL for list operations so `tool list -o json | grep` works.
- Use `--output yaml` for human eyeballing of complex nested data; rarely useful for pipes.

## Error message format

```
ERR  apply: deploy refused
     reason: dirty working tree (3 uncommitted files)
     hint:   commit or stash, then retry
```

Structured log example:

```
ERR  apply: deploy refused  reason="dirty working tree"  files=3  hint="commit or stash, then retry"
```

- No stack traces in user-facing errors. Log them at `debug` level (visible with `-vv`) or write to a file with a correlation id.
- Suggest the fix when the cause is unambiguous: `Did you mean: tool depoly → deploy?` (Levenshtein distance from known subcommands).
- Distinguish user errors from tool bugs. "Invalid flag value" → user error, exit 2. "Internal: nil pointer in handler" → bug, exit 1 with a "please file an issue at <repo>/issues" hint.

## Version output

```
tool 1.2.3 (rev abc1234, built 2026-05-21, go1.26)
```

Version + git short SHA + build date + runtime version. The SHA is critical for dev builds. `tool version --output json` for scripting.

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
