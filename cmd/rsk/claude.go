package main

import (
	"fmt"
	"slices"

	"github.com/ralvarezdev/ralvaskills/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/internal/tool"
	"github.com/ralvarezdev/ralvaskills/internal/ui"
	"github.com/spf13/cobra"
)

// availableClaudeTools lists all Claude Code tools that can be managed.
var availableClaudeTools = []string{
	"Agent",
	"AskUserQuestion",
	"Bash",
	"CronCreate",
	"CronDelete",
	"CronList",
	"Edit",
	"EnterPlanMode",
	"EnterWorktree",
	"ExitPlanMode",
	"ExitWorktree",
	"Glob",
	"Grep",
	"ListMcpResourcesTool",
	"LSP",
	"Monitor",
	"NotebookEdit",
	"PowerShell",
	"PushNotification",
	"Read",
	"ReadMcpResourceTool",
	"RemoteTrigger",
	"ScheduleWakeup",
	"SendMessage",
	"ShareOnboardingGuide",
	"Skill",
	"TaskCreate",
	"TaskGet",
	"TaskList",
	"TaskOutput",
	"TaskStop",
	"TaskUpdate",
	"TeamCreate",
	"TeamDelete",
	"TodoWrite",
	"ToolSearch",
	"WaitForMcpServers",
	"WebFetch",
	"WebSearch",
	"Write",
}

var (
	claudeCmd = &cobra.Command{
		Use:   "claude",
		Short: "Manage Claude Code configuration for this project.",
		Long: `Manage Claude Code configuration, including tool permissions
and other Claude-specific settings in .claude/settings.json.`,
	}

	claudeToolsCmd = &cobra.Command{
		Use:   "tools",
		Short: "Manage Claude Code tool permissions.",
		Long: `Manage which Claude Code tools (Bash, Read, Write, etc.) are
allowed or denied in this project. Changes are written to
.claude/settings.json and override global tool permissions.

Examples:
  rsk claude tools list
  rsk claude tools allow Bash(npm run *)
  rsk claude tools deny Write(**)
  rsk claude tools remove Bash(npm run *)`,
	}

	claudeToolsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List current tool permissions for this project.",
		RunE:  runClaudeToolsList,
	}

	claudeToolsAllowCmd = &cobra.Command{
		Use:   "allow [rule]",
		Short: "Allow a tool in this project.",
		Long: `Add a tool rule to the permissions.allow list in .claude/settings.json.
The rule format is Tool(specifier), for example:
  Bash(npm run *)
  Read(~/docs/**)
  WebFetch(domain:example.com)`,
		RunE: runClaudeToolsAllow,
	}

	claudeToolsDenyCmd = &cobra.Command{
		Use:   "deny [rule]",
		Short: "Deny a tool in this project.",
		Long: `Add a tool rule to the permissions.deny list in .claude/settings.json.
The rule format is Tool(specifier), for example:
  Write(**)
  Bash
  WebFetch`,
		RunE: runClaudeToolsDeny,
	}

	claudeToolsRemoveCmd = &cobra.Command{
		Use:   "remove [rule]",
		Short: "Remove a tool rule from permissions.",
		Long:  `Remove a tool rule from either the allow or deny list.`,
		RunE:  runClaudeToolsRemove,
	}
)

func init() {
	rootCmd.AddCommand(claudeCmd)
	claudeCmd.AddCommand(claudeToolsCmd)
	claudeToolsCmd.AddCommand(claudeToolsListCmd)
	claudeToolsCmd.AddCommand(claudeToolsAllowCmd)
	claudeToolsCmd.AddCommand(claudeToolsDenyCmd)
	claudeToolsCmd.AddCommand(claudeToolsRemoveCmd)
}

func runClaudeToolsList(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	cwd, err := manifest.ProjectFolderPath()
	if err != nil {
		return err
	}
	projectDir := cwd[:len(cwd)-len("/.rsk")]

	ct, _ := tool.Get(tool.ClaudeID)
	claudeTool := ct.(*tool.ClaudeTool)
	allow, deny, err := claudeTool.ReadPermissions(projectDir)
	if err != nil {
		return err
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Claude Code tools available:")
	fmt.Fprintln(out)

	// Show allowed tools first
	allowedTools := make([]string, 0)
	deniedTools := make([]string, 0)
	unconfiguredTools := make([]string, 0)

	for _, tool := range availableClaudeTools {
		if slices.Contains(allow, tool) {
			allowedTools = append(allowedTools, tool)
		} else if slices.Contains(deny, tool) {
			deniedTools = append(deniedTools, tool)
		} else {
			unconfiguredTools = append(unconfiguredTools, tool)
		}
	}

	if len(allowedTools) > 0 {
		fmt.Fprintln(out, "  Explicitly allowed:")
		for _, t := range allowedTools {
			fmt.Fprintf(out, "    ✓ %s\n", t)
		}
		fmt.Fprintln(out)
	}

	if len(deniedTools) > 0 {
		fmt.Fprintln(out, "  Explicitly denied:")
		for _, t := range deniedTools {
			fmt.Fprintf(out, "    ✗ %s\n", t)
		}
		fmt.Fprintln(out)
	}

	if len(unconfiguredTools) > 0 {
		fmt.Fprintln(out, "  Not configured (uses global settings):")
		for _, t := range unconfiguredTools {
			fmt.Fprintf(out, "    • %s\n", t)
		}
	}

	fmt.Fprintln(out)

	return nil
}

func runClaudeToolsAllow(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	rule, err := nameFromArgsOrPrompt(cmd, args, "Tool rule (e.g. Bash(npm run *))")
	if err != nil {
		return err
	}

	cwd, err := manifest.ProjectFolderPath()
	if err != nil {
		return err
	}
	projectDir := cwd[:len(cwd)-len("/.rsk")]

	ct, _ := tool.Get(tool.ClaudeID)
	claudeTool := ct.(*tool.ClaudeTool)
	allow, deny, err := claudeTool.ReadPermissions(projectDir)
	if err != nil {
		return err
	}

	if slices.Contains(allow, rule) {
		ui.Info(out, fmt.Sprintf("%s is already allowed", rule))
		return nil
	}

	// Remove from deny if present
	deny = slices.DeleteFunc(deny, func(v string) bool { return v == rule })

	allow = append(allow, rule)

	if err = claudeTool.WritePermissions(projectDir, allow, deny); err != nil {
		return err
	}

	fmt.Fprintln(out)
	ui.Success(out, fmt.Sprintf("allowed %s", rule))
	fmt.Fprintln(out)
	return nil
}

func runClaudeToolsDeny(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	rule, err := nameFromArgsOrPrompt(cmd, args, "Tool rule (e.g. Write(**) or Bash)")
	if err != nil {
		return err
	}

	cwd, err := manifest.ProjectFolderPath()
	if err != nil {
		return err
	}
	projectDir := cwd[:len(cwd)-len("/.rsk")]

	ct, _ := tool.Get(tool.ClaudeID)
	claudeTool := ct.(*tool.ClaudeTool)
	allow, deny, err := claudeTool.ReadPermissions(projectDir)
	if err != nil {
		return err
	}

	if slices.Contains(deny, rule) {
		ui.Info(out, fmt.Sprintf("%s is already denied", rule))
		return nil
	}

	// Remove from allow if present
	allow = slices.DeleteFunc(allow, func(v string) bool { return v == rule })

	deny = append(deny, rule)

	if err = claudeTool.WritePermissions(projectDir, allow, deny); err != nil {
		return err
	}

	fmt.Fprintln(out)
	ui.Success(out, fmt.Sprintf("denied %s", rule))
	fmt.Fprintln(out)
	return nil
}

func runClaudeToolsRemove(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	rule, err := nameFromArgsOrPrompt(cmd, args, "Tool rule to remove")
	if err != nil {
		return err
	}

	cwd, err := manifest.ProjectFolderPath()
	if err != nil {
		return err
	}
	projectDir := cwd[:len(cwd)-len("/.rsk")]

	ct, _ := tool.Get(tool.ClaudeID)
	claudeTool := ct.(*tool.ClaudeTool)
	allow, deny, err := claudeTool.ReadPermissions(projectDir)
	if err != nil {
		return err
	}

	initialAllow, initialDeny := len(allow), len(deny)

	allow = slices.DeleteFunc(allow, func(v string) bool { return v == rule })
	deny = slices.DeleteFunc(deny, func(v string) bool { return v == rule })

	if len(allow) == initialAllow && len(deny) == initialDeny {
		ui.Info(out, fmt.Sprintf("%s not found in permissions", rule))
		return nil
	}

	if err = claudeTool.WritePermissions(projectDir, allow, deny); err != nil {
		return err
	}

	fmt.Fprintln(out)
	ui.Success(out, fmt.Sprintf("removed %s", rule))
	fmt.Fprintln(out)
	return nil
}
