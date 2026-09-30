package main

import (
	"context"
	"fmt"
	"io"
	"sort"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/termkit"
	"github.com/ralvarezdev/termkit/output"

	"github.com/ralvarezdev/ralvaskills/v2/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/v2/internal/config"
	"github.com/ralvarezdev/ralvaskills/v2/internal/skill"
	"github.com/ralvarezdev/ralvaskills/v2/internal/source"
	"github.com/ralvarezdev/ralvaskills/v2/internal/ui"
)

var catalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Browse the catalog of available skills and bundles.",
	Args:  cobra.NoArgs,
	Long: `Browse what rsk knows about. By default lists every skill.

The bundles subcommand lists the bundles themselves; use
catalog bundle <name> to list the skills inside one bundle.

Examples:
  rsk catalog                       # all skills
  rsk catalog bundles               # all bundles
  rsk catalog bundle go-grpc        # skills inside the go-grpc bundle
  rsk catalog --source local        # filter by source
  rsk catalog --include personal    # include personal/ skills
  rsk catalog -o json`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		opts, err := catalogSkillsOpts(cmd)
		if err != nil {
			return err
		}
		return runCatalogSkills(cmd, opts)
	},
}

// catalogBundlesCmd lists the catalog's bundles and how much of each is
// installed.
var catalogBundlesCmd = &cobra.Command{
	Use:   "bundles",
	Short: "List the catalog's bundles and how many of their skills are installed.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runCatalogBundles(cmd, catalogOpts{format: printer.Format(cmd)})
	},
}

// catalogBundleCmd lists the skills that belong to one bundle.
var catalogBundleCmd = &cobra.Command{
	Use:   "bundle <name>",
	Short: "List the skills inside one bundle.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, err := catalogSkillsOpts(cmd)
		if err != nil {
			return err
		}
		opts.bundle = args[0]
		return runCatalogSkills(cmd, opts)
	},
}

// catalogSkillsOpts reads the flags shared by the skill-list views: the
// --include/--personal scope, the --source filter, and the --output format.
func catalogSkillsOpts(cmd *cobra.Command) (catalogOpts, error) {
	inc, err := cmdx.ReadIncludes(cmd, cmdx.IncludePersonal)
	if err != nil {
		return catalogOpts{}, err
	}
	sourceFilter, err := parseSourceFilter(cmdx.String(cmd, cmdx.FlagSource))
	if err != nil {
		return catalogOpts{}, err
	}
	return catalogOpts{personal: inc.Personal, source: sourceFilter, format: printer.Format(cmd)}, nil
}

type (
	catalogOpts struct {
		personal bool
		bundle   string
		source   skill.Source
		format   output.OutputFormat
	}

	skillEntry struct {
		Name     string `json:"name"`
		Version  string `json:"version"`
		Source   string `json:"source"`
		Personal bool   `json:"personal,omitempty"`
		Path     string `json:"path"`
	}

	bundleRow struct {
		config.Bundle
		linked int
		total  int
	}

	bundleEntry struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Skills      int    `json:"skills"`
		Linked      int    `json:"linked"`
	}
)

func runCatalogSkills(cmd *cobra.Command, opts catalogOpts) error {
	if err := checkOutput(cmd); err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	ctx := cmd.Context()
	capture := termkit.CaptureFromContext(ctx)

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n  Run 'rsk init' to set up rsk on this machine", err)
	}

	all := loadSkillsBySource(ctx, out, capture, cfg, opts.source)

	if opts.bundle != "" {
		all, err = filterByBundle(out, capture, all, opts.bundle)
		if err != nil {
			return err
		}
	}

	if !opts.personal {
		all = filterSkills(all, func(s skill.Skill) bool { return !s.IsPersonal })
	}

	sort.Slice(all, func(i, j int) bool {
		if all[i].Source != all[j].Source {
			return all[i].Source < all[j].Source
		}
		return all[i].Name < all[j].Name
	})

	if len(all) == 0 {
		if capture != nil {
			capture.AddMessage("No skills found.")
			return nil
		}
		ui.Info(out, "No skills found.")
		return nil
	}

	header := []string{headerSource, headerName, headerVersion}
	tableRows := termkit.Rows(all, catalogSkillTable)

	return renderList(cmd, opts.format, "", skillsToEntries(all), header, tableRows, skillNames(all))
}

// loadSkillsBySource walks the local and/or official sources per sourceFilter
// ("" means both), warning on (rather than failing for) a source that can't
// be walked.
func loadSkillsBySource(
	ctx context.Context, out io.Writer, capture *termkit.Capture, cfg config.Config, sourceFilter skill.Source,
) []skill.Skill {
	var all []skill.Skill

	if sourceFilter == "" || sourceFilter == skill.SourceLocal {
		local, walkErr := newLocalSource(cfg).All(ctx)
		if walkErr != nil {
			warnOrCapture(out, capture, fmt.Sprintf("walk local skills: %v", walkErr))
		} else {
			all = append(all, local...)
		}
	}

	if sourceFilter == "" || sourceFilter == skill.SourceOfficial {
		official, walkErr := source.NewOfficial(cfg.OfficialCache).All(ctx)
		if walkErr != nil {
			warnOrCapture(out, capture, walkErr.Error())
		} else {
			all = append(all, official...)
		}
	}

	return all
}

// filterByBundle keeps only the skills named in bundleName's catalog entry.
func filterByBundle(
	out io.Writer, capture *termkit.Capture, all []skill.Skill, bundleName string,
) ([]skill.Skill, error) {
	catalog, catalogWarn := config.LoadCatalog("")
	if catalogWarn != nil {
		warnOrCapture(out, capture, fmt.Sprintf("user catalog: %v", catalogWarn))
	}
	bundle, ok := config.FindBundle(catalog, bundleName)
	if !ok {
		return nil, fmt.Errorf("bundle %q not found — run 'rsk catalog bundles' to see all bundles", bundleName)
	}
	want := make(map[string]bool, len(bundle.Skills))
	for _, ref := range bundle.Skills {
		want[ref.Name] = true
	}
	return filterSkills(all, func(s skill.Skill) bool { return want[s.Name] }), nil
}

// skillNames extracts each skill's name, parallel to catalogSkillTable's
// rows, so a captured table can identify the row a RowAction fires on
// independently of how it's displayed.
func skillNames(skills []skill.Skill) []string {
	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Name
	}
	return names
}

// catalogSkillTable projects one catalog skill row.
var catalogSkillTable = termkit.Table[skill.Skill]{
	Row: func(s skill.Skill) []any {
		return []any{ui.SourceLabel(s.Source), ui.SkillName(s.Name), ui.SkillVersion(s.Version)}
	},
}

func skillsToEntries(skills []skill.Skill) []skillEntry {
	out := make([]skillEntry, len(skills))
	for i, s := range skills {
		out[i] = skillEntry{
			Name:     s.Name,
			Version:  s.Version,
			Source:   s.Source.String(),
			Personal: s.IsPersonal,
			Path:     s.Path,
		}
	}
	return out
}

func runCatalogBundles(cmd *cobra.Command, opts catalogOpts) error {
	if err := checkOutput(cmd); err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	capture := termkit.CaptureFromContext(cmd.Context())

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n  Run 'rsk init' to set up rsk on this machine", err)
	}

	catalog, catalogWarn := config.LoadCatalog("")
	if catalogWarn != nil {
		warnOrCapture(out, capture, fmt.Sprintf("user catalog: %v", catalogWarn))
	}

	targets := allTargetDirs(cfg)
	rows := make([]bundleRow, 0, len(catalog))
	for _, b := range catalog {
		linked := 0
		for _, ref := range b.Skills {
			if isLinkedAnywhere(ref.Name, targets) {
				linked++
			}
		}
		rows = append(rows, bundleRow{Bundle: b, linked: linked, total: len(b.Skills)})
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })

	if len(rows) == 0 {
		if capture != nil {
			capture.AddMessage("No bundles found.")
			return nil
		}
		ui.Info(out, "No bundles found.")
		return nil
	}

	header := []string{"", "Bundle", "Linked", "Description"}
	tableRows := termkit.Rows(rows, catalogBundleTable)

	return renderList(cmd, opts.format, "", bundleRowsToEntries(rows), header, tableRows, bundleNames(rows))
}

// bundleNames extracts each bundle's name, parallel to catalogBundleTable's
// rows, so a captured table can identify the row a RowAction fires on
// independently of how it's displayed.
func bundleNames(rows []bundleRow) []string {
	names := make([]string, len(rows))
	for i, r := range rows {
		names[i] = r.Name
	}
	return names
}

// catalogBundleTable projects one bundle row.
var catalogBundleTable = termkit.Table[bundleRow]{
	Row: func(r bundleRow) []any {
		mark := ui.ErrorMark
		switch {
		case r.linked == r.total && r.total > 0:
			mark = ui.SuccessMark
		case r.linked > 0:
			mark = ui.WarnMark
		}
		return []any{
			mark,
			ui.BundleTag(r.Name),
			ui.MutedStyle.Render(fmt.Sprintf("%d/%d", r.linked, r.total)),
			ui.MutedStyle.Render(r.Description),
		}
	},
}

func bundleRowsToEntries(rows []bundleRow) []bundleEntry {
	out := make([]bundleEntry, len(rows))
	for i, r := range rows {
		out[i] = bundleEntry{
			Name:        r.Name,
			Description: r.Description,
			Skills:      r.total,
			Linked:      r.linked,
		}
	}
	return out
}
