package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/termkit"

	"github.com/ralvarezdev/ralvaskills/v2/internal/ui"
)

// writeOrCaptureJSON renders v as indented JSON: under a capture (the
// picker's in-process run), it's recorded as the result's raw payload and
// shown in a scrollable viewport instead of a table; otherwise it's printed
// to out as it would be for a normal terminal invocation.
func writeOrCaptureJSON(out io.Writer, capture *termkit.Capture, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if capture != nil {
		capture.SetRaw(string(data))
		return nil
	}
	_, err = out.Write(append(data, '\n'))
	return err
}

// Names of the session form fields (form.Field.Name) that validation errors
// are blamed on: a positional argument is "arg:<name from the Use string>", a
// flag is its own name.
const (
	fieldArgName = "arg:name"
	fieldArgRule = "arg:rule"
)

// fieldError marks err as caused by the form field named field, so the
// interactive session shows it on that field instead of dropping back to the
// picker. On the command line it is the same error.
func fieldError(field, message string, args ...any) error {
	return termkit.NewFieldError(field, fmt.Errorf(message, args...))
}

// confirmProceed asks prompt on the terminal and reports whether to go ahead.
// Inside the interactive session it always does: a destructive command was
// already confirmed there (termkit.MarkDestructive), and the session owns the
// terminal, so there is nothing to ask on.
func confirmProceed(cmd *cobra.Command, out io.Writer, prompt string) bool {
	return ui.InSession(cmd.Context()) || ui.ConfirmYN(out, prompt)
}

// shouldConfirm reports whether a destructive command must ask "Proceed?":
// always, except with --yes, or inside the TUI session, whose confirm screen
// already asked. Piped runs still ask (a plain-text prompt that aborts on
// EOF), as before.
func shouldConfirm(cmd *cobra.Command, yes bool) bool {
	return !yes && !ui.InSession(cmd.Context())
}

// confirmDestructive asks "Proceed?" when shouldConfirm says so and reports
// whether the command may go ahead; every other case proceeds silently.
func confirmDestructive(cmd *cobra.Command, out io.Writer, yes bool) bool {
	if !shouldConfirm(cmd, yes) {
		return true
	}
	return ui.ConfirmYN(out, "Proceed?")
}

// nameFromArgsOrPrompt returns args[0] if provided, otherwise prompts
// interactively. Inside the session, which cannot prompt, the missing value is
// reported against the form field named field.
func nameFromArgsOrPrompt(cmd *cobra.Command, args []string, field, label string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	if ui.InSession(cmd.Context()) {
		return "", fieldError(field, "%s is required", strings.ToLower(label))
	}
	name, err := ui.Ask(cmd.OutOrStdout(), label, "")
	if err != nil {
		return "", err
	}
	if name == "" {
		return "", errors.New("specify a skill name")
	}
	return name, nil
}
