package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/huh"
)

// ConfirmYN prints a styled [y/N] prompt and returns true for "y" or "yes".
// Uses a huh Confirm field when running in a TTY; falls back to plain text otherwise.
func ConfirmYN(w io.Writer, prompt string) bool {
	if IsTTY() {
		return confirmHuh(prompt)
	}
	return confirmPlain(w, prompt)
}

func confirmHuh(prompt string) bool {
	result := false
	f := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title(prompt).
			Value(&result),
	)).WithTheme(huh.ThemeCharm())
	if err := f.Run(); err != nil {
		return false
	}
	return result
}

func confirmPlain(w io.Writer, prompt string) bool {
	_, _ = fmt.Fprintf(w, "\n"+Padding+"%s %s ", PromptStyle.Render(prompt), MutedStyle.Render("[y/N]"))

	raw, err := stdinReader.ReadString('\n')
	if err != nil {
		return false
	}
	_, _ = fmt.Fprintln(w)

	line := strings.ToLower(strings.TrimSpace(raw))
	return line == "y" || line == "yes"
}
