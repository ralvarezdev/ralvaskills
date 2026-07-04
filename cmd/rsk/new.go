package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/ralvarezdev/ralvaskills/internal"
	"github.com/ralvarezdev/ralvaskills/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/internal/tool"
	"github.com/ralvarezdev/ralvaskills/internal/ui"
	"github.com/spf13/cobra"
)

var newProjectFor string

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Initialize an rsk project in the current directory.",
	Long: `Create .rsk/rsk.mod and tool-specific config in the current directory so
this folder becomes an rsk project. The .claude/skills/ directory is created
lazily the first time a skill is installed.

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

	forFlag := newProjectFor
	if !cmd.Flags().Changed(cmdx.FlagFor) {
		choices := []string{string(tool.ClaudeID), string(tool.OpenCodeID), cmdx.ForAll}
		idx, err := ui.Select(out, "Tools to configure", choices, 0)
		if err != nil {
			return err
		}
		forFlag = choices[idx]
	}

	tools, err := toolsFromFlag(forFlag)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	existing, existErr := manifest.ProjectFolderPath()
	rskDir := filepath.Join(cwd, internal.ProjectFolderName)

	var m manifest.Mod
	var isNew bool

	if existErr == nil {
		// Project exists, read it
		m, err = manifest.ReadMod(existing)
		if err != nil {
			return err
		}

		// Find missing tools
		var added []tool.ID
		for _, t := range tools {
			found := slices.Contains(m.Tools, t)
			if !found {
				added = append(added, t)
				m.Tools = append(m.Tools, t)
			}
		}

		if len(added) == 0 {
			ui.Info(out, fmt.Sprintf("rsk project already configured for %s", newProjectFor))
			return nil
		}

		ui.Info(out, fmt.Sprintf("Adding tools to existing project: %v", added))
		rskDir = existing

	} else {
		isNew = true
		m = manifest.Mod{
			Tools:  tools,
			Skills: make(map[string]string),
			Pinned: []string{},
		}
	}

	if err = manifest.WriteMod(rskDir, m); err != nil {
		return err
	}

	// Initialize per-tool project config.
	for _, id := range tools {
		t, ok := tool.Get(id)
		if !ok {
			continue
		}
		// SyncPinned ensures the tool config (CLAUDE.md / opencode.json)
		// receives the currently pinned skills.
		if err = t.SyncPinned(cwd, m.Pinned); err != nil {
			return err
		}
	}

	fmt.Fprintln(out)
	if isNew {
		ui.Success(out, fmt.Sprintf("initialized .rsk/ for %s", newProjectFor))
	} else {
		ui.Success(out, fmt.Sprintf("updated .rsk/ for %s", newProjectFor))
	}
	ui.Indent(out, "Run 'rsk install <name>' to add skills to this project.")
	fmt.Fprintln(out)
	return nil
}
