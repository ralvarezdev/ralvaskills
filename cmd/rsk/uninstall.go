package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/ralvaskills/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/internal/config"
	"github.com/ralvarezdev/ralvaskills/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/internal/skill"
	"github.com/ralvarezdev/ralvaskills/internal/ui"
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

func runUninstall(cmd *cobra.Command, opts uninstallOpts, args []string) error {
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()

	args, err := resolveUninstallArgs(out, args)
	if err != nil {
		return err
	}
	if !opts.global && opts.forTool != "" {
		return errors.New("--for requires --global")
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

	if err = checkNoPersonalUninstall(names, targets, opts.personal); err != nil {
		return err
	}

	toRemove := findLinkedEntries(names, targets)
	if len(toRemove) == 0 {
		ui.Info(out, "No installed skills matched — nothing to remove.")
		return nil
	}

	printUninstallPreview(out, toRemove, opts.dryRun)
	if opts.dryRun {
		return nil
	}
	if !ui.ConfirmYN(out, "Proceed?") {
		fmt.Fprintln(out, "Aborted.")
		return nil
	}
	fmt.Fprintln(out)

	if failed := unlinkAll(out, errOut, toRemove); failed > 0 {
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

// resolveUninstallArgs prompts for a name when none was given on the command
// line.
func resolveUninstallArgs(out io.Writer, args []string) ([]string, error) {
	if len(args) > 0 {
		return args, nil
	}
	name, err := ui.Ask(out, "Skill or bundle to remove", "")
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, errors.New("specify at least one bundle or skill name")
	}
	return []string{name}, nil
}

// checkNoPersonalUninstall rejects the uninstall when any named skill is
// linked from personal/ and --personal wasn't passed. It checks via the
// symlink target path, since that's the only place "personal" is recorded.
func checkNoPersonalUninstall(names, targets []string, personal bool) error {
	if personal {
		return nil
	}
	for _, name := range names {
		for _, target := range targets {
			linkPath := filepath.Join(target, name)
			symlinkTarget, err := os.Readlink(linkPath)
			if err != nil {
				continue // not linked here, skip
			}
			if skill.IsPersonalPath(symlinkTarget) {
				return fmt.Errorf("skill %q is in personal/ — pass --personal to uninstall it", name)
			}
		}
	}
	return nil
}

// findLinkedEntries returns the (name, targetDir) pairs among names ×
// targets that are actually linked.
func findLinkedEntries(names, targets []string) []removeEntry {
	var toRemove []removeEntry
	for _, name := range names {
		for _, target := range targets {
			if skill.IsLinked(name, target) {
				toRemove = append(toRemove, removeEntry{name, target})
			}
		}
	}
	return toRemove
}

// printUninstallPreview renders the name → target table for the skills
// about to be (or that would be, under --dry-run) removed.
func printUninstallPreview(out io.Writer, toRemove []removeEntry, dryRun bool) {
	fmt.Fprintln(out)
	if dryRun {
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
}

// unlinkAll removes each entry's symlink, reporting (rather than aborting
// on) a failure so the rest of the batch still gets removed. It returns how
// many failed.
func unlinkAll(out, errOut io.Writer, toRemove []removeEntry) int {
	var failed int
	for _, e := range toRemove {
		if err := skill.Unlink(e.name, e.target); err != nil {
			ui.Failf(errOut, "remove %s: %v", e.name, err)
			failed++
		} else {
			ui.Success(out, fmt.Sprintf("%s removed from %s", e.name, e.target))
		}
	}
	return failed
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
