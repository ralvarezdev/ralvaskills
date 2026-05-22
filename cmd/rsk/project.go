package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/ralvarezdev/ralvaskills/internal"
	"github.com/ralvarezdev/ralvaskills/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/internal/ui"
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage the rsk project manifest in the current directory.",
	Long: `Manage the rsk.mod project manifest.

  rsk project init    Initialize a project manifest
  rsk project remove  Remove the manifest and .rsk/ directory`,
}

var projectInitFor string

var projectInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a project manifest in the current directory.",
	Long: `Create .rsk/rsk.mod, .rsk/skills/, and tool-specific config.

Use --for to select which tools to configure:
  --for claude-code  (default) Appends @.rsk/CLAUDE.md to ./CLAUDE.md
  --for opencode               Pins are written to opencode.json instructions
  --for all                    Both tools`,
	RunE: runProjectInit,
}

var projectRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove the rsk project manifest from the current directory.",
	Long:  `Removes .rsk/ and cleans up tool-specific config (CLAUDE.md import, opencode.json entries).`,
	RunE:  runProjectRemove,
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectInitCmd)
	projectCmd.AddCommand(projectRemoveCmd)

	projectInitCmd.Flags().StringVar(&projectInitFor, internal.FlagFor, string(internal.ToolClaudeCode), "Tools to configure: claude-code | opencode | all")
}

func toolsFromFlag(flag string) ([]internal.ToolID, error) {
	switch flag {
	case string(internal.ToolClaudeCode):
		return []internal.ToolID{internal.ToolClaudeCode}, nil
	case string(internal.ToolOpenCode):
		return []internal.ToolID{internal.ToolOpenCode}, nil
	case internal.ForAll:
		return []internal.ToolID{internal.ToolClaudeCode, internal.ToolOpenCode}, nil
	default:
		return nil, fmt.Errorf("--for must be %s, %s, or %s; got %q", internal.ToolClaudeCode, internal.ToolOpenCode, internal.ForAll, flag)
	}
}

func runProjectInit(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	tools, err := toolsFromFlag(projectInitFor)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	rskDir := filepath.Join(cwd, manifest.LocalConfigFolderName)
	skillsDir := manifest.LocalConfigSkillsPath(rskDir)
	if err := os.MkdirAll(skillsDir, manifest.LocalDirPermission); err != nil {
		return fmt.Errorf("create .rsk/skills: %w", err)
	}

	m := manifest.Mod{
		Version: "1",
		Tools:   tools,
		Skills:  make(map[string]string),
		Pinned:  []string{},
	}
	if err := manifest.WriteMod(rskDir, m); err != nil {
		return err
	}

	if slices.Contains(tools, internal.ToolClaudeCode) {
		if err := manifest.WritePinned(rskDir, nil); err != nil {
			return err
		}
		if err := manifest.AppendImport(filepath.Join(cwd, "CLAUDE.md")); err != nil {
			return err
		}
	}
	// OpenCode: opencode.json is managed lazily on first pin — nothing to do at init.

	fmt.Fprintln(out)
	ui.Success(out, fmt.Sprintf("initialized .rsk/ for %s", projectInitFor))
	ui.Indent(out, "Run 'rsk skill add <name>' to add skills to this project.")
	fmt.Fprintln(out)
	return nil
}

func runProjectRemove(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	rskDir := filepath.Join(cwd, manifest.LocalConfigFolderName)

	// Read tools from mod before deleting .rsk/ so we know what to clean up.
	tools := []internal.ToolID{internal.ToolClaudeCode} // safe default if mod is unreadable
	if m, modErr := manifest.ReadMod(rskDir); modErr == nil {
		tools = m.Tools
	}

	if slices.Contains(tools, internal.ToolClaudeCode) {
		if err := manifest.RemoveImport(filepath.Join(cwd, "CLAUDE.md")); err != nil {
			return err
		}
	}
	if slices.Contains(tools, internal.ToolOpenCode) {
		if err := manifest.RemoveOpenCodeInstructions(cwd); err != nil {
			return err
		}
	}

	if err := os.RemoveAll(rskDir); err != nil {
		return fmt.Errorf("remove .rsk: %w", err)
	}

	fmt.Fprintln(out)
	ui.Success(out, "removed .rsk/ and cleaned up tool configs")
	fmt.Fprintln(out)
	return nil
}
