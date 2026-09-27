package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/ralvaskills/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/internal/config"
	"github.com/ralvarezdev/ralvaskills/internal/tool"
	"github.com/ralvarezdev/ralvaskills/internal/ui"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up rsk for this machine.",
	Long: `Set up rsk for this machine by creating ~/.config/rsk/config.json.

Run once per machine. Prompts for:
  - Path to your local ralvaskills repo clone
  - Which AI tools to support (claude-code, opencode)
  - Global skills directory for each tool
  - Default install scope for --global

Set $RSK_CONFIG_HOME to sandbox the entire rsk config dir (config.json,
catalog.toml, caches) elsewhere — useful for testing without touching your
real machine config.

Examples:
  rsk init
  rsk init --force`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit(cmd, cmdx.Bool(cmd, cmdx.FlagForce))
	},
}

func runInit(cmd *cobra.Command, force bool) error {
	cfgPath := config.DefaultConfigFilePath()
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()

	if err := checkNoExistingConfig(out, errOut, cfgPath, force); err != nil {
		return err
	}

	ui.Header(out, "rsk init")
	fmt.Fprintln(out)

	repoPath, registryURL, err := promptSkillSource(out)
	if err != nil {
		return err
	}

	tools, selectedIndices, err := promptToolSelection(out)
	if err != nil {
		return err
	}

	globalTargets, err := promptGlobalTargets(out, tools, selectedIndices)
	if err != nil {
		return err
	}

	defaultScope, err := promptDefaultScope(out, tools, selectedIndices)
	if err != nil {
		return err
	}

	configDir := config.DefaultConfigFolderPath()
	cfg := config.Config{
		RepoPath:           repoPath,
		RegistryURL:        registryURL,
		GlobalTargets:      globalTargets,
		DefaultTargetScope: defaultScope,
		OfficialCache:      filepath.Join(configDir, "cache", "anthropic"),
		VersionsCache:      filepath.Join(configDir, "cache", "versions.json"),
	}
	if err = config.Save(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Fprintln(out)
	ui.Success(out, "Config written to "+cfgPath)
	fmt.Fprintln(out)
	ui.Info(out, "Next steps:")
	ui.Indent(out, "rsk install global --global   # install universal skills machine-wide")
	ui.Indent(out, "rsk new                       # initialize an rsk project in the current directory")
	ui.Indent(out, "rsk install <name>            # install a bundle or skill (e.g. go-grpc, gin, fastapi)")
	return nil
}

// checkNoExistingConfig fails with guidance when a config already exists and
// force wasn't passed; with force, it just warns that init will overwrite it.
func checkNoExistingConfig(out, errOut io.Writer, cfgPath string, force bool) error {
	if !config.Exists() {
		return nil
	}
	if !force {
		ui.Fail(errOut, "config already exists at "+cfgPath)
		ui.Fail(errOut, "  Use --force to overwrite.")
		return fmt.Errorf("config already exists: %s", cfgPath)
	}
	ui.Warn(out, "overwriting existing config at "+cfgPath)
	return nil
}

// promptSkillSource asks whether skills come from a local repo clone or the
// hosted registry, returning whichever of repoPath/registryURL applies.
func promptSkillSource(out io.Writer) (repoPath, registryURL string, err error) {
	modeIdx, err := ui.Select(out, "How do you want to fetch skills?",
		[]string{
			"Local clone  (repo already on disk)",
			"Registry     (download from skills.ralvarez.dev)",
		}, 0)
	if err != nil {
		return "", "", err
	}

	if modeIdx == 1 {
		registryURL, err = ui.Ask(out, "Registry URL", config.DefaultRegistryURL)
		return "", registryURL, err
	}

	repoPath, err = ui.Ask(out, "Path to your local ralvaskills repo clone", "")
	if err != nil {
		return "", "", err
	}
	repoPath, err = expandPath(repoPath)
	if err != nil {
		return "", "", fmt.Errorf("invalid repo path: %w", err)
	}
	if _, statErr := os.Stat(repoPath); os.IsNotExist(statErr) {
		ui.Warn(out, fmt.Sprintf(
			"path does not exist: %s (saved anyway — create it before running rsk install)",
			repoPath,
		))
	}
	return repoPath, "", nil
}

// promptToolSelection lets the user pick which of tool.All() to support,
// requiring at least one.
func promptToolSelection(out io.Writer) (tools []tool.Tool, selectedIndices []int, err error) {
	tools = tool.All()
	toolChoices := make([]string, len(tools))
	for i, t := range tools {
		toolChoices[i] = fmt.Sprintf("%s  (default: %s)", t.ID(), t.SkillsDir())
	}
	allIndices := make([]int, len(tools))
	for i := range tools {
		allIndices[i] = i
	}
	selectedIndices, err = ui.MultiSelect(out, "Which AI tools do you want to support?", toolChoices, allIndices)
	if err != nil {
		return nil, nil, err
	}
	if len(selectedIndices) == 0 {
		return nil, nil, errors.New("select at least one AI tool")
	}
	return tools, selectedIndices, nil
}

// promptGlobalTargets asks for the global skills directory of each selected
// tool, expanding any leading ~.
func promptGlobalTargets(out io.Writer, tools []tool.Tool, selectedIndices []int) (map[string]string, error) {
	globalTargets := make(map[string]string, len(selectedIndices))
	for _, idx := range selectedIndices {
		t := tools[idx]
		dir, err := ui.Ask(out, fmt.Sprintf("Global skills dir for %s", t.ID()), t.SkillsDir())
		if err != nil {
			return nil, err
		}
		expanded, err := expandPath(dir)
		if err != nil {
			return nil, fmt.Errorf("invalid path for %s: %w", t.ID(), err)
		}
		globalTargets[string(t.ID())] = expanded
	}
	return globalTargets, nil
}

// promptDefaultScope asks which tool (or "all") --global should target by
// default when --for is omitted.
func promptDefaultScope(out io.Writer, tools []tool.Tool, selectedIndices []int) (string, error) {
	scopeOptions := make([]string, 0, len(selectedIndices)+1)
	scopeOptions = append(scopeOptions, cmdx.ForAll)
	for _, idx := range selectedIndices {
		scopeOptions = append(scopeOptions, string(tools[idx].ID()))
	}
	scopeIdx, err := ui.Select(out, "Default scope when --global is passed without --for", scopeOptions, 0)
	if err != nil {
		return "", err
	}
	return scopeOptions[scopeIdx], nil
}

// expandPath expands a leading ~ to the user home directory.
func expandPath(p string) (string, error) {
	if !strings.HasPrefix(p, "~") {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, p[1:]), nil
}
