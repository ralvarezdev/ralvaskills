package source

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	rskgit "github.com/ralvarezdev/ralvaskills/v2/internal/git"
	"github.com/ralvarezdev/ralvaskills/v2/internal/skill"
)

// OfficialSkillsURL is the GitHub URL for the anthropics/skills repo used as
// the source for official skills, in both local-clone and registry mode.
const OfficialSkillsURL = "https://github.com/anthropics/skills"

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
// Returns an empty slice without error if the cache has not been fetched yet.
func (o *Official) All(ctx context.Context) ([]skill.Skill, error) {
	if _, err := os.Stat(o.fs.root); os.IsNotExist(err) {
		return nil, nil
	}
	return o.fs.all(ctx)
}

// Find returns the official skill with the given name. If the local
// anthropics/skills cache has never been fetched, it is cloned on demand
// (the same clone "rsk update --official" performs) so a fresh machine can
// install official skills on first try, regardless of local-clone vs
// registry mode.
func (o *Official) Find(ctx context.Context, name string) (skill.Skill, error) {
	if err := o.ensureCache(ctx); err != nil {
		return skill.Skill{}, fmt.Errorf("fetch official skills cache: %w", err)
	}
	return o.fs.find(ctx, name)
}

// ensureCache clones the anthropics/skills repo into the cache directory if
// it has not been fetched yet. No-op if the cache is already present.
func (o *Official) ensureCache(ctx context.Context) error {
	_, err := os.Stat(o.fs.root)
	switch {
	case err == nil:
		return nil
	case os.IsNotExist(err):
		return rskgit.Clone(ctx, OfficialSkillsURL, o.fs.root, io.Discard)
	default:
		return err
	}
}
