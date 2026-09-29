package main

import (
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/ralvaskills/v2/internal/cmdx"
)

// outputFormat is a validated --output value: text or json.
type outputFormat string

const (
	outputText outputFormat = "text"
	outputJSON outputFormat = "json"
)

// outputFormats is the accepted --output set, in help and completion order.
var outputFormats = []outputFormat{outputText, outputJSON}

// String returns the raw --output value.
func (o outputFormat) String() string {
	return string(o)
}

// valid reports whether o is one of the accepted output formats.
func (o outputFormat) valid() bool {
	return slices.Contains(outputFormats, o)
}

// outputUsage returns the --output usage string, derived from outputFormats so
// help, completion, and validation stay in step.
func outputUsage() string {
	names := make([]string, len(outputFormats))
	for i, f := range outputFormats {
		names[i] = f.String()
	}
	return "Output format: " + strings.Join(names, "|")
}

// parseOutputFormat validates a raw --output value, reporting an unknown value
// against the flag so the interactive session can show it on the field.
func parseOutputFormat(raw string) (outputFormat, error) {
	o := outputFormat(raw)
	if !o.valid() {
		return "", fieldError(cmdx.FlagOutput, "--output must be '%s' or '%s'", outputText, outputJSON)
	}
	return o, nil
}

// registerOutputCompletion wires --output's shell completion from outputFormats.
func registerOutputCompletion(cmd *cobra.Command) {
	values := make([]string, len(outputFormats))
	for i, f := range outputFormats {
		values[i] = f.String()
	}
	err := cmd.RegisterFlagCompletionFunc(cmdx.FlagOutput,
		func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
			return values, cobra.ShellCompDirectiveNoFileComp
		})
	if err != nil {
		panic("register --output completion on " + cmd.Name() + ": " + err.Error())
	}
}
