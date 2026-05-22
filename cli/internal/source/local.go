// Package source resolves skills from their respective origins.
package source

import (
	"fmt"
	"path/filepath"

	"github.com/ralvarezdev/ralvaskills/cli/internal/skill"
)

// Local resolves skills from the local ralvaskills repo clone.
type Local struct {
	repoPath string
}

// NewLocal returns a Local source rooted at repoPath.
func NewLocal(repoPath string) *Local {
	return &Local{repoPath: repoPath}
}

// SkillsRoot returns the absolute path to the skills/ directory.
func (l *Local) SkillsRoot() string {
	return filepath.Join(l.repoPath, "skills")
}

// All walks the skills/ directory and returns every discovered skill.
func (l *Local) All() ([]skill.Skill, error) {
	skills, err := skill.Walk(l.SkillsRoot(), skill.SourceLocal)
	if err != nil {
		return nil, fmt.Errorf("walk local skills: %w", err)
	}
	return skills, nil
}

// Find returns the local skill with the given name.
func (l *Local) Find(name string) (skill.Skill, error) {
	all, err := l.All()
	if err != nil {
		return skill.Skill{}, err
	}
	for _, s := range all {
		if s.Name == name {
			return s, nil
		}
	}
	return skill.Skill{}, fmt.Errorf("%w: %q in local repo at %s", ErrNotFound, name, l.repoPath)
}
