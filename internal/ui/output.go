// Package ui provides styled terminal output for the rsk CLI.
package ui

import (
	"fmt"
	"io"

	"github.com/ralvarezdev/ralvaskills/v3/internal/skill"
)

// Padding is the standard left/right padding for UI output.
const Padding = "  "

// quiet suppresses informational lines (Info, Indent, Dim); warnings, results,
// and errors always print. It is set by the --quiet flag.
var quiet bool

// verbose enables Debug diagnostics. It is set by the --verbose flag.
var verbose bool

// SetQuiet silences informational output.
func SetQuiet(v bool) {
	quiet = v
}

// SetVerbose enables debug diagnostics.
func SetVerbose(v bool) {
	verbose = v
}

// Debug prints a muted diagnostic line to w when --verbose is set.
func Debug(w io.Writer, msg string) {
	if !verbose {
		return
	}
	fmt.Fprintln(w, MutedStyle.Render("debug: "+msg))
}

// Header prints a bold section title followed by a divider line. rsk's
// TitleStyle is termkit's own TitleStyle re-exported (see style.go), so this
// delegates straight to termkit.WriteHeader rather than duplicating its body.
func Header(w io.Writer, msg string) {
	Theme.WriteHeader(w, msg)
}

// SectionHeader prints a bold title, an optional muted subtitle, and a
// divider. rsk's BoldStyle (Bold(true), no color) and its "  " padding match
// termkit.WriteSectionHeader's rendering exactly, so this delegates straight
// to it rather than duplicating its body.
func SectionHeader(w io.Writer, title, subtitle string) {
	Theme.WriteSectionHeader(w, title, subtitle)
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

// Info prints a plain line, unless --quiet is set.
func Info(w io.Writer, msg string) {
	if quiet {
		return
	}
	fmt.Fprintln(w, msg)
}

// Indent prints a muted indented line, unless --quiet is set.
func Indent(w io.Writer, msg string) {
	if quiet {
		return
	}
	fmt.Fprintln(w, Padding+MutedStyle.Render(msg))
}

// Dim prints a muted line with no indentation, unless --quiet is set.
func Dim(w io.Writer, msg string) {
	if quiet {
		return
	}
	fmt.Fprintln(w, MutedStyle.Render(msg))
}
