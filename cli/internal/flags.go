package internal

import "github.com/spf13/cobra"

// FlagBool reads a bool flag that is guaranteed to be registered on cmd.
func FlagBool(cmd *cobra.Command, name string) bool {
	v, _ := cmd.Flags().GetBool(name)
	return v
}

// FlagString reads a string flag that is guaranteed to be registered on cmd.
func FlagString(cmd *cobra.Command, name string) string {
	v, _ := cmd.Flags().GetString(name)
	return v
}
