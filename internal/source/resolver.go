package source

import (
	"context"

	"github.com/ralvarezdev/ralvaskills/internal/skill"
)

// Resolver is the common interface satisfied by Local, Official, and Registry.
// Commands depend on this interface rather than on concrete types so that the
// source-selection logic stays centralized in resolve.go.
type Resolver interface {
	// All returns every skill available from this source.
	All(ctx context.Context) ([]skill.Skill, error)

	// Find returns the skill with the given name. Returns an error wrapping
	// ErrNotFound when the skill is not present in this source.
	Find(ctx context.Context, name string) (skill.Skill, error)
}
