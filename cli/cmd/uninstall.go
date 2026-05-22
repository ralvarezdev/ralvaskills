package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ralvarezdev/ralvaskills/cli/internal/config"
	"github.com/ralvarezdev/ralvaskills/cli/internal/skill"
	"github.com/ralvarezdev/ralvaskills/cli/internal/ui"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall [bundle...] [flags]",
	Short: "Remove installed skills.",
	Long: `Remove symlinks for one or more bundles or a single skill.

Examples:
  rsk uninstall go-grpc
  rsk uninstall --skill go-architect
  rsk uninstall global --global
  rsk uninstall go-grpc --dry-run`,
	RunE: runUninstall,
}

var (
	uninstallGlobal   bool
	uninstallFor      string
	uninstallSkill    string
	uninstallPersonal bool
	uninstallDryRun   bool
)

func init() {
	rootCmd.AddCommand(uninstallCmd)
	f := uninstallCmd.Flags()
	f.BoolVar(&uninstallGlobal, "global", false, "Target global skills dir(s)")
	f.StringVar(&uninstallFor, "for", "", "Scope --global to a single tool (claude-code|opencode)")
	f.StringVar(&uninstallSkill, "skill", "", "Remove a single skill by name")
	f.BoolVar(&uninstallPersonal, "personal", false, "Allow removing personal/ skills")
	f.BoolVar(&uninstallDryRun, "dry-run", false, "Show what would be removed without doing it")
}

func runUninstall(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	if uninstallSkill != "" && len(args) > 0 {
		return fmt.Errorf("use either positional bundle names or --skill <name>, not both")
	}
	if uninstallSkill == "" && len(args) == 0 {
		return fmt.Errorf("specify at least one bundle name or use --skill <name>")
	}
	if !uninstallGlobal && uninstallFor != "" {
		return fmt.Errorf("--for requires --global")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n  Run 'rsk init' to set up rsk on this machine", err)
	}

	targets, err := resolveTargetDirs(cfg, uninstallGlobal, uninstallFor)
	if err != nil {
		return err
	}

	catalog := config.LoadCatalog("")
	names, err := skillNamesFromArgs(args, uninstallSkill, catalog)
	if err != nil {
		return err
	}

	// Personal guard: check via symlink target path.
	if !uninstallPersonal {
		for _, name := range names {
			for _, target := range targets {
				linkPath := filepath.Join(target, name)
				symlinkTarget, readErr := os.Readlink(linkPath)
				if readErr != nil {
					continue // not linked here, skip
				}
				if skill.IsPersonalPath(symlinkTarget) {
					return fmt.Errorf("skill %q is in personal/ — pass --personal to uninstall it", name)
				}
			}
		}
	}

	// Determine which (name, targetDir) pairs are actually linked.
	type removeEntry struct {
		name   string
		target string
	}
	var toRemove []removeEntry
	for _, name := range names {
		for _, target := range targets {
			if skill.IsLinked(name, target) {
				toRemove = append(toRemove, removeEntry{name, target})
			}
		}
	}

	if len(toRemove) == 0 {
		ui.Info(out, "No installed skills matched — nothing to remove.")
		return nil
	}

	fmt.Fprintln(out)
	if uninstallDryRun {
		ui.Header(out, "Dry run — would remove:")
	} else {
		ui.Header(out, "Skills to remove:")
	}

	nameWidth := 0
	for _, e := range toRemove {
		if len(e.name) > nameWidth {
			nameWidth = len(e.name)
		}
	}
	for _, e := range toRemove {
		fmt.Fprintf(out, "  %s  →  removed\n",
			ui.PadRight(e.name, nameWidth),
		)
		_ = filepath.Join(e.target, e.name) // referenced for future --verbose path display
	}
	fmt.Fprintln(out)

	if uninstallDryRun {
		return nil
	}

	if !ui.ConfirmYN(out, "Proceed?") {
		fmt.Fprintln(out, "Aborted.")
		return nil
	}
	fmt.Fprintln(out)

	var failed int
	for _, e := range toRemove {
		if unlinkErr := skill.Unlink(e.name, e.target); unlinkErr != nil {
			ui.Failf("remove %s: %v", e.name, unlinkErr)
			failed++
		} else {
			ui.Success(out, fmt.Sprintf("%s removed from %s", e.name, e.target))
		}
	}

	if failed > 0 {
		return fmt.Errorf("%d skill(s) failed to remove", failed)
	}
	return nil
}
