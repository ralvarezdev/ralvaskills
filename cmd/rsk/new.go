package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/termkit"

	"github.com/ralvarezdev/ralvaskills/v3/internal"
	"github.com/ralvarezdev/ralvaskills/v3/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/v3/internal/config"
	"github.com/ralvarezdev/ralvaskills/v3/internal/install"
	"github.com/ralvarezdev/ralvaskills/v3/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/v3/internal/source"
	"github.com/ralvarezdev/ralvaskills/v3/internal/tool"
	"github.com/ralvarezdev/ralvaskills/v3/internal/ui"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Initialize an rsk project in the current directory.",
	Long: `Create .rsk/rsk.mod and tool-specific config in the current directory so
this folder becomes an rsk project. The .claude/skills/ directory is created
lazily the first time a skill is installed.

Use --for to select which tools to configure:
  --for claude-code  (default) Appends @.rsk/CLAUDE.md to ./CLAUDE.md
  --for opencode               Writes pinned skills to opencode.json instructions
  --for all                    Both tools

Examples:
  rsk new
  rsk new --for opencode
  rsk new --for all`,
	RunE: runNew,
}

func toolsFromFlag(flag string) ([]tool.ID, error) {
	scope, err := cmdx.ParseTargetScope(flag)
	if err != nil {
		return nil, fieldError(cmdx.FlagFor, "%v", err)
	}
	switch scope {
	case cmdx.TargetScope(tool.ClaudeID):
		return []tool.ID{tool.ClaudeID}, nil
	case cmdx.TargetScope(tool.OpenCodeID):
		return []tool.ID{tool.OpenCodeID}, nil
	case cmdx.ScopeAll:
		return []tool.ID{tool.ClaudeID, tool.OpenCodeID}, nil
	default:
		return nil, fieldError(cmdx.FlagFor, "--for must be %s, %s, or %s; got %q",
			tool.ClaudeID, tool.OpenCodeID, cmdx.ForAll, flag)
	}
}

func runNew(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	forFlag, err := resolveForFlag(cmd, out)
	if err != nil {
		return err
	}

	tools, err := toolsFromFlag(forFlag)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	m, rskDir, isNew, err := loadOrCreateManifest(cwd, tools, forFlag, out)
	if err != nil || m == nil {
		return err
	}

	if err = manifest.WriteMod(rskDir, *m); err != nil {
		return err
	}
	if err = syncToolConfigs(tools, cwd, m.Pinned); err != nil {
		return err
	}

	if resolveMCPFlag(cmd) {
		if err = registerProjectMCP(out, cwd, tools); err != nil {
			return err
		}
	}

	if resolveGuideFlag(cmd) {
		if err = installRskGuide(cmd.Context(), out, cwd, rskDir); err != nil {
			return err
		}
	}

	fmt.Fprintln(out)
	if isNew {
		ui.Success(out, "initialized .rsk/ for "+forFlag)
	} else {
		ui.Success(out, "updated .rsk/ for "+forFlag)
	}
	ui.Indent(out, "Run 'rsk install <name>' to add skills to this project.")
	fmt.Fprintln(out)
	return nil
}

// resolveForFlag returns --for's value, prompting interactively when it
// wasn't set on the command line.
func resolveForFlag(cmd *cobra.Command, out io.Writer) (string, error) {
	forFlag := cmdx.String(cmd, cmdx.FlagFor)
	if cmd.Flags().Changed(cmdx.FlagFor) {
		return forFlag, nil
	}
	if ui.InSession(cmd.Context()) {
		return "", fieldError(cmdx.FlagFor, "choose which tools to configure")
	}

	choices := []string{string(tool.ClaudeID), string(tool.OpenCodeID), cmdx.ForAll}
	idx, err := ui.Select(out, "Tools to configure", choices, 0)
	if err != nil {
		return "", err
	}
	return choices[idx], nil
}

// loadOrCreateManifest reads the existing project manifest and adds any of
// tools it's missing, or builds a fresh one if no project exists yet. A nil
// *manifest.Mod with a nil error means the project is already configured for
// every requested tool and the caller should return without further work.
func loadOrCreateManifest(
	cwd string, tools []tool.ID, forFlag string, out io.Writer,
) (m *manifest.Mod, rskDir string, isNew bool, err error) {
	existing, existErr := manifest.ProjectFolderPath()
	if existErr == nil {
		loaded, readErr := manifest.ReadMod(existing)
		if readErr != nil {
			return nil, "", false, readErr
		}

		var added []tool.ID
		for _, t := range tools {
			if !slices.Contains(loaded.Tools, t) {
				added = append(added, t)
				loaded.Tools = append(loaded.Tools, t)
			}
		}
		if len(added) == 0 {
			ui.Info(out, "rsk project already configured for "+forFlag)
			return &loaded, existing, false, nil
		}

		ui.Info(out, fmt.Sprintf("Adding tools to existing project: %v", added))
		return &loaded, existing, false, nil
	}

	// No project yet — existErr just signals "not found", nothing to surface.
	return &manifest.Mod{
		Tools:  tools,
		Skills: make(map[string]string),
		Pinned: []string{},
	}, filepath.Join(cwd, internal.ProjectFolderName), true, nil
}

// syncToolConfigs pushes the currently pinned skills into each tool's
// project config (CLAUDE.md / opencode.json), skipping any ID that isn't a
// registered tool.
func syncToolConfigs(tools []tool.ID, cwd string, pinned []string) error {
	for _, id := range tools {
		t, ok := tool.Get(id)
		if !ok {
			continue
		}
		if err := t.SyncPinned(cwd, pinned); err != nil {
			return err
		}
	}
	return nil
}

// flagMCP is the boolean flag that registers the rsk MCP server without a
// prompt, for non-interactive use.
const flagMCP = "mcp"

// resolveMCPFlag returns whether to register the rsk MCP server: the --mcp flag
// when set, an interactive prompt otherwise.
func resolveMCPFlag(cmd *cobra.Command) bool {
	if cmd.Flags().Changed(flagMCP) {
		return cmdx.Bool(cmd, flagMCP)
	}
	if ui.InSession(cmd.Context()) {
		return false
	}
	return termkit.Confirm(cmd, false, termkit.StdioIsTerminal(), "Register the rsk MCP server in this project?")
}

// registerProjectMCP writes the rsk server into each configured tool's project
// config and, for Claude Code, the workflow pointer in ./CLAUDE.md. It skips
// (with a warning) when rsk is not on PATH, since the client would fail to
// launch the server.
func registerProjectMCP(out io.Writer, cwd string, tools []tool.ID) error {
	if _, err := exec.LookPath("rsk"); err != nil {
		ui.Warn(out, "'rsk' is not on PATH — skipping MCP registration. Put rsk on PATH "+
			"first, or the client will fail to launch the server.")
		//nolint:nilerr // not being on PATH is a skip with a warning, not a failure
		return nil
	}

	paths := make([]string, 0, len(tools))
	for _, id := range tools {
		t, ok := tool.Get(id)
		if !ok {
			continue
		}
		path, err := t.RegisterMCP(cwd)
		if err != nil {
			return err
		}
		paths = append(paths, path)
	}

	if slices.Contains(tools, tool.ClaudeID) {
		if err := tool.WriteMCPPointer(cwd); err != nil {
			return err
		}
	}

	ui.Success(out, "registered the rsk MCP server ("+strings.Join(paths, ", ")+")")
	return nil
}

// flagGuide is the boolean flag that installs rsk-guide without a prompt.
const flagGuide = "guide"

// resolveGuideFlag returns whether to install rsk-guide: the --guide flag when
// set, an interactive prompt otherwise.
func resolveGuideFlag(cmd *cobra.Command) bool {
	if cmd.Flags().Changed(flagGuide) {
		return cmdx.Bool(cmd, flagGuide)
	}
	if ui.InSession(cmd.Context()) {
		return false
	}
	return termkit.Confirm(cmd, false, termkit.StdioIsTerminal(), "Install the rsk-guide skill in this project?")
}

// installRskGuide installs rsk-guide into the project and pins it, so the agent
// loads the rsk policy automatically. It is best-effort: a missing config or an
// unresolvable skill warns instead of failing the whole command.
func installRskGuide(ctx context.Context, out io.Writer, cwd, rskDir string) error {
	cfg, err := config.Load()
	if err != nil {
		ui.Warn(out, "skipping rsk-guide install: "+err.Error()+"\n  Run 'rsk init' first.")
		//nolint:nilerr // rsk not initialized is a skip with a warning, not a failure
		return nil
	}

	bundles, catalogErr := config.LoadCatalog("")
	if catalogErr != nil {
		ui.Warn(out, "user catalog: "+catalogErr.Error())
	}

	skills, _, resolveErr := install.Resolve(
		ctx, []string{"rsk-guide"}, bundles,
		install.LocalSource(cfg),
		source.NewOfficial(cfg.OfficialCache),
	)
	if resolveErr != nil {
		ui.Warn(out, "could not resolve rsk-guide: "+resolveErr.Error())
		//nolint:nilerr // best-effort convenience install; the project is still set up
		return nil
	}

	m, err := manifest.ReadMod(rskDir)
	if err != nil {
		return err
	}
	results, err := install.Apply(skills, install.Options{
		Scope:   install.ScopeProject,
		Targets: install.ProjectSkillsDirs(cwd, m),
		RskDir:  rskDir,
		Pin:     true,
	})
	if err != nil {
		return err
	}
	for _, r := range results {
		if r.Err != nil {
			ui.Warn(out, "rsk-guide: "+r.Err.Error())
			continue
		}
		ui.Success(out, "installed and pinned rsk-guide")
	}
	return nil
}
