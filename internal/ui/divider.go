package ui

import "strings"

const (
	// DividerLen is the number of characters in the divider line.
	DividerLen = 48

	// Padding is the standard left/right padding for UI output.
	Padding = "  "
)

func divider() string {
	return DividerStyle.Render(strings.Repeat("─", DividerLen))
}
