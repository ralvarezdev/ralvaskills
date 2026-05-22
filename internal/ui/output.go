// Package ui provides styled terminal output for the rsk CLI.
package ui

import (
	"fmt"
	"io"

	"github.com/ralvarezdev/ralvaskills/internal/skill"
)

// Brand prints the rsk name + version header line.
func Brand(w io.Writer, version string) {
	fmt.Fprintf(w, "\n"+Padding+"%s"+Padding+"%s\n\n",
		BrandStyle.Render("rsk"),
		VersionStyle.Render("v"+version),
	)
}

// Header prints a bold section title with a divider.
func Header(w io.Writer, msg string) {
	fmt.Fprintln(w, TitleStyle.Render(msg))
	fmt.Fprintln(w, divider())
}

// SectionHeader prints a two-part header: bold title + muted subtitle + divider.
func SectionHeader(w io.Writer, title, subtitle string) {
	line := BoldStyle.Render(title)
	if subtitle != "" {
		line += Padding + MutedStyle.Render(subtitle)
	}
	fmt.Fprintln(w, line)
	fmt.Fprintln(w, divider())
}

// SourceLabel returns a styled source badge for table rows.
func SourceLabel(src skill.Source) string {
	switch src {
	case skill.SourceLocal:
		return LocalStyle.Render(src.CleanLabel())
	case skill.SourceOfficial:
		return OfficialStyle.Render(src.CleanLabel())
	case skill.SourceRegistry:
		return MutedStyle.Render(src.CleanLabel() + "  ")
	default:
		return MutedStyle.Render(src.CleanLabel())
	}
}

// BundleTag returns a styled bundle name for status/list rows.
func BundleTag(name string) string {
	return BundleStyle.Render(name)
}

// Success prints a green ✓ line.
func Success(w io.Writer, msg string) {
	fmt.Fprintln(w, Padding+SuccessMark+Padding+msg)
}

// Warn prints an amber ⚠ line.
func Warn(w io.Writer, msg string) {
	fmt.Fprintln(w, Padding+WarnMark+Padding+msg)
}

// Fail prints a red ✗ line to w.
func Fail(w io.Writer, msg string) {
	fmt.Fprintln(w, Padding+ErrorMark+Padding+msg)
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
	fmt.Fprintln(w, Padding+MutedStyle.Render(msg))
}

// Dim prints a muted line.
func Dim(w io.Writer, msg string) {
	fmt.Fprintln(w, MutedStyle.Render(msg))
}
