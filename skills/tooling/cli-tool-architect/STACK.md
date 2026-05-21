# Stack Versions

This skill pins **CLI-specific additions** beyond the core language stacks. Core CLI libraries (cobra, pflag, viper, typer) are canonical in the language-architect STACK files.

| Dependency | Pinned version | Purpose |
|---|---|---|
| spf13/cobra | 1.10 | Go CLI framework — see [go-architect](../../languages/go-architect/STACK.md) |
| spf13/pflag | 1.0 | POSIX-style flags for cobra — see [go-architect](../../languages/go-architect/STACK.md) |
| spf13/viper | 1.21 | Layered config (flag/env/file/default precedence) — see [go-architect](../../languages/go-architect/STACK.md) |
| typer | 0.25 | Python CLI framework — see [python-architect](../../languages/python-architect/STACK.md) |
| charmbracelet/log | 1.0 | **Optional** Go — prettier human-facing structured logs alongside `slog` |
| charmbracelet/lipgloss | 1.1 | **Optional** Go — tables, boxes, padded layouts for `text` output mode |
| charmbracelet/bubbletea | 1.3 | **Optional** Go — full TUI when you need interactive screens (not a default) |
| rich | 15.0 | **Default** Python — color, tables, progress bars; TTY-aware |
| click | 8.4 | **Alternative** Python CLI framework — `typer` is built on it; use directly only for very low-level needs |

## Notes

- **Config file format:** **TOML** by default (`config.toml` in XDG location). Use the stdlib `tomllib` (Python 3.11+) or `viper.SetConfigType("toml")` (Go).
- **XDG paths:** use `github.com/adrg/xdg` (Go) or `platformdirs` (Python) — both handle Windows / macOS / Linux paths correctly. Don't hand-roll.
- **Discovery order for config:** `--config <path>` flag → `TOOL_CONFIG` env var → `./<tool>.toml` → `$XDG_CONFIG_HOME/<tool>/config.toml`. Wire via viper / typer-config.
- **Output formats:** every list/get subcommand supports `--output text|json|yaml` (`-o` short form). JSON is JSONL for list outputs (one record per line).
- **Color & TTY:** auto-detect TTY; respect `NO_COLOR` and `FORCE_COLOR` env vars; expose `--color=auto|always|never` flag.
- **Distribution:** ship multi-arch static binaries (linux/macOS/windows × amd64/arm64) — `goreleaser` for Go, `uv tool install` for Python.
- **`rsk` uses JSON** for its config file (legacy); TOML is the recommended default for new CLIs authored under this skill.

_Last reviewed: 2026-05-21_
_Skill version at last review: 1.0.0_
