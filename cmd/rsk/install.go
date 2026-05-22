package main

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/ralvarezdev/ralvaskills/internal"
	"github.com/ralvarezdev/ralvaskills/internal/config"
	"github.com/ralvarezdev/ralvaskills/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/internal/skill"
	"github.com/ralvarezdev/ralvaskills/internal/source"
	"github.com/ralvarezdev/ralvaskills/internal/ui"
	"github.com/spf13/cobra"
)

type installOpts struct {
	global, dryRun, personal bool
	forTool, skill, version  string
}

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
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInstall(cmd, installOpts{
			global:   internal.FlagBool(cmd, internal.FlagGlobal),
			dryRun:   internal.FlagBool(cmd, internal.FlagDryRun),
			personal: internal.FlagBool(cmd, internal.FlagPersonal),
			forTool:  internal.FlagString(cmd, internal.FlagFor),
			skill:    internal.FlagString(cmd, internal.FlagSkill),
			version:  internal.FlagString(cmd, internal.FlagVersion),
		}, args)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	f := installCmd.Flags()
	f.Bool(internal.FlagGlobal, false, "Install to global skills dir(s)")
	f.String(internal.FlagFor, "", "Scope --global to a single tool (claude-code|opencode)")
	f.String(internal.FlagSkill, "", "Install a single skill by name (skips bundle resolution)")
	f.Bool(internal.FlagPersonal, false, "Allow installing personal/ skills")
	f.String(internal.FlagVersion, "", "Pin to a specific repo tag (local skills only)")
	f.Bool(internal.FlagDryRun, false, "Show what would be installed without doing it")
}

func runInstall(cmd *cobra.Command, opts installOpts, args []string) error {
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()
	ctx := cmd.Context()

	// Validate flag combinations before any dispatch.
	if opts.skill != "" && len(args) > 0 {
		return fmt.Errorf("use either positional bundle names or --skill <name>, not both")
	}
	if !opts.global && opts.forTool != "" {
		return fmt.Errorf("--for requires --global")
	}
	if opts.version != "" {
		return fmt.Errorf("--version is not yet supported; skills are symlinked from the local repo HEAD")
	}

	// No explicit targets: install from rsk.mod in the current project.
	if opts.skill == "" && len(args) == 0 && !opts.global {
		return runInstallFromMod(cmd, opts)
	}

	if opts.skill == "" && len(args) == 0 && opts.global {
		return fmt.Errorf("specify at least one bundle name or use --skill <name>")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n  Run 'rsk init' to set up rsk on this machine", err)
	}

	targets, err := resolveTargetDirs(cfg, opts.global, opts.forTool)
	if err != nil {
		return err
	}

	localSrc := newLocalSource(cfg)
	officialSrc := source.NewOfficial(cfg.OfficialCache)

	var skills []skill.Skill
	var warnings []string

	if opts.skill != "" {
		s, findErr := findSkillByName(ctx, opts.skill, localSrc, officialSrc)
		if findErr != nil {
			return findErr
		}
		skills = append(skills, s)
	} else {
		catalog, catalogWarn := config.LoadCatalog("")
		if catalogWarn != nil {
			ui.Warn(out, fmt.Sprintf("user catalog: %v", catalogWarn))
		}
		for _, bundleName := range args {
			bundle, ok := config.FindBundle(catalog, bundleName)
			if !ok {
				return fmt.Errorf("bundle %q not found — run 'rsk list' to see available bundles", bundleName)
			}
			ss, ws, resolveErr := resolveBundleSkills(ctx, bundle, localSrc, officialSrc)
			if resolveErr != nil {
				return resolveErr
			}
			skills = append(skills, ss...)
			warnings = append(warnings, ws...)
		}
	}

	if !opts.personal {
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
	if opts.dryRun {
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
				suffix = "  " + ui.ReLink
			}
			fmt.Fprintf(out, "  %s  %s  %s  %s  %s%s\n",
				ui.SourceLabel(s.Source),
				ui.PadRight(ui.SkillName(s.Name), nameWidth),
				ui.PadRight(ui.SkillVersion(s.Version), 7),
				ui.Arrow,
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

	if opts.dryRun {
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
				ui.Failf(errOut, "link %s: %v", s.Name, linkErr)
				failed++
			} else {
				ui.Success(out, fmt.Sprintf("%s  %s  %s",
					ui.SkillName(s.Name),
					ui.Arrow,
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

// runInstallFromMod reads rsk.mod in the current project, resolves all skills,
// symlinks them into .rsk/skills/, updates rsk.lock, and rewrites .rsk/CLAUDE.md.
func runInstallFromMod(cmd *cobra.Command, opts installOpts) error {
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()
	ctx := cmd.Context()

	rskDir, err := manifest.LocalConfigFolderPath()
	if err != nil {
		return err
	}

	m, err := manifest.ReadMod(rskDir)
	if err != nil {
		return err
	}

	if len(m.Skills) == 0 {
		ui.Warn(out, "rsk.mod has no skills. Run 'rsk skill add <name>' to add one.")
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n  Run 'rsk init' to set up rsk on this machine", err)
	}

	localSrc := newLocalSource(cfg)
	officialSrc := source.NewOfficial(cfg.OfficialCache)

	skillsDir := manifest.LocalConfigSkillsPath(rskDir)

	sortedNames := make([]string, 0, len(m.Skills))
	for name := range m.Skills {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	var skills []skill.Skill
	var warnings []string
	for _, name := range sortedNames {
		s, findErr := findSkillByName(ctx, name, localSrc, officialSrc)
		if findErr != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v — skipped", name, findErr))
			continue
		}
		skills = append(skills, s)
	}

	fmt.Fprintln(out)
	ui.Header(out, "Skills to install:")

	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Name
	}
	nameWidth := ui.MaxWidth(names)

	for _, s := range skills {
		suffix := ""
		if skill.IsLinked(s.Name, skillsDir) {
			suffix = "  " + ui.ReLink
		}
		fmt.Fprintf(out, "  %s  %s  %s  %s  %s%s\n",
			ui.SourceLabel(s.Source),
			ui.PadRight(ui.SkillName(s.Name), nameWidth),
			ui.PadRight(ui.SkillVersion(s.Version), 7),
			ui.Arrow,
			ui.MutedPath(filepath.Join(skillsDir, s.Name)),
			suffix,
		)
	}

	for _, w := range warnings {
		fmt.Fprintln(out)
		ui.Warn(out, w)
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

	lock, err := manifest.ReadLock(rskDir)
	if err != nil {
		return err
	}

	var failed int
	for _, s := range skills {
		if linkErr := skill.Link(s, skillsDir); linkErr != nil {
			ui.Failf(errOut, "link %s: %v", s.Name, linkErr)
			failed++
			continue
		}
		lock = manifest.UpsertLockEntry(lock, manifest.LockEntry{
			Name:    s.Name,
			Version: s.Version,
			Source:  s.Source,
			Path:    s.Path,
		})
		ui.Success(out, fmt.Sprintf("%s  %s  %s",
			ui.SkillName(s.Name),
			ui.Arrow,
			ui.MutedPath(filepath.Join(skillsDir, s.Name)),
		))
	}

	if failed > 0 {
		return fmt.Errorf("%d skill(s) failed to install", failed)
	}

	if err := manifest.WriteLock(rskDir, lock); err != nil {
		return err
	}
	if err := syncPinnedAllTools(rskDir, m); err != nil {
		return err
	}

	fmt.Fprintln(out)
	ui.Info(out, "Run 'rsk status' to verify installed skills.")
	return nil
}
