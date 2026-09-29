package main

import (
	"github.com/spf13/cobra"

	"github.com/ralvarezdev/termkit"

	"github.com/ralvarezdev/ralvaskills/v2/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/v2/internal/skill"
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
	f.Bool(cmdx.FlagBundles, false, "Alias for --bundle with no name")
	f.String(cmdx.FlagBundle, "", "List bundles; with a name (--bundle NAME), list that bundle's skills")
	f.Lookup(cmdx.FlagBundle).NoOptDefVal = bundleListSentinel
	f.Lookup(cmdx.FlagBundles).Hidden = true
	cmdx.RegisterInclude(catalogCmd, "Include extra skills in the output", cmdx.IncludePersonal)
	f.String(cmdx.FlagSource, "", "Filter by source: local|official")
	err := catalogCmd.RegisterFlagCompletionFunc(cmdx.FlagSource,
		func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
			return []string{skill.SourceLocal.String(), skill.SourceOfficial.String()},
				cobra.ShellCompDirectiveNoFileComp
		})
	if err != nil {
		panic("register --source completion: " + err.Error())
	}
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
	cmdx.RegisterInclude(installCmd, "Allow installing extra skills", cmdx.IncludePersonal)
	f.Bool(cmdx.FlagPin, false, "Also pin installed skills in the project (project scope only)")
	f.String(cmdx.FlagVersion, "", "Pin to a specific repo tag (local skills only)")
	f.Bool(cmdx.FlagDryRun, false, "Show what would be installed without doing it")

	// list command
	rootCmd.AddCommand(listCmd)
	ui.MarkTableView(listCmd)
	f = listCmd.Flags()
	f.Bool(cmdx.FlagGlobal, false, "List global skills")
	f.String(cmdx.FlagFor, "", "Scope --global to a single tool (claude-code|opencode)")
	cmdx.RegisterInclude(listCmd, "Include extra skills in the output", cmdx.IncludePersonal)
	f.StringP(cmdx.FlagOutput, "o", string(outputText), "Output format: text|json")

	// new command
	rootCmd.AddCommand(newCmd)
	f = newCmd.Flags()
	f.String(cmdx.FlagFor, "", "Tools to configure: claude-code|opencode|all")

	// pin and unpin commands
	rootCmd.AddCommand(pinCmd)
	rootCmd.AddCommand(unpinCmd)
	pinCmd.Flags().Bool(cmdx.FlagRemove, false, "Unpin the skill instead of pinning it")

	// status command
	rootCmd.AddCommand(statusCmd)
	ui.MarkTableView(statusCmd)
	f = statusCmd.Flags()
	f.Bool(cmdx.FlagGlobal, false, "Show global skills only")
	f.String(cmdx.FlagFor, "", "Scope --global to a single tool (claude-code|opencode)")
	f.Bool("project", false, "Show project skills only")
	f.Bool(cmdx.FlagStack, false, "Fetch latest versions and show STACK.md drift (network, opt-in)")
	f.Bool(cmdx.FlagRefresh, false, "With --stack: bypass the 24h cache and force a re-fetch")
	cmdx.RegisterInclude(statusCmd, "Include extra skills in the output", cmdx.IncludePersonal)
	f.StringP(cmdx.FlagOutput, "o", string(outputText), "Output format: text|json")

	// uninstall command
	rootCmd.AddCommand(uninstallCmd)
	f = uninstallCmd.Flags()
	f.Bool(cmdx.FlagGlobal, false, "Uninstall from the configured global skills dir(s)")
	f.String(cmdx.FlagFor, "", "With --global, scope to a single tool (claude-code|opencode)")
	cmdx.RegisterInclude(uninstallCmd, "Allow uninstalling extra skills", cmdx.IncludePersonal)
	f.Bool(cmdx.FlagDryRun, false, "Show what would be uninstalled without doing it")
	f.BoolP(cmdx.FlagYes, "y", false, "Skip the confirmation prompt")

	// update command
	rootCmd.AddCommand(updateCmd)
	f = updateCmd.Flags()
	f.Bool(cmdx.FlagGlobal, false, "Update global skills")
	f.String(cmdx.FlagFor, "", "Scope --global to a single tool (claude-code|opencode)")
	cmdx.RegisterInclude(
		updateCmd,
		"Also update extra skills; official also syncs the official skill cache",
		cmdx.IncludePersonal,
		cmdx.IncludeOfficial,
	)
	f.Bool(cmdx.FlagDryRun, false, "Show what would be updated without doing it")

	markDestructive()
	assignGroups()
}

// markDestructive flags the commands that change or delete files, so the
// interactive session asks before running them (unless --yes is given).
func markDestructive() {
	for _, cmd := range []*cobra.Command{destroyCmd, installCmd, uninstallCmd, updateCmd} {
		termkit.MarkDestructive(cmd)
	}
}

// assignGroups files each top-level command under its help section.
func assignGroups() {
	for id, cmds := range map[string][]*cobra.Command{
		groupMachine:   {initCmd},
		groupProject:   {newCmd, destroyCmd},
		groupInstall:   {installCmd, uninstallCmd, updateCmd},
		groupPinning:   {pinCmd},
		groupToolsConf: {canonicalTools.root},
		groupViews:     {listCmd, catalogCmd, statusCmd},
	} {
		for _, cmd := range cmds {
			cmd.GroupID = id
		}
	}
}
