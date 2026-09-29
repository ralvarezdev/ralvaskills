package cmdx

import (
	"fmt"

	"github.com/ralvarezdev/ralvaskills/v2/internal/tool"
)

// ParseTargetScope converts raw into a TargetScope: a registered tool ID, the
// ScopeAll sentinel, or the empty scope when raw is empty (the flag was not
// given). It is the single place the --for value set is validated, so help,
// completion, and runtime checks can all derive from the same source.
func ParseTargetScope(raw string) (TargetScope, error) {
	if raw == "" {
		return "", nil
	}
	if TargetScope(raw) == ScopeAll {
		return ScopeAll, nil
	}
	id, err := tool.ParseID(raw)
	if err != nil {
		return "", fmt.Errorf("--%s: %w", FlagFor, err)
	}
	return TargetScope(id), nil
}

// String returns the scope's raw value.
func (s TargetScope) String() string {
	return string(s)
}
