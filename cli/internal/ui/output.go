// Package ui provides styled terminal output for the rsk CLI.
package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ralvarezdev/ralvaskills/cli/internal/skill"
)

// ── palette ──────────────────────────────────────────────────────────────────

var (
	clrBrand    = lipgloss.Color("#2DD4BF") // teal  — rsk brand
	clrLocal    = lipgloss.Color("#A78BFA") // violet — ralva skills
	clrOfficial = lipgloss.Color("#60A5FA") // blue   — anthr skills
	clrSuccess  = lipgloss.Color("#4ADE80") // green
	clrWarn     = lipgloss.Color("#FBBF24") // amber
	clrError    = lipgloss.Color("#F87171") // red
	clrMuted    = lipgloss.Color("#6B7280") // gray
	clrText     = lipgloss.Color("#F9FAFB") // near-white
	clrBundle   = lipgloss.Color("#34D399") // emerald
)

// ── styles ───────────────────────────────────────────────────────────────────

var (
	brandStyle      = lipgloss.NewStyle().Foreground(clrBrand).Bold(true)
	versionStyle    = lipgloss.NewStyle().Foreground(clrMuted)
	localStyle      = lipgloss.NewStyle().Foreground(clrLocal).Bold(true)
	officialStyle   = lipgloss.NewStyle().Foreground(clrOfficial).Bold(true)
	successStyle    = lipgloss.NewStyle().Foreground(clrSuccess).Bold(true)
	warnStyle       = lipgloss.NewStyle().Foreground(clrWarn).Bold(true)
	errorStyle      = lipgloss.NewStyle().Foreground(clrError).Bold(true)
	mutedStyle      = lipgloss.NewStyle().Foreground(clrMuted)
	boldStyle       = lipgloss.NewStyle().Foreground(clrText).Bold(true)
	bundleStyle     = lipgloss.NewStyle().Foreground(clrBundle)
	titleStyle      = lipgloss.NewStyle().Foreground(clrText).Bold(true)
	divStyle        = lipgloss.NewStyle().Foreground(clrMuted)
	promptStyle     = lipgloss.NewStyle().Foreground(clrBrand).Bold(true)
	arrowStyle      = lipgloss.NewStyle().Foreground(clrMuted)
	relinkStyle     = lipgloss.NewStyle().Foreground(clrWarn)
)

const dividerLen = 48

func divider() string {
	return divStyle.Render(strings.Repeat("─", dividerLen))
}

// ── brand ────────────────────────────────────────────────────────────────────

// Brand prints the rsk name + version header line.
func Brand(w io.Writer, version string) {
	fmt.Fprintf(w, "\n  %s  %s\n\n",
		brandStyle.Render("rsk"),
		versionStyle.Render("v"+version),
	)
}

// ── headers ───────────────────────────────────────────────────────────────────

// Header prints a bold section title with a divider.
func Header(w io.Writer, msg string) {
	fmt.Fprintln(w, titleStyle.Render(msg))
	fmt.Fprintln(w, divider())
}

// SectionHeader prints a two-part header: bold title + muted subtitle + divider.
func SectionHeader(w io.Writer, title, subtitle string) {
	line := boldStyle.Render(title)
	if subtitle != "" {
		line += "  " + mutedStyle.Render(subtitle)
	}
	fmt.Fprintln(w, line)
	fmt.Fprintln(w, divider())
}

// ── source labels ─────────────────────────────────────────────────────────────

// SourceLabel returns a styled source badge for table rows.
func SourceLabel(src skill.Source) string {
	switch src {
	case skill.SourceLocal:
		return localStyle.Render("ralva")
	case skill.SourceOfficial:
		return officialStyle.Render("anthr")
	case skill.SourceRegistry:
		return mutedStyle.Render("reg  ")
	default:
		return mutedStyle.Render("?????")
	}
}

// BundleTag returns a styled bundle name for status/list rows.
func BundleTag(name string) string {
	return bundleStyle.Render(name)
}

// ── marks & symbols ───────────────────────────────────────────────────────────

func SuccessMark() string { return successStyle.Render("✓") }
func WarnMark() string    { return warnStyle.Render("⚠") }
func ErrorMark() string   { return errorStyle.Render("✗") }
func Arrow() string       { return arrowStyle.Render("→") }
func ReLink() string      { return relinkStyle.Render("re-link") }

// ── print helpers ─────────────────────────────────────────────────────────────

// Success prints a green ✓ line.
func Success(w io.Writer, msg string) {
	fmt.Fprintln(w, "  "+SuccessMark()+"  "+msg)
}

// Warn prints an amber ⚠ line.
func Warn(w io.Writer, msg string) {
	fmt.Fprintln(w, "  "+WarnMark()+"  "+msg)
}

// Fail prints a red ✗ line to w.
func Fail(w io.Writer, msg string) {
	fmt.Fprintln(w, "  "+ErrorMark()+"  "+msg)
}

// Failf prints a formatted red ✗ line to w.
func Failf(w io.Writer, format string, args ...any) {
	Fail(w, fmt.Sprintf(format, args...))
}

// Info prints a plain line.
func Info(w io.Writer, msg string) {
	fmt.Fprintln(w, msg)
}

// Indent prints a muted indented line.
func Indent(w io.Writer, msg string) {
	fmt.Fprintln(w, "  "+mutedStyle.Render(msg))
}

// Dim prints a muted line.
func Dim(w io.Writer, msg string) {
	fmt.Fprintln(w, mutedStyle.Render(msg))
}

// ── confirm ───────────────────────────────────────────────────────────────────

// ConfirmYN prints a styled [y/N] prompt and returns true for "y" or "yes".
func ConfirmYN(w io.Writer, prompt string) bool {
	fmt.Fprintf(w, "\n  %s %s ", promptStyle.Render(prompt), mutedStyle.Render("[y/N]"))
	r := bufio.NewReader(os.Stdin)
	line, _ := r.ReadString('\n')
	fmt.Fprintln(w)
	line = strings.ToLower(strings.TrimSpace(strings.TrimRight(line, "\r\n")))
	return line == "y" || line == "yes"
}

// ── table helpers ─────────────────────────────────────────────────────────────

// PadRight pads s to visual width n, correctly handling ANSI escape codes.
func PadRight(s string, n int) string {
	w := lipgloss.Width(s)
	if w >= n {
		return s
	}
	return s + strings.Repeat(" ", n-w)
}

// MaxWidth returns the visual width of the longest string in items.
func MaxWidth(items []string) int {
	w := 0
	for _, s := range items {
		if vw := lipgloss.Width(s); vw > w {
			w = vw
		}
	}
	return w
}

// SkillName returns a styled skill name for table rows.
func SkillName(name string) string {
	return boldStyle.Render(name)
}

// SkillVersion returns a styled version string for table rows.
func SkillVersion(version string) string {
	if version == "" {
		return mutedStyle.Render("-")
	}
	return mutedStyle.Render(version)
}

// MutedPath returns a dimmed path string.
func MutedPath(p string) string {
	return mutedStyle.Render(p)
}
