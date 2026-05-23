package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ralvarezdev/ralvaskills/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/internal/config"
	rskgit "github.com/ralvarezdev/ralvaskills/internal/git"
	"github.com/ralvarezdev/ralvaskills/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/internal/skill"
	"github.com/ralvarezdev/ralvaskills/internal/source"
	"github.com/ralvarezdev/ralvaskills/internal/ui"
	"github.com/spf13/cobra"
)

// officialSkillsURL is the GitHub URL for the anthropics/skills repo used as a
// source for official skills in local-repo mode.
const officialSkillsURL = "https://github.com/anthropics/skills"

type updateOpts struct {
	global, dryRun, personal, official bool
	forTool                            string
}

var updateCmd = &cobra.Command{
	Use:   "update [name...] [flags]",
	Short: "Pull the latest skills and re-link.",
	Long: `Update installed bundles or skills to their latest versions.

In local-repo mode: runs git pull on the ralvaskills clone (symlinks update
automatically). In registry mode: fetches the latest index and re-downloads any
skills with new versions.

Names are resolved against the catalog: a name that matches a bundle expands
to that bundle's skills; otherwise it's treated as a single skill.

Use --official to also refresh the anthropics/skills cache.

Examples:
  rsk update                       # local mode: git pull the local clone
  rsk update --official            # also refresh the anthropics/skills cache
  rsk update grpc-architect        # update one skill
  rsk update docs                  # update everything in the docs bundle
  rsk update go-grpc --global`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpdate(cmd, updateOpts{
			global:   cmdx.Bool(cmd, cmdx.FlagGlobal),
			dryRun:   cmdx.Bool(cmd, cmdx.FlagDryRun),
			personal: cmdx.Bool(cmd, cmdx.FlagPersonal),
			official: cmdx.Bool(cmd, cmdx.FlagOfficial),
			forTool:  cmdx.String(cmd, cmdx.FlagFor),
		}, args)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	f := updateCmd.Flags()
	f.Bool(cmdx.FlagGlobal, false, "Target the configured global skills dir(s)")
	f.String(cmdx.FlagFor, "", "With --global, scope to a single tool (claude-code|opencode)")
	f.Bool(cmdx.FlagOfficial, false, "Also re-fetch the anthropics/skills cache")
	f.Bool(cmdx.FlagPersonal, false, "Include personal/ skills in update")
	f.Bool(cmdx.FlagDryRun, false, "Show what would change without applying it")
}

func runUpdate(cmd *cobra.Command, opts updateOpts, args []string) error {
	if !opts.global && opts.forTool != "" {
		return fmt.Errorf("--for requires --global")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n  Run 'rsk init' to set up rsk on this machine", err)
	}

	if cfg.LocalMode() {
		return runUpdateLocal(cmd, args, cfg, opts)
	}
	return runUpdateRegistry(cmd, args, cfg, opts)
}

// runUpdateLocal handles update for local-repo mode: git pull + optional official cache refresh.
func runUpdateLocal(cmd *cobra.Command, args []string, cfg config.Config, opts updateOpts) error {
	out := cmd.OutOrStdout()
	ctx := cmd.Context()

	needsOfficial := opts.official
	if !needsOfficial && len(args) > 0 {
		catalog, catalogWarn := config.LoadCatalog("")
		if catalogWarn != nil {
			ui.Warn(out, fmt.Sprintf("user catalog: %v", catalogWarn))
		}
		names, namesErr := skillNamesFromArgs(args, catalog)
		if namesErr != nil {
			return namesErr
		}
		for _, name := range names {
			for _, b := range catalog {
				for _, ref := range b.Skills {
					if ref.Name == name && ref.Source == skill.SourceOfficial {
						needsOfficial = true
					}
				}
			}
		}
	}

	officialCacheDir := filepath.Join(cfg.OfficialCache, skill.SkillsFolderName)

	if opts.dryRun {
		fmt.Fprintln(out)
		ui.Header(out, "Dry run — would run:")
		fmt.Fprintf(out, "  git pull  in  %s\n", cfg.RepoPath)
		if needsOfficial {
			if _, statErr := os.Stat(officialCacheDir); os.IsNotExist(statErr) {
				fmt.Fprintf(out, "  git clone %s  →  %s\n", officialSkillsURL, officialCacheDir)
			} else {
				fmt.Fprintf(out, "  git pull  in  %s\n", officialCacheDir)
			}
		}
		fmt.Fprintln(out)
		return nil
	}

	if !ui.ConfirmYN(out, "Proceed?") {
		fmt.Fprintln(out, "Aborted.")
		return nil
	}
	fmt.Fprintln(out)

	ui.Info(out, fmt.Sprintf("Pulling %s …", cfg.RepoPath))
	if err := rskgit.Pull(ctx, cfg.RepoPath, out); err != nil {
		return fmt.Errorf("git pull (local repo): %w", err)
	}

	if needsOfficial {
		if err := refreshOfficialCache(cmd, officialCacheDir); err != nil {
			return err
		}
	}

	// If we're in a project and arg-targeted, also bump rsk.lock entries.
	if !opts.global && len(args) > 0 {
		if err := bumpLockEntries(cmd, args, cfg); err != nil {
			ui.Warn(out, fmt.Sprintf("update rsk.lock: %v", err))
		}
	}

	fmt.Fprintln(out)
	ui.Success(out, "Update complete.")
	ui.Info(out, "  Run 'rsk status' to see installed skill versions.")
	return nil
}

// bumpLockEntries re-resolves each named skill (expanding bundle args) and
// updates the project's rsk.lock entries. Used after a local-repo pull.
func bumpLockEntries(cmd *cobra.Command, args []string, cfg config.Config) error {
	ctx := cmd.Context()

	rskDir, err := manifest.ProjectFolderPath()
	if err != nil {
		return err
	}
	catalog, catalogWarn := config.LoadCatalog("")
	if catalogWarn != nil {
		ui.Warn(cmd.OutOrStdout(), fmt.Sprintf("user catalog: %v", catalogWarn))
	}
	names, err := skillNamesFromArgs(args, catalog)
	if err != nil {
		return err
	}

	localSrc := newLocalSource(cfg)
	officialSrc := source.NewOfficial(cfg.OfficialCache)

	m, err := manifest.ReadMod(rskDir)
	if err != nil {
		return err
	}
	targets := projectSkillsDirs(filepath.Dir(rskDir), m)

	lock, err := manifest.ReadLock(rskDir)
	if err != nil {
		return err
	}
	for _, name := range names {
		s, findErr := findSkillByName(ctx, name, localSrc, officialSrc)
		if findErr != nil {
			continue
		}
		for _, target := range targets {
			if linkErr := skill.Link(s, target); linkErr != nil {
				return linkErr
			}
		}
		lock = manifest.UpsertLockEntry(lock, manifest.LockEntry{
			Name:    s.Name,
			Version: s.Version,
			Source:  s.Source,
			Path:    s.Path,
		})
	}
	return manifest.WriteLock(rskDir, lock)
}

// refreshOfficialCache clones or pulls the official anthropics/skills cache.
func refreshOfficialCache(cmd *cobra.Command, dir string) error {
	out := cmd.OutOrStdout()
	ctx := cmd.Context()

	if _, statErr := os.Stat(dir); os.IsNotExist(statErr) {
		ui.Info(out, fmt.Sprintf("Cloning %s → %s …", officialSkillsURL, dir))
		if err := rskgit.Clone(ctx, officialSkillsURL, dir, out); err != nil {
			return fmt.Errorf("git clone (official cache): %w", err)
		}
		return nil
	}

	ui.Info(out, fmt.Sprintf("Pulling %s …", dir))
	if err := rskgit.Pull(ctx, dir, out); err != nil {
		return fmt.Errorf("git pull (official cache): %w", err)
	}
	return nil
}

// runUpdateRegistry handles update for registry mode: fetch the index, compare
// installed skills against latest versions, re-download and re-link any that changed.
func runUpdateRegistry(cmd *cobra.Command, args []string, cfg config.Config, opts updateOpts) error {
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()
	ctx := cmd.Context()

	targets, err := resolveTargetDirs(cfg, opts.global, opts.forTool)
	if err != nil {
		return err
	}

	reg := source.NewRegistry(cfg.RegistryURL, cfg.RegistryCache())

	ui.Info(out, fmt.Sprintf("Fetching index from %s …", cfg.RegistryURL))
	index, err := reg.Index(ctx)
	if err != nil {
		return fmt.Errorf("fetch registry index: %w", err)
	}

	// Determine which skill names to check (all installed, or just the args).
	var namesToCheck []string
	if len(args) > 0 {
		catalog, catalogWarn := config.LoadCatalog("")
		if catalogWarn != nil {
			ui.Warn(out, fmt.Sprintf("user catalog: %v", catalogWarn))
		}
		namesToCheck, err = skillNamesFromArgs(args, catalog)
		if err != nil {
			return err
		}
	} else {
		// Check all skills installed in any target dir.
		seen := make(map[string]bool)
		for _, target := range targets {
			entries, readErr := os.ReadDir(target)
			if readErr != nil {
				continue
			}
			for _, e := range entries {
				if !seen[e.Name()] {
					seen[e.Name()] = true
					namesToCheck = append(namesToCheck, e.Name())
				}
			}
		}
	}

	// Find skills with a newer version available.
	type updatePair struct {
		name      string
		installed string
		latest    string
		newSkill  skill.Skill
	}
	var toUpdate []updatePair

	for _, name := range namesToCheck {
		entry, ok := index[name]
		if !ok {
			continue
		}
		// Determine installed version from symlink target path.
		installedVer := installedVersionFromTargets(name, targets, cfg.RegistryCache())
		if installedVer == entry.Latest {
			continue
		}
		s, findErr := reg.FindVersion(ctx, name, entry.Latest)
		if findErr != nil {
			ui.Warn(out, fmt.Sprintf("skip %s: %v", name, findErr))
			continue
		}
		toUpdate = append(toUpdate, updatePair{
			name:      name,
			installed: installedVer,
			latest:    entry.Latest,
			newSkill:  s,
		})
	}

	if len(toUpdate) == 0 {
		ui.Info(out, "All installed skills are up to date.")
		return nil
	}

	fmt.Fprintln(out)
	ui.Header(out, "Skills to update:")
	nameWidth := 0
	for _, u := range toUpdate {
		if len(u.name) > nameWidth {
			nameWidth = len(u.name)
		}
	}
	for _, u := range toUpdate {
		from := u.installed
		if from == "" {
			from = "?"
		}
		fmt.Fprintf(out, "  %s  %s  %s  %s\n",
			ui.PadRight(ui.SkillName(u.name), nameWidth),
			ui.SkillVersion(from),
			ui.Arrow,
			ui.SkillVersion(u.latest),
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
	for _, u := range toUpdate {
		for _, target := range targets {
			if !skill.IsLinked(u.name, target) {
				continue
			}
			if linkErr := skill.Link(u.newSkill, target); linkErr != nil {
				ui.Failf(errOut, "re-link %s: %v", u.name, linkErr)
				failed++
			} else {
				ui.Success(out, fmt.Sprintf("%s  %s  %s  %s",
					ui.SkillName(u.name),
					ui.SkillVersion(u.installed),
					ui.Arrow,
					ui.SkillVersion(u.latest),
				))
			}
		}
	}

	if failed > 0 {
		return fmt.Errorf("%d skill(s) failed to update", failed)
	}

	fmt.Fprintln(out)
	ui.Success(out, "Update complete.")
	return nil
}

// installedVersionFromTargets reads the symlink target for name in each target dir
// and extracts the version segment from a registry cache path.
func installedVersionFromTargets(name string, targets []string, registryCacheDir string) string {
	for _, target := range targets {
		linkPath := filepath.Join(target, name)
		dest, err := os.Readlink(linkPath)
		if err != nil {
			continue
		}
		// Registry cache layout: <registryCacheDir>/<name>/<version>/
		prefix := filepath.Join(registryCacheDir, name) + string(filepath.Separator)
		if len(dest) > len(prefix) && dest[:len(prefix)] == prefix {
			rest := dest[len(prefix):]
			// rest is "<version>" or "<version>/<something>"
			for i, c := range rest {
				if c == filepath.Separator {
					return rest[:i]
				}
			}
			return rest
		}
	}
	return ""
}
