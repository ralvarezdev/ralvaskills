package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ralvarezdev/ralvaskills/cli/internal/config"
	"github.com/ralvarezdev/ralvaskills/cli/internal/ui"
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
	RunE: runInit,
}

var initForce bool

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&initForce, "force", false, "Overwrite existing config without prompting")
}

func runInit(cmd *cobra.Command, _ []string) error {
	cfgPath := config.DefaultPath()
	out := cmd.OutOrStdout()

	if config.Exists() {
		if !initForce {
			ui.Fail(fmt.Sprintf("config already exists at %s", cfgPath))
			ui.Fail("  Use --force to overwrite.")
			return fmt.Errorf("config already exists: %s", cfgPath)
		}
		ui.Warn(out, fmt.Sprintf("overwriting existing config at %s", cfgPath))
	}

	r := bufio.NewReader(os.Stdin)

	ui.Header(out, "rsk init")
	fmt.Fprintln(out)

	// 1. Skill source: local clone or hosted registry.
	fmt.Fprintln(out, "How do you want to fetch local skills?")
	fmt.Fprintln(out, "  [1] Local clone  (repo already on disk)")
	fmt.Fprintln(out, "  [2] Registry     (download from skills.ralvarez.dev)")
	modeStr, err := prompt(r, out, "Select", "1")
	if err != nil {
		return err
	}

	var repoPath, registryURL string
	if parseIntOrDefault(modeStr, 1) == 2 {
		registryURL, err = prompt(r, out, "Registry URL", "https://skills.ralvarez.dev")
		if err != nil {
			return err
		}
	} else {
		repoPath, err = prompt(r, out, "Path to your local ralvaskills repo clone", "")
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

	// 2. AI tool selection
	type toolDef struct {
		id         string
		defaultDir string
	}
	tools := []toolDef{
		{id: "claude-code", defaultDir: claudeSkillsDir()},
		{id: "opencode", defaultDir: openCodeSkillsDir()},
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Which AI tools do you want to support?")
	for i, t := range tools {
		fmt.Fprintf(out, "  [%d] %s  (default: %s)\n", i+1, t.id, t.defaultDir)
	}
	selStr, err := prompt(r, out, "Select (comma-separated)", "1,2")
	if err != nil {
		return err
	}
	selected := parseMultiSelect(selStr, len(tools))
	if len(selected) == 0 {
		return fmt.Errorf("select at least one AI tool")
	}

	// 3. Global skills directories
	globalTargets := make(map[string]string, len(selected))
	fmt.Fprintln(out)
	var (
		dir      string
		expanded string
	)
	for _, idx := range selected {
		t := tools[idx]
		dir, err = prompt(r, out, fmt.Sprintf("Global skills dir for %s", t.id), t.defaultDir)
		if err != nil {
			return err
		}
		expanded, err = expandPath(dir)
		if err != nil {
			return fmt.Errorf("invalid path for %s: %w", t.id, err)
		}
		globalTargets[t.id] = expanded
	}

	// 4. Default target scope
	scopeOptions := make([]string, 0, len(selected)+1)
	scopeOptions = append(scopeOptions, "all")
	for _, idx := range selected {
		scopeOptions = append(scopeOptions, tools[idx].id)
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Default scope when --global is passed without --for:")
	for i, s := range scopeOptions {
		fmt.Fprintf(out, "  [%d] %s\n", i+1, s)
	}
	scopeStr, err := prompt(r, out, "Select", "1")
	if err != nil {
		return err
	}
	scopeIdx := clampInt(parseIntOrDefault(scopeStr, 1), 1, len(scopeOptions))
	defaultScope := scopeOptions[scopeIdx-1]

	// Build and save config.
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home dir: %w", err)
	}
	cfg := config.Config{
		RepoPath:           repoPath,
		RegistryURL:        registryURL,
		GlobalTargets:      globalTargets,
		DefaultTargetScope: defaultScope,
		OfficialCache:      filepath.Join(home, ".ralvaskills", "cache", "anthropic"),
		VersionsCache:      filepath.Join(home, ".ralvaskills", "cache", "versions.json"),
	}
	err = config.Save(cfg)
	if err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Fprintln(out)
	ui.Success(out, fmt.Sprintf("Config written to %s", cfgPath))
	fmt.Fprintln(out)
	ui.Info(out, "Next steps:")
	ui.Indent(out, "rsk install global --global   # install universal skills machine-wide")
	ui.Indent(out, "rsk install <bundle>          # install skills for a project (e.g. go-grpc, gin, fastapi)")
	return nil
}

// prompt prints a labeled prompt to w and reads a trimmed line from r.
// If the user presses enter with no input, defaultVal is returned.
func prompt(r *bufio.Reader, w io.Writer, label, defaultVal string) (string, error) {
	if defaultVal != "" {
		fmt.Fprintf(w, "%s [%s]: ", label, defaultVal)
	} else {
		fmt.Fprintf(w, "%s: ", label)
	}
	line, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read input: %w", err)
	}
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return defaultVal, nil
	}
	return line, nil
}

// parseMultiSelect parses a comma-separated list of 1-based indices into 0-based indices.
func parseMultiSelect(s string, count int) []int {
	seen := make(map[int]bool)
	var out []int
	for _, part := range strings.Split(s, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || n < 1 || n > count || seen[n] {
			continue
		}
		seen[n] = true
		out = append(out, n-1)
	}
	return out
}

func parseIntOrDefault(s string, def int) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return def
	}
	return n
}

func clampInt(n, lo, hi int) int {
	if n < lo {
		return lo
	}
	if n > hi {
		return hi
	}
	return n
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

func claudeSkillsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".claude", "skills")
	}
	return filepath.Join(home, ".claude", "skills")
}

func openCodeSkillsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".config", "opencode", "skills")
	}
	return filepath.Join(home, ".config", "opencode", "skills")
}
