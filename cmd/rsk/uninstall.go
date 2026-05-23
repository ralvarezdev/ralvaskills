package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/ralvarezdev/ralvaskills/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/internal/config"
	"github.com/ralvarezdev/ralvaskills/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/internal/skill"
	"github.com/ralvarezdev/ralvaskills/internal/ui"
	"github.com/spf13/cobra"
)

type (
	removeEntry struct {
		name   string
		target string
	}

	uninstallOpts struct {
		global, dryRun, personal bool
		forTool                  string
	}
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <name...> [flags]",
	Short: "Remove installed bundles or skills.",
	Long: `Remove symlinks for one or more bundles or skills.

Names are resolved against the catalog: a name that matches a bundle expands
to that bundle's skills; otherwise the name is treated as a single skill.

By default operates on the current project (.rsk/skills/) and cleans up
.rsk/rsk.mod, .rsk/rsk.lock, and pinned skill imports. With --global the
operation targets the configured global skills directories instead.

Examples:
  rsk uninstall go-grpc
  rsk uninstall go-architect
  rsk uninstall global --global
  rsk uninstall go-grpc --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUninstall(cmd, uninstallOpts{
			global:   cmdx.Bool(cmd, cmdx.FlagGlobal),
			dryRun:   cmdx.Bool(cmd, cmdx.FlagDryRun),
			personal: cmdx.Bool(cmd, cmdx.FlagPersonal),
			forTool:  cmdx.String(cmd, cmdx.FlagFor),
		}, args)
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
	f := uninstallCmd.Flags()
	f.Bool(cmdx.FlagGlobal, false, "Target the configured global skills dir(s)")
	f.String(cmdx.FlagFor, "", "With --global, scope to a single tool (claude-code|opencode)")
	f.Bool(cmdx.FlagPersonal, false, "Allow removing personal/ skills")
	f.Bool(cmdx.FlagDryRun, false, "Show what would be removed without doing it")
}

func runUninstall(cmd *cobra.Command, opts uninstallOpts, args []string) error {
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()

	if len(args) == 0 {
		return fmt.Errorf("specify at least one bundle or skill name")
	}
	if !opts.global && opts.forTool != "" {
		return fmt.Errorf("--for requires --global")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n  Run 'rsk init' to set up rsk on this machine", err)
	}

	targets, err := resolveTargetDirs(cfg, opts.global, opts.forTool)
	if err != nil {
		return err
	}

	catalog, catalogWarn := config.LoadCatalog("")
	if catalogWarn != nil {
		ui.Warn(out, fmt.Sprintf("user catalog: %v", catalogWarn))
	}
	names, err := skillNamesFromArgs(args, catalog)
	if err != nil {
		return err
	}

	// Personal guard: check via symlink target path.
	if !opts.personal {
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
	if opts.dryRun {
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
		fmt.Fprintf(out, "  %s  %s  %s\n",
			ui.PadRight(ui.SkillName(e.name), nameWidth),
			ui.Arrow,
			ui.MutedPath(filepath.Join(e.target, e.name)),
		)
	}
	fmt.Fprintln(out)

	if opts.dryRun {
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
			ui.Failf(errOut, "remove %s: %v", e.name, unlinkErr)
			failed++
		} else {
			ui.Success(out, fmt.Sprintf("%s removed from %s", e.name, e.target))
		}
	}

	if failed > 0 {
		return fmt.Errorf("%d skill(s) failed to remove", failed)
	}

	if !opts.global {
		var rskDir string
		rskDir, err = manifest.ProjectFolderPath()
		if err != nil {
			return err
		}

		if cleanErr := cleanupManifest(rskDir, toRemove); cleanErr != nil {
			ui.Warn(out, fmt.Sprintf("update manifest: %v", cleanErr))
		}
	}

	return nil
}

// cleanupManifest removes uninstalled skills from rsk.mod, rsk.lock, and tool configs.
// Returns nil without error if rsk.mod does not exist.
func cleanupManifest(rskDir string, removed []removeEntry) error {
	m, err := manifest.ReadMod(rskDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}

	seen := make(map[string]bool, len(removed))
	var toClean []string
	for _, e := range removed {
		if !seen[e.name] {
			seen[e.name] = true
			toClean = append(toClean, e.name)
		}
	}
	if len(toClean) == 0 {
		return nil
	}

	for _, name := range toClean {
		delete(m.Skills, name)
		m.Pinned = slices.DeleteFunc(m.Pinned, func(v string) bool { return v == name })
	}
	if err = manifest.WriteMod(rskDir, m); err != nil {
		return fmt.Errorf("write rsk.mod: %w", err)
	}

	var lock manifest.Lock
	lock, err = manifest.ReadLock(rskDir)
	if err != nil {
		return fmt.Errorf("read rsk.lock: %w", err)
	}
	for _, name := range toClean {
		lock = manifest.RemoveLockEntry(lock, name)
	}
	if err = manifest.WriteLock(rskDir, lock); err != nil {
		return fmt.Errorf("write rsk.lock: %w", err)
	}

	if err = syncPinnedAllTools(rskDir, m); err != nil {
		return fmt.Errorf("sync tool configs: %w", err)
	}

	return nil
}
