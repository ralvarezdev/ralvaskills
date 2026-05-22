package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall [bundle|skill] [flags]",
	Short: "Remove installed skills.",
	Long: `Remove symlinks for a bundle or individual skill.

Examples:
  rsk uninstall go-grpc
  rsk uninstall --skill go-architect
  rsk uninstall global --global`,
	RunE: runUninstall,
}

var (
	uninstallGlobal   bool
	uninstallFor      string
	uninstallPersonal bool
)

func init() {
	rootCmd.AddCommand(uninstallCmd)
	f := uninstallCmd.Flags()
	f.BoolVar(&uninstallGlobal, "global", false, "Target global skills dir(s)")
	f.StringVar(&uninstallFor, "for", "", "Scope --global to a single tool (claude-code|opencode)")
	f.BoolVar(&uninstallPersonal, "personal", false, "Allow uninstalling personal/ skills")
}

func runUninstall(_ *cobra.Command, _ []string) error {
	return fmt.Errorf("not yet implemented — coming soon")
}
