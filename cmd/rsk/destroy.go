package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/termkit"

	"github.com/ralvarezdev/ralvaskills/v3/internal"
	"github.com/ralvarezdev/ralvaskills/v3/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/v3/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/v3/internal/skill"
	"github.com/ralvarezdev/ralvaskills/v3/internal/tool"
	"github.com/ralvarezdev/ralvaskills/v3/internal/ui"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Remove the rsk project from the current directory.",
	Long: `Remove .rsk/ from the current directory and clean up tool-specific config
(CLAUDE.md import, opencode.json entries).

This does not touch globally installed skills or the user-level config.

It asks for confirmation first (a plain-text prompt when piped, which aborts
if no answer is given); pass --yes to skip the prompt.

Examples:
  rsk destroy
  rsk destroy --yes`,
	RunE: runDestroy,
}

func runDestroy(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	rskDir := filepath.Join(cwd, internal.ProjectFolderName)

	// Read mod before deleting .rsk/ so we know which tools and skills to clean up.
	tools := []tool.ID{tool.ClaudeID}
	var installedNames []string
	if m, modErr := manifest.ReadMod(rskDir); modErr == nil {
		tools = m.Tools
		for name := range m.Skills {
			installedNames = append(installedNames, name)
		}
	}

	printDestroyPlan(out, rskDir, cwd, tools, len(installedNames) > 0)

	if !termkit.Confirm(cmd, cmdx.Bool(cmd, cmdx.FlagYes), termkit.StdioIsTerminal(), "Proceed?") {
		fmt.Fprintln(out, "Aborted.")
		return nil
	}
	fmt.Fprintln(out)

	if err = unlinkAndRemovePinned(out, cwd, tools, installedNames); err != nil {
		return err
	}
	if err = unregisterMCP(cwd, tools); err != nil {
		return err
	}
	if err = os.RemoveAll(rskDir); err != nil {
		return fmt.Errorf("remove .rsk: %w", err)
	}

	fmt.Fprintln(out)
	ui.Success(out, "removed .rsk/ and cleaned up tool configs")
	fmt.Fprintln(out)
	return nil
}

// destroyPlanRow is one thing runDestroy is about to remove: a path (or tool
// ID) and what it holds.
type destroyPlanRow struct {
	path, what string
}

// destroyPlanTable projects one destroyPlanRow, both columns muted alike
// since nothing here is more or less severe than anything else.
var destroyPlanTable = termkit.Table[destroyPlanRow]{
	Row: func(r destroyPlanRow) []any {
		return []any{ui.MutedStyle.Render(r.path), ui.MutedStyle.Render(r.what)}
	},
}

// printDestroyPlan lists what runDestroy is about to remove: the .rsk/
// directory, each tool's pinned-skill imports, and (when hasSkills) its
// skill symlink directory.
func printDestroyPlan(out io.Writer, rskDir, cwd string, tools []tool.ID, hasSkills bool) {
	rows := []destroyPlanRow{{path: rskDir, what: ""}}
	for _, id := range tools {
		t, ok := tool.Get(id)
		if !ok {
			continue
		}
		rows = append(rows, destroyPlanRow{path: string(id), what: "pinned skill imports"})
		if hasSkills {
			rows = append(rows, destroyPlanRow{path: t.ProjectSkillsDir(cwd), what: "skill symlinks"})
		}
	}

	fmt.Fprintln(out)
	ui.Header(out, "This will remove:")
	termkit.WriteTableStyled(
		out, []string{"", ""}, termkit.Rows(rows, destroyPlanTable),
		false, nil, termkit.TableBorderless, ui.Theme,
	)
	fmt.Fprintln(out)
}

// unlinkAndRemovePinned removes each installed skill's symlink from every
// distinct tool skills dir, then clears each tool's pinned-skill imports.
func unlinkAndRemovePinned(out io.Writer, cwd string, tools []tool.ID, installedNames []string) error {
	seenDirs := make(map[string]bool)
	for _, id := range tools {
		t, ok := tool.Get(id)
		if !ok {
			continue
		}
		dir := t.ProjectSkillsDir(cwd)
		if !seenDirs[dir] {
			seenDirs[dir] = true
			for _, name := range installedNames {
				if unlinkErr := skill.Unlink(name, dir); unlinkErr != nil {
					ui.Warn(out, fmt.Sprintf("unlink %s: %v", name, unlinkErr))
				}
			}
		}
		if err := t.RemovePinned(cwd); err != nil {
			return err
		}
	}
	return nil
}

// unregisterMCP removes the rsk MCP server from each configured tool's project
// config and drops the workflow pointer from ./CLAUDE.md.
func unregisterMCP(cwd string, tools []tool.ID) error {
	for _, id := range tools {
		t, ok := tool.Get(id)
		if !ok {
			continue
		}
		if err := t.UnregisterMCP(cwd); err != nil {
			return err
		}
	}
	return tool.RemoveMCPPointer(cwd)
}
