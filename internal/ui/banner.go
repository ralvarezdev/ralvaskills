package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
)

// Capability describes what the current terminal can render: whether color is
// wanted, whether ASCII art is appropriate, and the terminal width at
// start-up. It is decided once, not re-checked per call, so a single process
// renders consistently even if something changes stdout mid-run.
type Capability struct {
	color bool
	art   bool
	width int
}

// NewCapability inspects the environment and returns the Capability for the
// current process. Color honors NO_COLOR and TERM=dumb; art additionally
// requires stdout to be a real terminal (IsTTY), since a wide pipe should
// still get plain text.
func NewCapability() Capability {
	capa := Capability{color: useColor(), art: IsTTY() && os.Getenv("TERM") != "dumb"}
	if width, _, err := term.GetSize(os.Stdout.Fd()); err == nil {
		capa.width = width
	}
	return capa
}

// useColor reports whether ANSI styling should be emitted. NO_COLOR and a
// dumb terminal both win over the TTY check.
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
	capa := NewCapability()

	if !capa.art {
		return plainHeadline
	}

	width := capa.width
	if width <= 0 {
		width = fallbackTermWidth
	}

	return capa.Header(width, minArtHeight) + "\n" + capa.center(capa.muted(tagline), width)
}

// Header renders the picker's top: the wordmark centered in width columns
// with a margin row above and below when there is room for it and height,
// otherwise a single compact line.
func (c Capability) Header(width, height int) string {
	top := strings.Repeat("\n", headerMarginTop)
	bottom := strings.Repeat("\n", headerMarginBottom)

	if width < minArtWidth || height < minArtHeight {
		line := c.accent(brandGlyph+" "+compactHeadline, true)
		return top + c.center(line, width) + bottom
	}

	return top + c.center(c.wordmarkBlock(), width) + bottom
}

// wordmarkBlock renders the wordmark down the gradient with the tagline
// beneath it, as one block whose rows share a left edge.
func (c Capability) wordmarkBlock() string {
	rows := strings.Split(wordmark, "\n")
	for i, row := range rows {
		padded := fmt.Sprintf("%-*s", wordmarkWidth, row)
		rows[i] = c.paint(padded, lipgloss.NewStyle().Foreground(wordmarkGradient[i%len(wordmarkGradient)]))
	}

	line := c.accent(brandGlyph+" ", false) + c.paint(tagline, lipgloss.NewStyle().Bold(true))

	return lipgloss.JoinVertical(lipgloss.Center, strings.Join(rows, "\n"), line)
}

// accent paints text in the accent color, bold when asked.
func (c Capability) accent(text string, bold bool) string {
	return c.paint(text, lipgloss.NewStyle().Bold(bold).Foreground(ColorAccent))
}

// muted paints secondary text.
func (c Capability) muted(text string) string {
	return c.paint(text, lipgloss.NewStyle().Foreground(ColorMuted))
}

// paint applies style only when color is wanted.
func (c Capability) paint(text string, style lipgloss.Style) string {
	if !c.color {
		return text
	}
	return style.Render(text)
}

// center places a block in the middle of width columns, as a unit, so a
// multi-row block keeps its own alignment.
func (c Capability) center(block string, width int) string {
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, block)
}
