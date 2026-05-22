package cmd

import "github.com/spf13/cobra"

// flagBool reads a bool flag that is guaranteed to be registered on cmd.
func flagBool(cmd *cobra.Command, name string) bool {
	v, _ := cmd.Flags().GetBool(name)
	return v
}

// flagString reads a string flag that is guaranteed to be registered on cmd.
func flagString(cmd *cobra.Command, name string) string {
	v, _ := cmd.Flags().GetString(name)
	return v
}
