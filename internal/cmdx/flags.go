// Package cmdx provides small helpers for working with cobra commands.
package cmdx

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Bool reads a bool flag that is guaranteed to be registered on cmd.
// Panics if the flag was not registered — this is a programmer error.
func Bool(cmd *cobra.Command, name string) bool {
	v, err := cmd.Flags().GetBool(name)
	if err != nil {
		panic(fmt.Sprintf("flag %q not registered on %q: %v", name, cmd.Name(), err))
	}
	return v
}

// String reads a string flag that is guaranteed to be registered on cmd.
// Panics if the flag was not registered — this is a programmer error.
func String(cmd *cobra.Command, name string) string {
	v, err := cmd.Flags().GetString(name)
	if err != nil {
		panic(fmt.Sprintf("flag %q not registered on %q: %v", name, cmd.Name(), err))
	}
	return v
}
