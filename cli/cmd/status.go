package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ralvarezdev/ralvaskills/cli/internal/config"
	"github.com/ralvarezdev/ralvaskills/cli/internal/skill"
	"github.com/ralvarezdev/ralvaskills/cli/internal/ui"
	"github.com/spf13/cobra"
)

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
	RunE: runStatus,
}

var (
	statusGlobal   bool
	statusFor      string
	statusProject  bool
	statusStack    bool
	statusRefresh  bool
	statusPersonal bool
)

func init() {
	rootCmd.AddCommand(statusCmd)
	f := statusCmd.Flags()
	f.BoolVar(&statusGlobal, "global", false, "Show global skills only")
	f.StringVar(&statusFor, "for", "", "Scope --global to a single tool (claude-code|opencode)")
	f.BoolVar(&statusProject, "project", false, "Show project skills only")
	f.BoolVar(&statusStack, "stack", false, "Fetch latest versions and show STACK.md drift (network, opt-in)")
	f.BoolVar(&statusRefresh, "refresh", false, "With --stack: bypass the 24h cache and force a re-fetch")
	f.BoolVar(&statusPersonal, "personal", false, "Include personal/ skills in output")
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

func runStatus(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	if statusStack {
		return fmt.Errorf("--stack is not yet implemented")
	}
	if !statusStack && statusRefresh {
		return fmt.Errorf("--refresh requires --stack")
	}
	if statusGlobal && statusProject {
		return fmt.Errorf("--global and --project are mutually exclusive")
	}
	if !statusGlobal && statusFor != "" {
		return fmt.Errorf("--for requires --global")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n  Run 'rsk init' to set up rsk on this machine", err)
	}

	catalog := config.LoadCatalog("")
	membership := bundleMembershipIndex(catalog)
	sections := buildStatusSections(cfg, statusGlobal, statusProject, statusFor)

	for i := range sections {
		entries, scanErr := scanLinked(
			sections[i].dir,
			cfg.RepoPath, cfg.OfficialCache, cfg.RegistryCache(),
			membership, statusPersonal,
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
			bundleTag := ""
			if len(e.bundles) > 0 {
				tags := make([]string, len(e.bundles))
				for i, b := range e.bundles {
					tags[i] = ui.BundleTag(b)
				}
				bundleTag = "  " + strings.Join(tags, " ")
			}
			fmt.Fprintf(out, "  %s  %s  %s  %s%s\n",
				ui.SourceLabel(e.source),
				ui.PadRight(ui.SkillName(e.name), nameWidth),
				ui.PadRight(ui.SkillVersion(e.version), 7),
				ui.SuccessMark(),
				bundleTag,
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

func buildStatusSections(cfg config.Config, globalOnly, projectOnly bool, forTool string) []statusSection {
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

	if !globalOnly {
		cwd, err := os.Getwd()
		if err == nil {
			projectDir := filepath.Join(cwd, ".claude", "skills")
			sections = append(sections, statusSection{
				title:    "Project",
				subtitle: projectDir,
				dir:      projectDir,
			})
		}
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
