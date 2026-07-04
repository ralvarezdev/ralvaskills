package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ralvarezdev/ralvaskills/internal"
	"github.com/ralvarezdev/ralvaskills/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/internal/skill"
	"github.com/ralvarezdev/ralvaskills/internal/tool"
	"github.com/ralvarezdev/ralvaskills/internal/ui"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Remove the rsk project from the current directory.",
	Long: `Remove .rsk/ from the current directory and clean up tool-specific config
(CLAUDE.md import, opencode.json entries).

This does not touch globally installed skills or the user-level config.`,
	RunE: runDestroy,
}

func init() {
	rootCmd.AddCommand(destroyCmd)
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

	hasSkills := len(installedNames) > 0

	fmt.Fprintln(out)
	ui.Header(out, "This will remove:")
	fmt.Fprintf(out, "  %s\n", ui.MutedStyle.Render(rskDir))
	for _, id := range tools {
		t, ok := tool.Get(id)
		if !ok {
			continue
		}
		fmt.Fprintf(out, "  %s  %s\n",
			ui.MutedStyle.Render(string(id)),
			ui.MutedStyle.Render("pinned skill imports"),
		)
		if hasSkills {
			fmt.Fprintf(out, "  %s  %s\n",
				ui.MutedStyle.Render(t.ProjectSkillsDir(cwd)),
				ui.MutedStyle.Render("skill symlinks"),
			)
		}
	}
	fmt.Fprintln(out)

	if !ui.ConfirmYN(out, "Proceed?") {
		fmt.Fprintln(out, "Aborted.")
		return nil
	}
	fmt.Fprintln(out)

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
