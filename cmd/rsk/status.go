package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/termkit"

	"github.com/ralvarezdev/ralvaskills/v2/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/v2/internal/config"
	"github.com/ralvarezdev/ralvaskills/v2/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/v2/internal/skill"
	"github.com/ralvarezdev/ralvaskills/v2/internal/ui"
)

var statusCmd = &cobra.Command{
	Use:   "status [flags]",
	Short: "Show installed skills and version drift (see also: list).",
	Long: `List installed skills with their versions and drift. Unlike 'rsk list'
(the offline manifest/installed view), status compares versions across
project + global.

Without --stack, no network calls are made.
With --stack, fetches latest versions from proxy.golang.org and pypi.org
and highlights skills whose STACK.md may be outdated (results cached 24h).

Examples:
  rsk status
  rsk status --global
  rsk status --project
  rsk status --stack
  rsk status --stack --refresh
  rsk status -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		inc, err := cmdx.ReadIncludes(cmd, cmdx.IncludePersonal)
		if err != nil {
			return err
		}
		return runStatus(cmd, statusOpts{
			global:   cmdx.Bool(cmd, cmdx.FlagGlobal),
			project:  cmdx.Bool(cmd, "project"),
			stack:    cmdx.Bool(cmd, cmdx.FlagStack),
			refresh:  cmdx.Bool(cmd, cmdx.FlagRefresh),
			personal: inc.Personal,
			forTool:  cmdx.String(cmd, cmdx.FlagFor),
			output:   outputFormat(cmdx.String(cmd, cmdx.FlagOutput)),
		})
	},
}

type (
	// statusOpts holds the options for the status command.
	statusOpts struct {
		global, project, stack, refresh, personal bool
		forTool                                   string
		output                                    outputFormat
	}

	// statusSection groups linked skills under a single target directory.
	statusSection struct {
		title    string
		subtitle string
		dir      string
		skills   []linkedEntry
	}

	// linkedEntry describes one symlink found in a target directory.
	linkedEntry struct {
		name    string
		version string
		source  skill.Source
		bundles []string
	}

	// statusSkillEntry is one skill row inside a statusSectionEntry, serialized
	// as JSON when -o json is set.
	statusSkillEntry struct {
		Name    string   `json:"name"`
		Version string   `json:"version,omitempty"`
		Source  string   `json:"source,omitempty"`
		Pinned  bool     `json:"pinned,omitempty"`
		Bundles []string `json:"bundles,omitempty"`
	}

	// statusSectionEntry is one section (a scanned target directory) in the
	// JSON status output, mirroring statusSection's text rendering.
	statusSectionEntry struct {
		Title    string             `json:"title"`
		Subtitle string             `json:"subtitle,omitempty"`
		Dir      string             `json:"dir"`
		Skills   []statusSkillEntry `json:"skills"`
	}
)

func runStatus(cmd *cobra.Command, opts statusOpts) error {
	out := cmd.OutOrStdout()
	capture := ui.CaptureFromContext(cmd.Context())

	if err := validateStatusOpts(opts); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n  Run 'rsk init' to set up rsk on this machine", err)
	}

	// --global means "report on the global install only" — no project
	// (rsk.mod) needs to exist in the current directory, so skip resolving
	// it entirely, matching how resolveTargetDirs branches for list/install.
	var rskDir string
	if !opts.global {
		rskDir, err = manifest.ProjectFolderPath()
		if err != nil {
			return err
		}
	}

	catalog, catalogWarn := config.LoadCatalog("")
	if catalogWarn != nil {
		warnOrCapture(out, capture, fmt.Sprintf("user catalog: %v", catalogWarn))
	}
	membership := bundleMembershipIndex(catalog)

	m, modErr := manifest.ReadMod(rskDir)
	var projectDirs []string
	if modErr == nil {
		projectDirs = projectSkillsDirs(filepath.Dir(rskDir), m)
	}
	sections := buildStatusSections(cfg, opts.global, opts.project, opts.forTool, projectDirs)

	pinnedSet := make(map[string]bool)
	if modErr == nil {
		for _, p := range m.Pinned {
			pinnedSet[p] = true
		}
	}

	scanStatusSections(out, capture, sections, cfg, membership, opts.personal)

	if opts.output == outputJSON {
		return writeOrCaptureJSON(out, capture, statusSectionsToEntries(sections, pinnedSet))
	}
	return printStatusText(out, capture, cmd, sections, pinnedSet)
}

// warnOrCapture reports msg on the capture (when a captured in-process run is
// in progress) instead of writing it to out, since nothing written to out is
// seen while the TUI owns the terminal.
func warnOrCapture(out io.Writer, capture *termkit.Capture, msg string) {
	if capture != nil {
		capture.AddMessage(msg)
		return
	}
	ui.Warn(out, msg)
}

// validateStatusOpts checks runStatus's flag combinations before doing any
// scanning work.
func validateStatusOpts(opts statusOpts) error {
	if !opts.output.valid() {
		return fmt.Errorf("--output must be '%s' or '%s'", outputText, outputJSON)
	}
	if opts.stack {
		return errors.New("--stack is not yet implemented")
	}
	if !opts.stack && opts.refresh {
		return errors.New("--refresh requires --stack")
	}
	if opts.global && opts.project {
		return errors.New("--global and --project are mutually exclusive")
	}
	if !opts.global && opts.forTool != "" {
		return errors.New("--for requires --global")
	}
	return nil
}

// scanStatusSections fills in each section's skills in place by scanning its
// target directory, warning (rather than failing) on a directory that can't
// be scanned.
func scanStatusSections(
	out io.Writer, capture *termkit.Capture,
	sections []statusSection, cfg config.Config, membership map[string][]string, includePersonal bool,
) {
	for i := range sections {
		entries, err := scanLinked(
			sections[i].dir,
			cfg.RepoPath, cfg.OfficialCache, cfg.RegistryCache(),
			membership, includePersonal,
		)
		if err != nil {
			warnOrCapture(out, capture, fmt.Sprintf("scan %s: %v", sections[i].dir, err))
		}
		sections[i].skills = entries
	}
}

// printStatusText renders every non-empty section as a table, or a "no
// skills installed" notice when all sections are empty.
func printStatusText(
	out io.Writer, capture *termkit.Capture, cmd *cobra.Command, sections []statusSection, pinnedSet map[string]bool,
) error {
	hasAny := false
	for _, sec := range sections {
		if len(sec.skills) == 0 {
			continue
		}
		hasAny = true
		if capture == nil {
			fmt.Fprintln(out)
			ui.SectionHeader(out, sec.title, sec.subtitle)
		}
		printStatusSectionRows(out, capture, cmd, sec.title, sec.skills, pinnedSet)
	}

	if capture != nil {
		if !hasAny {
			capture.AddMessage("No skills installed. Run 'rsk install <bundle>' to get started.")
		}
		return nil
	}

	fmt.Fprintln(out)
	if !hasAny {
		ui.Warn(out, "No skills installed.")
		ui.Indent(out, "Run 'rsk install <bundle>' to get started.")
	}
	return nil
}

// printStatusSectionRows renders one row per linked skill, tagged with
// [pinned] and/or its bundle memberships. title becomes the captured table's
// Title, so the result view's tab line identifies which section is active
// when a captured status run has more than one.
func printStatusSectionRows(
	out io.Writer, capture *termkit.Capture, cmd *cobra.Command, title string,

	skills []linkedEntry, pinnedSet map[string]bool,
) {
	header := []string{headerSource, headerName, headerVersion, "", ""}
	rows := termkit.Rows(skills, statusRowTable(pinnedSet))
	if capture != nil {
		capture.AddTable(termkit.Data{
			Title: title,

			Headers: header,

			Rows:    rows,
			IDs:     linkedSkillNames(skills),
			Actions: ui.RowActionsFor(cmd),
		})
		return
	}
	termkit.WriteTableStyled(out, header, rows, false, nil, true, false)
}

// linkedSkillNames extracts each linked skill's name, parallel to
// statusRowTable's rows, so a captured status table can identify the row a
// RowAction fires on independently of how it's displayed.
func linkedSkillNames(skills []linkedEntry) []string {
	names := make([]string, len(skills))
	for i, e := range skills {
		names[i] = e.name
	}
	return names
}

// statusRowActionScope resolves the captured status table a row action fired
// on to the uninstall scope its rows belong to: a "Global — <tool>" section
// runs uninstall with --global --for <tool>, while a "Project" section keeps
// uninstall's project default. This is what lets a bare `rsk status`, which
// renders one table per scanned directory, uninstall from the right one.
func statusRowActionScope(table termkit.Data) map[string]string {
	if tool, ok := strings.CutPrefix(table.Title, "Global — "); ok {
		return map[string]string{cmdx.FlagGlobal: "true", cmdx.FlagFor: tool}
	}
	return nil
}

// statusRowTable projects one linked skill into its status row. The pinned set
// is captured because it is per-section, not per-entry.
func statusRowTable(pinnedSet map[string]bool) termkit.Table[linkedEntry] {
	return termkit.Table[linkedEntry]{
		Row: func(e linkedEntry) []any {
			return []any{
				ui.SourceLabel(e.source),
				ui.SkillName(e.name),
				ui.SkillVersion(e.version),
				ui.SuccessMark,
				statusRowTags(e, pinnedSet),
			}
		},
	}
}

// statusRowTags builds the trailing "  [pinned]  bundle-tag ..." suffix for
// one status row.
func statusRowTags(e linkedEntry, pinnedSet map[string]bool) string {
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
	return tags
}

// statusSectionsToEntries converts the scanned statusSections into their
// JSON-serializable shape, carrying over the pinned marker (which is tracked
// separately from statusSection/linkedEntry) onto each skill row.
func statusSectionsToEntries(sections []statusSection, pinnedSet map[string]bool) []statusSectionEntry {
	entries := make([]statusSectionEntry, len(sections))
	for i, sec := range sections {
		skills := make([]statusSkillEntry, len(sec.skills))
		for j, e := range sec.skills {
			skills[j] = statusSkillEntry{
				Name:    e.name,
				Version: e.version,
				Source:  e.source.String(),
				Pinned:  pinnedSet[e.name],
				Bundles: e.bundles,
			}
		}
		entries[i] = statusSectionEntry{
			Title:    sec.title,
			Subtitle: sec.subtitle,
			Dir:      sec.dir,
			Skills:   skills,
		}
	}
	return entries
}

func buildStatusSections(
	cfg config.Config,
	globalOnly, projectOnly bool,
	forTool string,
	projectDirs []string,
) []statusSection {
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
		for _, dir := range projectDirs {
			sections = append(sections, statusSection{
				title:    "Project",
				subtitle: dir,
				dir:      dir,
			})
		}
	}

	return sections
}

func scanLinked(
	targetDir, repoPath, officialCachePath, registryCachePath string,
	membership map[string][]string,
	includePersonal bool,
) ([]linkedEntry, error) {
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
		if statErr != nil || !skill.IsLink(fi) {
			continue
		}
		symlinkTarget, readErr := os.Readlink(linkPath)
		if readErr != nil {
			continue
		}
		if !includePersonal && skill.IsPersonalPath(symlinkTarget) {
			continue
		}
		skillVersion, readVersionErr := skill.ReadVersion(symlinkTarget)
		if readVersionErr != nil {
			skillVersion = ""
		}
		src := detectSkillSource(symlinkTarget, repoPath, officialCachePath, registryCachePath)
		result = append(result, linkedEntry{
			name:    e.Name(),
			version: skillVersion,
			source:  src,
			bundles: membership[e.Name()],
		})
	}
	return result, nil
}
