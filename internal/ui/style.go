package ui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/ralvarezdev/termkit"
)

// Styles for the rsk CLI. The generic ones (accent, danger, success, warning,
// info, muted, title) come from termkit's palette; the rsk-specific badges
// below compose termkit colors with a local bold/underline where needed.
var (
	// BrandStyle is the primary branded style for the rsk name in the header.
	BrandStyle = termkit.BrandStyle

	// VersionStyle is the style for the version string next to the brand name.
	VersionStyle = termkit.MutedStyle

	// LocalStyle is the style for local (ralva) source labels in skill tables.
	LocalStyle = lipgloss.NewStyle().Foreground(termkit.ColorBrand).Bold(true)

	// OfficialStyle is the style for official (anthr) source labels in skill tables.
	OfficialStyle = lipgloss.NewStyle().Foreground(termkit.ColorInfo).Bold(true)

	// SuccessStyle is the style for success messages and check marks.
	SuccessStyle = termkit.SuccessStyle

	// WarnStyle is the style for warnings and caution messages.
	WarnStyle = termkit.WarningStyle

	// ErrorStyle is the style for error messages and failure marks.
	ErrorStyle = termkit.ErrorStyle

	// MutedStyle is the style for secondary text, dividers, and registry labels.
	MutedStyle = termkit.MutedStyle

	// BoldStyle is a generic bold style for titles and emphasis.
	BoldStyle = lipgloss.NewStyle().Bold(true)

	// BundleStyle is the style for bundle tags in status and list output.
	BundleStyle = lipgloss.NewStyle().Foreground(termkit.ColorSuccess)

	// TitleStyle is the style for section titles.
	TitleStyle = termkit.TitleStyle

	// DividerStyle is the style for divider lines between sections.
	DividerStyle = termkit.MutedStyle

	// PromptStyle is the style for interactive prompts and input labels.
	PromptStyle = termkit.AccentStyle

	// ArrowStyle is the style for the "→" transition arrow.
	ArrowStyle = termkit.MutedStyle

	// ReLinkStyle is the style for the "re-link" badge shown when a skill's
	// source path has changed and needs to be re-linked.
	ReLinkStyle = termkit.WarningStyle
)
