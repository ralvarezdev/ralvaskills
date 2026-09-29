package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"

	"github.com/ralvarezdev/termkit"
)

// Theme is the termkit theme every rsk style, mark and view renders with.
var Theme = termkit.TokyoNight

// Adaptive palette, re-exported from the Theme so the banner and picker share
// one palette with every other consumer.
var (
	// ColorAccent is the brand accent: the wordmark, the picker's selection,
	// and picker titles.
	ColorAccent = Theme.Accent

	// ColorDanger marks errors and destructive actions.
	ColorDanger = Theme.Danger

	// ColorWarning marks warnings and caution messages.
	ColorWarning = Theme.Warning

	// ColorMuted is for secondary text: hints, descriptions, key help.
	ColorMuted = Theme.Muted
)

// wordmarkGradientSteps is how far each successive wordmark row darkens
// toward black, roughly matching the falloff of the old hand-picked teal
// ramp this replaced.
var wordmarkGradientSteps = []float64{0, 0.18, 0.34, 0.48}

// wordmarkGradient shades the wordmark row by row, darkened down from the
// active termkit Palette's accent color — so the banner follows whatever
// palette termkit is set to instead of a color baked in for one specific
// scheme.
var wordmarkGradient = buildWordmarkGradient()

func buildWordmarkGradient() []lipgloss.AdaptiveColor {
	out := make([]lipgloss.AdaptiveColor, len(wordmarkGradientSteps))
	for i, t := range wordmarkGradientSteps {
		out[i] = lipgloss.AdaptiveColor{
			Light: darken(Theme.Accent.Light, t),
			Dark:  darken(Theme.Accent.Dark, t),
		}
	}
	return out
}

// darken blends hex toward black by t (0 = unchanged, 1 = black) in Lab
// space, which keeps the blend perceptually even. An unparseable hex is
// returned unchanged rather than panicking.
func darken(hex string, t float64) string {
	c, err := colorful.Hex(hex)
	if err != nil {
		return hex
	}
	return c.BlendLab(colorful.Color{}, t).Hex()
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
)
