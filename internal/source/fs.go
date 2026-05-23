package source

import (
	"context"
	"fmt"

	"github.com/ralvarezdev/ralvaskills/internal/skill"
)

// fsSource is a shared base for filesystem-backed skill sources (Local and
// Official). It walks a directory and caches nothing — suitable for sources
// where the directory is small enough to stat on every call.
type fsSource struct {
	root     string
	src      skill.Source
	notFound func(name string) error
}

// all walks root and returns every discovered skill with Source set to src.
func (f *fsSource) all(_ context.Context) ([]skill.Skill, error) {
	skills, err := skill.Walk(f.root, f.src)
	if err != nil {
		return nil, fmt.Errorf("walk %s skills: %w", f.src, err)
	}
	return skills, nil
}

// find walks root and returns the skill with the given name. Calls
// f.notFound(name) when the name is not present.
func (f *fsSource) find(ctx context.Context, name string) (skill.Skill, error) {
	all, err := f.all(ctx)
	if err != nil {
		return skill.Skill{}, err
	}
	if s, ok := skill.FindByName(all, name); ok {
		return s, nil
	}
	return skill.Skill{}, f.notFound(name)
}
