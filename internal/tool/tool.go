// Package tool defines the Tool interface, the ID type, and the global registry
// of supported AI tools. Adding a new tool requires only one new file in this
// package with an init() function that calls Register.
package tool

import (
	"fmt"
	"slices"
	"strings"
)

// ID identifies a supported AI tool by its string name in rsk.mod.
type ID string

// Tool describes an AI tool that rsk can manage skills for.
type Tool interface {
	// ToolID returns the unique identifier for this tool.
	ID() ID

	// SkillsDir returns the canonical path to this tool's global skills
	// directory, resolved using the current XDG environment.
	SkillsDir() string

	// ProjectSkillsDir returns the path to the project-local skills directory
	// that this tool discovers. projectRoot is the project root (parent of .rsk/).
	ProjectSkillsDir(projectRoot string) string

	// SyncPinned writes tool-specific skill imports so that the tool sees
	// exactly the skills in pinnedNames. projectDir is the project root
	// (the parent of .rsk/).
	SyncPinned(projectDir string, pinnedNames []string) error

	// RemovePinned removes all rsk-managed skill imports from the tool's
	// config. Returns nil if the config file does not exist.
	RemovePinned(projectDir string) error
}

// registry is the global map of registered tools, keyed by ID.
var registry = make(map[ID]Tool)

func init() {
	Register(&ClaudeTool{})
	Register(&openCodeTool{})
}

// String returns the string representation of the tool identifier.
func (id ID) String() string {
	return string(id)
}

// MarshalText implements encoding.TextMarshaler.
func (id ID) MarshalText() ([]byte, error) {
	return []byte(id), nil
}

// UnmarshalText implements encoding.TextUnmarshaler, rejecting unknown tool IDs.
func (id *ID) UnmarshalText(b []byte) error {
	parsed, err := ParseID(string(b))
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// ParseID converts s to an ID. Returns an error if s is not a registered tool.
func ParseID(s string) (ID, error) {
	if _, ok := registry[ID(s)]; ok {
		return ID(s), nil
	}
	return "", fmt.Errorf("unknown tool %q: must be one of %s", s, strings.Join(Names(), ", "))
}

// Register adds t to the global tool registry. It is called from each tool's
// init() function and panics if the same ID is registered twice.
func Register(t Tool) {
	if _, dup := registry[t.ID()]; dup {
		panic(fmt.Sprintf("tool: duplicate registration for ID %q", t.ID()))
	}
	registry[t.ID()] = t
}

// Get returns the Tool with the given ID and whether it was found.
func Get(id ID) (Tool, bool) {
	t, ok := registry[id]
	return t, ok
}

// All returns all registered tools in ID-sorted order.
func All() []Tool {
	ids := make([]ID, 0, len(registry))
	for id := range registry {
		ids = append(ids, id)
	}
	slices.Sort(ids)

	tools := make([]Tool, len(ids))
	for i, id := range ids {
		tools[i] = registry[id]
	}
	return tools
}

// Names returns all registered tool IDs as strings, in sorted order.
func Names() []string {
	tools := All()
	names := make([]string, len(tools))
	for i, t := range tools {
		names[i] = string(t.ID())
	}
	return names
}
