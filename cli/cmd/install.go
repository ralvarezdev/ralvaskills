package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/ralvarezdev/ralvaskills/cli/internal/config"
	"github.com/ralvarezdev/ralvaskills/cli/internal/skill"
	"github.com/ralvarezdev/ralvaskills/cli/internal/source"
	"github.com/ralvarezdev/ralvaskills/cli/internal/ui"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [bundle...] [flags]",
	Short: "Install a bundle or skill.",
	Long: `Install one or more bundles or a single skill by name.

By default installs to ./.claude/skills/ (project-local).
Use --global to install to the configured global skills directories.

Examples:
  rsk install go-grpc
  rsk install design docs
  rsk install global --global
  rsk install --skill go-architect
  rsk install go-grpc --global --for claude-code
  rsk install --skill demo-script-architect --personal
  rsk install go-grpc --dry-run`,
	RunE: runInstall,
}

var (
	installGlobal   bool
	installFor      string
	installSkill    string
	installPersonal bool
	installVersion  string
	installDryRun   bool
)

func init() {
	rootCmd.AddCommand(installCmd)
	f := installCmd.Flags()
	f.BoolVar(&installGlobal, "global", false, "Install to global skills dir(s)")
	f.StringVar(&installFor, "for", "", "Scope --global to a single tool (claude-code|opencode)")
	f.StringVar(&installSkill, "skill", "", "Install a single skill by name (skips bundle resolution)")
	f.BoolVar(&installPersonal, "personal", false, "Allow installing personal/ skills")
	f.StringVar(&installVersion, "version", "", "Pin to a specific repo tag (local skills only)")
	f.BoolVar(&installDryRun, "dry-run", false, "Show what would be installed without doing it")
}

func runInstall(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	if installSkill != "" && len(args) > 0 {
		return fmt.Errorf("use either positional bundle names or --skill <name>, not both")
	}
	if installSkill == "" && len(args) == 0 {
		return fmt.Errorf("specify at least one bundle name or use --skill <name>")
	}
	if !installGlobal && installFor != "" {
		return fmt.Errorf("--for requires --global")
	}
	if installVersion != "" {
		return fmt.Errorf("--version is not yet supported; skills are symlinked from the local repo HEAD")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n  Run 'rsk init' to set up rsk on this machine", err)
	}

	targets, err := resolveTargetDirs(cfg, installGlobal, installFor)
	if err != nil {
		return err
	}

	localSrc := newLocalSource(cfg)
	officialSrc := source.NewOfficial(cfg.OfficialCache)

	var skills []skill.Skill
	var warnings []string

	if installSkill != "" {
		s, findErr := findSkillByName(installSkill, localSrc, officialSrc)
		if findErr != nil {
			return findErr
		}
		skills = append(skills, s)
	} else {
		catalog := config.LoadCatalog("")
		for _, bundleName := range args {
			bundle, ok := config.FindBundle(catalog, bundleName)
			if !ok {
				return fmt.Errorf("bundle %q not found — run 'rsk list' to see available bundles", bundleName)
			}
			ss, ws, resolveErr := resolveBundleSkills(bundle, localSrc, officialSrc)
			if resolveErr != nil {
				return resolveErr
			}
			skills = append(skills, ss...)
			warnings = append(warnings, ws...)
		}
	}

	if !installPersonal {
		for _, s := range skills {
			if s.IsPersonal {
				return fmt.Errorf("skill %q is in personal/ — pass --personal to install it", s.Name)
			}
		}
	}

	skills = dedupSkills(skills)

	if len(skills) == 0 {
		ui.Warn(out, "nothing to install")
		for _, w := range warnings {
			ui.Warn(out, w)
		}
		return nil
	}

	fmt.Fprintln(out)
	if installDryRun {
		ui.Header(out, "Dry run — would install:")
	} else {
		ui.Header(out, "Skills to install:")
	}

	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Name
	}
	nameWidth := ui.MaxWidth(names)

	for _, s := range skills {
		for _, target := range targets {
			suffix := ""
			if skill.IsLinked(s.Name, target) {
				suffix = "  " + ui.ReLink()
			}
			fmt.Fprintf(out, "  %s  %s  %s  %s  %s%s\n",
				ui.SourceLabel(s.Source),
				ui.PadRight(ui.SkillName(s.Name), nameWidth),
				ui.PadRight(ui.SkillVersion(s.Version), 7),
				ui.Arrow(),
				ui.MutedPath(filepath.Join(target, s.Name)),
				suffix,
			)
		}
	}

	for _, w := range warnings {
		fmt.Fprintln(out)
		ui.Warn(out, w)
	}
	fmt.Fprintln(out)

	if installDryRun {
		return nil
	}

	if !ui.ConfirmYN(out, "Proceed?") {
		fmt.Fprintln(out, "Aborted.")
		return nil
	}
	fmt.Fprintln(out)

	var failed int
	for _, s := range skills {
		for _, target := range targets {
			if linkErr := skill.Link(s, target); linkErr != nil {
				ui.Failf("link %s: %v", s.Name, linkErr)
				failed++
			} else {
				ui.Success(out, fmt.Sprintf("%s  %s  %s",
					ui.SkillName(s.Name),
					ui.Arrow(),
					ui.MutedPath(filepath.Join(target, s.Name)),
				))
			}
		}
	}

	if failed > 0 {
		return fmt.Errorf("%d skill(s) failed to install", failed)
	}

	fmt.Fprintln(out)
	ui.Info(out, "Run 'rsk status' to verify installed skills.")
	return nil
}
