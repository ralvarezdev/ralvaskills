package main

import (
	"fmt"

	"github.com/ralvarezdev/ralvaskills/internal/ui"
	"github.com/spf13/cobra"
)

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
		return "", fmt.Errorf("specify a skill name")
	}
	return name, nil
}
