package source

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ralvarezdev/ralvaskills/cli/internal/skill"
)

// Official resolves skills from the cached anthropics/skills clone.
type Official struct {
	cacheDir string // path to the skills/ subdirectory inside the anthropic cache
}

// NewOfficial returns an Official source backed by cacheDir (the official_cache config value).
func NewOfficial(cacheDir string) *Official {
	return &Official{cacheDir: filepath.Join(cacheDir, "skills")}
}

// All walks the cache and returns every discovered official skill.
func (o *Official) All(_ context.Context) ([]skill.Skill, error) {
	skills, err := skill.Walk(o.cacheDir, skill.SourceOfficial)
	if err != nil {
		return nil, fmt.Errorf("walk official skills: %w", err)
	}
	return skills, nil
}

// Find returns the official skill with the given name.
func (o *Official) Find(ctx context.Context, name string) (skill.Skill, error) {
	if _, err := os.Stat(o.cacheDir); os.IsNotExist(err) {
		return skill.Skill{}, fmt.Errorf(
			"%w: official skill %q — cache not found; run 'rsk update --official' to fetch it",
			ErrNotFound, name,
		)
	}
	all, err := o.All(ctx)
	if err != nil {
		return skill.Skill{}, err
	}
	for _, s := range all {
		if s.Name == name {
			return s, nil
		}
	}
	return skill.Skill{}, fmt.Errorf(
		"%w: official skill %q — run 'rsk update --official' to refresh the cache",
		ErrNotFound, name,
	)
}
