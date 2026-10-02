// Package mcp implements rsk's Model Context Protocol server.
//
// This package starts with the part that carries the most design risk and the
// least infrastructure: detecting a project's traits from its files, with the
// artifact that proves each one, and mapping them to the skills that apply. The
// detection is deterministic and reads only the filesystem — no network, no
// catalog, no transport — so it can be reviewed and tested long before an MCP
// server exists. Later slices add the catalog, the stdio server and the tools
// that expose these results to an agent.
package mcp
