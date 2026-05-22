package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ralvarezdev/ralvaskills/cli/internal/config"
	"github.com/ralvarezdev/ralvaskills/cli/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/cli/internal/skill"
	"github.com/ralvarezdev/ralvaskills/cli/internal/ui"
	"github.com/spf13/cobra"
)

type statusOpts struct {
	global, project, stack, refresh, personal bool
	forTool                                   string
}

var statusCmd = &cobra.Command{
	Use:   "status [flags]",
	Short: "Show installed skills and version drift.",
	Long: `List installed skills with their versions.

Without --stack, no network calls are made.
With --stack, fetches latest versions from proxy.golang.org and pypi.org
and highlights skills whose STACK.md may be outdated (results cached 24h).

Examples:
  rsk status
  rsk status --global
  rsk status --project
  rsk status --stack
  rsk status --stack --refresh`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus(cmd, statusOpts{
			global:   flagBool(cmd, "global"),
			project:  flagBool(cmd, "project"),
			stack:    flagBool(cmd, "stack"),
			refresh:  flagBool(cmd, "refresh"),
			personal: flagBool(cmd, "personal"),
			forTool:  flagString(cmd, "for"),
		})
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	f := statusCmd.Flags()
	f.Bool("global", false, "Show global skills only")
	f.String("for", "", "Scope --global to a single tool (claude-code|opencode)")
	f.Bool("project", false, "Show project skills only")
	f.Bool("stack", false, "Fetch latest versions and show STACK.md drift (network, opt-in)")
	f.Bool("refresh", false, "With --stack: bypass the 24h cache and force a re-fetch")
	f.Bool("personal", false, "Include personal/ skills in output")
}

// statusSection groups linked skills under a single target directory.
type statusSection struct {
	title    string
	subtitle string
	dir      string
	skills   []linkedEntry
}

// linkedEntry describes one symlink found in a target directory.
type linkedEntry struct {
	name    string
	version string
	source  skill.Source
	bundles []string
}

func runStatus(cmd *cobra.Command, opts statusOpts) error {
	out := cmd.OutOrStdout()

	if opts.stack {
		return fmt.Errorf("--stack is not yet implemented")
	}
	if !opts.stack && opts.refresh {
		return fmt.Errorf("--refresh requires --stack")
	}
	if opts.global && opts.project {
		return fmt.Errorf("--global and --project are mutually exclusive")
	}
	if !opts.global && opts.forTool != "" {
		return fmt.Errorf("--for requires --global")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n  Run 'rsk init' to set up rsk on this machine", err)
	}

	cwd, _ := os.Getwd()

	catalog, catalogWarn := config.LoadCatalog("")
	if catalogWarn != nil {
		ui.Warn(out, fmt.Sprintf("user catalog: %v", catalogWarn))
	}
	membership := bundleMembershipIndex(catalog)
	sections := buildStatusSections(cfg, opts.global, opts.project, opts.forTool, cwd)

	pinnedSet := make(map[string]bool)
	if cwd != "" {
		if m, modErr := manifest.ReadMod(filepath.Join(cwd, ".rsk")); modErr == nil {
			for _, p := range m.Pinned {
				pinnedSet[p] = true
			}
		}
	}

	for i := range sections {
		entries, scanErr := scanLinked(
			sections[i].dir,
			cfg.RepoPath, cfg.OfficialCache, cfg.RegistryCache(),
			membership, opts.personal,
		)
		if scanErr != nil {
			ui.Warn(out, fmt.Sprintf("scan %s: %v", sections[i].dir, scanErr))
		}
		sections[i].skills = entries
	}

	any := false
	for _, sec := range sections {
		if len(sec.skills) == 0 {
			continue
		}
		any = true
		fmt.Fprintln(out)
		ui.SectionHeader(out, sec.title, sec.subtitle)

		names := make([]string, len(sec.skills))
		for i, e := range sec.skills {
			names[i] = e.name
		}
		nameWidth := ui.MaxWidth(names)

		for _, e := range sec.skills {
			tags := ""
			if pinnedSet[e.name] {
				tags += "  [pinned]"
			}
			if len(e.bundles) > 0 {
				bundleTags := make([]string, len(e.bundles))
				for i, b := range e.bundles {
					bundleTags[i] = ui.BundleTag(b)
				}
				tags += "  " + strings.Join(bundleTags, " ")
			}
			fmt.Fprintf(out, "  %s  %s  %s  %s%s\n",
				ui.SourceLabel(e.source),
				ui.PadRight(ui.SkillName(e.name), nameWidth),
				ui.PadRight(ui.SkillVersion(e.version), 7),
				ui.SuccessMark(),
				tags,
			)
		}
	}

	fmt.Fprintln(out)

	if !any {
		ui.Warn(out, "No skills installed.")
		ui.Indent(out, "Run 'rsk install <bundle>' to get started.")
	}

	return nil
}

func buildStatusSections(cfg config.Config, globalOnly, projectOnly bool, forTool, cwd string) []statusSection {
	var sections []statusSection

	if !projectOnly {
		if forTool != "" {
			if dir, ok := cfg.GlobalTargets[forTool]; ok {
				sections = append(sections, statusSection{
					title:    "Global — " + forTool,
					subtitle: dir,
					dir:      dir,
				})
			}
		} else {
			for tool, dir := range cfg.GlobalTargets {
				sections = append(sections, statusSection{
					title:    "Global — " + tool,
					subtitle: dir,
					dir:      dir,
				})
			}
		}
	}

	if !globalOnly && cwd != "" {
		projectDir := filepath.Join(cwd, ".rsk", "skills")
		sections = append(sections, statusSection{
			title:    "Project",
			subtitle: projectDir,
			dir:      projectDir,
		})
	}

	return sections
}

func scanLinked(targetDir, repoPath, officialCachePath, registryCachePath string, membership map[string][]string, includePersonal bool) ([]linkedEntry, error) {
	entries, err := os.ReadDir(targetDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var result []linkedEntry
	for _, e := range entries {
		linkPath := filepath.Join(targetDir, e.Name())
		fi, statErr := os.Lstat(linkPath)
		if statErr != nil || fi.Mode()&os.ModeSymlink == 0 {
			continue
		}
		symlinkTarget, readErr := os.Readlink(linkPath)
		if readErr != nil {
			continue
		}
		if !includePersonal && skill.IsPersonalPath(symlinkTarget) {
			continue
		}
		version, _ := skill.ReadVersion(symlinkTarget)
		src := detectSkillSource(symlinkTarget, repoPath, officialCachePath, registryCachePath)
		result = append(result, linkedEntry{
			name:    e.Name(),
			version: version,
			source:  src,
			bundles: membership[e.Name()],
		})
	}
	return result, nil
}
