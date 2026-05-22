// Package ui provides styled terminal output for the rsk CLI.
package ui

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	// ColorTeal is a bright teal used for branding and prompts.
	ColorTeal = "#2DD4BF"

	// ColorViolet is a vibrant violet used for local (ralva) skill labels.
	ColorViolet = "#A78BFA"

	// ColorBlue is a bright blue used for official (anthr) skill labels.
	ColorBlue = "#60A5FA"

	// ColorGreen is a fresh green used for success messages and badges.
	ColorGreen = "#4ADE80"

	// ColorAmber is a warm amber used for warnings and cautions.
	ColorAmber = "#FBBF24"

	// ColorRed is a bold red used for errors and failures.
	ColorRed = "#F87171"

	// ColorGray is a neutral gray used for secondary text and dividers.
	ColorGray = "#6B7280"

	// ColorWhite is a near-white used for primary text on dark backgrounds.
	ColorWhite = "#F9FAFB"

	// ColorEmerald is a bright emerald used for bundle tags.
	ColorEmerald = "#34D399"
)

var (
	// ColorBrand is the primary color used for rsk branding and prompts.
	ColorBrand    = lipgloss.Color(ColorTeal)

	// ColorLocal is the color used for local (ralva) skills in source labels.
	ColorLocal    = lipgloss.Color(ColorViolet)

	// ColorOfficial is the color used for official (anthr) skills in source labels.
	ColorOfficial = lipgloss.Color(ColorBlue)

	// ColorSuccess is the color used for success messages and badges.
	ColorSuccess  = lipgloss.Color(ColorGreen)

	// ColorWarn is the color used for warnings and caution messages.
	ColorWarn     = lipgloss.Color(ColorAmber)

	// ColorError is the color used for error messages and badges.
	ColorError    = lipgloss.Color(ColorRed)

	// ColorMuted is the color used for secondary text, dividers, and registry skills.
	ColorMuted    = lipgloss.Color(ColorGray)

	// ColorText is the color used for primary text on dark backgrounds.
	ColorText     = lipgloss.Color(ColorWhite)

	// ColorBundle is the color used for bundle tags in status and list outputs.
	ColorBundle   = lipgloss.Color(ColorEmerald)
)
