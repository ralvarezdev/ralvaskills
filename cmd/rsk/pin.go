package main

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/ralvaskills/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/internal/ui"
)

var (
	pinCmd = &cobra.Command{
		Use:   "pin [name]",
		Short: "Pin a skill so it is imported into each tool's project config.",
		Long: `Pin a skill that is already tracked in rsk.mod. Pinning imports the skill
into every configured tool's project config (.rsk/CLAUDE.md for Claude Code,
opencode.json for OpenCode) so it is auto-loaded by agents in this project.

The skill must already be in rsk.mod — run 'rsk install <name>' first.`,
		RunE: runPin,
	}

	unpinCmd = &cobra.Command{
		Use:   "unpin [name]",
		Short: "Remove a skill from the pinned list.",
		Long: `Remove a skill from the pinned list. The skill stays installed (still
symlinked into .rsk/skills/) but is no longer auto-imported into each tool's
project config.`,
		RunE: runUnpin,
	}
)

func runPin(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	name, err := nameFromArgsOrPrompt(cmd, args, "Skill to pin")
	if err != nil {
		return err
	}

	rskDir, err := manifest.ProjectFolderPath()
	if err != nil {
		return err
	}

	m, err := manifest.ReadMod(rskDir)
	if err != nil {
		return err
	}

	if _, ok := m.Skills[name]; !ok {
		return fmt.Errorf("skill %q is not in the manifest — run 'rsk install %s' first", name, name)
	}

	if slices.Contains(m.Pinned, name) {
		ui.Info(out, name+" is already pinned")
		return nil
	}

	m.Pinned = append(m.Pinned, name)
	if err = manifest.WriteMod(rskDir, m); err != nil {
		return err
	}
	if err = syncPinnedAllTools(rskDir, m); err != nil {
		return err
	}

	fmt.Fprintln(out)
	ui.Success(out, "pinned "+ui.SkillName(name))
	fmt.Fprintln(out)
	return nil
}

func runUnpin(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	name, err := nameFromArgsOrPrompt(cmd, args, "Skill to unpin")
	if err != nil {
		return err
	}

	rskDir, err := manifest.ProjectFolderPath()
	if err != nil {
		return err
	}

	m, err := manifest.ReadMod(rskDir)
	if err != nil {
		return err
	}

	if !slices.Contains(m.Pinned, name) {
		ui.Info(out, name+" is not pinned")
		return nil
	}

	m.Pinned = slices.DeleteFunc(m.Pinned, func(v string) bool { return v == name })
	if err = manifest.WriteMod(rskDir, m); err != nil {
		return err
	}
	if err = syncPinnedAllTools(rskDir, m); err != nil {
		return err
	}

	fmt.Fprintln(out)
	ui.Success(out, "unpinned "+ui.SkillName(name))
	fmt.Fprintln(out)
	return nil
}
