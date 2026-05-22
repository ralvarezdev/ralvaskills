// Package cmd implements the rsk CLI commands.
package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

// Build metadata — injected via -ldflags at release time.
var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "rsk",
	Short: "Manage ralvaskills — install, update, and check AI skill bundles.",
	Long: `rsk manages your local and official AI skills for Claude Code and OpenCode.

  rsk init              Set up rsk for this machine
  rsk install <bundle>  Install a bundle or skill
  rsk update            Pull the latest skills
  rsk status            Show installed skills and version drift
  rsk list              Browse the skill catalog
  rsk uninstall         Remove installed skills

Use 'rsk <command> --help' for command-specific help.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command and exits on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "rsk: "+err.Error())
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = fmt.Sprintf(
		"%s (rev %s, built %s, %s)",
		version, commit, buildDate, runtime.Version(),
	)
	rootCmd.SetVersionTemplate("rsk {{.Version}}\n")
}
