package ui

import "github.com/charmbracelet/lipgloss"

// Adaptive palette. Each color picks between a light-background and a
// dark-background value so the banner and picker stay legible either way.
// These mirror the fixed colors in style.go, which most output helpers still
// use directly — this palette exists for the new banner/picker surfaces.
var (
	// ColorAccent is the brand accent: the wordmark, the picker's selection,
	// and picker titles.
	ColorAccent = lipgloss.AdaptiveColor{Light: "#0F766E", Dark: "#2DD4BF"}

	// ColorDanger marks errors and destructive actions.
	ColorDanger = lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#F87171"}

	// ColorWarning marks warnings and caution messages.
	ColorWarning = lipgloss.AdaptiveColor{Light: "#B45309", Dark: "#FBBF24"}

	// ColorMuted is for secondary text: hints, descriptions, key help.
	ColorMuted = lipgloss.AdaptiveColor{Light: "#52525B", Dark: "#A1A1AA"}
)

// wordmarkGradient shades the wordmark row by row along the accent hue.
var wordmarkGradient = []lipgloss.AdaptiveColor{
	{Light: "#14B8A6", Dark: "#5EEAD4"},
	{Light: "#0F9488", Dark: "#2DD4BF"},
	{Light: "#0D7A70", Dark: "#14B8A6"},
	{Light: "#0F766E", Dark: "#0D9488"},
}

const (
	// wordmark is the rsk banner, wordmarkWidth columns by six rows. Rows are
	// padded to the full width at render time so centering keeps them aligned.
	wordmark = `██████╗  ███████╗ ██╗  ██╗
██╔══██╗ ██╔════╝ ██║ ██╔╝
██████╔╝ ███████╗ █████╔╝
██╔══██╗ ╚════██║ ██╔═██╗
██║  ██║ ███████║ ██║  ██╗
╚═╝  ╚═╝ ╚══════╝ ╚═╝  ╚═╝`
	wordmarkWidth = 26

	brandGlyph      = "◆"
	tagline         = "manage your AI skills"
	compactHeadline = "rsk · AI skills manager"
	plainHeadline   = "rsk: manage your AI skills"
)

// Layout. The header is centered in the terminal with a blank row above and
// below; under minArtWidth x minArtHeight it drops to a compact one-liner.
const (
	headerMarginTop    = 1
	headerMarginBottom = 1
	headerSideMargin   = 2
	minArtWidth        = wordmarkWidth + 2*headerSideMargin + 2
	minArtHeight       = 18

	// fallbackTermWidth centers the static banner when the terminal will not
	// report its size.
	fallbackTermWidth = 80

	// defaultPickerWidth and defaultPickerHeight size the picker's list
	// before the terminal reports its size; minPickerListHeight is the floor
	// so a tiny terminal still shows some rows.
	defaultPickerWidth  = 80
	defaultPickerHeight = 24
	minPickerListHeight = 6
)
