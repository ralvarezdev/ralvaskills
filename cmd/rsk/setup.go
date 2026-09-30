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

	// Cobra's built-in help/completion commands would otherwise land in an
	// untitled block; devtrack and finance file theirs under a group too.
	rootCmd.SetHelpCommandGroupID(groupOther)
	rootCmd.SetCompletionCommandGroupID(groupOther)

	// catalog command
	rootCmd.AddCommand(catalogCmd)
	termkit.MarkTableView(catalogCmd)
	f := catalogCmd.Flags()
	f.Bool(cmdx.FlagStack, false, "Fetch and display dependency metadata alongside skills")
	f.String(cmdx.FlagBundle, "", "List bundles; with a name (--bundle NAME), list that bundle's skills")
	f.Lookup(cmdx.FlagBundle).NoOptDefVal = bundleListSentinel
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
	f.StringP(cmdx.FlagOutput, "o", string(outputText), outputUsage())
	registerOutputCompletion(catalogCmd)

	// tools command (canonical) and the hidden legacy `claude tools` path,
	// both built by the same factory.
	rootCmd.AddCommand(canonicalTools.root)
	termkit.MarkTableView(canonicalTools.list)

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
	f.BoolP(cmdx.FlagYes, "y", false, "Skip the confirmation prompt")
	registerForCompletion(installCmd, false)

	// list command
	rootCmd.AddCommand(listCmd)
	termkit.MarkTableView(listCmd)
	f = listCmd.Flags()
	f.Bool(cmdx.FlagGlobal, false, "List global skills")
	f.String(cmdx.FlagFor, "", "Scope --global to a single tool (claude-code|opencode)")
	cmdx.RegisterInclude(listCmd, "Include extra skills in the output", cmdx.IncludePersonal)
	f.StringP(cmdx.FlagOutput, "o", string(outputText), outputUsage())
	registerOutputCompletion(listCmd)
	registerForCompletion(listCmd, false)

	// new command
	rootCmd.AddCommand(newCmd)
	f = newCmd.Flags()
	f.String(cmdx.FlagFor, "", "Tools to configure: claude-code|opencode|all")
	registerForCompletion(newCmd, true)

	// pin and unpin commands
	rootCmd.AddCommand(pinCmd)
	pinCmd.Flags().Bool(cmdx.FlagRemove, false, "Unpin the skill instead of pinning it")

	// status command
	rootCmd.AddCommand(statusCmd)
	termkit.MarkTableView(statusCmd)
	f = statusCmd.Flags()
	f.Bool(cmdx.FlagGlobal, false, "Show global skills only")
	f.String(cmdx.FlagFor, "", "Scope --global to a single tool (claude-code|opencode)")
	f.Bool("project", false, "Show project skills only")
	f.Bool(cmdx.FlagStack, false, "Fetch latest versions and show STACK.md drift (network, opt-in)")
	f.Bool(cmdx.FlagRefresh, false, "With --stack: bypass the 24h cache and force a re-fetch")
	cmdx.RegisterInclude(statusCmd, "Include extra skills in the output", cmdx.IncludePersonal)
	f.StringP(cmdx.FlagOutput, "o", string(outputText), outputUsage())
	registerOutputCompletion(statusCmd)
	registerForCompletion(statusCmd, false)

	// uninstall command
	rootCmd.AddCommand(uninstallCmd)
	f = uninstallCmd.Flags()
	f.Bool(cmdx.FlagGlobal, false, "Uninstall from the configured global skills dir(s)")
	f.String(cmdx.FlagFor, "", "With --global, scope to a single tool (claude-code|opencode)")
	cmdx.RegisterInclude(uninstallCmd, "Allow uninstalling extra skills", cmdx.IncludePersonal)
	f.Bool(cmdx.FlagDryRun, false, "Show what would be uninstalled without doing it")
	f.BoolP(cmdx.FlagYes, "y", false, "Skip the confirmation prompt")
	registerForCompletion(uninstallCmd, false)

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
	f.BoolP(cmdx.FlagYes, "y", false, "Skip the confirmation prompt")
	registerForCompletion(updateCmd, false)

	// row actions: install/uninstall directly from a catalog/list table row
	// in the TUI, instead of leaving the view to type the name.
	ui.MarkRowAction(catalogCmd, "i", "install", installCmd)
	ui.MarkRowAction(listCmd, "u", "uninstall", uninstallCmd)
	ui.MarkRowAction(listCmd, "p", "pin", pinCmd)
	ui.MarkRowActionArgs(listCmd, "n", "unpin", pinCmd, func(id string, _ termkit.Data) []string {
		return []string{id, "--" + cmdx.FlagRemove}
	})

	// row actions for the Claude tools view: allow, deny, or remove the tool
	// under the cursor without leaving the table to type its rule.
	for _, t := range []toolsCmds{canonicalTools} {
		registerOutputCompletion(t.list)
		ui.MarkRowAction(t.list, "a", "allow", t.allow)
		ui.MarkRowAction(t.list, "d", "deny", t.deny)
		ui.MarkRowAction(t.list, "x", "remove", t.remove)
	}

	// status renders one table per scanned directory, so its uninstall action
	// resolves --global/--for from the table the row came from.
	ui.MarkRowActionScoped(statusCmd, "u", "uninstall", uninstallCmd, statusRowActionScope)

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
