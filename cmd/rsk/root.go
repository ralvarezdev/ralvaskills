// Command rsk manages local and official AI skills for Claude Code and OpenCode.
package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/ralvaskills/v2/internal/ui"
)

const exitAborted = 130 // SIGINT + 128 per POSIX convention

// Cobra group IDs for the root command's sections, in help/picker order.
const (
	groupMachine   = "machine"
	groupProject   = "project"
	groupInstall   = "install"
	groupPinning   = "pinning"
	groupToolsConf = "tools"
	groupViews     = "views"
)

// Build metadata — injected via -ldflags at release time.
var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

// rootGroups are the root command's help sections, in display order. The TUI
// picker orders its top-level rows by the same list.
var rootGroups = []*cobra.Group{
	{ID: groupMachine, Title: "Machine setup:"},
	{ID: groupProject, Title: "Project lifecycle:"},
	{ID: groupInstall, Title: "Install / uninstall / update (use --global for system-wide):"},
	{ID: groupPinning, Title: "Project pinning (auto-import into each tool's project config):"},
	{ID: groupToolsConf, Title: "Tool configuration:"},
	{ID: groupViews, Title: "Read-only views:"},
}

var rootCmd = &cobra.Command{
	Use:           "rsk",
	Short:         "Manage ralvaskills — install, update, and check AI skill bundles.",
	Long:          "rsk manages your local and official AI skills for Claude Code and OpenCode.",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, _ []string) error {
		notice := updateNotice()
		maybeSpawnUpdateCheck()

		if !ui.IsTTY() {
			return printHome(cmd, notice)
		}
		return ui.RunSession(cmd.Context(), ui.SessionOptions{
			Root:     cmd.Root(),
			Runnable: sessionRunnable,
			Notice:   notice,
		})
	},
}

// sessionRunnable offers a command in the interactive picker unless it is
// cobra's own help/completion machinery or `init`, whose setup wizard prompts
// on the terminal, which the session owns while it runs.
func sessionRunnable(cmd *cobra.Command) bool {
	top := cmd
	for top.HasParent() && top.Parent().HasParent() {
		top = top.Parent()
	}
	switch top.Name() {
	case "help", "completion", initCmd.Name():
		return false
	}
	return true
}

// printHome prints the plain banner + cobra help shown for a bare `rsk`
// invocation outside a TTY (pipes, CI, scripts). notice, when non-empty, is
// an updates-available flag printed above the banner.
func printHome(cmd *cobra.Command, notice string) error {
	if notice != "" {
		ui.Warn(cmd.OutOrStdout(), notice)
	}
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
func buildInfoFallback(curVersion, curCommit, curBuildDate string) (version, commit, buildDate string) {
	version, commit, buildDate = curVersion, curCommit, curBuildDate

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return version, commit, buildDate
	}

	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		// Module versions carry a leading "v" (e.g. "v2.0.3"); strip it to
		// match the unprefixed version goreleaser's ldflags inject, so
		// `--version` reads the same regardless of install method.
		version = strings.TrimPrefix(info.Main.Version, "v")
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
