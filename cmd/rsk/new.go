package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/ralvaskills/internal"
	"github.com/ralvarezdev/ralvaskills/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/internal/tool"
	"github.com/ralvarezdev/ralvaskills/internal/ui"
)

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

	forFlag, err := resolveForFlag(cmd, out)
	if err != nil {
		return err
	}

	tools, err := toolsFromFlag(forFlag)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	m, rskDir, isNew, err := loadOrCreateManifest(cwd, tools, forFlag, out)
	if err != nil || m == nil {
		return err
	}

	if err = manifest.WriteMod(rskDir, *m); err != nil {
		return err
	}
	if err = syncToolConfigs(tools, cwd, m.Pinned); err != nil {
		return err
	}

	fmt.Fprintln(out)
	if isNew {
		ui.Success(out, "initialized .rsk/ for "+forFlag)
	} else {
		ui.Success(out, "updated .rsk/ for "+forFlag)
	}
	ui.Indent(out, "Run 'rsk install <name>' to add skills to this project.")
	fmt.Fprintln(out)
	return nil
}

// resolveForFlag returns --for's value, prompting interactively when it
// wasn't set on the command line.
func resolveForFlag(cmd *cobra.Command, out io.Writer) (string, error) {
	forFlag := cmdx.String(cmd, cmdx.FlagFor)
	if cmd.Flags().Changed(cmdx.FlagFor) {
		return forFlag, nil
	}

	choices := []string{string(tool.ClaudeID), string(tool.OpenCodeID), cmdx.ForAll}
	idx, err := ui.Select(out, "Tools to configure", choices, 0)
	if err != nil {
		return "", err
	}
	return choices[idx], nil
}

// loadOrCreateManifest reads the existing project manifest and adds any of
// tools it's missing, or builds a fresh one if no project exists yet. A nil
// *manifest.Mod with a nil error means the project is already configured for
// every requested tool and the caller should return without further work.
func loadOrCreateManifest(
	cwd string, tools []tool.ID, forFlag string, out io.Writer,
) (m *manifest.Mod, rskDir string, isNew bool, err error) {
	existing, existErr := manifest.ProjectFolderPath()
	if existErr == nil {
		loaded, readErr := manifest.ReadMod(existing)
		if readErr != nil {
			return nil, "", false, readErr
		}

		var added []tool.ID
		for _, t := range tools {
			if !slices.Contains(loaded.Tools, t) {
				added = append(added, t)
				loaded.Tools = append(loaded.Tools, t)
			}
		}
		if len(added) == 0 {
			ui.Info(out, "rsk project already configured for "+forFlag)
			return nil, "", false, nil
		}

		ui.Info(out, fmt.Sprintf("Adding tools to existing project: %v", added))
		return &loaded, existing, false, nil
	}

	// No project yet — existErr just signals "not found", nothing to surface.
	return &manifest.Mod{
		Tools:  tools,
		Skills: make(map[string]string),
		Pinned: []string{},
	}, filepath.Join(cwd, internal.ProjectFolderName), true, nil
}

// syncToolConfigs pushes the currently pinned skills into each tool's
// project config (CLAUDE.md / opencode.json), skipping any ID that isn't a
// registered tool.
func syncToolConfigs(tools []tool.ID, cwd string, pinned []string) error {
	for _, id := range tools {
		t, ok := tool.Get(id)
		if !ok {
			continue
		}
		if err := t.SyncPinned(cwd, pinned); err != nil {
			return err
		}
	}
	return nil
}
