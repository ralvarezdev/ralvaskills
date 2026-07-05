package source

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/ralvarezdev/ralvaskills/internal/skill"
)

// Local resolves skills from the local ralvaskills repo clone.
type Local struct {
	repoPath string
	fs       fsSource
}

// NewLocal returns a Local source rooted at repoPath.
func NewLocal(repoPath string) *Local {
	root := filepath.Join(repoPath, skill.SkillsFolderName)
	return &Local{
		repoPath: repoPath,
		fs: fsSource{
			root: root,
			src:  skill.SourceLocal,
			notFound: func(name string) error {
				return fmt.Errorf("%w: local skill %q not found in repo at %s", ErrNotFound, name, repoPath)
			},
		},
	}
}

// SkillsRoot returns the absolute path to the skills/ directory.
func (l *Local) SkillsRoot() string {
	return l.fs.root
}

// All walks the local skills directory and returns every discovered skill.
func (l *Local) All(ctx context.Context) ([]skill.Skill, error) {
	return l.fs.all(ctx)
}

// Find returns the local skill with the given name.
func (l *Local) Find(ctx context.Context, name string) (skill.Skill, error) {
	return l.fs.find(ctx, name)
}
