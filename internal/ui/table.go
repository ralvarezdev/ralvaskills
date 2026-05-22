package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// PadRight pads s to visual width n, correctly handling ANSI escape codes.
func PadRight(s string, n int) string {
	w := lipgloss.Width(s)
	if w >= n {
		return s
	}
	return s + strings.Repeat(" ", n-w)
}

// MaxWidth returns the visual width of the longest string in items.
func MaxWidth(items []string) int {
	w := 0
	for _, s := range items {
		if vw := lipgloss.Width(s); vw > w {
			w = vw
		}
	}
	return w
}

// SkillName returns a styled skill name for table rows.
func SkillName(name string) string {
	return BoldStyle.Render(name)
}

// SkillVersion returns a styled version string for table rows.
func SkillVersion(version string) string {
	if version == "" {
		return MutedStyle.Render("-")
	}
	return MutedStyle.Render(version)
}

// MutedPath returns a dimmed path string.
func MutedPath(p string) string {
	return MutedStyle.Render(p)
}
