package cmd

import (
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/ralvarezdev/ralvaskills/cli/internal"
	"github.com/ralvarezdev/ralvaskills/cli/internal/config"
	"github.com/ralvarezdev/ralvaskills/cli/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/cli/internal/skill"
	"github.com/ralvarezdev/ralvaskills/cli/internal/source"
	"github.com/ralvarezdev/ralvaskills/cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	skillCmd = &cobra.Command{
		Use:   "skill",
		Short: "Manage skills in the project manifest.",
		Long: `Add, remove, pin, and list skills in the project rsk.mod.

  rsk skill add <name>     Add a skill and install it
  rsk skill pin <name>     Import the skill in .rsk/CLAUDE.md
  rsk skill unpin <name>   Remove the import
  rsk skill list           Show all skills in the manifest
  rsk skill upgrade <name> Re-resolve and re-link a skill`,
	}

	skillAddCmd = &cobra.Command{
		Use:   "add <name[@version]>",
		Short: "Add a skill to the project manifest and install it.",
		Args:  cobra.ExactArgs(1),
		RunE:  runSkillAdd,
	}

	skillPinCmd = &cobra.Command{
		Use:   "pin <name>",
		Short: "Pin a skill so it is imported in .rsk/CLAUDE.md.",
		Args:  cobra.ExactArgs(1),
		RunE:  runSkillPin,
	}

	skillUnpinCmd = &cobra.Command{
		Use:   "unpin <name>",
		Short: "Remove a skill from the pinned list.",
		Args:  cobra.ExactArgs(1),
		RunE:  runSkillUnpin,
	}

	skillListCmd = &cobra.Command{
		Use:   "list",
		Short: "List skills in the project manifest.",
		RunE:  runSkillList,
	}

	skillUpgradeCmd = &cobra.Command{
		Use:   "upgrade <name>",
		Short: "Re-resolve and re-link a skill to its latest available version.",
		Args:  cobra.ExactArgs(1),
		RunE:  runSkillUpgrade,
	}
)

func init() {
	rootCmd.AddCommand(skillCmd)
	skillCmd.AddCommand(skillAddCmd)
	skillCmd.AddCommand(skillPinCmd)
	skillCmd.AddCommand(skillUnpinCmd)
	skillCmd.AddCommand(skillListCmd)
	skillCmd.AddCommand(skillUpgradeCmd)

	skillAddCmd.Flags().Bool(internal.FlagPin, false, fmt.Sprintf("Also pin the skill in %s", manifest.RelativeClaudeFilePath))
}

func runSkillAdd(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	ctx := cmd.Context()

	name, version, err := parseNameVersion(args[0])
	if err != nil {
		return err
	}

	rskDir, err := manifest.LocalConfigFolderPath()
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n  Run 'rsk init' to set up rsk on this machine", err)
	}

	localSrc := newLocalSource(cfg)
	officialSrc := source.NewOfficial(cfg.OfficialCache)

	s, err := findSkillByName(ctx, name, localSrc, officialSrc)
	if err != nil {
		return err
	}

	skillsDir := manifest.LocalConfigSkillsPath(rskDir)
	if err := skill.Link(s, skillsDir); err != nil {
		return fmt.Errorf("link %s: %w", name, err)
	}

	m, err := manifest.ReadMod(rskDir)
	if err != nil {
		return err
	}

	constraint := version
	if constraint == "" {
		constraint = "*"
	}
	m.Skills[name] = constraint

	if internal.FlagBool(cmd, internal.FlagPin) && !slices.Contains(m.Pinned, name) {
		m.Pinned = append(m.Pinned, name)
	}

	if err := manifest.WriteMod(rskDir, m); err != nil {
		return err
	}

	lock, err := manifest.ReadLock(rskDir)
	if err != nil {
		return err
	}
	lock = manifest.UpsertLockEntry(lock, manifest.LockEntry{
		Name:    name,
		Version: s.Version,
		Source:  s.Source,
		Path:    s.Path,
	})
	if err := manifest.WriteLock(rskDir, lock); err != nil {
		return err
	}

	if slices.Contains(m.Pinned, name) {
		if err := syncPinnedAllTools(rskDir, m); err != nil {
			return err
		}
	}

	fmt.Fprintln(out)
	ui.Success(out, fmt.Sprintf("added %s  (%s)", ui.SkillName(name), ui.SkillVersion(constraint)))
	fmt.Fprintln(out)
	return nil
}

func runSkillPin(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	name := args[0]

	rskDir, err := manifest.LocalConfigFolderPath()
	if err != nil {
		return err
	}

	m, err := manifest.ReadMod(rskDir)
	if err != nil {
		return err
	}

	if _, ok := m.Skills[name]; !ok {
		return fmt.Errorf("skill %q is not in the manifest — run 'rsk skill add %s' first", name, name)
	}

	if slices.Contains(m.Pinned, name) {
		ui.Info(out, fmt.Sprintf("%s is already pinned", name))
		return nil
	}

	m.Pinned = append(m.Pinned, name)
	if err := manifest.WriteMod(rskDir, m); err != nil {
		return err
	}
	if err := syncPinnedAllTools(rskDir, m); err != nil {
		return err
	}

	fmt.Fprintln(out)
	ui.Success(out, fmt.Sprintf("pinned %s", ui.SkillName(name)))
	fmt.Fprintln(out)
	return nil
}

func runSkillUnpin(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	name := args[0]

	rskDir, err := manifest.LocalConfigFolderPath()
	if err != nil {
		return err
	}

	m, err := manifest.ReadMod(rskDir)
	if err != nil {
		return err
	}

	if !slices.Contains(m.Pinned, name) {
		ui.Info(out, fmt.Sprintf("%s is not pinned", name))
		return nil
	}

	m.Pinned = slices.DeleteFunc(m.Pinned, func(v string) bool { return v == name })
	if err := manifest.WriteMod(rskDir, m); err != nil {
		return err
	}
	if err := syncPinnedAllTools(rskDir, m); err != nil {
		return err
	}

	fmt.Fprintln(out)
	ui.Success(out, fmt.Sprintf("unpinned %s", ui.SkillName(name)))
	fmt.Fprintln(out)
	return nil
}

func runSkillList(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	rskDir, err := manifest.LocalConfigFolderPath()
	if err != nil {
		return err
	}

	m, err := manifest.ReadMod(rskDir)
	if err != nil {
		return err
	}

	if len(m.Skills) == 0 {
		fmt.Fprintln(out)
		ui.Warn(out, "No skills in manifest. Run 'rsk skill add <name>' to add one.")
		fmt.Fprintln(out)
		return nil
	}

	fmt.Fprintln(out)
	ui.Header(out, "Project skills:")

	skillsDir := manifest.LocalConfigSkillsPath(rskDir)
	pinnedSet := make(map[string]bool, len(m.Pinned))
	for _, p := range m.Pinned {
		pinnedSet[p] = true
	}

	names := make([]string, 0, len(m.Skills))
	for n := range m.Skills {
		names = append(names, n)
	}
	sort.Strings(names)
	nameWidth := ui.MaxWidth(names)

	for _, name := range names {
		constraint := m.Skills[name]
		installed := skill.IsLinked(name, skillsDir)

		mark := ui.SuccessMark
		if !installed {
			mark = ui.ErrorMark
		}
		pinnedTag := ""
		if pinnedSet[name] {
			pinnedTag = "  [pinned]"
		}

		fmt.Fprintf(out, "  %s  %s  %s%s\n",
			mark,
			ui.PadRight(ui.SkillName(name), nameWidth),
			ui.SkillVersion(constraint),
			pinnedTag,
		)
	}

	fmt.Fprintln(out)
	return nil
}

func runSkillUpgrade(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	ctx := cmd.Context()
	name := args[0]

	rskDir, err := manifest.LocalConfigFolderPath()
	if err != nil {
		return err
	}

	m, err := manifest.ReadMod(rskDir)
	if err != nil {
		return err
	}

	if _, ok := m.Skills[name]; !ok {
		return fmt.Errorf("skill %q is not in the manifest — run 'rsk skill add %s' first", name, name)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n  Run 'rsk init' to set up rsk on this machine", err)
	}

	localSrc := newLocalSource(cfg)
	officialSrc := source.NewOfficial(cfg.OfficialCache)

	s, err := findSkillByName(ctx, name, localSrc, officialSrc)
	if err != nil {
		return err
	}

	skillsDir := manifest.LocalConfigSkillsPath(rskDir)
	if err := skill.Link(s, skillsDir); err != nil {
		return fmt.Errorf("link %s: %w", name, err)
	}

	lock, err := manifest.ReadLock(rskDir)
	if err != nil {
		return err
	}
	lock = manifest.UpsertLockEntry(lock, manifest.LockEntry{
		Name:    name,
		Version: s.Version,
		Source:  s.Source,
		Path:    s.Path,
	})
	if err := manifest.WriteLock(rskDir, lock); err != nil {
		return err
	}

	fmt.Fprintln(out)
	ui.Success(out, fmt.Sprintf("upgraded %s  %s  %s",
		ui.SkillName(name),
		ui.Arrow,
		ui.SkillVersion(s.Version),
	))
	fmt.Fprintln(out)
	return nil
}

// syncPinnedAllTools writes pinned skill entries for every tool in m.Tools.
// rskDir is .rsk/, projectDir is its parent.
func syncPinnedAllTools(rskDir string, m manifest.Mod) error {
	projectDir := filepath.Dir(rskDir)
	for _, tool := range m.Tools {
		switch tool {
		case internal.ToolClaudeCode:
			if err := manifest.WritePinned(rskDir, m.Pinned); err != nil {
				return fmt.Errorf("sync claude-code: %w", err)
			}
		case internal.ToolOpenCode:
			if err := manifest.SyncOpenCodeInstructions(projectDir, m.Pinned); err != nil {
				return fmt.Errorf("sync opencode: %w", err)
			}
		default:
			return fmt.Errorf("unknown tool %q in rsk.mod", tool)
		}
	}
	return nil
}

func parseNameVersion(arg string) (name, version string, err error) {
	if idx := strings.Index(arg, "@"); idx >= 0 {
		name = arg[:idx]
		version = arg[idx+1:]
	} else {
		name = arg
	}
	if name == "" {
		return "", "", fmt.Errorf("skill name must not be empty (got %q)", arg)
	}
	return name, version, nil
}
