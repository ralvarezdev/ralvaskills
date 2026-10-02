package main

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/termkit"

	"github.com/ralvarezdev/ralvaskills/v3/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/v3/internal/config"
	"github.com/ralvarezdev/ralvaskills/v3/internal/install"
	"github.com/ralvarezdev/ralvaskills/v3/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/v3/internal/skill"
	"github.com/ralvarezdev/ralvaskills/v3/internal/source"
	"github.com/ralvarezdev/ralvaskills/v3/internal/tool"
	"github.com/ralvarezdev/ralvaskills/v3/internal/ui"
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
  rsk install <name>                           # install a bundle or skill by name
  rsk install                                  # project: install everything in rsk.mod
  rsk install go-grpc                          # project: bundle + manifest track
  rsk install go-architect --pin               # project: install + pin in CLAUDE.md
  rsk install global --global                  # global: bundle to every tool dir
  rsk install go-grpc --global --for claude-code
  rsk install demo-script-architect --include personal
  rsk install go-grpc --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		inc, err := cmdx.ReadIncludes(cmd, cmdx.IncludePersonal)
		if err != nil {
			return err
		}
		forTool, err := forToolFlag(cmd)
		if err != nil {
			return err
		}
		return runInstall(cmd, installOpts{
			global:   cmdx.Bool(cmd, cmdx.FlagGlobal),
			dryRun:   cmdx.Bool(cmd, cmdx.FlagDryRun),
			personal: inc.Personal,
			pin:      cmdx.Bool(cmd, cmdx.FlagPin),
			yes:      cmdx.Bool(cmd, cmdx.FlagYes),
			forTool:  forTool,
			version:  cmdx.String(cmd, cmdx.FlagVersion),
		}, args)
	},
}

type installOpts struct {
	global, dryRun, personal, pin, yes bool
	forTool                            tool.ID
	version                            string
}

func runInstall(cmd *cobra.Command, opts installOpts, args []string) error {
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()
	ctx := cmd.Context()

	if err := validateInstallOpts(opts); err != nil {
		return err
	}

	// Project-bundle install with no args: read rsk.mod and install everything.
	if !opts.global && len(args) == 0 {
		return runInstallFromMod(cmd, opts)
	}

	if opts.global && len(args) == 0 {
		name, err := promptInstallName(cmd, out)
		if err != nil {
			return err
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

	if err = checkNoPersonalSkills(skills, opts.personal); err != nil {
		return err
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

// validateInstallOpts checks the flag combinations runInstall requires
// before doing any resolution work.
func validateInstallOpts(opts installOpts) error {
	if !opts.global && opts.forTool != "" {
		return fieldError(cmdx.FlagFor, "--for requires --global")
	}
	if opts.global && opts.pin {
		return fieldError(cmdx.FlagPin, "--pin only applies to project installs (drop --global or --pin)")
	}
	if opts.version != "" {
		return fieldError(
			cmdx.FlagVersion,
			"--version is not yet supported; skills are symlinked from the local repo HEAD",
		)
	}
	return nil
}

// promptInstallName asks for a bundle or skill name when `rsk install
// --global` was run with no positional args.
func promptInstallName(cmd *cobra.Command, out io.Writer) (string, error) {
	if ui.InSession(cmd.Context()) {
		return "", fieldError(fieldArgName, "specify at least one bundle or skill name")
	}
	name, err := ui.Ask(out, "Skill or bundle name", "")
	if err != nil {
		return "", err
	}
	if name == "" {
		return "", errors.New("specify at least one bundle or skill name")
	}
	return name, nil
}

// checkNoPersonalSkills rejects the install when any resolved skill lives
// under personal/ and --personal wasn't passed.
func checkNoPersonalSkills(skills []skill.Skill, personal bool) error {
	if personal {
		return nil
	}
	for _, s := range skills {
		if s.IsPersonal {
			return fmt.Errorf("skill %q is in personal/ — pass --include personal to install it", s.Name)
		}
	}
	return nil
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
	if !termkit.Confirm(cmd, opts.yes, termkit.StdioIsTerminal(), "Proceed?") {
		fmt.Fprintln(out, "Aborted.")
		return nil
	}
	fmt.Fprintln(out)

	results, err := install.Apply(skills, install.Options{
		Scope:   install.ScopeGlobal,
		Targets: targets,
	})
	if err != nil {
		return err
	}
	return reportInstall(out, errOut, results)
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

	if !previewAndConfirm(cmd, out, skills, targets, warnings, opts.dryRun, opts.yes) {
		return nil
	}

	constraints, err := constraintsFromArgs(args)
	if err != nil {
		return err
	}

	results, err := install.Apply(skills, install.Options{
		Scope:       install.ScopeProject,
		Targets:     targets,
		RskDir:      rskDir,
		Pin:         opts.pin,
		Constraints: constraints,
	})
	if err != nil {
		return err
	}
	return reportInstall(out, errOut, results)
}

// previewAndConfirm renders the install preview and, unless dryRun, asks the
// user to confirm. It returns false whenever the caller should stop (dry run
// finished, or the user declined).
func previewAndConfirm(
	cmd *cobra.Command, out io.Writer, skills []skill.Skill, targets, warnings []string, dryRun, yes bool,
) bool {
	printInstallPreview(out, skills, targets, warnings, dryRun)
	if dryRun {
		return false
	}
	if !termkit.Confirm(cmd, yes, termkit.StdioIsTerminal(), "Proceed?") {
		fmt.Fprintln(out, "Aborted.")
		return false
	}
	fmt.Fprintln(out)
	return true
}

// constraintsFromArgs returns name → version constraints from "name@version"
// args. Bundle names are passed through unchanged.
func constraintsFromArgs(args []string) (map[string]string, error) {
	return install.ConstraintsFromArgs(args)
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

	rows := make([]installPlanRow, 0, len(targets)*len(skills))
	for _, s := range skills {
		for _, target := range targets {
			rows = append(rows, installPlanRow{
				source: ui.SourceLabel(s.Source),
				name:   ui.SkillName(s.Name),
				ver:    ui.SkillVersion(s.Version),
				target: ui.MutedPath(filepath.Join(target, s.Name)),
				relink: skill.IsLinked(s.Name, target),
			})
		}
	}
	termkit.WriteTableStyled(out, []string{headerSource, headerName, headerVersion, "", headerPath, ""},
		termkit.Rows(rows, installPlanTable), false, nil, termkit.TableBorderless, ui.Theme)

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

	if !previewAndConfirm(cmd, out, skills, targets, warnings, opts.dryRun, opts.yes) {
		return nil
	}

	results, err := install.Apply(skills, install.Options{
		Scope:   install.ScopeProject,
		Targets: targets,
		RskDir:  rskDir,
	})
	if err != nil {
		return err
	}
	return reportInstall(out, errOut, results)
}

// reportInstall renders the outcome of an install.Apply call and returns an
// error when any skill failed, so the caller exits non-zero.
func reportInstall(out, errOut io.Writer, results []install.Result) error {
	var failed int
	for _, r := range results {
		if r.Err != nil {
			ui.Failf(errOut, "link %s: %v", r.Name, r.Err)
			failed++
			continue
		}
		for _, file := range r.Files {
			ui.Success(out, fmt.Sprintf("%s  %s  %s",
				ui.SkillName(r.Name),
				ui.Arrow,
				ui.MutedPath(file),
			))
		}
	}

	if failed > 0 {
		return fmt.Errorf("%d skill(s) failed to install", failed)
	}
	fmt.Fprintln(out)
	ui.Info(out, "Run 'rsk status' to verify installed skills.")
	return nil
}

// installPlanRow is one row of the "skills to install" preview.
type installPlanRow struct {
	source string
	name   string
	ver    string
	target string
	relink bool
}

// installPlanTable projects one install-preview row.
var installPlanTable = termkit.Table[installPlanRow]{
	Row: func(r installPlanRow) []any {
		suffix := ""
		if r.relink {
			suffix = ui.ReLink
		}
		return []any{r.source, r.name, r.ver, ui.Arrow, r.target, suffix}
	},
}
