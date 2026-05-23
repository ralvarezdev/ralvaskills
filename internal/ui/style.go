package ui

import "github.com/charmbracelet/lipgloss"

// muted is shared by VersionStyle, MutedStyle, DividerStyle, and ArrowStyle.
var muted = lipgloss.Color("#6B7280")

var (
	// BrandStyle is the primary branded style for the rsk name in the header.
	BrandStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#2DD4BF")).Bold(true)

	// VersionStyle is the style for the version string next to the brand name.
	VersionStyle = lipgloss.NewStyle().Foreground(muted)

	// LocalStyle is the style for local (ralva) source labels in skill tables.
	LocalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#A78BFA")).Bold(true)

	// OfficialStyle is the style for official (anthr) source labels in skill tables.
	OfficialStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA")).Bold(true)

	// SuccessStyle is the style for success messages and check marks.
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ADE80")).Bold(true)

	// WarnStyle is the style for warnings and caution messages.
	WarnStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24")).Bold(true)

	// ErrorStyle is the style for error messages and failure marks.
	ErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")).Bold(true)

	// MutedStyle is the style for secondary text, dividers, and registry labels.
	MutedStyle = lipgloss.NewStyle().Foreground(muted)

	// BoldStyle is a generic bold style for titles and emphasis.
	BoldStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F9FAFB")).Bold(true)

	// BundleStyle is the style for bundle tags in status and list output.
	BundleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#34D399"))

	// TitleStyle is the style for section titles.
	TitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F9FAFB")).Bold(true)

	// DividerStyle is the style for divider lines between sections.
	DividerStyle = lipgloss.NewStyle().Foreground(muted)

	// PromptStyle is the style for interactive prompts and input labels.
	PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#2DD4BF")).Bold(true)

	// ArrowStyle is the style for the "→" transition arrow.
	ArrowStyle = lipgloss.NewStyle().Foreground(muted)

	// ReLinkStyle is the style for the "re-link" badge shown when a skill's
	// source path has changed and needs to be re-linked.
	ReLinkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24"))
)
