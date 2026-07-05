// Package ui provides styled terminal output for the rsk CLI.
package ui

import (
	"fmt"
	"io"

	"github.com/ralvarezdev/ralvaskills/internal/skill"
)

// Brand prints the rsk name and version header.
func Brand(w io.Writer, version string) {
	_, _ = fmt.Fprintf(w, "\n"+Padding+"%s"+Padding+"%s\n\n",
		BrandStyle.Render("rsk"),
		VersionStyle.Render("v"+version),
	)
}

// Header prints a bold section title followed by a divider line.
func Header(w io.Writer, msg string) {
	_, _ = fmt.Fprintln(w, TitleStyle.Render(msg))
	_, _ = fmt.Fprintln(w, divider())
}

// SectionHeader prints a bold title, an optional muted subtitle, and a divider.
func SectionHeader(w io.Writer, title, subtitle string) {
	line := BoldStyle.Render(title)
	if subtitle != "" {
		line += Padding + MutedStyle.Render(subtitle)
	}
	fmt.Fprintln(w, line)
	fmt.Fprintln(w, divider())
}

// SourceLabel returns a styled source badge for table rows. All labels are
// padded to the same visual width (5 chars) so columns align.
func SourceLabel(src skill.Source) string {
	switch src {
	case skill.SourceLocal:
		return LocalStyle.Render(src.Label())
	case skill.SourceOfficial:
		return OfficialStyle.Render(src.Label())
	default:
		return MutedStyle.Render(src.Label())
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

// Fail prints a red ✗ line.
func Fail(w io.Writer, msg string) {
	fmt.Fprintln(w, Padding+ErrorMark+Padding+msg)
}

// Failf prints a formatted red ✗ line.
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

// Dim prints a muted line with no indentation.
func Dim(w io.Writer, msg string) {
	fmt.Fprintln(w, MutedStyle.Render(msg))
}
