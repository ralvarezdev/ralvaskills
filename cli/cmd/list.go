package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/ralvarezdev/ralvaskills/cli/internal"
	"github.com/ralvarezdev/ralvaskills/cli/internal/config"
	"github.com/ralvarezdev/ralvaskills/cli/internal/skill"
	"github.com/ralvarezdev/ralvaskills/cli/internal/source"
	"github.com/ralvarezdev/ralvaskills/cli/internal/ui"
	"github.com/spf13/cobra"
)

type listOpts struct {
	installed, personal bool
	bundle, source      string
	output              internal.OutputFormat
}

var listCmd = &cobra.Command{
	Use:   "list [flags]",
	Short: "Browse the skill catalog.",
	Long: `List all available skills and bundles.

Examples:
  rsk list
  rsk list --bundle go-grpc
  rsk list --source local
  rsk list --installed
  rsk list --personal
  rsk list -o json`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runList(cmd, listOpts{
			installed: internal.FlagBool(cmd, internal.FlagInstalled),
			personal:  internal.FlagBool(cmd, internal.FlagPersonal),
			bundle:    internal.FlagString(cmd, internal.FlagBundle),
			source:    internal.FlagString(cmd, internal.FlagSource),
			output:    internal.OutputFormat(internal.FlagString(cmd, internal.FlagOutput)),
		})
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	f := listCmd.Flags()
	f.String(internal.FlagBundle, "", "Show skills in a specific bundle")
	f.String(internal.FlagSource, "", "Filter by source: local | official")
	f.Bool(internal.FlagInstalled, false, "Show only installed skills")
	f.Bool(internal.FlagPersonal, false, "Include personal/ skills in listing")
	f.StringP(internal.FlagOutput, "o", string(internal.OutputText), "Output format: text | json")
}

func runList(cmd *cobra.Command, opts listOpts) error {
	out := cmd.OutOrStdout()
	ctx := cmd.Context()

	if opts.source != "" && opts.source != skill.SourceLocal.String() && opts.source != skill.SourceOfficial.String() {
		return fmt.Errorf("--source must be '%s' or '%s'", skill.SourceLocal, skill.SourceOfficial)
	}
	if !opts.output.Valid() {
		return fmt.Errorf("--output must be '%s' or '%s'", internal.OutputText, internal.OutputJSON)
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
			if !os.IsNotExist(walkErr) {
				ui.Warn(out, fmt.Sprintf("walk official skills: %v", walkErr))
			}
		} else {
			all = append(all, official...)
		}
	}

	// --bundle filter: keep only skills that appear in that bundle.
	if opts.bundle != "" {
		catalog, catalogWarn := config.LoadCatalog("")
		if catalogWarn != nil {
			ui.Warn(out, fmt.Sprintf("user catalog: %v", catalogWarn))
		}
		bundle, ok := config.FindBundle(catalog, opts.bundle)
		if !ok {
			return fmt.Errorf("bundle %q not found — run 'rsk list' to see all bundles", opts.bundle)
		}
		want := make(map[string]bool, len(bundle.Skills))
		for _, ref := range bundle.Skills {
			want[ref.Name] = true
		}
		all = filterSkills(all, func(s skill.Skill) bool { return want[s.Name] })
	}

	// --personal filter: drop personal skills unless opted in.
	if !opts.personal {
		all = filterSkills(all, func(s skill.Skill) bool { return !s.IsPersonal })
	}

	// --installed filter: keep only skills linked in any configured target dir.
	if opts.installed {
		targets := allTargetDirs(cfg)
		all = filterSkills(all, func(s skill.Skill) bool {
			for _, t := range targets {
				if skill.IsLinked(s.Name, t) {
					return true
				}
			}
			return false
		})
	}

	// Stable sort: official before local, then alphabetical within each.
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

	if opts.output == internal.OutputJSON {
		return writeJSON(out, skillsToEntries(all))
	}
	return printListTable(out, all)
}

func printListTable(out io.Writer, skills []skill.Skill) error {
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

type skillEntry struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Source   string `json:"source"`
	Personal bool   `json:"personal,omitempty"`
	Path     string `json:"path"`
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
