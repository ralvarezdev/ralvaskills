package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/ralvarezdev/ralvaskills/cli/internal/config"
	"github.com/ralvarezdev/ralvaskills/cli/internal/skill"
	"github.com/ralvarezdev/ralvaskills/cli/internal/source"
	"github.com/ralvarezdev/ralvaskills/cli/internal/ui"
	"github.com/spf13/cobra"
)

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
	RunE: runList,
}

var (
	listBundle    string
	listSource    string
	listInstalled bool
	listPersonal  bool
	listOutput    string
)

func init() {
	rootCmd.AddCommand(listCmd)
	f := listCmd.Flags()
	f.StringVar(&listBundle, "bundle", "", "Show skills in a specific bundle")
	f.StringVar(&listSource, "source", "", "Filter by source: local | official")
	f.BoolVar(&listInstalled, "installed", false, "Show only installed skills")
	f.BoolVar(&listPersonal, "personal", false, "Include personal/ skills in listing")
	f.StringVarP(&listOutput, "output", "o", "text", "Output format: text | json")
}

func runList(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	if listSource != "" && listSource != "local" && listSource != "official" {
		return fmt.Errorf("--source must be 'local' or 'official'")
	}
	if listOutput != "text" && listOutput != "json" {
		return fmt.Errorf("--output must be 'text' or 'json'")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n  Run 'rsk init' to set up rsk on this machine", err)
	}

	localSrc := newLocalSource(cfg)
	officialSrc := source.NewOfficial(cfg.OfficialCache)

	var all []skill.Skill

	if listSource == "" || listSource == "local" {
		local, walkErr := localSrc.All()
		if walkErr != nil {
			ui.Warn(out, fmt.Sprintf("walk local skills: %v", walkErr))
		} else {
			all = append(all, local...)
		}
	}

	if listSource == "" || listSource == "official" {
		official, walkErr := officialSrc.All()
		if walkErr != nil {
			if !os.IsNotExist(walkErr) {
				ui.Warn(out, fmt.Sprintf("walk official skills: %v", walkErr))
			}
		} else {
			all = append(all, official...)
		}
	}

	// --bundle filter: keep only skills that appear in that bundle.
	if listBundle != "" {
		catalog := config.LoadCatalog("")
		bundle, ok := config.FindBundle(catalog, listBundle)
		if !ok {
			return fmt.Errorf("bundle %q not found — run 'rsk list' to see all bundles", listBundle)
		}
		want := make(map[string]bool, len(bundle.Skills))
		for _, ref := range bundle.Skills {
			want[ref.Name] = true
		}
		filtered := all[:0]
		for _, s := range all {
			if want[s.Name] {
				filtered = append(filtered, s)
			}
		}
		all = filtered
	}

	// --personal filter: drop personal skills unless opted in.
	if !listPersonal {
		filtered := all[:0]
		for _, s := range all {
			if !s.IsPersonal {
				filtered = append(filtered, s)
			}
		}
		all = filtered
	}

	// --installed filter: keep only skills linked in any configured target dir.
	if listInstalled {
		targets := allTargetDirs(cfg)
		filtered := all[:0]
		for _, s := range all {
			for _, t := range targets {
				if skill.IsLinked(s.Name, t) {
					filtered = append(filtered, s)
					break
				}
			}
		}
		all = filtered
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

	if listOutput == "json" {
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

	for _, s := range skills {
		ver := s.Version
		if ver == "" {
			ver = "-"
		}
		fmt.Fprintf(out, "  %s  %s  %s\n",
			s.Source.Label(),
			ui.PadRight(s.Name, nameWidth),
			ver,
		)
	}
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
