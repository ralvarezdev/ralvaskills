package main

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/ralvarezdev/ralvaskills/v3/internal/config"
	"github.com/ralvarezdev/ralvaskills/v3/internal/install"
	"github.com/ralvarezdev/ralvaskills/v3/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/v3/internal/skill"
	"github.com/ralvarezdev/ralvaskills/v3/internal/source"
	"github.com/ralvarezdev/ralvaskills/v3/internal/tool"
)

// newLocalSource returns the appropriate source.Resolver for local skills.
//
//nolint:ireturn // returns interface for dependency injection
func newLocalSource(cfg config.Config) source.Resolver {
	return install.LocalSource(cfg)
}

// projectSkillsDirs returns deduplicated per-tool project skills directories.
func projectSkillsDirs(projectRoot string, m manifest.Mod) []string {
	return install.ProjectSkillsDirs(projectRoot, m)
}

// resolveTargetDirs determines which skill directories an operation should act on.
// Without --global it returns the per-tool project-local skills directories.
func resolveTargetDirs(cfg config.Config, global bool, forTool tool.ID) ([]string, error) {
	return install.Targets(cfg, global, forTool)
}

// findSkillByName looks up a skill by name, trying local first then official.
func findSkillByName(ctx context.Context, name string, localSrc, officialSrc source.Resolver) (skill.Skill, error) {
	return install.FindSkill(ctx, name, localSrc, officialSrc)
}

// resolveNames expands bundle-or-skill names into skills, with soft warnings for
// bundle members that could not be resolved.
func resolveNames(
	ctx context.Context,
	names []string,
	catalog []config.Bundle,
	localSrc, officialSrc source.Resolver,
) ([]skill.Skill, []string, error) {
	return install.Resolve(ctx, names, catalog, localSrc, officialSrc)
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
	return install.DedupSkills(skills)
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
	if rskDir, err := manifest.ProjectFolderPath(); err == nil {
		if m, modErr := manifest.ReadMod(rskDir); modErr == nil {
			dirs = append(dirs, projectSkillsDirs(filepath.Dir(rskDir), m)...)
		}
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

// parseNameVersion splits a "name" or "name@version" argument.
func parseNameVersion(arg string) (name, version string, err error) {
	return install.ParseNameVersion(arg)
}

// syncPinnedAllTools writes pinned skill entries for every tool in rsk.mod.
func syncPinnedAllTools(rskDir string, m manifest.Mod) error {
	return install.SyncPinnedAllTools(rskDir, m)
}
