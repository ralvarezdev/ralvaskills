package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/termkit"

	"github.com/ralvarezdev/ralvaskills/v2/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/v2/internal/config"
	"github.com/ralvarezdev/ralvaskills/v2/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/v2/internal/skill"
	"github.com/ralvarezdev/ralvaskills/v2/internal/ui"
)

var listCmd = &cobra.Command{
	Use:   "list [flags]",
	Short: "Show installed skills.",
	Long: `List installed skills.

Without --global, shows the skills tracked in the current project's rsk.mod
(with install + pin marks). With --global, shows skills symlinked into the
configured global skills directories.

Examples:
  rsk list                              # project manifest contents
  rsk list --global                     # globally installed skills (all tools)
  rsk list --global --for claude-code   # one tool's global skills
  rsk list -o json`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runList(cmd, listOpts{
			global:  cmdx.Bool(cmd, cmdx.FlagGlobal),
			forTool: cmdx.String(cmd, cmdx.FlagFor),
			output:  outputFormat(cmdx.String(cmd, cmdx.FlagOutput)),
		})
	},
}

type (
	listOpts struct {
		global  bool
		forTool string
		output  outputFormat
	}

	// listedSkill is one row in the list output — used for both project-manifest
	// and global views and serialized as JSON when -o json is set.
	listedSkill struct {
		Name      string `json:"name"`
		Version   string `json:"version,omitempty"`
		Source    string `json:"source,omitempty"`
		Tool      string `json:"tool,omitempty"`
		Installed bool   `json:"installed"`
		Pinned    bool   `json:"pinned,omitempty"`
		Path      string `json:"path,omitempty"`
	}
)

func runList(cmd *cobra.Command, opts listOpts) error {
	if !opts.output.valid() {
		return fmt.Errorf("--output must be '%s' or '%s'", outputText, outputJSON)
	}
	if !opts.global && opts.forTool != "" {
		return errors.New("--for requires --global")
	}
	if opts.global {
		return runListGlobal(cmd, opts)
	}
	return runListProject(cmd, opts)
}

func runListProject(cmd *cobra.Command, opts listOpts) error {
	out := cmd.OutOrStdout()
	capture := ui.CaptureFromContext(cmd.Context())

	rskDir, err := manifest.ProjectFolderPath()
	if err != nil {
		return err
	}

	m, err := manifest.ReadMod(rskDir)
	if err != nil {
		return err
	}

	projectDirs := projectSkillsDirs(filepath.Dir(rskDir), m)
	pinnedSet := make(map[string]bool, len(m.Pinned))
	for _, p := range m.Pinned {
		pinnedSet[p] = true
	}

	names := make([]string, 0, len(m.Skills))
	for n := range m.Skills {
		names = append(names, n)
	}
	sort.Strings(names)

	rows := make([]listedSkill, 0, len(names))
	for _, name := range names {
		rows = append(rows, listedSkill{
			Name:      name,
			Version:   m.Skills[name],
			Installed: isLinkedAnywhere(name, projectDirs),
			Pinned:    pinnedSet[name],
		})
	}

	if len(rows) == 0 {
		if capture != nil {
			capture.AddMessage("No skills in manifest. Run 'rsk install <name>' to add one.")
			return nil
		}
		fmt.Fprintln(out)
		ui.Warn(out, "No skills in manifest. Run 'rsk install <name>' to add one.")
		fmt.Fprintln(out)
		return nil
	}

	if opts.output == outputJSON {
		return writeOrCaptureJSON(out, capture, rows)
	}

	if capture != nil {
		printProjectListTable(out, capture, rows)
		return nil
	}

	fmt.Fprintln(out)
	ui.Header(out, "Project skills:")
	printProjectListTable(out, capture, rows)
	fmt.Fprintln(out)
	return nil
}

func printProjectListTable(out io.Writer, capture *termkit.Capture, rows []listedSkill) {
	header := []string{"", headerName, headerVersion, ""}
	tableRows := termkit.Rows(rows, listProjectTable)
	if capture != nil {
		capture.AddTable(termkit.Data{Headers: header, Rows: tableRows})
		return
	}
	termkit.WriteTableStyled(out, header, tableRows, false, nil, true, false)
}

// listProjectTable projects one project-manifest row.
var listProjectTable = termkit.Table[listedSkill]{
	Row: func(r listedSkill) []any {
		mark := ui.SuccessMark
		if !r.Installed {
			mark = ui.ErrorMark
		}
		pinnedTag := ""
		if r.Pinned {
			pinnedTag = "[pinned]"
		}
		return []any{mark, ui.SkillName(r.Name), ui.SkillVersion(r.Version), pinnedTag}
	},
}

func runListGlobal(cmd *cobra.Command, opts listOpts) error {
	out := cmd.OutOrStdout()
	capture := ui.CaptureFromContext(cmd.Context())

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%w\n  Run 'rsk init' to set up rsk on this machine", err)
	}

	targets, err := resolveTargetDirs(cfg, true, opts.forTool)
	if err != nil {
		return err
	}

	// resolveTargetDirs returns bare directories, so map each one back to
	// its owning tool ID (via cfg.GlobalTargets) to label rows below —
	// otherwise a skill installed for multiple tools prints as indistinguishable
	// duplicate rows.
	dirToTool := make(map[string]string, len(cfg.GlobalTargets))
	for toolID, dir := range cfg.GlobalTargets {
		dirToTool[dir] = toolID
	}

	var rows []listedSkill
	for _, target := range targets {
		entries, scanErr := scanLinked(target, cfg.RepoPath, cfg.OfficialCache, cfg.RegistryCache(), nil, false)
		if scanErr != nil && !errors.Is(scanErr, os.ErrNotExist) {
			ui.Warn(out, fmt.Sprintf("scan %s: %v", target, scanErr))
			continue
		}
		for _, e := range entries {
			rows = append(rows, listedSkill{
				Name:      e.name,
				Version:   e.version,
				Source:    e.source.String(),
				Tool:      dirToTool[target],
				Installed: true,
			})
		}
	}

	if len(rows) == 0 {
		if capture != nil {
			capture.AddMessage("No global skills installed.")
			return nil
		}
		ui.Info(out, "No global skills installed.")
		return nil
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Tool != rows[j].Tool {
			return rows[i].Tool < rows[j].Tool
		}
		return rows[i].Name < rows[j].Name
	})

	if opts.output == outputJSON {
		return writeOrCaptureJSON(out, capture, rows)
	}

	if capture != nil {
		printGlobalListTable(out, capture, rows)
		return nil
	}

	fmt.Fprintln(out)
	ui.Header(out, "Global skills:")
	printGlobalListTable(out, capture, rows)
	fmt.Fprintln(out)
	return nil
}

func printGlobalListTable(out io.Writer, capture *termkit.Capture, rows []listedSkill) {
	header := []string{headerSource, "Tool", headerName, headerVersion}
	tableRows := termkit.Rows(rows, listGlobalTable)
	if capture != nil {
		capture.AddTable(termkit.Data{Headers: header, Rows: tableRows})
		return
	}
	termkit.WriteTableStyled(out, header, tableRows, false, nil, true, false)
}

// listGlobalTable projects one global-install row.
var listGlobalTable = termkit.Table[listedSkill]{
	Row: func(r listedSkill) []any {
		return []any{
			ui.SourceLabel(skill.Source(r.Source)),
			r.Tool,
			ui.SkillName(r.Name),
			ui.SkillVersion(r.Version),
		}
	},
}
