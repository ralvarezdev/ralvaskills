package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [bundle...] [flags]",
	Short: "Install a bundle or skill.",
	Long: `Install one or more bundles or a single skill by name.

By default installs to ./.claude/skills/ (project-local).
Use --global to install to the configured global skills directories.

Examples:
  rsk install go-grpc
  rsk install design docs
  rsk install global --global
  rsk install --skill go-architect
  rsk install go-grpc --global --for claude-code
  rsk install --skill demo-script-architect --personal
  rsk install go-grpc --dry-run`,
	RunE: runInstall,
}

var (
	installGlobal   bool
	installFor      string
	installSkill    string
	installPersonal bool
	installVersion  string
	installDryRun   bool
)

func init() {
	rootCmd.AddCommand(installCmd)
	f := installCmd.Flags()
	f.BoolVar(&installGlobal, "global", false, "Install to global skills dir(s)")
	f.StringVar(&installFor, "for", "", "Scope --global to a single tool (claude-code|opencode)")
	f.StringVar(&installSkill, "skill", "", "Install a single skill by name (skips bundle resolution)")
	f.BoolVar(&installPersonal, "personal", false, "Allow installing personal/ skills")
	f.StringVar(&installVersion, "version", "", "Pin to a specific repo tag (local skills only)")
	f.BoolVar(&installDryRun, "dry-run", false, "Show what would be installed without doing it")
}

func runInstall(_ *cobra.Command, _ []string) error {
	return fmt.Errorf("not yet implemented — coming soon")
}
