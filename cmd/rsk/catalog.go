package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/ralvaskills/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/internal/config"
	"github.com/ralvarezdev/ralvaskills/internal/skill"
	"github.com/ralvarezdev/ralvaskills/internal/source"
	"github.com/ralvarezdev/ralvaskills/internal/ui"
)

const (
	outputText outputFormat = "text"
	outputJSON outputFormat = "json"
)

var catalogCmd = &cobra.Command{
	Use:   "catalog [flags]",
	Short: "Browse the catalog of available skills and bundles.",
	Long: `Browse what rsk knows about. By default lists every skill. Use --bundles to
list bundles instead, or --bundle <name> to list the skills inside a single
bundle.

Examples:
  rsk catalog                       # all skills
  rsk catalog --bundles             # all bundles
  rsk catalog --bundle go-grpc      # skills inside the go-grpc bundle
  rsk catalog --source local        # filter by source
  rsk catalog --personal            # include personal/ skills
  rsk catalog -o json`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runCatalog(cmd, catalogOpts{
			bundles:  cmdx.Bool(cmd, cmdx.FlagBundles),
			personal: cmdx.Bool(cmd, cmdx.FlagPersonal),
			bundle:   cmdx.String(cmd, cmdx.FlagBundle),
			source:   cmdx.String(cmd, cmdx.FlagSource),
			output:   outputFormat(cmdx.String(cmd, cmdx.FlagOutput)),
		})
	},
}

type (
	outputFormat string

	catalogOpts struct {
		bundles, personal bool
		bundle, source    string
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

func (o outputFormat) valid() bool {
	return o == outputText || o == outputJSON
}

func runCatalog(cmd *cobra.Command, opts catalogOpts) error {
	if !opts.output.valid() {
		return fmt.Errorf("--output must be '%s' or '%s'", outputText, outputJSON)
	}
	if opts.bundles && (opts.bundle != "" || opts.personal || opts.source != "") {
		return errors.New("--bundle, --personal, and --source apply when listing skills, not --bundles")
	}
	if opts.bundles {
		return runCatalogBundles(cmd, opts)
	}
	return runCatalogSkills(cmd, opts)
}

func runCatalogSkills(cmd *cobra.Command, opts catalogOpts) error {
	out := cmd.OutOrStdout()
	ctx := cmd.Context()

	if opts.source != "" && opts.source != skill.SourceLocal.String() && opts.source != skill.SourceOfficial.String() {
		return fmt.Errorf("--source must be '%s' or '%s'", skill.SourceLocal, skill.SourceOfficial)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n  Run 'rsk init' to set up rsk on this machine", err)
	}

	localSrc := newLocalSource(cfg)
	officialSrc := source.NewOfficial(cfg.OfficialCache)

	var all []skill.Skill

	if opts.source == "" || opts.source == skill.SourceLocal.String() {
		local, walkErr := localSrc.All(ctx)
		if walkErr != nil {
			ui.Warn(out, fmt.Sprintf("walk local skills: %v", walkErr))
		} else {
			all = append(all, local...)
		}
	}

	if opts.source == "" || opts.source == skill.SourceOfficial.String() {
		official, walkErr := officialSrc.All(ctx)
		if walkErr != nil {
			ui.Warn(out, walkErr.Error())
		} else {
			all = append(all, official...)
		}
	}

	if opts.bundle != "" {
		catalog, catalogWarn := config.LoadCatalog("")
		if catalogWarn != nil {
			ui.Warn(out, fmt.Sprintf("user catalog: %v", catalogWarn))
		}
		bundle, ok := config.FindBundle(catalog, opts.bundle)
		if !ok {
			return fmt.Errorf("bundle %q not found — run 'rsk catalog --bundles' to see all bundles", opts.bundle)
		}
		want := make(map[string]bool, len(bundle.Skills))
		for _, ref := range bundle.Skills {
			want[ref.Name] = true
		}
		all = filterSkills(all, func(s skill.Skill) bool { return want[s.Name] })
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
		ui.Info(out, "No skills found.")
		return nil
	}

	if opts.output == outputJSON {
		return writeJSON(out, skillsToEntries(all))
	}
	return printCatalogSkillTable(out, all)
}

func printCatalogSkillTable(out io.Writer, skills []skill.Skill) error {
	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Name
	}
	nameWidth := ui.MaxWidth(names)

	fmt.Fprintln(out)
	for _, s := range skills {
		fmt.Fprintf(out, "  %s  %s  %s\n",
			ui.SourceLabel(s.Source),
			ui.PadRight(ui.SkillName(s.Name), nameWidth),
			ui.SkillVersion(s.Version),
		)
	}
	fmt.Fprintln(out)
	return nil
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

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func runCatalogBundles(cmd *cobra.Command, opts catalogOpts) error {
	out := cmd.OutOrStdout()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n  Run 'rsk init' to set up rsk on this machine", err)
	}

	catalog, catalogWarn := config.LoadCatalog("")
	if catalogWarn != nil {
		ui.Warn(out, fmt.Sprintf("user catalog: %v", catalogWarn))
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
		ui.Info(out, "No bundles found.")
		return nil
	}

	if opts.output == outputJSON {
		return writeJSON(out, bundleRowsToEntries(rows))
	}
	return printCatalogBundleTable(out, rows)
}

func printCatalogBundleTable(out io.Writer, rows []bundleRow) error {
	names := make([]string, len(rows))
	counts := make([]string, len(rows))
	for i, r := range rows {
		names[i] = r.Name
		counts[i] = fmt.Sprintf("%d/%d", r.linked, r.total)
	}
	nameWidth := ui.MaxWidth(names)
	countWidth := ui.MaxWidth(counts)

	fmt.Fprintln(out)
	for i, r := range rows {
		mark := ui.ErrorMark
		switch {
		case r.linked == r.total && r.total > 0:
			mark = ui.SuccessMark
		case r.linked > 0:
			mark = ui.WarnMark
		}
		fmt.Fprintf(out, "  %s  %s  %s  %s\n",
			mark,
			ui.PadRight(ui.BundleTag(r.Name), nameWidth),
			ui.PadRight(ui.MutedStyle.Render(counts[i]), countWidth),
			ui.MutedStyle.Render(r.Description),
		)
	}
	fmt.Fprintln(out)
	return nil
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
