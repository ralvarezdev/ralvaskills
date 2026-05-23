package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ralvarezdev/ralvaskills/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/internal/config"
	"github.com/ralvarezdev/ralvaskills/internal/tool"
	"github.com/ralvarezdev/ralvaskills/internal/ui"
	"github.com/spf13/cobra"
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

Examples:
  rsk init
  rsk init --force`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit(cmd, cmdx.Bool(cmd, cmdx.FlagForce))
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().Bool(cmdx.FlagForce, false, "Overwrite existing config without prompting")
}

func runInit(cmd *cobra.Command, force bool) error {
	cfgPath := config.DefaultConfigFilePath()
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()

	if config.Exists() {
		if !force {
			ui.Fail(errOut, fmt.Sprintf("config already exists at %s", cfgPath))
			ui.Fail(errOut, "  Use --force to overwrite.")
			return fmt.Errorf("config already exists: %s", cfgPath)
		}
		ui.Warn(out, fmt.Sprintf("overwriting existing config at %s", cfgPath))
	}

	ui.Header(out, "rsk init")
	fmt.Fprintln(out)

	// 1. Skill source: local clone or hosted registry.
	modeIdx, err := ui.Select(out, "How do you want to fetch skills?",
		[]string{
			"Local clone  (repo already on disk)",
			"Registry     (download from skills.ralvarez.dev)",
		}, 0)
	if err != nil {
		return err
	}

	var repoPath, registryURL string
	if modeIdx == 1 {
		registryURL, err = ui.Ask(out, "Registry URL", config.DefaultRegistryURL)
		if err != nil {
			return err
		}
	} else {
		repoPath, err = ui.Ask(out, "Path to your local ralvaskills repo clone", "")
		if err != nil {
			return err
		}
		repoPath, err = expandPath(repoPath)
		if err != nil {
			return fmt.Errorf("invalid repo path: %w", err)
		}
		if _, statErr := os.Stat(repoPath); os.IsNotExist(statErr) {
			ui.Warn(out, fmt.Sprintf(
				"path does not exist: %s (saved anyway — create it before running rsk install)",
				repoPath,
			))
		}
	}

	// 2. AI tool selection.
	tools := tool.All()
	toolChoices := make([]string, len(tools))
	for i, t := range tools {
		toolChoices[i] = fmt.Sprintf("%s  (default: %s)", t.ID(), t.SkillsDir())
	}
	allIndices := make([]int, len(tools))
	for i := range tools {
		allIndices[i] = i
	}
	selectedIndices, err := ui.MultiSelect(out, "Which AI tools do you want to support?", toolChoices, allIndices)
	if err != nil {
		return err
	}
	if len(selectedIndices) == 0 {
		return fmt.Errorf("select at least one AI tool")
	}

	// 3. Global skills directories.
	globalTargets := make(map[string]string, len(selectedIndices))
	for _, idx := range selectedIndices {
		t := tools[idx]
		dir, err := ui.Ask(out, fmt.Sprintf("Global skills dir for %s", t.ID()), t.SkillsDir())
		if err != nil {
			return err
		}
		expanded, err := expandPath(dir)
		if err != nil {
			return fmt.Errorf("invalid path for %s: %w", t.ID(), err)
		}
		globalTargets[string(t.ID())] = expanded
	}

	// 4. Default target scope.
	scopeOptions := make([]string, 0, len(selectedIndices)+1)
	scopeOptions = append(scopeOptions, cmdx.ForAll)
	for _, idx := range selectedIndices {
		scopeOptions = append(scopeOptions, string(tools[idx].ID()))
	}
	scopeIdx, err := ui.Select(out, "Default scope when --global is passed without --for", scopeOptions, 0)
	if err != nil {
		return err
	}
	defaultScope := scopeOptions[scopeIdx]

	// Build and save config.
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
	ui.Success(out, fmt.Sprintf("Config written to %s", cfgPath))
	fmt.Fprintln(out)
	ui.Info(out, "Next steps:")
	ui.Indent(out, "rsk install global --global   # install universal skills machine-wide")
	ui.Indent(out, "rsk new                       # initialize an rsk project in the current directory")
	ui.Indent(out, "rsk install <name>            # install a bundle or skill (e.g. go-grpc, gin, fastapi)")
	return nil
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
