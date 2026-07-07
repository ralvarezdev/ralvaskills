package main

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/ralvaskills/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/internal/tool"
	"github.com/ralvarezdev/ralvaskills/internal/ui"
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

//nolint:gochecknoinits // init() used for command registration
func init() {
	rootCmd.AddCommand(claudeCmd)
	claudeCmd.AddCommand(claudeToolsCmd)
	claudeToolsCmd.AddCommand(claudeToolsListCmd)
	claudeToolsCmd.AddCommand(claudeToolsAllowCmd)
	claudeToolsCmd.AddCommand(claudeToolsDenyCmd)
	claudeToolsCmd.AddCommand(claudeToolsRemoveCmd)
}

func claudeToolGet() (*tool.ClaudeTool, error) {
	t, ok := tool.Get(tool.ClaudeID)
	if !ok {
		return nil, fmt.Errorf("tool %s not registered", tool.ClaudeID)
	}
	ct, ok := t.(*tool.ClaudeTool)
	if !ok {
		return nil, fmt.Errorf("tool %s has unexpected type %T", tool.ClaudeID, t)
	}
	return ct, nil
}

func runClaudeToolsList(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	cwd, err := manifest.ProjectFolderPath()
	if err != nil {
		return err
	}
	projectDir := cwd[:len(cwd)-len("/.rsk")]

	claudeTool, err := claudeToolGet()
	if err != nil {
		return err
	}
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
		switch {
		case slices.Contains(allow, tool):
			allowedTools = append(allowedTools, tool)
		case slices.Contains(deny, tool):
			deniedTools = append(deniedTools, tool)
		default:
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

	claudeTool, err := claudeToolGet()
	if err != nil {
		return err
	}
	allow, deny, err := claudeTool.ReadPermissions(projectDir)
	if err != nil {
		return err
	}

	if slices.Contains(allow, rule) {
		ui.Info(out, rule+" is already allowed")
		return nil
	}

	// Remove from deny if present
	deny = slices.DeleteFunc(deny, func(v string) bool { return v == rule })

	allow = append(allow, rule)

	if err = claudeTool.WritePermissions(projectDir, allow, deny); err != nil {
		return err
	}

	fmt.Fprintln(out)
	ui.Success(out, "allowed "+rule)
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

	claudeTool, err := claudeToolGet()
	if err != nil {
		return err
	}
	allow, deny, err := claudeTool.ReadPermissions(projectDir)
	if err != nil {
		return err
	}

	if slices.Contains(deny, rule) {
		ui.Info(out, rule+" is already denied")
		return nil
	}

	// Remove from allow if present
	allow = slices.DeleteFunc(allow, func(v string) bool { return v == rule })

	deny = append(deny, rule)

	if err = claudeTool.WritePermissions(projectDir, allow, deny); err != nil {
		return err
	}

	fmt.Fprintln(out)
	ui.Success(out, "denied "+rule)
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

	claudeTool, err := claudeToolGet()
	if err != nil {
		return err
	}
	allow, deny, err := claudeTool.ReadPermissions(projectDir)
	if err != nil {
		return err
	}

	initialAllow, initialDeny := len(allow), len(deny)

	allow = slices.DeleteFunc(allow, func(v string) bool { return v == rule })
	deny = slices.DeleteFunc(deny, func(v string) bool { return v == rule })

	if len(allow) == initialAllow && len(deny) == initialDeny {
		ui.Info(out, rule+" not found in permissions")
		return nil
	}

	if err = claudeTool.WritePermissions(projectDir, allow, deny); err != nil {
		return err
	}

	fmt.Fprintln(out)
	ui.Success(out, "removed "+rule)
	fmt.Fprintln(out)
	return nil
}
