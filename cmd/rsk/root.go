// Command rsk manages local and official AI skills for Claude Code and OpenCode.
package main

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

Machine setup:
  rsk init                       Set up rsk for this machine

Project lifecycle:
  rsk new                        Initialize an rsk project in this directory
  rsk destroy                    Remove .rsk/ and clean up tool configs

Install / uninstall / update (use --global for system-wide):
  rsk install                    Install everything tracked in rsk.mod
  rsk install <name>             Install bundles or skills by name
  rsk uninstall <name>           Remove installed bundles or skills
  rsk update [name]              Pull latest and re-link

Project pinning (auto-import into each tool's project config):
  rsk pin <name>                 Pin an installed skill
  rsk unpin <name>               Unpin a skill

Tool configuration:
  rsk claude tools               Manage Claude Code tool permissions

Read-only views:
  rsk list                       Show installed skills (project / --global)
  rsk catalog                    Browse available skills and bundles
  rsk status                     Combined view across project + global

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
