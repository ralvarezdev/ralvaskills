// Package ui provides terminal output helpers with color and TTY-aware formatting.
package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	ansiReset  = "\033[0m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
)

// ColorEnabled reports whether colored output is allowed for w.
// Respects NO_COLOR, FORCE_COLOR, and TTY detection.
func ColorEnabled(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func colorize(w io.Writer, code, msg string) string {
	if ColorEnabled(w) {
		return code + msg + ansiReset
	}
	return msg
}

// Success prints a green ✓ prefixed line to w.
func Success(w io.Writer, msg string) {
	fmt.Fprintln(w, colorize(w, ansiGreen, "✓")+" "+msg)
}

// Warn prints a yellow ⚠ prefixed line to w.
func Warn(w io.Writer, msg string) {
	fmt.Fprintln(w, colorize(w, ansiYellow, "⚠")+" "+msg)
}

// Fail prints a red ✗ prefixed line to stderr.
func Fail(msg string) {
	fmt.Fprintln(os.Stderr, colorize(os.Stderr, ansiRed, "✗")+" "+msg)
}

// Failf prints a formatted red ✗ prefixed line to stderr.
func Failf(format string, args ...any) {
	Fail(fmt.Sprintf(format, args...))
}

// Info prints a plain line to w.
func Info(w io.Writer, msg string) {
	fmt.Fprintln(w, msg)
}

// Indent prints msg with two-space indent to w.
func Indent(w io.Writer, msg string) {
	fmt.Fprintln(w, "  "+msg)
}

// Header prints a bold section header to w.
func Header(w io.Writer, msg string) {
	fmt.Fprintln(w, colorize(w, ansiBold, msg))
}

// Dim prints a faint line to w.
func Dim(w io.Writer, msg string) {
	fmt.Fprintln(w, colorize(w, ansiDim, msg))
}

// PadRight pads s with spaces on the right to width n.
func PadRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

// MaxWidth returns the length of the longest string in items.
func MaxWidth(items []string) int {
	width := 0
	for _, s := range items {
		if len(s) > width {
			width = len(s)
		}
	}
	return width
}
