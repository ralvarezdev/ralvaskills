package source

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ralvarezdev/ralvaskills/internal/skill"
)

// Official resolves skills from the cached anthropic/skills clone.
type Official struct {
	cacheDir string
	fs       fsSource
}

// NewOfficial returns an Official source backed by cacheDir (the
// official_cache config value).
func NewOfficial(cacheDir string) *Official {
	root := filepath.Join(cacheDir, skill.SkillsFolderName)
	return &Official{
		cacheDir: cacheDir,
		fs: fsSource{
			root: root,
			src:  skill.SourceOfficial,
			notFound: func(name string) error {
				return fmt.Errorf(
					"%w: official skill %q — run 'rsk update --official' to refresh the cache",
					ErrNotFound, name,
				)
			},
		},
	}
}

// All walks the official cache and returns every discovered skill.
func (o *Official) All(ctx context.Context) ([]skill.Skill, error) {
	return o.fs.all(ctx)
}

// Find returns the official skill with the given name. Returns a descriptive
// error if the cache directory does not exist, prompting the user to run
// "rsk update --official".
func (o *Official) Find(ctx context.Context, name string) (skill.Skill, error) {
	if _, err := os.Stat(o.fs.root); os.IsNotExist(err) {
		return skill.Skill{}, fmt.Errorf(
			"%w: official skill %q — cache not found; run 'rsk update --official' to fetch it",
			ErrNotFound, name,
		)
	}
	return o.fs.find(ctx, name)
}
