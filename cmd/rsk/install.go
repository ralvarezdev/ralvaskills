package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/ralvaskills/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/internal/config"
	"github.com/ralvarezdev/ralvaskills/internal/fsperm"
	"github.com/ralvarezdev/ralvaskills/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/internal/skill"
	"github.com/ralvarezdev/ralvaskills/internal/source"
	"github.com/ralvarezdev/ralvaskills/internal/ui"
)

var installCmd = &cobra.Command{
	Use:   "install [name...] [flags]",
	Short: "Install bundles or skills.",
	Long: `Install one or more bundles or skills by name.

Names are resolved against the catalog: if a name matches a bundle, the bundle
expands to its skills; otherwise the name is treated as a single skill.

By default installs to the current project (.rsk/skills/) and writes entries
into .rsk/rsk.mod and .rsk/rsk.lock.

With --global the install targets the configured global skills directories
without touching any project manifest.

Examples:
  rsk install                                  # project: install everything in rsk.mod
  rsk install go-grpc                          # project: bundle + manifest track
  rsk install go-architect --pin               # project: install + pin in CLAUDE.md
  rsk install global --global                  # global: bundle to every tool dir
  rsk install go-grpc --global --for claude-code
  rsk install demo-script-architect --personal
  rsk install go-grpc --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInstall(cmd, installOpts{
			global:   cmdx.Bool(cmd, cmdx.FlagGlobal),
			dryRun:   cmdx.Bool(cmd, cmdx.FlagDryRun),
			personal: cmdx.Bool(cmd, cmdx.FlagPersonal),
			pin:      cmdx.Bool(cmd, cmdx.FlagPin),
			forTool:  cmdx.String(cmd, cmdx.FlagFor),
			version:  cmdx.String(cmd, cmdx.FlagVersion),
		}, args)
	},
}

type installOpts struct {
	global, dryRun, personal, pin bool
	forTool, version              string
}

func runInstall(cmd *cobra.Command, opts installOpts, args []string) error {
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()
	ctx := cmd.Context()

	if !opts.global && opts.forTool != "" {
		return errors.New("--for requires --global")
	}
	if opts.global && opts.pin {
		return errors.New("--pin only applies to project installs (drop --global or --pin)")
	}
	if opts.version != "" {
		return errors.New("--version is not yet supported; skills are symlinked from the local repo HEAD")
	}

	// Project-bundle install with no args: read rsk.mod and install everything.
	if !opts.global && len(args) == 0 {
		return runInstallFromMod(cmd, opts)
	}

	if opts.global && len(args) == 0 {
		name, err := ui.Ask(out, "Skill or bundle name", "")
		if err != nil {
			return err
		}
		if name == "" {
			return errors.New("specify at least one bundle or skill name")
		}
		args = []string{name}
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n  Run 'rsk init' to set up rsk on this machine", err)
	}

	localSrc := newLocalSource(cfg)
	officialSrc := source.NewOfficial(cfg.OfficialCache)

	catalog, catalogWarn := config.LoadCatalog("")
	if catalogWarn != nil {
		ui.Warn(out, fmt.Sprintf("user catalog: %v", catalogWarn))
	}

	skills, warnings, err := resolveNames(ctx, args, catalog, localSrc, officialSrc)
	if err != nil {
		return err
	}
	skills = dedupSkills(skills)

	if !opts.personal {
		for _, s := range skills {
			if s.IsPersonal {
				return fmt.Errorf("skill %q is in personal/ — pass --personal to install it", s.Name)
			}
		}
	}

	if len(skills) == 0 {
		ui.Warn(out, "nothing to install")
		for _, w := range warnings {
			ui.Warn(out, w)
		}
		return nil
	}

	if opts.global {
		return runInstallGlobal(cmd, cfg, opts, skills, warnings)
	}
	return runInstallProject(cmd, args, opts, skills, warnings, errOut)
}

// runInstallGlobal symlinks each resolved skill into the configured global
// skills dir(s). No manifest is touched.
func runInstallGlobal(
	cmd *cobra.Command, cfg config.Config, opts installOpts,
	skills []skill.Skill, warnings []string,
) error {
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()

	targets, err := resolveTargetDirs(cfg, true, opts.forTool)
	if err != nil {
		return err
	}

	printInstallPreview(out, skills, targets, warnings, opts.dryRun)
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

// runInstallProject links each resolved skill into each configured tool's
// project skills directory and writes entries into rsk.mod and rsk.lock. If
// --pin is set, the top-level argument names are pinned for every configured
// tool.
func runInstallProject(
	cmd *cobra.Command, args []string, opts installOpts,
	skills []skill.Skill, warnings []string, errOut io.Writer,
) error {
	out := cmd.OutOrStdout()

	rskDir, err := manifest.ProjectFolderPath()
	if err != nil {
		return err
	}
	projectRoot := filepath.Dir(rskDir)

	m, err := manifest.ReadMod(rskDir)
	if err != nil {
		return err
	}
	targets := projectSkillsDirs(projectRoot, m)

	printInstallPreview(out, skills, targets, warnings, opts.dryRun)
	if opts.dryRun {
		return nil
	}
	if !ui.ConfirmYN(out, "Proceed?") {
		fmt.Fprintln(out, "Aborted.")
		return nil
	}
	fmt.Fprintln(out)

	for _, target := range targets {
		if mkdirErr := os.MkdirAll(target, fsperm.Dir); mkdirErr != nil {
			return fmt.Errorf("create %s: %w", target, mkdirErr)
		}
	}

	lock, err := manifest.ReadLock(rskDir)
	if err != nil {
		return err
	}

	// Per-argument version constraints, parsed from any "name@version" arg.
	argConstraints, err := constraintsFromArgs(args)
	if err != nil {
		return err
	}

	var failed int
	var installedNames []string
	for _, s := range skills {
		linkFailed := false
		for _, target := range targets {
			if linkErr := skill.Link(s, target); linkErr != nil {
				ui.Failf(errOut, "link %s → %s: %v", s.Name, target, linkErr)
				linkFailed = true
			}
		}
		if linkFailed {
			failed++
			continue
		}
		constraint, ok := argConstraints[s.Name]
		if !ok || constraint == "" {
			constraint = "*"
		}
		m.Skills[s.Name] = constraint
		installedNames = append(installedNames, s.Name)
		lock = manifest.UpsertLockEntry(lock, manifest.LockEntry{
			Name:    s.Name,
			Version: s.Version,
			Source:  s.Source,
			Path:    s.Path,
		})
		ui.Success(out, fmt.Sprintf("%s  %s  %s",
			ui.SkillName(s.Name),
			ui.Arrow,
			ui.MutedPath(filepath.Join(targets[0], s.Name)),
		))
	}

	if failed > 0 {
		return fmt.Errorf("%d skill(s) failed to install", failed)
	}

	// --pin pins every skill resolved from args (bundles expand to all their
	// skills), not just names that happen to match a top-level arg.
	if opts.pin {
		for _, name := range installedNames {
			if !slices.Contains(m.Pinned, name) {
				m.Pinned = append(m.Pinned, name)
			}
		}
	}

	if err = manifest.WriteMod(rskDir, m); err != nil {
		return err
	}
	if err = manifest.WriteLock(rskDir, lock); err != nil {
		return err
	}
	if err = syncPinnedAllTools(rskDir, m); err != nil {
		return err
	}

	fmt.Fprintln(out)
	ui.Info(out, "Run 'rsk status' to verify installed skills.")
	return nil
}

// constraintsFromArgs returns a map of name → version constraint pulled from
// any "name@version" arg. Bundle names are passed through unchanged.
func constraintsFromArgs(args []string) (constraints map[string]string, err error) {
	constraints = make(map[string]string, len(args))
	for _, raw := range args {
		name, version, parseErr := parseNameVersion(raw)
		if parseErr != nil {
			return nil, parseErr
		}
		if version != "" {
			constraints[name] = version
		}
	}
	return constraints, nil
}

// printInstallPreview renders the skill × target preview block shared by the
// global and project install flows.
func printInstallPreview(out io.Writer, skills []skill.Skill, targets, warnings []string, dryRun bool) {
	fmt.Fprintln(out)
	if dryRun {
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
				ui.PadRight(ui.SkillVersion(s.Version), versionColumnWidth),
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
}

// runInstallFromMod reads rsk.mod in the current project, resolves all skills,
// symlinks them into the per-tool project skills dirs, updates rsk.lock, and
// rewrites tool configs.
func runInstallFromMod(cmd *cobra.Command, opts installOpts) error {
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()
	ctx := cmd.Context()

	rskDir, err := manifest.ProjectFolderPath()
	if err != nil {
		return err
	}

	m, err := manifest.ReadMod(rskDir)
	if err != nil {
		return err
	}

	if len(m.Skills) == 0 {
		ui.Warn(out, "rsk.mod has no skills. Run 'rsk install <name>' to add one.")
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n  Run 'rsk init' to set up rsk on this machine", err)
	}

	localSrc := newLocalSource(cfg)
	officialSrc := source.NewOfficial(cfg.OfficialCache)

	projectRoot := filepath.Dir(rskDir)
	targets := projectSkillsDirs(projectRoot, m)

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

	printInstallPreview(out, skills, targets, warnings, opts.dryRun)
	if opts.dryRun {
		return nil
	}
	if !ui.ConfirmYN(out, "Proceed?") {
		fmt.Fprintln(out, "Aborted.")
		return nil
	}
	fmt.Fprintln(out)

	for _, target := range targets {
		if mkdirErr := os.MkdirAll(target, fsperm.Dir); mkdirErr != nil {
			return fmt.Errorf("create %s: %w", target, mkdirErr)
		}
	}

	lock, err := manifest.ReadLock(rskDir)
	if err != nil {
		return err
	}

	var failed int
	for _, s := range skills {
		linkFailed := false
		for _, target := range targets {
			if linkErr := skill.Link(s, target); linkErr != nil {
				ui.Failf(errOut, "link %s → %s: %v", s.Name, target, linkErr)
				linkFailed = true
			}
		}
		if linkFailed {
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
			ui.MutedPath(filepath.Join(targets[0], s.Name)),
		))
	}

	if failed > 0 {
		return fmt.Errorf("%d skill(s) failed to install", failed)
	}

	if err = manifest.WriteLock(rskDir, lock); err != nil {
		return err
	}
	if err = syncPinnedAllTools(rskDir, m); err != nil {
		return err
	}

	fmt.Fprintln(out)
	ui.Info(out, "Run 'rsk status' to verify installed skills.")
	return nil
}
