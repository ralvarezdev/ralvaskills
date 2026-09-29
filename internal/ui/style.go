package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles for the rsk CLI. The generic ones (accent, danger, success, warning,
// info, muted, title) come from termkit's palette; the rsk-specific badges
// below compose termkit colors with a local bold/underline where needed.
var (
	// BrandStyle is the primary branded style for the rsk name in the header.
	BrandStyle = Theme.BrandStyle()

	// VersionStyle is the style for the version string next to the brand name.
	VersionStyle = Theme.MutedStyle()

	// LocalStyle is the style for local (ralva) source labels in skill tables.
	LocalStyle = lipgloss.NewStyle().Foreground(Theme.Brand).Bold(true)

	// OfficialStyle is the style for official (anthr) source labels in skill tables.
	OfficialStyle = lipgloss.NewStyle().Foreground(Theme.Info).Bold(true)

	// SuccessStyle is the style for success messages and check marks.
	SuccessStyle = Theme.SuccessStyle()

	// WarnStyle is the style for warnings and caution messages.
	WarnStyle = Theme.WarningStyle()

	// ErrorStyle is the style for error messages and failure marks.
	ErrorStyle = Theme.ErrorStyle()

	// MutedStyle is the style for secondary text, dividers, and registry labels.
	MutedStyle = Theme.MutedStyle()

	// BoldStyle is a generic bold style for titles and emphasis.
	BoldStyle = lipgloss.NewStyle().Bold(true)

	// BundleStyle is the style for bundle tags in status and list output.
	BundleStyle = lipgloss.NewStyle().Foreground(Theme.Success)

	// TitleStyle is the style for section titles.
	TitleStyle = Theme.TitleStyle()

	// DividerStyle is the style for divider lines between sections.
	DividerStyle = Theme.MutedStyle()

	// PromptStyle is the style for interactive prompts and input labels.
	PromptStyle = Theme.AccentStyle()

	// ArrowStyle is the style for the "→" transition arrow.
	ArrowStyle = Theme.MutedStyle()

	// ReLinkStyle is the style for the "re-link" badge shown when a skill's
	// source path has changed and needs to be re-linked.
	ReLinkStyle = Theme.WarningStyle()
)
