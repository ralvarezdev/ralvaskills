package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/termkit"

	"github.com/ralvarezdev/ralvaskills/v2/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/v2/internal/config"
	rskgit "github.com/ralvarezdev/ralvaskills/v2/internal/git"
	"github.com/ralvarezdev/ralvaskills/v2/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/v2/internal/skill"
	"github.com/ralvarezdev/ralvaskills/v2/internal/source"
	"github.com/ralvarezdev/ralvaskills/v2/internal/tool"
	"github.com/ralvarezdev/ralvaskills/v2/internal/ui"
	updatecheck "github.com/ralvarezdev/ralvaskills/v2/internal/update"
)

// updatePair is a skill with a newer version available in the registry.
type updatePair struct {
	name      string
	installed string
	latest    string
	newSkill  skill.Skill
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

Use --include official (legacy: --official) to also refresh the anthropics/skills cache.

Examples:
  rsk update [name]                # pull latest and re-link (all when no name)
  rsk update                       # local mode: git pull the local clone
  rsk update --include official    # also refresh the anthropics/skills cache
  rsk update grpc-architect        # update one skill
  rsk update docs                  # update everything in the docs bundle
  rsk update go-grpc --global`,
	RunE: func(cmd *cobra.Command, args []string) error {
		inc, err := cmdx.ReadIncludes(cmd, cmdx.IncludePersonal, cmdx.IncludeOfficial)
		if err != nil {
			return err
		}
		forTool, err := forToolFlag(cmd)
		if err != nil {
			return err
		}
		return runUpdate(cmd, updateOpts{
			global:   cmdx.Bool(cmd, cmdx.FlagGlobal),
			dryRun:   cmdx.Bool(cmd, cmdx.FlagDryRun),
			personal: inc.Personal,
			official: inc.Official,
			yes:      cmdx.Bool(cmd, cmdx.FlagYes),
			forTool:  forTool,
		}, args)
	},
}

type updateOpts struct {
	global, dryRun, personal, official, yes bool
	forTool                                 tool.ID
}

func runUpdate(cmd *cobra.Command, opts updateOpts, args []string) error {
	if !opts.global && opts.forTool != "" {
		return fieldError(cmdx.FlagFor, "--for requires --global")
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

	needsOfficial, err := needsOfficialCache(out, args, opts.official)
	if err != nil {
		return err
	}
	officialCacheDir := filepath.Join(cfg.OfficialCache, skill.SkillsFolderName)

	if opts.dryRun {
		printLocalUpdateDryRun(out, cfg, needsOfficial, officialCacheDir)
		return nil
	}

	if !termkit.Confirm(cmd, opts.yes, termkit.StdioIsTerminal(), "Proceed?") {
		fmt.Fprintln(out, "Aborted.")
		return nil
	}
	fmt.Fprintln(out)

	return applyLocalUpdate(cmd, args, cfg, opts, needsOfficial, officialCacheDir)
}

// needsOfficialCache reports whether the anthropics/skills cache also needs
// refreshing: explicitly via --official, or implicitly because any named arg
// resolves to an official-source skill.
func needsOfficialCache(out io.Writer, args []string, official bool) (bool, error) {
	if official || len(args) == 0 {
		return official, nil
	}

	catalog, catalogWarn := config.LoadCatalog("")
	if catalogWarn != nil {
		ui.Warn(out, fmt.Sprintf("user catalog: %v", catalogWarn))
	}
	names, err := skillNamesFromArgs(args, catalog)
	if err != nil {
		return false, err
	}
	for _, name := range names {
		for _, b := range catalog {
			for _, ref := range b.Skills {
				if ref.Name == name && ref.Source == skill.SourceOfficial {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// printLocalUpdateDryRun previews the git operations a real run would perform.
func printLocalUpdateDryRun(out io.Writer, cfg config.Config, needsOfficial bool, officialCacheDir string) {
	fmt.Fprintln(out)
	ui.Header(out, "Dry run — would run:")
	fmt.Fprintf(out, "  git pull  in  %s\n", cfg.RepoPath)
	if needsOfficial {
		if _, err := os.Stat(officialCacheDir); os.IsNotExist(err) {
			fmt.Fprintf(out, "  git clone %s  →  %s\n", source.OfficialSkillsURL, officialCacheDir)
		} else {
			fmt.Fprintf(out, "  git pull  in  %s\n", officialCacheDir)
		}
	}
	fmt.Fprintln(out)
}

// applyLocalUpdate pulls the local repo clone, refreshes the official cache
// if needed, and bumps the project lock — the non-dry-run body of
// runUpdateLocal, run only after the user has confirmed.
func applyLocalUpdate(
	cmd *cobra.Command, args []string, cfg config.Config, opts updateOpts, needsOfficial bool, officialCacheDir string,
) error {
	out := cmd.OutOrStdout()
	ctx := cmd.Context()

	ui.Info(out, fmt.Sprintf("Pulling %s …", cfg.RepoPath))
	if err := rskgit.Pull(ctx, cfg.RepoPath, out); err != nil {
		return fmt.Errorf("git pull (local repo): %w", err)
	}

	// The clone is now current with its remote — refresh the passive
	// update-availability cache so the next bare `rsk` invocation reflects
	// that immediately rather than waiting on the background worker.
	if err := updatecheck.Save(cfg, updatecheck.Result{Mode: updatecheck.ModeLocal}); err != nil {
		ui.Warn(out, fmt.Sprintf("update-check cache: %v", err))
	}

	if needsOfficial {
		if err := refreshOfficialCache(cmd, officialCacheDir); err != nil {
			return err
		}
	}

	// If we're in a project, bump rsk.lock entries for any changed skills.
	if !opts.global {
		if err := bumpLockForProject(cmd, args, cfg); err != nil {
			ui.Warn(out, fmt.Sprintf("update rsk.lock: %v", err))
		}
	}

	fmt.Fprintln(out)
	ui.Success(out, "Update complete.")
	ui.Info(out, "  Run 'rsk status' to see installed skill versions.")
	return nil
}

// tryRskDir returns the project .rsk directory. Second return is false when
// the caller is not inside a project — no error, just nothing to do.
func tryRskDir() (string, bool) {
	d, err := manifest.ProjectFolderPath()
	return d, err == nil
}

// tryReadMod reads rsk.mod from rskDir. Second return is false on any read
// error (no mod means nothing to bump).
func tryReadMod(rskDir string) (manifest.Mod, bool) {
	m, err := manifest.ReadMod(rskDir)
	return m, err == nil
}

// bumpLockForProject re-resolves skills and updates the project's rsk.lock.
// When args is empty, all skills in rsk.mod are refreshed. No-op if no project.
func bumpLockForProject(cmd *cobra.Command, args []string, cfg config.Config) error {
	ctx := cmd.Context()

	rskDir, ok := tryRskDir()
	if !ok {
		return nil
	}

	catalog, warn := config.LoadCatalog("")
	if warn != nil {
		ui.Warn(cmd.OutOrStdout(), fmt.Sprintf("user catalog: %v", warn))
	}

	names, err := skillNamesFromArgs(args, catalog)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		m, modOK := tryReadMod(rskDir)
		if !modOK {
			return nil
		}
		for name := range m.Skills {
			names = append(names, name)
		}
		if len(names) == 0 {
			return nil
		}
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
		ui.Info(out, fmt.Sprintf("Cloning %s → %s …", source.OfficialSkillsURL, dir))
		if err := rskgit.Clone(ctx, source.OfficialSkillsURL, dir, out); err != nil {
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

// runOfficialCacheRefresh handles --official for registry mode: it clones or
// pulls the anthropics/skills cache independently of the registry index
// check below, respecting --dry-run and prompting for confirmation before
// touching the filesystem.
func runOfficialCacheRefresh(cmd *cobra.Command, cfg config.Config, dryRun, yes bool) error {
	out := cmd.OutOrStdout()
	dir := filepath.Join(cfg.OfficialCache, skill.SkillsFolderName)

	if dryRun {
		fmt.Fprintln(out)
		ui.Header(out, "Dry run — would run:")
		if _, statErr := os.Stat(dir); os.IsNotExist(statErr) {
			fmt.Fprintf(out, "  git clone %s  →  %s\n", source.OfficialSkillsURL, dir)
		} else {
			fmt.Fprintf(out, "  git pull  in  %s\n", dir)
		}
		fmt.Fprintln(out)
		return nil
	}

	if !termkit.Confirm(cmd, yes, termkit.StdioIsTerminal(), "Refresh anthropics/skills cache?") {
		fmt.Fprintln(out, "Skipped official cache refresh.")
		return nil
	}
	fmt.Fprintln(out)
	return refreshOfficialCache(cmd, dir)
}

// runUpdateRegistry handles update for registry mode: fetch the index, compare
// installed skills against latest versions, re-download and re-link any that changed.
func runUpdateRegistry(cmd *cobra.Command, args []string, cfg config.Config, opts updateOpts) error {
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()
	ctx := cmd.Context()

	if opts.official {
		if err := runOfficialCacheRefresh(cmd, cfg, opts.dryRun, opts.yes); err != nil {
			return err
		}
	}

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

	namesToCheck, err := registryNamesToCheck(out, args, targets)
	if err != nil {
		return err
	}

	toUpdate := findOutdatedSkills(ctx, out, reg, index, namesToCheck, targets, cfg.RegistryCache())
	if len(toUpdate) == 0 {
		saveUpdateCache(out, cfg, nil)
		ui.Info(out, "All installed skills are up to date.")
		return nil
	}

	printRegistryUpdatePlan(out, toUpdate)

	if opts.dryRun {
		saveUpdateCache(out, cfg, updatePairNames(toUpdate))
		return nil
	}
	if !termkit.Confirm(cmd, opts.yes, termkit.StdioIsTerminal(), "Proceed?") {
		fmt.Fprintln(out, "Aborted.")
		saveUpdateCache(out, cfg, updatePairNames(toUpdate))
		return nil
	}
	fmt.Fprintln(out)

	failed, failedNames := relinkOutdated(out, errOut, toUpdate, targets)

	// Refresh the passive update-availability cache with the actual outcome —
	// after this point only skills that failed to re-link are still
	// outdated — so the next bare `rsk` invocation reflects the truth
	// immediately rather than waiting on the background worker.
	saveUpdateCache(out, cfg, failedNames)

	if failed > 0 {
		return fmt.Errorf("%d skill(s) failed to update", failed)
	}

	// Bump project lock for updated skills (no-op if not in a project).
	if !opts.global {
		if lockErr := bumpLockForProject(cmd, args, cfg); lockErr != nil {
			ui.Warn(out, fmt.Sprintf("update rsk.lock: %v", lockErr))
		}
	}

	fmt.Fprintln(out)
	ui.Success(out, "Update complete.")
	return nil
}

// registryNamesToCheck resolves args to skill names when given, or otherwise
// collects every skill name currently linked in any target dir.
func registryNamesToCheck(out io.Writer, args, targets []string) ([]string, error) {
	if len(args) > 0 {
		catalog, catalogWarn := config.LoadCatalog("")
		if catalogWarn != nil {
			ui.Warn(out, fmt.Sprintf("user catalog: %v", catalogWarn))
		}
		return skillNamesFromArgs(args, catalog)
	}

	seen := make(map[string]bool)
	var names []string
	for _, target := range targets {
		entries, err := os.ReadDir(target)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !seen[e.Name()] {
				seen[e.Name()] = true
				names = append(names, e.Name())
			}
		}
	}
	return names, nil
}

// findOutdatedSkills checks each name in namesToCheck against index, skipping
// any that aren't in the registry, aren't actually behind, or fail to
// resolve at their latest version (warned, not fatal).
func findOutdatedSkills(
	ctx context.Context, out io.Writer, reg *source.Registry, index map[string]*source.IndexEntry,
	namesToCheck, targets []string, registryCache string,
) []updatePair {
	var toUpdate []updatePair
	for _, name := range namesToCheck {
		entry, ok := index[name]
		if !ok {
			continue
		}
		// Determine installed version from symlink target path.
		installedVer := updatecheck.InstalledVersionFromTargets(name, targets, registryCache)
		if installedVer == entry.Latest {
			continue
		}
		s, err := reg.FindVersion(ctx, name, entry.Latest)
		if err != nil {
			ui.Warn(out, fmt.Sprintf("skip %s: %v", name, err))
			continue
		}
		toUpdate = append(toUpdate, updatePair{
			name:      name,
			installed: installedVer,
			latest:    entry.Latest,
			newSkill:  s,
		})
	}
	return toUpdate
}

// printRegistryUpdatePlan renders the name/installed/latest table of skills
// about to be updated.
func printRegistryUpdatePlan(out io.Writer, toUpdate []updatePair) {
	fmt.Fprintln(out)
	ui.Header(out, "Skills to update:")
	termkit.WriteTableStyled(out, []string{headerName, "From", "", "To"},
		termkit.Rows(toUpdate, updatePlanTable), false, nil, termkit.TableBorderless, ui.Theme)
	fmt.Fprintln(out)
}

// updatePlanTable projects one update-plan row.
var updatePlanTable = termkit.Table[updatePair]{
	Row: func(u updatePair) []any {
		from := u.installed
		if from == "" {
			from = "?"
		}
		return []any{ui.SkillName(u.name), ui.SkillVersion(from), ui.Arrow, ui.SkillVersion(u.latest)}
	},
}

// relinkOutdated re-links each outdated skill wherever it's currently
// linked, reporting (rather than aborting on) a failure so the rest of the
// batch still updates. It returns how many failed and their names.
func relinkOutdated(out, errOut io.Writer, toUpdate []updatePair, targets []string) (failed int, failedNames []string) {
	for _, u := range toUpdate {
		skillFailed := false
		for _, target := range targets {
			if !skill.IsLinked(u.name, target) {
				continue
			}
			if err := skill.Link(u.newSkill, target); err != nil {
				ui.Failf(errOut, "re-link %s: %v", u.name, err)
				failed++
				skillFailed = true
			} else {
				ui.Success(out, fmt.Sprintf("%s  %s  %s  %s",
					ui.SkillName(u.name),
					ui.SkillVersion(u.installed),
					ui.Arrow,
					ui.SkillVersion(u.latest),
				))
			}
		}
		if skillFailed {
			failedNames = append(failedNames, u.name)
		}
	}
	return failed, failedNames
}

// updatePairNames extracts the skill names from toUpdate, for caching the
// update-availability state when nothing was actually applied (dry-run or
// an aborted confirm).
func updatePairNames(toUpdate []updatePair) []string {
	names := make([]string, len(toUpdate))
	for i, u := range toUpdate {
		names[i] = u.name
	}
	return names
}

// saveUpdateCache refreshes the passive update-availability cache with
// outdated, warning (not failing) on error since this is best-effort.
func saveUpdateCache(out io.Writer, cfg config.Config, outdated []string) {
	result := updatecheck.Result{Mode: updatecheck.ModeRegistry, Outdated: outdated}
	if cacheErr := updatecheck.Save(cfg, result); cacheErr != nil {
		ui.Warn(out, fmt.Sprintf("update-check cache: %v", cacheErr))
	}
}
