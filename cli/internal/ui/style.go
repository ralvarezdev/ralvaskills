package ui

import "github.com/charmbracelet/lipgloss"

var (
	// BrandStyle is the style used for the rsk name in the header.
	BrandStyle    = lipgloss.NewStyle().Foreground(ColorBrand).Bold(true)

	// VersionStyle is the style used for the version string in the header.
	VersionStyle  = lipgloss.NewStyle().Foreground(ColorMuted)

	// LocalStyle is the style used for local (ralva) skills in source labels.
	LocalStyle    = lipgloss.NewStyle().Foreground(ColorLocal).Bold(true)

	// OfficialStyle is the style used for official (anthr) skills in source labels.
	OfficialStyle = lipgloss.NewStyle().Foreground(ColorOfficial).Bold(true)

	// SuccessStyle is the style used for success messages and badges.
	SuccessStyle  = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)

	// WarnStyle is the style used for warnings and caution messages.
	WarnStyle     = lipgloss.NewStyle().Foreground(ColorWarn).Bold(true)

	// ErrorStyle is the style used for error messages and badges.
	ErrorStyle    = lipgloss.NewStyle().Foreground(ColorError).Bold(true)

	// MutedStyle is the style used for secondary text, dividers, and registry skills.
	MutedStyle    = lipgloss.NewStyle().Foreground(ColorMuted)

	// BoldStyle is a generic bold style for titles and important text.
	BoldStyle     = lipgloss.NewStyle().Foreground(ColorText).Bold(true)

	// BundleStyle is the style used for bundle tags in status and list outputs.
	BundleStyle   = lipgloss.NewStyle().Foreground(ColorBundle)

	// TitleStyle is the style used for titles in the UI.
	TitleStyle    = lipgloss.NewStyle().Foreground(ColorText).Bold(true)

	// DividerStyle is the style used for divider lines between sections and tables.
	DividerStyle  = lipgloss.NewStyle().Foreground(ColorMuted)

	// PromptStyle is the style used for user prompts and input labels.
	PromptStyle   = lipgloss.NewStyle().Foreground(ColorBrand).Bold(true)

	// ArrowStyle is the style used for the "→" arrow in relinking messages and other transitions.
	ArrowStyle    = lipgloss.NewStyle().Foreground(ColorMuted)

	// ReLinkStyle is the style used for the "relinking" message when a local skill's source path has changed.
	ReLinkStyle   = lipgloss.NewStyle().Foreground(ColorWarn)
)
