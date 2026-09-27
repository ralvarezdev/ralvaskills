package ui

import (
	"github.com/ralvarezdev/termkit"
)

// PadRight pads s to visual width n, correctly handling ANSI escape codes.
//
// Deprecated: use termkit.PadRight.
func PadRight(s string, n int) string { return termkit.PadRight(s, n) }

// MaxWidth returns the visual width of the longest string in items.
//
// Deprecated: use termkit.MaxWidth.
func MaxWidth(items []string) int { return termkit.MaxWidth(items) }

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
