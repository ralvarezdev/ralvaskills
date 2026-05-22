package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [bundle|skill] [flags]",
	Short: "Pull the latest skills and re-symlink.",
	Long: `Run git pull on the ralvaskills repo and re-create symlinks.

With no arguments, updates everything. Use --official to also
re-fetch the anthropics/skills cache.

Examples:
  rsk update
  rsk update --official
  rsk update docs
  rsk update --skill grpc-architect`,
	RunE: runUpdate,
}

var (
	updateGlobal   bool
	updateFor      string
	updateOfficial bool
	updatePersonal bool
	updateDryRun   bool
)

func init() {
	rootCmd.AddCommand(updateCmd)
	f := updateCmd.Flags()
	f.BoolVar(&updateGlobal, "global", false, "Target global skills dir(s)")
	f.StringVar(&updateFor, "for", "", "Scope --global to a single tool (claude-code|opencode)")
	f.BoolVar(&updateOfficial, "official", false, "Also re-fetch the anthropics/skills cache")
	f.BoolVar(&updatePersonal, "personal", false, "Include personal/ skills in update")
	f.BoolVar(&updateDryRun, "dry-run", false, "Show what would change without applying it")
}

func runUpdate(_ *cobra.Command, _ []string) error {
	return fmt.Errorf("not yet implemented — coming soon")
}
