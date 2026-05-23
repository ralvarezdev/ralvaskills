package ui

import (
	"errors"
	"os"

	"github.com/mattn/go-isatty"
)

// ErrAborted is returned when the user cancels an interactive prompt (Ctrl+C / Esc).
// Callers should propagate it unchanged; root.Execute handles it as a silent exit 130.
var ErrAborted = errors.New("user aborted")

// IsTTY reports whether both stdin and stdout are attached to a real terminal.
// When false, interactive TUI widgets are replaced with plain-text fallbacks so
// the CLI works correctly in scripts, CI pipelines, and piped invocations.
func IsTTY() bool {
	return (isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())) &&
		(isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()))
}
