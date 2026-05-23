package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ralvarezdev/ralvaskills/internal"
	"github.com/ralvarezdev/ralvaskills/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/internal/fsperm"
	"github.com/ralvarezdev/ralvaskills/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/internal/tool"
	"github.com/ralvarezdev/ralvaskills/internal/ui"
	"github.com/spf13/cobra"
)

var newProjectFor string

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Initialize an rsk project in the current directory.",
	Long: `Create .rsk/rsk.mod, .rsk/skills/, and tool-specific config in the current
directory so this folder becomes an rsk project.

Use --for to select which tools to configure:
  --for claude-code  (default) Appends @.rsk/CLAUDE.md to ./CLAUDE.md
  --for opencode               Writes pinned skills to opencode.json instructions
  --for all                    Both tools

Examples:
  rsk new
  rsk new --for opencode
  rsk new --for all`,
	RunE: runNew,
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringVar(
		&newProjectFor, cmdx.FlagFor, string(tool.ClaudeID),
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

func runNew(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	tools, err := toolsFromFlag(newProjectFor)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	if existing, existErr := manifest.ProjectFolderPath(); existErr == nil {
		return fmt.Errorf("rsk project already exists at %s", existing)
	}
	rskDir := filepath.Join(cwd, internal.ProjectFolderName)
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
	ui.Success(out, fmt.Sprintf("initialized .rsk/ for %s", newProjectFor))
	ui.Indent(out, "Run 'rsk install <name>' to add skills to this project.")
	fmt.Fprintln(out)
	return nil
}
