package main

import (
	"github.com/spf13/cobra"

	"github.com/ralvarezdev/ralvaskills/v2/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/v2/internal/ui"
)

// setupCommands registers all subcommands and their flags with the root command.
func setupCommands() {
	rootCmd.AddGroup(rootGroups...)

	// catalog command
	rootCmd.AddCommand(catalogCmd)
	ui.MarkTableView(catalogCmd)
	f := catalogCmd.Flags()
	f.Bool(cmdx.FlagStack, false, "Fetch and display dependency metadata alongside skills")
	f.Bool(cmdx.FlagBundles, false, "Show bundles instead of individual skills")
	f.Bool(cmdx.FlagPersonal, false, "Include personal/ skills in output")
	f.String(cmdx.FlagBundle, "", "List the skills inside a single bundle")
	f.String(cmdx.FlagSource, "", "Filter by source: local|official")
	f.StringP(cmdx.FlagOutput, "o", string(outputText), "Output format: text|json")

	// tools command (canonical) and the hidden legacy `claude tools` path,
	// both built by the same factory.
	rootCmd.AddCommand(canonicalTools.root)
	ui.MarkTableView(canonicalTools.list)
	rootCmd.AddCommand(claudeCmd)
	claudeCmd.AddCommand(legacyTools.root)
	ui.MarkTableView(legacyTools.list)

	// destroy command
	rootCmd.AddCommand(destroyCmd)
	destroyCmd.Flags().BoolP(cmdx.FlagYes, "y", false, "Skip the confirmation prompt")

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
	ui.MarkTableView(listCmd)
	f = listCmd.Flags()
	f.Bool(cmdx.FlagGlobal, false, "List global skills")
	f.String(cmdx.FlagFor, "", "Scope --global to a single tool (claude-code|opencode)")
	f.Bool(cmdx.FlagPersonal, false, "Include personal/ skills in output")
	f.StringP(cmdx.FlagOutput, "o", string(outputText), "Output format: text|json")

	// new command
	rootCmd.AddCommand(newCmd)
	f = newCmd.Flags()
	f.String(cmdx.FlagFor, "", "Tools to configure: claude-code|opencode|all")

	// pin and unpin commands
	rootCmd.AddCommand(pinCmd)
	rootCmd.AddCommand(unpinCmd)

	// status command
	rootCmd.AddCommand(statusCmd)
	ui.MarkTableView(statusCmd)
	f = statusCmd.Flags()
	f.Bool(cmdx.FlagGlobal, false, "Show global skills only")
	f.String(cmdx.FlagFor, "", "Scope --global to a single tool (claude-code|opencode)")
	f.Bool("project", false, "Show project skills only")
	f.Bool(cmdx.FlagStack, false, "Fetch latest versions and show STACK.md drift (network, opt-in)")
	f.Bool(cmdx.FlagRefresh, false, "With --stack: bypass the 24h cache and force a re-fetch")
	f.Bool(cmdx.FlagPersonal, false, "Include personal/ skills in output")
	f.StringP(cmdx.FlagOutput, "o", string(outputText), "Output format: text|json")

	// uninstall command
	rootCmd.AddCommand(uninstallCmd)
	f = uninstallCmd.Flags()
	f.Bool(cmdx.FlagGlobal, false, "Uninstall from the configured global skills dir(s)")
	f.String(cmdx.FlagFor, "", "With --global, scope to a single tool (claude-code|opencode)")
	f.Bool(cmdx.FlagPersonal, false, "Allow uninstalling personal/ skills")
	f.Bool(cmdx.FlagDryRun, false, "Show what would be uninstalled without doing it")
	f.BoolP(cmdx.FlagYes, "y", false, "Skip the confirmation prompt")

	// update command
	rootCmd.AddCommand(updateCmd)
	f = updateCmd.Flags()
	f.Bool(cmdx.FlagGlobal, false, "Update global skills")
	f.String(cmdx.FlagFor, "", "Scope --global to a single tool (claude-code|opencode)")
	f.Bool(cmdx.FlagPersonal, false, "Include personal/ skills in the update")
	f.Bool(cmdx.FlagOfficial, false, "Sync the official skill cache")
	f.Bool(cmdx.FlagDryRun, false, "Show what would be updated without doing it")

	// row actions: install/uninstall directly from a catalog/list table row
	// in the TUI, instead of leaving the view to type the name.
	ui.MarkRowAction(catalogCmd, "i", "install", installCmd)
	ui.MarkRowAction(listCmd, "u", "uninstall", uninstallCmd)

	// row actions for the Claude tools view: allow, deny, or remove the tool
	// under the cursor without leaving the table to type its rule.
	for _, t := range []toolsCmds{canonicalTools, legacyTools} {
		ui.MarkRowAction(t.list, "a", "allow", t.allow)
		ui.MarkRowAction(t.list, "d", "deny", t.deny)
		ui.MarkRowAction(t.list, "x", "remove", t.remove)
	}

	// status renders one table per scanned directory, so its uninstall action
	// resolves --global/--for from the table the row came from.
	ui.MarkRowActionScoped(statusCmd, "u", "uninstall", uninstallCmd, statusRowActionScope)

	assignGroups()
}

// assignGroups files each top-level command under its help section.
func assignGroups() {
	for id, cmds := range map[string][]*cobra.Command{
		groupMachine:   {initCmd},
		groupProject:   {newCmd, destroyCmd},
		groupInstall:   {installCmd, uninstallCmd, updateCmd},
		groupPinning:   {pinCmd, unpinCmd},
		groupToolsConf: {canonicalTools.root},
		groupViews:     {listCmd, catalogCmd, statusCmd},
	} {
		for _, cmd := range cmds {
			cmd.GroupID = id
		}
	}
}
