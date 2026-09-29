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
