package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ralvarezdev/ralvaskills/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/internal/fsperm"
	"github.com/ralvarezdev/ralvaskills/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/internal/tool"
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

	projectInitCmd.Flags().StringVar(
		&projectInitFor, cmdx.FlagFor, string(tool.ClaudeID),
		"Tools to configure: claude-code | opencode | all",
	)
}

func toolsFromFlag(flag string) ([]tool.ID, error) {
	switch flag {
	case string(tool.ClaudeID):
		return []tool.ID{tool.ClaudeID}, nil
	case string(tool.OpenCodeID):
		return []tool.ID{tool.OpenCodeID}, nil
	case cmdx.ForAll:
		return []tool.ID{tool.ClaudeID, tool.OpenCodeID}, nil
	default:
		return nil, fmt.Errorf("--for must be %s, %s, or %s; got %q",
			tool.ClaudeID, tool.OpenCodeID, cmdx.ForAll, flag)
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

	rskDir := filepath.Join(cwd, manifest.ProjectFolderName)
	skillsDir := manifest.ProjectSkillsPath(rskDir)
	if err = os.MkdirAll(skillsDir, fsperm.Dir); err != nil {
		return fmt.Errorf("create .rsk/skills: %w", err)
	}

	m := manifest.Mod{
		Tools:  tools,
		Skills: make(map[string]string),
		Pinned: []string{},
	}
	if err = manifest.WriteMod(rskDir, m); err != nil {
		return err
	}

	// Initialize per-tool project config. Empty pinned list on first init.
	for _, id := range tools {
		t, ok := tool.Get(id)
		if !ok {
			continue
		}
		if err = t.SyncPinned(cwd, nil); err != nil {
			return err
		}
	}

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

	rskDir := filepath.Join(cwd, manifest.ProjectFolderName)

	// Read tools from mod before deleting .rsk/ so we know what to clean up.
	tools := []tool.ID{tool.ClaudeID}
	if m, modErr := manifest.ReadMod(rskDir); modErr == nil {
		tools = m.Tools
	}

	for _, id := range tools {
		t, ok := tool.Get(id)
		if !ok {
			continue
		}
		if err = t.RemovePinned(cwd); err != nil {
			return err
		}
	}

	if err = os.RemoveAll(rskDir); err != nil {
		return fmt.Errorf("remove .rsk: %w", err)
	}

	fmt.Fprintln(out)
	ui.Success(out, "removed .rsk/ and cleaned up tool configs")
	fmt.Fprintln(out)
	return nil
}
