package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ralvarezdev/ralvaskills/cli/internal/config"
	"github.com/ralvarezdev/ralvaskills/cli/internal/skill"
	"github.com/ralvarezdev/ralvaskills/cli/internal/source"
	"github.com/ralvarezdev/ralvaskills/cli/internal/ui"
	"github.com/spf13/cobra"
)

const officialSkillsURL = "https://github.com/anthropics/skills"

var updateCmd = &cobra.Command{
	Use:   "update [bundle...] [flags]",
	Short: "Pull the latest skills and re-symlink.",
	Long: `Update installed skills to their latest versions.

In local-repo mode: runs git pull on the ralvaskills clone (symlinks update automatically).
In registry mode: fetches the latest index and re-downloads any skills with new versions.

Use --official to also refresh the anthropics/skills cache.

Examples:
  rsk update
  rsk update --official
  rsk update docs
  rsk update --skill grpc-architect`,
	RunE: runUpdate,
}

var (
	updateGlobal   bool
	updateFor      string
	updateSkill    string
	updateOfficial bool
	updatePersonal bool
	updateDryRun   bool
)

func init() {
	rootCmd.AddCommand(updateCmd)
	f := updateCmd.Flags()
	f.BoolVar(&updateGlobal, "global", false, "Target global skills dir(s)")
	f.StringVar(&updateFor, "for", "", "Scope --global to a single tool (claude-code|opencode)")
	f.StringVar(&updateSkill, "skill", "", "Update a single skill by name")
	f.BoolVar(&updateOfficial, "official", false, "Also re-fetch the anthropics/skills cache")
	f.BoolVar(&updatePersonal, "personal", false, "Include personal/ skills in update")
	f.BoolVar(&updateDryRun, "dry-run", false, "Show what would change without applying it")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if updateSkill != "" && len(args) > 0 {
		return fmt.Errorf("use either positional bundle names or --skill <name>, not both")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n  Run 'rsk init' to set up rsk on this machine", err)
	}

	if cfg.LocalMode() {
		return runUpdateLocal(cmd, args, cfg)
	}
	return runUpdateRegistry(cmd, args, cfg)
}

// runUpdateLocal handles update for local-repo mode: git pull + optional official cache refresh.
func runUpdateLocal(cmd *cobra.Command, args []string, cfg config.Config) error {
	out := cmd.OutOrStdout()

	needsOfficial := updateOfficial
	if !needsOfficial && (len(args) > 0 || updateSkill != "") {
		catalog := config.LoadCatalog("")
		names, namesErr := skillNamesFromArgs(args, updateSkill, catalog)
		if namesErr != nil {
			return namesErr
		}
		for _, name := range names {
			for _, b := range catalog {
				for _, ref := range b.Skills {
					if ref.Name == name && ref.Source == config.SourceOfficial {
						needsOfficial = true
					}
				}
			}
		}
	}

	officialCacheDir := filepath.Join(cfg.OfficialCache, "skills")

	if updateDryRun {
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
	if err := gitPull(cfg.RepoPath, out); err != nil {
		return fmt.Errorf("git pull (local repo): %w", err)
	}

	if needsOfficial {
		if _, statErr := os.Stat(officialCacheDir); os.IsNotExist(statErr) {
			ui.Info(out, fmt.Sprintf("Cloning %s → %s …", officialSkillsURL, officialCacheDir))
			if err := gitClone(officialSkillsURL, officialCacheDir, out); err != nil {
				return fmt.Errorf("git clone (official cache): %w", err)
			}
		} else {
			ui.Info(out, fmt.Sprintf("Pulling %s …", officialCacheDir))
			if err := gitPull(officialCacheDir, out); err != nil {
				return fmt.Errorf("git pull (official cache): %w", err)
			}
		}
	}

	fmt.Fprintln(out)
	ui.Success(out, "Update complete.")
	ui.Info(out, "  Run 'rsk status' to see installed skill versions.")
	return nil
}

// runUpdateRegistry handles update for registry mode: fetch the index, compare
// installed skills against latest versions, re-download and re-link any that changed.
func runUpdateRegistry(cmd *cobra.Command, args []string, cfg config.Config) error {
	out := cmd.OutOrStdout()

	targets, err := resolveTargetDirs(cfg, updateGlobal, updateFor)
	if err != nil {
		return err
	}

	reg := source.NewRegistry(cfg.RegistryURL, cfg.RegistryCache())

	// Fetch the registry index to find latest versions.
	ui.Info(out, fmt.Sprintf("Fetching index from %s …", cfg.RegistryURL))
	index, err := reg.Index()
	if err != nil {
		return fmt.Errorf("fetch registry index: %w", err)
	}

	// Determine which skill names to check (all installed, or just the args).
	var namesToCheck []string
	if updateSkill != "" || len(args) > 0 {
		catalog := config.LoadCatalog("")
		namesToCheck, err = skillNamesFromArgs(args, updateSkill, catalog)
		if err != nil {
			return err
		}
	} else {
		// Check all skills installed in any target dir.
		seen := make(map[string]bool)
		for _, target := range targets {
			entries, _ := os.ReadDir(target)
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
		name       string
		installed  string
		latest     string
		newSkill   skill.Skill
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
		s, findErr := reg.FindVersion(name, entry.Latest)
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
		fmt.Fprintf(out, "  %s  %s → %s\n",
			ui.PadRight(u.name, nameWidth), from, u.latest)
	}
	fmt.Fprintln(out)

	if updateDryRun {
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
				ui.Failf("re-link %s: %v", u.name, linkErr)
				failed++
			} else {
				ui.Success(out, fmt.Sprintf("%s  %s → %s", u.name, u.installed, u.latest))
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

func gitPull(dir string, out interface{ Write([]byte) (int, error) }) error {
	c := exec.Command("git", "-C", dir, "pull")
	c.Stdout = out
	c.Stderr = os.Stderr
	return c.Run()
}

func gitClone(url, dest string, out interface{ Write([]byte) (int, error) }) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o750); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}
	c := exec.Command("git", "clone", url, dest)
	c.Stdout = out
	c.Stderr = os.Stderr
	return c.Run()
}
