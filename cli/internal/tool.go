package internal

import (
	"fmt"
	"strings"

	"github.com/ralvarezdev/ralvaskills/cli/internal/config"
)

// ToolID identifies a supported AI tool by its string name in rsk.mod.
type ToolID string

// Tool name constants used in Mod.Tools and in cmd-layer dispatch.
const (
	ToolClaudeCode ToolID = "claude-code"
	ToolOpenCode   ToolID = "opencode"
)

// MarshalText implements encoding.TextMarshaler.
func (t ToolID) MarshalText() ([]byte, error) { return []byte(t), nil }

// UnmarshalText implements encoding.TextUnmarshaler, rejecting unknown tool names.
// Validation is derived from SupportedTools so adding a tool only requires one change.
func (t *ToolID) UnmarshalText(b []byte) error {
	tools := SupportedTools()
	for _, tool := range tools {
		if ToolID(b) == tool.ID {
			*t = ToolID(b)
			return nil
		}
	}
	names := make([]string, len(tools))
	for i, tool := range tools {
		names[i] = string(tool.ID)
	}
	return fmt.Errorf("unknown tool %q: must be one of %s", string(b), strings.Join(names, ", "))
}

// ToolDef defines the properties of a tool, including its identifier and default directory.
type ToolDef struct {
	ID         ToolID
	DefaultDir string
}

// SupportedTools returns all supported AI tools with their default directories.
// Evaluated lazily so XDG paths are resolved at call time, not at package init.
func SupportedTools() []ToolDef {
	return []ToolDef{
		{ID: ToolClaudeCode, DefaultDir: config.ClaudeSkillsDir()},
		{ID: ToolOpenCode, DefaultDir: config.OpenCodeSkillsDir()},
	}
}
