package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/termkit"
	"github.com/ralvarezdev/termkit/cli"
	"github.com/ralvarezdev/termkit/output"

	"github.com/ralvarezdev/ralvaskills/v3/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/v3/internal/ui"
)

// printer owns the shared --output/-o flag (text|json): registration,
// validation and completion. Rendering goes through newFormatter.
var printer = cli.NewPrinter(cli.OutputConfig{
	Formats:       []output.OutputFormat{output.FormatText, output.FormatJSON},
	DefaultFormat: output.FormatText,
	Errorf: func(*cobra.Command, string, []output.OutputFormat) error {
		return fieldError(
			cmdx.FlagOutput, "--output must be '%s' or '%s'",
			output.FormatText, output.FormatJSON,
		)
	},
})

// addOutputFlag registers --output/-o on cmd.
func addOutputFlag(cmd *cobra.Command) {
	printer.AddFlag(cmd)
}

// checkOutput validates cmd's --output value, blaming the form field so the
// picker shows it. The printer also validates at PreRunE; this keeps a direct
// run* call (as the tests make) honest.
func checkOutput(cmd *cobra.Command) error {
	switch raw := cmdx.String(cmd, cmdx.FlagOutput); raw {
	case "", string(output.FormatText), string(output.FormatJSON):
		return nil
	default:
		return fieldError(
			cmdx.FlagOutput, "--output must be '%s' or '%s'",
			output.FormatText, output.FormatJSON,
		)
	}
}

// outputOptions is how every rsk table renders: borderless with a plain
// header; the cells carry their own styling (see ui.SkillName and friends).
var outputOptions = output.Options{Theme: ui.Theme, Layout: termkit.TableBorderless, NoColor: true}

// newFormatter returns cmd's formatter for format, wired to the session
// capture so a picker run records its table/JSON instead of writing behind
// the TUI.
func newFormatter(cmd *cobra.Command, format output.OutputFormat) *output.Formatter {
	f := output.NewWithOptions(cmd.OutOrStdout(), format, outputOptions)
	if capture := termkit.CaptureFromContext(cmd.Context()); capture != nil {
		f.SetCapture(capture)
	}

	return f
}

// renderList renders one list view through the shared formatter: the
// borderless terminal table, the picker's captured table (with row actions),
// or the JSON payload v. section is the bold title printed above the table in
// terminal mode; an empty section prints only the surrounding blank lines.
func renderList(
	cmd *cobra.Command, format output.OutputFormat, section string, v any,
	header []string, rows [][]any, ids []string,
) error {
	capture := termkit.CaptureFromContext(cmd.Context())
	f := newFormatter(cmd, format)
	terminal := capture == nil && format == output.FormatText
	out := cmd.OutOrStdout()

	if terminal {
		fmt.Fprintln(out)
		if section != "" {
			ui.Header(out, section)
		}
	}
	if err := output.RenderRowsWith(f, v, header, rows, ids, ui.RowActionsFor(cmd)); err != nil {
		return err
	}
	if terminal {
		fmt.Fprintln(out)
	}

	return nil
}

// renderTitledTable is renderList for a table the picker shows under a title
// (status renders one per scanned section). It always records the title on the
// captured table, and renders the borderless table on a terminal.
func renderTitledTable(
	cmd *cobra.Command, format output.OutputFormat, title string, v any,
	header []string, rows [][]any, ids []string,
) error {
	f := newFormatter(cmd, format)
	if capture := termkit.CaptureFromContext(cmd.Context()); capture != nil {
		f.SetCaptureTitle(title)
	}

	return output.RenderRowsWith(f, v, header, rows, ids, ui.RowActionsFor(cmd))
}

// emptyHint reports an empty result: recorded on the capture under the picker,
// otherwise written to out. warn styles it as a warning with blank lines.
func emptyHint(cmd *cobra.Command, warn bool, msg string) {
	if capture := termkit.CaptureFromContext(cmd.Context()); capture != nil {
		capture.AddMessage(msg)

		return
	}
	out := cmd.OutOrStdout()
	if warn {
		fmt.Fprintln(out)
		ui.Warn(out, msg)
		fmt.Fprintln(out)

		return
	}
	ui.Info(out, msg)
}
