package main

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/termkit"
	"github.com/ralvarezdev/termkit/output"

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

// toolsForHint is the --for flag's help text, shared by every tools subcommand.
const toolsForHint = "Tool to configure (only claude-code is supported for now)"

type (
	// toolsCmds is one registered copy of the tools command tree: the parent
	// and its four subcommands.
	toolsCmds struct {
		root, list, allow, deny, remove *cobra.Command
	}
)

// canonicalTools is `rsk tools`.
var canonicalTools = newToolsCmds("rsk tools")

// newToolsCmds builds a tools command tree whose help examples are written
// with prefix (e.g. "rsk tools"). Every subcommand carries its own --for flag
// so the TUI parameter form and row actions see it.
func newToolsCmds(prefix string) toolsCmds {
	root := &cobra.Command{
		Use:   "tools",
		Short: "Manage tool permissions (allow/deny/list/remove).",
		Long: fmt.Sprintf(`Manage which tools (Bash, Read, Write, etc.) are allowed or denied in
this project. For Claude Code, changes are written to .claude/settings.json
and override global tool permissions.

Use --for to pick the AI tool (default claude-code; other tools are not
supported yet).

Examples:
  %[1]s list
  %[1]s allow Bash(npm run *)
  %[1]s deny Write(**)
  %[1]s remove Bash(npm run *)`, prefix),
	}

	list := &cobra.Command{
		Use:   "list",
		Short: "List current tool permissions for this project.",
		Long: fmt.Sprintf(`List current tool permissions for this project.

Examples:
  %[1]s list
  %[1]s list -o json`, prefix),
		RunE: runClaudeToolsList,
	}
	allow := &cobra.Command{
		Use:   "allow [rule]",
		Short: "Allow a tool in this project.",
		Long: `Add a tool rule to the permissions.allow list in .claude/settings.json.
The rule format is Tool(specifier), for example:
  Bash(npm run *)
  Read(~/docs/**)
  WebFetch(domain:example.com)`,
		RunE: runClaudeToolsAllow,
	}
	deny := &cobra.Command{
		Use:   "deny [rule]",
		Short: "Deny a tool in this project.",
		Long: `Add a tool rule to the permissions.deny list in .claude/settings.json.
The rule format is Tool(specifier), for example:
  Write(**)
  Bash
  WebFetch`,
		RunE: runClaudeToolsDeny,
	}
	remove := &cobra.Command{
		Use:   "remove [rule]",
		Short: "Remove a tool rule from permissions.",
		Long:  `Remove a tool rule from either the allow or deny list.`,
		RunE:  runClaudeToolsRemove,
	}

	root.AddCommand(list, allow, deny, remove)
	for _, c := range []*cobra.Command{list, allow, deny, remove} {
		c.Flags().String(cmdx.FlagFor, string(tool.ClaudeID), toolsForHint)
	}
	addOutputFlag(list)
	return toolsCmds{root: root, list: list, allow: allow, deny: deny, remove: remove}
}

// requireClaudeTarget validates the --for flag of a tools subcommand. Only
// Claude Code is supported today: other known tools (or "all") get a "not
// supported yet" error, anything else an unknown-tool error.
func requireClaudeTarget(cmd *cobra.Command) error {
	target := cmdx.String(cmd, cmdx.FlagFor)
	if target == "" {
		return nil
	}
	scope, err := cmdx.ParseTargetScope(target)
	if err != nil {
		return fmt.Errorf("unknown tool %q for --for (supported: %s)", target, tool.ClaudeID)
	}
	if scope != cmdx.TargetScope(tool.ClaudeID) {
		return fmt.Errorf("--for %s is not supported yet: tools subcommands only support %s", target, tool.ClaudeID)
	}
	return nil
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

// claudeToolsPermissions is the JSON shape for `rsk claude tools list -o json`
// — the raw allow/deny rule lists as configured in .claude/settings.json.
type claudeToolsPermissions struct {
	Allow []string `json:"allow"`
	Deny  []string `json:"deny"`
}

func runClaudeToolsList(cmd *cobra.Command, args []string) error {
	if err := checkOutput(cmd); err != nil {
		return err
	}
	if err := requireClaudeTarget(cmd); err != nil {
		return err
	}
	format := printer.Format(cmd)

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

	if format == output.FormatJSON {
		return newFormatter(cmd, format).JSON(claudeToolsPermissions{Allow: allow, Deny: deny})
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

	header := []string{headerName, "Status"}
	tableRows := termkit.Rows(rows, claudeToolsTable)

	return renderList(cmd, format, "Claude Code tools available:", rows, header, tableRows, claudeToolNames())
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

// claudeToolNames lists the tools table's row IDs — the bare tool name
// allow/deny/remove accept as a rule — parallel to claudeToolsTable's rows,
// which are built by walking availableClaudeTools in order.
func claudeToolNames() []string {
	names := make([]string, len(availableClaudeTools))
	copy(names, availableClaudeTools)
	return names
}

func runClaudeToolsAllow(cmd *cobra.Command, args []string) error {
	if err := requireClaudeTarget(cmd); err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	rule, err := nameFromArgsOrPrompt(cmd, args, fieldArgRule, "Tool rule (e.g. Bash(npm run *))")
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
	if err := requireClaudeTarget(cmd); err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	rule, err := nameFromArgsOrPrompt(cmd, args, fieldArgRule, "Tool rule (e.g. Write(**) or Bash)")
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
	if err := requireClaudeTarget(cmd); err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	rule, err := nameFromArgsOrPrompt(cmd, args, fieldArgRule, "Tool rule to remove")
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
