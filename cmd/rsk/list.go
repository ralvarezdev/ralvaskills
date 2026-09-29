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
	"github.com/ralvarezdev/ralvaskills/v2/internal/tool"
	"github.com/ralvarezdev/ralvaskills/v2/internal/ui"
)

var listCmd = &cobra.Command{
	Use:   "list [flags]",
	Short: "Show installed skills (project / --global).",
	Long: `List installed skills (project / --global). This is the offline view of what
the manifest and the skills directories contain; for version drift see
'rsk status'.

Without --global, shows the skills tracked in the current project's rsk.mod
(with install + pin marks). With --global, shows skills symlinked into the
configured global skills directories.

Examples:
  rsk list                              # project manifest contents
  rsk list --global                     # globally installed skills (all tools)
  rsk list --global --for claude-code   # one tool's global skills
  rsk list -o json`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// --include personal is accepted (and validated) for symmetry; list
		// has never filtered on it.
		if _, err := cmdx.ReadIncludes(cmd, cmdx.IncludePersonal); err != nil {
			return err
		}
		forTool, err := forToolFlag(cmd)
		if err != nil {
			return err
		}
		output, err := parseOutputFormat(cmdx.String(cmd, cmdx.FlagOutput))
		if err != nil {
			return err
		}
		return runList(cmd, listOpts{
			global:  cmdx.Bool(cmd, cmdx.FlagGlobal),
			forTool: forTool,
			output:  output,
		})
	},
}

type (
	listOpts struct {
		global  bool
		forTool tool.ID
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
	capture := termkit.CaptureFromContext(cmd.Context())

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
		printProjectListTable(out, capture, cmd, rows)
		return nil
	}

	fmt.Fprintln(out)
	ui.Header(out, "Project skills:")
	printProjectListTable(out, capture, cmd, rows)
	fmt.Fprintln(out)
	return nil
}

func printProjectListTable(out io.Writer, capture *termkit.Capture, cmd *cobra.Command, rows []listedSkill) {
	header := []string{"", headerName, headerVersion, ""}
	tableRows := termkit.Rows(rows, listProjectTable)
	if capture != nil {
		capture.AddTable(termkit.Data{
			Headers: header,
			Rows:    tableRows,
			IDs:     listedSkillNames(rows),
			Actions: ui.RowActionsFor(cmd),
		})
		return
	}
	termkit.WriteTableStyled(out, header, tableRows, false, nil, termkit.TableBorderless, ui.Theme)
}

// listedSkillNames extracts each row's skill name, parallel to
// listProjectTable/listGlobalTable's rows, so a captured table can identify
// the row a RowAction fires on independently of how it's displayed.
func listedSkillNames(rows []listedSkill) []string {
	names := make([]string, len(rows))
	for i, r := range rows {
		names[i] = r.Name
	}
	return names
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
	capture := termkit.CaptureFromContext(cmd.Context())

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
		printGlobalListTable(out, capture, cmd, rows)
		return nil
	}

	fmt.Fprintln(out)
	ui.Header(out, "Global skills:")
	printGlobalListTable(out, capture, cmd, rows)
	fmt.Fprintln(out)
	return nil
}

func printGlobalListTable(out io.Writer, capture *termkit.Capture, cmd *cobra.Command, rows []listedSkill) {
	header := []string{headerSource, "Tool", headerName, headerVersion}
	tableRows := termkit.Rows(rows, listGlobalTable)
	if capture != nil {
		capture.AddTable(termkit.Data{
			Headers: header,
			Rows:    tableRows,
			IDs:     listedSkillNames(rows),
			Actions: ui.RowActionsFor(cmd),
		})
		return
	}
	termkit.WriteTableStyled(out, header, tableRows, false, nil, termkit.TableBorderless, ui.Theme)
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
