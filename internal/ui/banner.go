package ui

import (
	"os"

	"github.com/ralvarezdev/termkit"
)

// rskBannerSpec is rsk's BannerSpec: the wordmark art, its gradient, copy,
// and layout thresholds, built from theme.go's constants. termkit owns the
// mechanism (tier selection, centering, margins); this spec owns everything
// the banner says and looks like.
var rskBannerSpec = termkit.BannerSpec{
	Wordmark:         wordmark,
	WordmarkWidth:    wordmarkWidth,
	WordmarkGradient: wordmarkGradient,
	BrandGlyph:       brandGlyph,
	Tagline:          tagline,
	CompactHeadline:  compactHeadline,
	PlainHeadline:    plainHeadline,
	MinArtWidth:      minArtWidth,
	MinArtHeight:     minArtHeight,
	MarginTop:        headerMarginTop,
	MarginBottom:     headerMarginBottom,
	FallbackWidth:    fallbackTermWidth,
}

// NewCapability inspects the environment and returns the Capability for the
// current process. Color honors NO_COLOR and TERM=dumb; art additionally
// requires stdout to be a real terminal (IsTTY), since a wide pipe should
// still get plain text.
func NewCapability() termkit.Capability {
	return termkit.NewCapability(useColor())
}

// useColor reports whether ANSI styling should be emitted. NO_COLOR and a
// dumb terminal both win over the TTY check. termkit has no direct
// equivalent of this decision (it takes color as an already-resolved
// input), so rsk keeps it locally.
func useColor() bool {
	if _, disabled := os.LookupEnv("NO_COLOR"); disabled {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	return IsTTY()
}

// Banner renders the header shown by a bare `rsk`: the full gradient wordmark
// centered when the terminal is wide and tall enough, a compact one-line
// glyph + headline when it isn't, and a fully plain line when there is no
// color capability at all.
func Banner() string {
	return NewCapability().Banner(rskBannerSpec)
}
