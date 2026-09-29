package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/termkit"

	"github.com/ralvarezdev/ralvaskills/v2/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/v2/internal/config"
	"github.com/ralvarezdev/ralvaskills/v2/internal/skill"
	"github.com/ralvarezdev/ralvaskills/v2/internal/source"
	"github.com/ralvarezdev/ralvaskills/v2/internal/ui"
)

var catalogCmd = &cobra.Command{
	Use:   "catalog [flags]",
	Short: "Browse the catalog of available skills and bundles.",
	Args:  cobra.MaximumNArgs(1),
	Long: `Browse what rsk knows about. By default lists every skill. Use --bundle
alone to list bundles instead, or --bundle <name> to list the skills inside a
single bundle.

The older --bundles and --personal flags still work as hidden aliases of
--bundle and --include personal.

Examples:
  rsk catalog                       # all skills
  rsk catalog --bundle              # all bundles
  rsk catalog --bundle go-grpc      # skills inside the go-grpc bundle
  rsk catalog --source local        # filter by source
  rsk catalog --include personal    # include personal/ skills
  rsk catalog -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		inc, err := cmdx.ReadIncludes(cmd, cmdx.IncludePersonal)
		if err != nil {
			return err
		}
		sourceFilter, err := parseSourceFilter(cmdx.String(cmd, cmdx.FlagSource))
		if err != nil {
			return err
		}
		bundles, bundle, err := resolveBundleFlag(cmdx.String(cmd, cmdx.FlagBundle), args)
		if err != nil {
			return err
		}
		output, err := parseOutputFormat(cmdx.String(cmd, cmdx.FlagOutput))
		if err != nil {
			return err
		}
		return runCatalog(cmd, catalogOpts{
			bundles:  bundles,
			personal: inc.Personal,
			bundle:   bundle,
			source:   sourceFilter,
			output:   output,
		})
	},
}

// bundleListSentinel is --bundle's NoOptDefVal: the value the flag takes when
// given with no explicit name, meaning "list the bundles themselves".
const bundleListSentinel = "*"

// resolveBundleFlag folds --bundle [name] and an optional positional name into
// (list bundles?, bundle name). pflag only binds a NoOptDefVal flag's value
// with --bundle=NAME, so a lone positional argument after a bare --bundle is
// taken as the name (--bundle NAME).
func resolveBundleFlag(flagValue string, args []string) (listBundles bool, name string, err error) {
	if len(args) > 0 && flagValue != bundleListSentinel {
		return false, "", fmt.Errorf("unexpected argument %q (a bundle name goes after --bundle)", args[0])
	}
	if len(args) > 0 {
		flagValue = args[0]
	}
	if flagValue == bundleListSentinel {
		return true, "", nil
	}
	return false, flagValue, nil
}

type (
	catalogOpts struct {
		bundles, personal bool
		bundle            string
		source            skill.Source
		output            outputFormat
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

func runCatalog(cmd *cobra.Command, opts catalogOpts) error {
	if !opts.output.valid() {
		return fmt.Errorf("--output must be '%s' or '%s'", outputText, outputJSON)
	}
	if opts.bundles && (opts.bundle != "" || opts.personal || opts.source != "") {
		return errors.New("--bundle <name>, --include, and --source apply when listing skills, not bundles")
	}
	if opts.bundles {
		return runCatalogBundles(cmd, opts)
	}
	return runCatalogSkills(cmd, opts)
}

func runCatalogSkills(cmd *cobra.Command, opts catalogOpts) error {
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

	if opts.output == outputJSON {
		return writeOrCaptureJSON(out, capture, skillsToEntries(all))
	}
	return printCatalogSkillTable(out, capture, cmd, all)
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
		return nil, fmt.Errorf("bundle %q not found — run 'rsk catalog --bundles' to see all bundles", bundleName)
	}
	want := make(map[string]bool, len(bundle.Skills))
	for _, ref := range bundle.Skills {
		want[ref.Name] = true
	}
	return filterSkills(all, func(s skill.Skill) bool { return want[s.Name] }), nil
}

func printCatalogSkillTable(out io.Writer, capture *termkit.Capture, cmd *cobra.Command, skills []skill.Skill) error {
	header := []string{headerSource, headerName, headerVersion}
	rows := termkit.Rows(skills, catalogSkillTable)
	if capture != nil {
		capture.AddTable(termkit.Data{
			Headers: header,
			Rows:    rows,
			IDs:     skillNames(skills),
			Actions: ui.RowActionsFor(cmd),
		})
		return nil
	}
	fmt.Fprintln(out)
	termkit.WriteTableStyled(out, header, rows, false, nil, termkit.TableBorderless, ui.Theme)
	fmt.Fprintln(out)
	return nil
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

	if opts.output == outputJSON {
		return writeOrCaptureJSON(out, capture, bundleRowsToEntries(rows))
	}
	return printCatalogBundleTable(out, capture, cmd, rows)
}

func printCatalogBundleTable(out io.Writer, capture *termkit.Capture, cmd *cobra.Command, rows []bundleRow) error {
	header := []string{"", "Bundle", "Linked", "Description"}
	tableRows := termkit.Rows(rows, catalogBundleTable)
	if capture != nil {
		capture.AddTable(termkit.Data{
			Headers: header,
			Rows:    tableRows,
			IDs:     bundleNames(rows),
			Actions: ui.RowActionsFor(cmd),
		})
		return nil
	}
	fmt.Fprintln(out)
	termkit.WriteTableStyled(out, header, tableRows, false, nil, termkit.TableBorderless, ui.Theme)
	fmt.Fprintln(out)
	return nil
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
