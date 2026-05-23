package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ralvarezdev/ralvaskills/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/internal/config"
	"github.com/ralvarezdev/ralvaskills/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/internal/skill"
	"github.com/ralvarezdev/ralvaskills/internal/source"
	"github.com/ralvarezdev/ralvaskills/internal/tool"
)

// newLocalSource returns the appropriate source.Resolver for local skills based
// on whether the config points to a local repo clone or the hosted registry.
func newLocalSource(cfg config.Config) source.Resolver {
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
		rskDir := filepath.Join(cwd, manifest.ProjectFolderName)
		return []string{manifest.ProjectSkillsPath(rskDir)}, nil
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
	if scope == cmdx.ForAll {
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
func findSkillByName(ctx context.Context, name string, localSrc, officialSrc source.Resolver) (skill.Skill, error) {
	s, localErr := localSrc.Find(ctx, name)
	if localErr == nil {
		return s, nil
	}
	if !errors.Is(localErr, source.ErrNotFound) {
		return skill.Skill{}, localErr
	}

	s, officialErr := officialSrc.Find(ctx, name)
	if officialErr == nil {
		return s, nil
	}
	if !errors.Is(officialErr, source.ErrNotFound) {
		return skill.Skill{}, officialErr
	}

	return skill.Skill{}, fmt.Errorf(
		"skill %q not found in local repo or official cache\n  Run 'rsk catalog' to browse available skills",
		name,
	)
}

// resolveNames takes a list of bundle-or-skill names and returns the resolved
// skills (deduplicated) along with any soft warnings (skills inside a bundle
// that couldn't be resolved). A name that matches a bundle expands to that
// bundle's skills; otherwise the name is treated as a skill.
func resolveNames(
	ctx context.Context,
	names []string,
	catalog []config.Bundle,
	localSrc, officialSrc source.Resolver,
) ([]skill.Skill, []string, error) {
	var skills []skill.Skill
	var warnings []string

	for _, raw := range names {
		bareName, _, err := parseNameVersion(raw)
		if err != nil {
			return nil, nil, err
		}
		if bundle, ok := config.FindBundle(catalog, bareName); ok {
			ss, ws, resolveErr := resolveBundleSkills(ctx, bundle, localSrc, officialSrc)
			if resolveErr != nil {
				return nil, nil, resolveErr
			}
			skills = append(skills, ss...)
			warnings = append(warnings, ws...)
			continue
		}
		s, findErr := findSkillByName(ctx, bareName, localSrc, officialSrc)
		if findErr != nil {
			return nil, nil, findErr
		}
		skills = append(skills, s)
	}

	return skills, warnings, nil
}

// resolveBundleSkills resolves a bundle's skill refs into concrete Skill values.
// Planned skills (not on disk) and missing official cache entries produce
// warnings rather than errors — the operation continues for available skills.
func resolveBundleSkills(
	ctx context.Context, bundle config.Bundle, localSrc, officialSrc source.Resolver,
) ([]skill.Skill, []string, error) {
	var skills []skill.Skill
	var warnings []string

	for _, ref := range bundle.Skills {
		src, notFoundMsg := sourceForRef(ref.Source, localSrc, officialSrc)
		s, err := src.Find(ctx, ref.Name)
		if err == nil {
			skills = append(skills, s)
			continue
		}
		if !errors.Is(err, source.ErrNotFound) {
			return nil, nil, fmt.Errorf("resolve %s skill %q: %w", ref.Source, ref.Name, err)
		}
		warnings = append(warnings, fmt.Sprintf(notFoundMsg, ref.Name))
	}

	return skills, warnings, nil
}

// sourceForRef returns the Resolver and a not-found warning format string for
// the given source type. Format string receives the skill name as its argument.
func sourceForRef(
	src skill.Source, localSrc, officialSrc source.Resolver,
) (resolver source.Resolver, notFoundMsg string) {
	if src == skill.SourceOfficial {
		return officialSrc, "%s (official) not in cache — run 'rsk update --official' to fetch it"
	}
	return localSrc, "%s is not yet available (planned) — skipped"
}

// filterSkills returns a new slice containing only the skills for which keep returns true.
func filterSkills(in []skill.Skill, keep func(skill.Skill) bool) []skill.Skill {
	out := make([]skill.Skill, 0, len(in))
	for _, s := range in {
		if keep(s) {
			out = append(out, s)
		}
	}
	return out
}

// isLinkedAnywhere reports whether a skill named name is symlinked into any of
// the given target directories.
func isLinkedAnywhere(name string, targets []string) bool {
	for _, t := range targets {
		if skill.IsLinked(name, t) {
			return true
		}
	}
	return false
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

// skillNamesFromArgs resolves bundle-or-skill args to a deduplicated list of
// skill names by reading the catalog. A name matching a bundle expands to that
// bundle's skill names; otherwise the name is treated as a skill name.
func skillNamesFromArgs(args []string, catalog []config.Bundle) ([]string, error) {
	seen := make(map[string]bool)
	var names []string
	for _, raw := range args {
		bareName, _, err := parseNameVersion(raw)
		if err != nil {
			return nil, err
		}
		if b, ok := config.FindBundle(catalog, bareName); ok {
			for _, ref := range b.Skills {
				if !seen[ref.Name] {
					seen[ref.Name] = true
					names = append(names, ref.Name)
				}
			}
			continue
		}
		if !seen[bareName] {
			seen[bareName] = true
			names = append(names, bareName)
		}
	}
	return names, nil
}

// detectSkillSource determines the origin of a skill by matching its path against
// known cache roots. Falls back to SourceRegistry in registry mode (repoPath == "")
// and SourceLocal in local-clone mode.
func detectSkillSource(skillPath, repoPath, officialCachePath, registryCachePath string) skill.Source {
	sp := filepath.ToSlash(filepath.Clean(skillPath))
	if repoPath != "" && strings.HasPrefix(sp, filepath.ToSlash(filepath.Clean(repoPath))+"/") {
		return skill.SourceLocal
	}
	if registryCachePath != "" && strings.HasPrefix(sp, filepath.ToSlash(filepath.Clean(registryCachePath))+"/") {
		return skill.SourceRegistry
	}
	if officialCachePath != "" && strings.HasPrefix(sp, filepath.ToSlash(filepath.Clean(officialCachePath))+"/") {
		return skill.SourceOfficial
	}
	if repoPath == "" {
		return skill.SourceRegistry
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
		rskDir := filepath.Join(cwd, manifest.ProjectFolderName)
		dirs = append(dirs, manifest.ProjectSkillsPath(rskDir))
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

// parseNameVersion splits a "name" or "name@version" argument. Returns an
// error if the name half is empty.
func parseNameVersion(arg string) (name, version string, err error) {
	if idx := strings.Index(arg, "@"); idx >= 0 {
		name = arg[:idx]
		version = arg[idx+1:]
	} else {
		name = arg
	}
	if name == "" {
		return "", "", fmt.Errorf("skill name must not be empty (got %q)", arg)
	}
	return name, version, nil
}

// syncPinnedAllTools writes pinned skill entries for every tool listed in
// m.Tools. rskDir is the .rsk/ directory; the project root is its parent.
func syncPinnedAllTools(rskDir string, m manifest.Mod) error {
	projectDir := filepath.Dir(rskDir)
	for _, id := range m.Tools {
		t, ok := tool.Get(id)
		if !ok {
			return fmt.Errorf("unknown tool %q in rsk.mod", id)
		}
		if err := t.SyncPinned(projectDir, m.Pinned); err != nil {
			return err
		}
	}
	return nil
}
