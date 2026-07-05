package main

import "github.com/ralvarezdev/ralvaskills/internal/cmdx"

// setupCommands registers all subcommands and their flags with the root command.
func setupCommands() {
	// catalog command
	rootCmd.AddCommand(catalogCmd)
	f := catalogCmd.Flags()
	f.Bool(cmdx.FlagStack, false, "Fetch and display dependency metadata alongside skills")
	f.Bool(cmdx.FlagBundles, false, "Show bundles instead of individual skills")
	f.Bool(cmdx.FlagPersonal, false, "Include personal/ skills in output")

	// claude command and subcommands
	rootCmd.AddCommand(claudeCmd)
	claudeCmd.AddCommand(claudeToolsCmd)
	claudeToolsCmd.AddCommand(claudeToolsListCmd)
	claudeToolsCmd.AddCommand(claudeToolsAllowCmd)
	claudeToolsCmd.AddCommand(claudeToolsDenyCmd)
	claudeToolsCmd.AddCommand(claudeToolsRemoveCmd)

	// destroy command
	rootCmd.AddCommand(destroyCmd)

	// init command
	rootCmd.AddCommand(initCmd)
	f = initCmd.Flags()
	f.Bool(cmdx.FlagForce, false, "Overwrite existing rsk.mod if present")

	// install command
	rootCmd.AddCommand(installCmd)
	f = installCmd.Flags()
	f.Bool(cmdx.FlagGlobal, false, "Install to the configured global skills dir(s)")
	f.String(cmdx.FlagFor, "", "With --global, scope to a single tool (claude-code|opencode)")
	f.Bool(cmdx.FlagPersonal, false, "Allow installing personal/ skills")
	f.Bool(cmdx.FlagPin, false, "Also pin installed skills in the project (project scope only)")
	f.String(cmdx.FlagVersion, "", "Pin to a specific repo tag (local skills only)")
	f.Bool(cmdx.FlagDryRun, false, "Show what would be installed without doing it")

	// list command
	rootCmd.AddCommand(listCmd)
	f = listCmd.Flags()
	f.Bool(cmdx.FlagGlobal, false, "List global skills")
	f.String(cmdx.FlagFor, "", "Scope --global to a single tool (claude-code|opencode)")
	f.Bool(cmdx.FlagPersonal, false, "Include personal/ skills in output")

	// new command
	rootCmd.AddCommand(newCmd)

	// pin and unpin commands
	rootCmd.AddCommand(pinCmd)
	rootCmd.AddCommand(unpinCmd)

	// status command
	rootCmd.AddCommand(statusCmd)
	f = statusCmd.Flags()
	f.Bool(cmdx.FlagGlobal, false, "Show global skills only")
	f.String(cmdx.FlagFor, "", "Scope --global to a single tool (claude-code|opencode)")
	f.Bool("project", false, "Show project skills only")
	f.Bool(cmdx.FlagStack, false, "Fetch latest versions and show STACK.md drift (network, opt-in)")
	f.Bool(cmdx.FlagRefresh, false, "With --stack: bypass the 24h cache and force a re-fetch")
	f.Bool(cmdx.FlagPersonal, false, "Include personal/ skills in output")

	// uninstall command
	rootCmd.AddCommand(uninstallCmd)
	f = uninstallCmd.Flags()
	f.Bool(cmdx.FlagGlobal, false, "Uninstall from the configured global skills dir(s)")
	f.String(cmdx.FlagFor, "", "With --global, scope to a single tool (claude-code|opencode)")
	f.Bool(cmdx.FlagDryRun, false, "Show what would be uninstalled without doing it")

	// update command
	rootCmd.AddCommand(updateCmd)
	f = updateCmd.Flags()
	f.Bool(cmdx.FlagGlobal, false, "Update global skills")
	f.String(cmdx.FlagFor, "", "Scope --global to a single tool (claude-code|opencode)")
	f.Bool(cmdx.FlagOfficial, false, "Sync the official skill cache")
	f.Bool(cmdx.FlagDryRun, false, "Show what would be updated without doing it")
}
