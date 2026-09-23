// Command rsk manages local and official AI skills for Claude Code and OpenCode.
package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/ralvaskills/internal/ui"
)

const exitAborted = 130 // SIGINT + 128 per POSIX convention

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
	RunE: func(cmd *cobra.Command, _ []string) error {
		if !ui.IsTTY() {
			return printHome(cmd)
		}
		return ui.RunPicker(cmd.Commands(), cmd.OutOrStdout())
	},
}

// printHome prints the plain banner + cobra help shown for a bare `rsk`
// invocation outside a TTY (pipes, CI, scripts).
func printHome(cmd *cobra.Command) error {
	if _, err := fmt.Fprintln(cmd.OutOrStdout(), ui.Banner()); err != nil {
		return fmt.Errorf("write banner: %w", err)
	}
	return cmd.Help()
}

//nolint:gochecknoinits // init() used for root command setup
func init() {
	// go install and plain `go build` skip -ldflags, so version/commit/buildDate
	// stay at their zero-value defaults; fall back to the module/VCS info Go
	// embeds in the binary so `--version` isn't stuck reporting "dev/none/unknown".
	if version == "dev" {
		version, commit, buildDate = buildInfoFallback(version, commit, buildDate)
	}

	rootCmd.Version = fmt.Sprintf(
		"%s (rev %s, built %s, %s)",
		version, commit, buildDate, runtime.Version(),
	)
	rootCmd.SetVersionTemplate("rsk {{.Version}}\n")
	setupCommands()
}

// buildInfoFallback fills in version/commit/buildDate from runtime/debug's
// build info when the ldflags-injected defaults were never overridden, i.e.
// the binary wasn't produced by the release pipeline (e.g. `go install ...@latest`).
func buildInfoFallback(version, commit, buildDate string) (string, string, string) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return version, commit, buildDate
	}

	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		version = info.Main.Version
	}

	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			commit = setting.Value
		case "vcs.time":
			buildDate = setting.Value
		}
	}

	return version, commit, buildDate
}

// Execute runs the root command and exits on error.
// A user cancellation (Ctrl+C during a prompt) exits 130 silently per SIGINT convention.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if errors.Is(err, ui.ErrAborted) {
			os.Exit(exitAborted)
		}
		fmt.Fprintln(os.Stderr, "rsk: "+err.Error())
		os.Exit(1)
	}
}
