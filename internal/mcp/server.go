package mcp

import (
	"context"
	"fmt"
	"log/slog"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ralvarezdev/mcpkit"

	"github.com/ralvarezdev/ralvaskills/v3/internal/catalog"
	"github.com/ralvarezdev/ralvaskills/v3/internal/config"
	"github.com/ralvarezdev/ralvaskills/v3/internal/source"
)

// Deps is the environment the MCP server needs, injected so the server is
// testable without the CLI.
type Deps struct {
	Cfg     config.Config
	Logger  *slog.Logger
	Version string
}

// NewServer builds the MCP server with every tool and resource registered. The
// result is transport-independent: the caller decides whether to run it over
// stdio or in memory.
func NewServer(deps Deps) *sdk.Server {
	srv := sdk.NewServer(
		&sdk.Implementation{Name: "rsk", Title: "rsk", Version: deps.Version},
		&sdk.ServerOptions{Logger: deps.Logger},
	)
	srv.AddReceivingMiddleware(mcpkit.RecoverWith(mcpkit.RecoverConfig{Logger: deps.Logger}))
	registerTools(srv, deps)
	registerCatalogResource(srv, deps)
	return srv
}

// ServeStdio runs srv over stdin/stdout until the client closes the connection.
// Protocol frames go to stdout; logs must stay on stderr.
func ServeStdio(ctx context.Context, srv *sdk.Server) error {
	if err := srv.Run(ctx, &sdk.StdioTransport{}); err != nil {
		return fmt.Errorf("run mcp server: %w", err)
	}
	return nil
}

// catalog lists the skills available to this machine: from the local clone in
// local mode, or from the cached registry index otherwise.
func (d Deps) catalog(ctx context.Context) ([]catalog.Entry, error) {
	if d.Cfg.LocalMode() {
		return catalog.Build(ctx, source.NewLocal(d.Cfg.RepoPath))
	}

	reg := source.NewRegistry(d.Cfg.RegistryURL, d.Cfg.RegistryCache())
	snap, err := NewIndexCache(d.Cfg.RegistryCache(), DefaultIndexTTL, reg.Index).Load(ctx)
	if err != nil {
		return nil, err
	}
	return catalog.FromIndex(snap.Skills), nil
}
