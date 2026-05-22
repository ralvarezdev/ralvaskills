package cmd

import (
	"fmt"

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
  rsk list --personal`,
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
	f.StringVarP(&listOutput, "output", "o", "text", "Output format: text | json | yaml")
}

func runList(_ *cobra.Command, _ []string) error {
	return fmt.Errorf("not yet implemented — coming soon")
}
