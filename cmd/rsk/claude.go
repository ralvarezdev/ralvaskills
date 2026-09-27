package main

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/termkit"

	"github.com/ralvarezdev/ralvaskills/v2/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/v2/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/v2/internal/tool"
	"github.com/ralvarezdev/ralvaskills/v2/internal/ui"
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
		Long: `List current tool permissions for this project.

Examples:
  rsk claude tools list
  rsk claude tools list -o json`,
		RunE: runClaudeToolsList,
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

// claudeToolsPermissions is the JSON shape for `rsk claude tools list -o json`
// — the raw allow/deny rule lists as configured in .claude/settings.json.
type claudeToolsPermissions struct {
	Allow []string `json:"allow"`
	Deny  []string `json:"deny"`
}

func runClaudeToolsList(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	output := outputFormat(cmdx.String(cmd, cmdx.FlagOutput))
	if !output.valid() {
		return fmt.Errorf("--output must be '%s' or '%s'", outputText, outputJSON)
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

	if output == outputJSON {
		return writeJSON(out, claudeToolsPermissions{Allow: allow, Deny: deny})
	}

	rows := make([]claudeToolRow, 0, len(availableClaudeTools))
	for _, t := range availableClaudeTools {
		switch {
		case slices.Contains(allow, t):
			rows = append(rows, claudeToolRow{name: t, status: claudeToolAllowed})
		case slices.Contains(deny, t):
			rows = append(rows, claudeToolRow{name: t, status: claudeToolDenied})
		default:
			rows = append(rows, claudeToolRow{name: t, status: claudeToolUnconfigured})
		}
	}

	fmt.Fprintln(out)
	ui.Header(out, "Claude Code tools available:")
	termkit.WriteTableStyled(
		out,
		[]string{headerName, "Status"},
		termkit.Rows(rows, claudeToolsTable),
		false,
		nil,
		true,
	)
	fmt.Fprintln(out)

	return nil
}

// claudeToolStatus is a Claude tool's permission state relative to the
// project's settings.
type claudeToolStatus int

const (
	claudeToolAllowed claudeToolStatus = iota
	claudeToolDenied
	claudeToolUnconfigured
)

// claudeToolRow is one row of the tools-available table.
type claudeToolRow struct {
	name   string
	status claudeToolStatus
}

// claudeToolsTable projects one claudeToolRow, marking each tool's
// permission state the same way install/uninstall mark their own rows.
var claudeToolsTable = termkit.Table[claudeToolRow]{
	Row: func(r claudeToolRow) []any {
		switch r.status {
		case claudeToolAllowed:
			return []any{r.name, ui.SuccessMark + " explicitly allowed"}
		case claudeToolDenied:
			return []any{r.name, ui.ErrorMark + " explicitly denied"}
		default:
			return []any{r.name, ui.MutedStyle.Render("• not configured (uses global settings)")}
		}
	},
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
