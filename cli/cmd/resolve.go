package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ralvarezdev/ralvaskills/cli/internal/config"
	"github.com/ralvarezdev/ralvaskills/cli/internal/skill"
	"github.com/ralvarezdev/ralvaskills/cli/internal/source"
)

// skillResolver is the common interface satisfied by source.Local, source.Registry,
// and source.Official. Commands depend on this interface, not on concrete types.
type skillResolver interface {
	All() ([]skill.Skill, error)
	Find(name string) (skill.Skill, error)
}

// newLocalSource returns the appropriate skillResolver for local skills based on
// whether the config points to a local repo clone or the hosted registry.
func newLocalSource(cfg config.Config) skillResolver {
	if cfg.LocalMode() {
		return source.NewLocal(cfg.RepoPath)
	}
	return source.NewRegistry(cfg.RegistryURL, cfg.RegistryCache())
}

// resolveTargetDirs determines which skill directories an operation should act on.
// Without --global it returns the project-local .claude/skills/ directory.
func resolveTargetDirs(cfg config.Config, global bool, forTool string) ([]string, error) {
	if !global {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get working directory: %w", err)
		}
		return []string{filepath.Join(cwd, ".claude", "skills")}, nil
	}

	if forTool != "" {
		dir, ok := cfg.GlobalTargets[forTool]
		if !ok {
			return nil, fmt.Errorf(
				"tool %q is not configured — configured tools: %s",
				forTool, joinKeys(cfg.GlobalTargets),
			)
		}
		return []string{dir}, nil
	}

	scope := cfg.DefaultTargetScope
	if scope == "all" {
		dirs := make([]string, 0, len(cfg.GlobalTargets))
		for _, dir := range cfg.GlobalTargets {
			dirs = append(dirs, dir)
		}
		return dirs, nil
	}

	dir, ok := cfg.GlobalTargets[scope]
	if !ok {
		return nil, fmt.Errorf(
			"default_target_scope %q does not match any configured tool — run 'rsk init' to fix",
			scope,
		)
	}
	return []string{dir}, nil
}

// findSkillByName looks up a skill by name, trying local first then official.
func findSkillByName(name string, localSrc skillResolver, officialSrc skillResolver) (skill.Skill, error) {
	s, localErr := localSrc.Find(name)
	if localErr == nil {
		return s, nil
	}
	if !errors.Is(localErr, source.ErrNotFound) {
		return skill.Skill{}, localErr
	}

	s, officialErr := officialSrc.Find(name)
	if officialErr == nil {
		return s, nil
	}
	if !errors.Is(officialErr, source.ErrNotFound) {
		return skill.Skill{}, officialErr
	}

	return skill.Skill{}, fmt.Errorf(
		"skill %q not found in local repo or official cache\n  Run 'rsk list' to browse available skills",
		name,
	)
}

// resolveBundleSkills resolves a bundle's skill refs into concrete Skill values.
// Planned skills (not on disk) and missing official cache entries produce warnings
// rather than errors — the operation continues for all available skills.
func resolveBundleSkills(bundle config.Bundle, localSrc skillResolver, officialSrc skillResolver) ([]skill.Skill, []string, error) {
	var skills []skill.Skill
	var warnings []string

	for _, ref := range bundle.Skills {
		switch ref.Source {
		case config.SourceLocal:
			s, err := localSrc.Find(ref.Name)
			if err != nil {
				if errors.Is(err, source.ErrNotFound) {
					warnings = append(warnings, fmt.Sprintf(
						"%s is not yet available (planned) — skipped",
						ref.Name,
					))
					continue
				}
				return nil, nil, fmt.Errorf("resolve local skill %q: %w", ref.Name, err)
			}
			skills = append(skills, s)

		case config.SourceOfficial:
			s, err := officialSrc.Find(ref.Name)
			if err != nil {
				if errors.Is(err, source.ErrNotFound) {
					warnings = append(warnings, fmt.Sprintf(
						"%s (official) not in cache — run 'rsk update --official' to fetch it",
						ref.Name,
					))
					continue
				}
				return nil, nil, fmt.Errorf("resolve official skill %q: %w", ref.Name, err)
			}
			skills = append(skills, s)
		}
	}

	return skills, warnings, nil
}

func dedupSkills(skills []skill.Skill) []skill.Skill {
	seen := make(map[string]bool, len(skills))
	result := make([]skill.Skill, 0, len(skills))
	for _, s := range skills {
		if !seen[s.Name] {
			seen[s.Name] = true
			result = append(result, s)
		}
	}
	return result
}

// skillNamesFromArgs resolves bundle or --skill args to a deduplicated list of skill names.
// It does not require skills to exist on disk — it only reads the catalog.
func skillNamesFromArgs(bundleArgs []string, skillFlag string, catalog []config.Bundle) ([]string, error) {
	if skillFlag != "" {
		return []string{skillFlag}, nil
	}
	seen := make(map[string]bool)
	var names []string
	for _, bundleName := range bundleArgs {
		b, ok := config.FindBundle(catalog, bundleName)
		if !ok {
			return nil, fmt.Errorf("bundle %q not found — run 'rsk list' to see available bundles", bundleName)
		}
		for _, ref := range b.Skills {
			if !seen[ref.Name] {
				seen[ref.Name] = true
				names = append(names, ref.Name)
			}
		}
	}
	return names, nil
}

// detectSkillSource determines the origin of a skill by matching its path against
// known cache roots. Registry-cached skills are treated as local source.
func detectSkillSource(skillPath, repoPath, officialCachePath, registryCachePath string) skill.Source {
	sp := filepath.ToSlash(filepath.Clean(skillPath))
	if repoPath != "" && strings.HasPrefix(sp, filepath.ToSlash(filepath.Clean(repoPath))+"/") {
		return skill.SourceLocal
	}
	if registryCachePath != "" && strings.HasPrefix(sp, filepath.ToSlash(filepath.Clean(registryCachePath))+"/") {
		return skill.SourceLocal
	}
	if officialCachePath != "" && strings.HasPrefix(sp, filepath.ToSlash(filepath.Clean(officialCachePath))+"/") {
		return skill.SourceOfficial
	}
	return skill.SourceLocal
}

// allTargetDirs returns all configured target dirs (all global targets + project-local).
func allTargetDirs(cfg config.Config) []string {
	dirs := make([]string, 0, len(cfg.GlobalTargets)+1)
	for _, dir := range cfg.GlobalTargets {
		dirs = append(dirs, dir)
	}
	if cwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, filepath.Join(cwd, ".claude", "skills"))
	}
	return dirs
}

// bundleMembershipIndex returns a map from skill name to the bundle names that include it.
func bundleMembershipIndex(catalog []config.Bundle) map[string][]string {
	idx := make(map[string][]string)
	for _, b := range catalog {
		for _, ref := range b.Skills {
			idx[ref.Name] = append(idx[ref.Name], b.Name)
		}
	}
	return idx
}

func joinKeys(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}
