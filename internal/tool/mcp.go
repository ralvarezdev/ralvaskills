package tool

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/ralvarezdev/ralvaskills/v3/internal/fsx"
)

const (
	// MCPFileName is the project file Claude Code reads for MCP servers.
	MCPFileName = ".mcp.json"

	// rskMCPServerName is the key rsk registers its server under.
	rskMCPServerName = "rsk"

	// openCodeMCPKey is the opencode.json key holding local MCP servers.
	openCodeMCPKey = "mcp"

	mcpPointerStart = "<!-- rsk-mcp:start -->"
	mcpPointerEnd   = "<!-- rsk-mcp:end -->"
)

// claudeMCPServer is the Claude Code MCP entry rsk registers.
func claudeMCPServer() map[string]any {
	return map[string]any{"command": "rsk", "args": []any{"mcp"}}
}

// openCodeMCPServer is the opencode MCP entry rsk registers.
func openCodeMCPServer() map[string]any {
	return map[string]any{"type": "local", "command": []any{"rsk", "mcp"}, "enabled": true}
}

// RegisterMCP writes the rsk server into the project's .mcp.json, preserving
// other servers and keys. It is idempotent: a matching entry is left untouched.
// It returns the path written.
func (*ClaudeTool) RegisterMCP(projectDir string) (string, error) {
	return registerMCPServer(
		filepath.Join(projectDir, MCPFileName), "mcpServers",
		claudeMCPServer(), readClaudeConfig, writeClaudeConfig,
	)
}

// UnregisterMCP removes the rsk server from .mcp.json, leaving others intact.
func (*ClaudeTool) UnregisterMCP(projectDir string) error {
	return removeMCPServer(
		filepath.Join(projectDir, MCPFileName), "mcpServers",
		readClaudeConfig, writeClaudeConfig,
	)
}

// RegisterMCP writes the rsk server into the project's opencode.json, preserving
// other servers and keys. It is idempotent and returns the path written.
func (*openCodeTool) RegisterMCP(projectDir string) (string, error) {
	return registerMCPServer(
		filepath.Join(projectDir, OpenCodeFileName), openCodeMCPKey,
		openCodeMCPServer(), readOCConfig, writeOCConfig,
	)
}

// UnregisterMCP removes the rsk server from opencode.json, leaving others intact.
func (*openCodeTool) UnregisterMCP(projectDir string) error {
	return removeMCPServer(
		filepath.Join(projectDir, OpenCodeFileName), openCodeMCPKey,
		readOCConfig, writeOCConfig,
	)
}

func registerMCPServer(
	path, key string,
	server map[string]any,
	read func(string) (map[string]any, error),
	write func(string, map[string]any) error,
) (string, error) {
	cfg, err := read(path)
	if err != nil {
		return "", err
	}

	servers, ok := cfg[key].(map[string]any)
	if !ok {
		servers = make(map[string]any)
	}
	if reflect.DeepEqual(servers[rskMCPServerName], server) {
		return path, nil
	}

	servers[rskMCPServerName] = server
	cfg[key] = servers
	if writeErr := write(path, cfg); writeErr != nil {
		return "", writeErr
	}
	return path, nil
}

func removeMCPServer(
	path, key string,
	read func(string) (map[string]any, error),
	write func(string, map[string]any) error,
) error {
	cfg, err := read(path)
	if err != nil {
		return err
	}

	servers, ok := cfg[key].(map[string]any)
	if !ok {
		return nil
	}
	if _, present := servers[rskMCPServerName]; !present {
		return nil
	}

	delete(servers, rskMCPServerName)
	if len(servers) == 0 {
		delete(cfg, key)
	} else {
		cfg[key] = servers
	}
	return write(path, cfg)
}

const mcpPointerBody = "## Project standards\n\n" +
	"Before scaffolding or a non-trivial change, call the rsk MCP `project_profile` " +
	"tool to learn which skills apply here. Installed skills appear automatically " +
	"in your skill list — invoke them with the Skill tool. Never install without " +
	"asking.\n"

const mcpPointerBlock = mcpPointerStart + "\n" + mcpPointerBody + mcpPointerEnd + "\n"

// WriteMCPPointer adds the rsk workflow pointer to the project's ./CLAUDE.md,
// replacing any previous block so repeated calls are idempotent. Other content
// is preserved.
func WriteMCPPointer(projectDir string) error {
	path := filepath.Join(projectDir, ClaudeFileName)
	if err := refuseIfLink(path); err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", path, err)
	}

	content := strings.TrimRight(removeMCPPointerBlock(string(data)), "\n")
	if content != "" {
		content += "\n\n"
	}
	return writeMCPPointerFile(path, content+mcpPointerBlock)
}

// RemoveMCPPointer removes the rsk workflow pointer block from ./CLAUDE.md.
func RemoveMCPPointer(projectDir string) error {
	path := filepath.Join(projectDir, ClaudeFileName)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if linkErr := refuseIfLink(path); linkErr != nil {
		return linkErr
	}
	return writeMCPPointerFile(path, removeMCPPointerBlock(string(data)))
}

func removeMCPPointerBlock(s string) string {
	start := strings.Index(s, mcpPointerStart)
	if start < 0 {
		return s
	}
	end := strings.Index(s[start:], mcpPointerEnd)
	if end < 0 {
		return s
	}
	end = start + end + len(mcpPointerEnd)
	if end < len(s) && s[end] == '\n' {
		end++
	}
	return s[:start] + s[end:]
}

func writeMCPPointerFile(path, content string) error {
	return fsx.WriteAtomic(path, rskTempPattern, func(w io.Writer) error {
		_, err := io.WriteString(w, content)
		return err
	})
}
