package ui

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/termkit"
)

// MarkTableView, IsTableView, WithCapture, and CaptureFromContext are
// re-exported from termkit so callers keep their familiar names.

// MarkTableView flags cmd as safe to render inside the TUI: read-only and
// non-interactive, so the session can run it in-process and render its
// captured table natively instead of handing it the terminal.
func MarkTableView(cmd *cobra.Command) {
	termkit.MarkTableView(cmd)
}

// IsTableView reports whether cmd was marked with MarkTableView.
func IsTableView(cmd *cobra.Command) bool {
	return termkit.IsTableView(cmd)
}

// WithCapture returns a context carrying c, so a command's own print helpers
// can record their table into it instead of rendering to stdout. A nil c
// leaves the context unchanged.
func WithCapture(ctx context.Context, c *termkit.Capture) context.Context {
	return termkit.WithCapture(ctx, c)
}

// CaptureFromContext returns the capture stored by WithCapture, or nil when
// none is present (a normal CLI run).
func CaptureFromContext(ctx context.Context) *termkit.Capture {
	return termkit.CaptureFromContext(ctx)
}
