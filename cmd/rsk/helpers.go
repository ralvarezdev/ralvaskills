package main

import (
	"encoding/json"
	"errors"
	"io"

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

// nameFromArgsOrPrompt returns args[0] if provided, otherwise prompts interactively.
func nameFromArgsOrPrompt(cmd *cobra.Command, args []string, label string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
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
