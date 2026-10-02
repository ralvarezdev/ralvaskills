package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/ralvaskills/v3/internal/config"
	"github.com/ralvarezdev/ralvaskills/v3/internal/mcp"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run the MCP server over stdio.",
	Long: `Run the rsk Model Context Protocol server over stdio.

The server exposes project profiling and skill search to an MCP client such as
Claude Code or opencode. It is launched by the client, not run by hand, and
speaks JSON-RPC on stdout — logs go to stderr.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("%w\n  Run 'rsk init' to set up rsk on this machine", err)
		}
		logger := slog.New(slog.NewTextHandler(cmd.ErrOrStderr(), nil))
		srv := mcp.NewServer(mcp.Deps{Cfg: cfg, Logger: logger, Version: version})
		return mcp.ServeStdio(cmd.Context(), srv)
	},
}
