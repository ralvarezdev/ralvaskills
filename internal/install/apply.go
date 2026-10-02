package install

import (
	"path/filepath"
	"slices"

	"github.com/ralvarezdev/ralvaskills/v3/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/v3/internal/skill"
)

// Scope is where an install writes its skills.
type Scope string

const (
	// ScopeProject installs into the project manifest (.rsk/).
	ScopeProject Scope = "project"

	// ScopeGlobal installs into the configured global skills dirs.
	ScopeGlobal Scope = "global"
)

// Options configures Apply.
type Options struct {
	Constraints map[string]string
	Targets     []string
	RskDir      string
	Scope       Scope
	Pin         bool
}

// Result is the outcome for one skill. Files lists the links created; Err is
// set when the skill could not be linked into every target.
type Result struct {
	Err     error
	Name    string
	Version string
	Scope   Scope
	Files   []string
}

// Apply links each skill into every target and, for the project scope, upserts
// rsk.mod and rsk.lock and rewrites the pinned tool configs. A link failure is
// reported per skill rather than aborting, so the rest of the batch still
// lands; a manifest write failure is returned as err.
func Apply(skills []skill.Skill, opts Options) ([]Result, error) {
	results := make([]Result, 0, len(skills))
	linked := make([]skill.Skill, 0, len(skills))

	for _, s := range skills {
		res := Result{Name: s.Name, Version: s.Version, Scope: opts.Scope}
		for _, target := range opts.Targets {
			if linkErr := skill.Link(s, target); linkErr != nil {
				res.Err = linkErr
				continue
			}
			res.Files = append(res.Files, filepath.Join(target, s.Name))
		}
		results = append(results, res)
		if res.Err == nil {
			linked = append(linked, s)
		}
	}

	if opts.Scope == ScopeProject && len(linked) > 0 {
		if _, err := applyProject(linked, opts); err != nil {
			return results, err
		}
	}
	return results, nil
}

// applyProject upserts the linked skills into rsk.mod and rsk.lock, applies the
// pin, and rewrites the pinned tool configs.
func applyProject(linked []skill.Skill, opts Options) ([]string, error) {
	rskDir := opts.RskDir
	if rskDir == "" {
		var err error
		rskDir, err = manifest.ProjectFolderPath()
		if err != nil {
			return nil, err
		}
	}

	m, err := manifest.ReadMod(rskDir)
	if err != nil {
		return nil, err
	}
	lock, err := manifest.ReadLock(rskDir)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(linked))
	for _, s := range linked {
		constraint := "*"
		if c, ok := opts.Constraints[s.Name]; ok && c != "" {
			constraint = c
		}
		m.Skills[s.Name] = constraint
		names = append(names, s.Name)
		lock = manifest.UpsertLockEntry(lock, manifest.LockEntry{
			Name:    s.Name,
			Version: s.Version,
			Source:  s.Source,
			Path:    s.Path,
		})
	}

	if opts.Pin {
		for _, name := range names {
			if !slices.Contains(m.Pinned, name) {
				m.Pinned = append(m.Pinned, name)
			}
		}
	}

	if writeErr := manifest.WriteMod(rskDir, m); writeErr != nil {
		return nil, writeErr
	}
	if writeErr := manifest.WriteLock(rskDir, lock); writeErr != nil {
		return nil, writeErr
	}
	if syncErr := SyncPinnedAllTools(rskDir, m); syncErr != nil {
		return nil, syncErr
	}
	return []string{manifest.ModPath(rskDir), manifest.LockPath(rskDir)}, nil
}
