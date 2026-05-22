package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [flags]",
	Short: "Show installed skills and version drift.",
	Long: `List installed skills with their versions and update availability.

Without --stack, no network calls are made.
With --stack, fetches latest versions from proxy.golang.org and pypi.org
and highlights skills whose STACK.md may be outdated (results cached 24h).

Examples:
  rsk status
  rsk status --global
  rsk status --stack
  rsk status --stack --refresh`,
	RunE: runStatus,
}

var (
	statusGlobal   bool
	statusFor      string
	statusProject  bool
	statusStack    bool
	statusRefresh  bool
	statusPersonal bool
)

func init() {
	rootCmd.AddCommand(statusCmd)
	f := statusCmd.Flags()
	f.BoolVar(&statusGlobal, "global", false, "Show global skills only")
	f.StringVar(&statusFor, "for", "", "Scope --global to a single tool (claude-code|opencode)")
	f.BoolVar(&statusProject, "project", false, "Show project skills only")
	f.BoolVar(&statusStack, "stack", false, "Fetch latest versions and show STACK.md drift (network, opt-in)")
	f.BoolVar(&statusRefresh, "refresh", false, "With --stack: bypass the 24h cache and force a re-fetch")
	f.BoolVar(&statusPersonal, "personal", false, "Include personal/ skills in output")
}

func runStatus(_ *cobra.Command, _ []string) error {
	return fmt.Errorf("not yet implemented — coming soon")
}
