// Package install resolves and installs skills into a project or global scope.
// It is the single implementation shared by `rsk install` and the MCP server,
// so the CLI and the agent cannot diverge on what gets written.
package install

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ralvarezdev/ralvaskills/v3/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/v3/internal/config"
	"github.com/ralvarezdev/ralvaskills/v3/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/v3/internal/skill"
	"github.com/ralvarezdev/ralvaskills/v3/internal/source"
	"github.com/ralvarezdev/ralvaskills/v3/internal/tool"
)

// LocalSource returns the resolver for the configured local source: a repo
// clone in local mode, the hosted registry otherwise.
//
//nolint:ireturn // returns interface for dependency injection
func LocalSource(cfg config.Config) source.Resolver {
	if cfg.LocalMode() {
		return source.NewLocal(cfg.RepoPath)
	}
	return source.NewRegistry(cfg.RegistryURL, cfg.RegistryCache())
}

// FindSkill returns the skill named name, trying the local source first and the
// official cache second.
func FindSkill(ctx context.Context, name string, localSrc, officialSrc source.Resolver) (skill.Skill, error) {
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

// Resolve expands bundle-or-skill names into concrete, deduplicated skills,
// along with soft warnings for bundle members that could not be resolved.
func Resolve(
	ctx context.Context,
	names []string,
	catalog []config.Bundle,
	localSrc, officialSrc source.Resolver,
) ([]skill.Skill, []string, error) {
	var skills []skill.Skill
	var warnings []string

	for _, raw := range names {
		bareName, _, err := ParseNameVersion(raw)
		if err != nil {
			return nil, nil, err
		}
		if bundle, ok := config.FindBundle(catalog, bareName); ok {
			bundleSkills, bundleWarnings, resolveErr := resolveBundleSkills(
				ctx, bundle, localSrc, officialSrc,
			)
			if resolveErr != nil {
				return nil, nil, resolveErr
			}
			skills = append(skills, bundleSkills...)
			warnings = append(warnings, bundleWarnings...)
			continue
		}
		s, findErr := FindSkill(ctx, bareName, localSrc, officialSrc)
		if findErr != nil {
			return nil, nil, findErr
		}
		skills = append(skills, s)
	}

	return DedupSkills(skills), warnings, nil
}

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

//
//nolint:ireturn // returns interface for dependency injection
func sourceForRef(
	src skill.Source, localSrc, officialSrc source.Resolver,
) (resolver source.Resolver, notFoundMsg string) {
	if src == skill.SourceOfficial {
		return officialSrc, "%s (official) not in cache — run 'rsk update --include official' to fetch it"
	}
	return localSrc, "%s is not yet available (planned) — skipped"
}

// DedupSkills returns skills with duplicates removed, keeping first occurrences.
func DedupSkills(skills []skill.Skill) []skill.Skill {
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

// ParseNameVersion splits "name" or "name@version".
func ParseNameVersion(arg string) (name, version string, err error) {
	if before, after, ok := strings.Cut(arg, "@"); ok {
		name = before
		version = after
	} else {
		name = arg
	}
	if name == "" {
		return "", "", fmt.Errorf("skill name must not be empty (got %q)", arg)
	}
	return name, version, nil
}

// ConstraintsFromArgs returns name → version constraints from "name@version" args.
func ConstraintsFromArgs(args []string) (map[string]string, error) {
	constraints := make(map[string]string, len(args))
	for _, raw := range args {
		name, version, err := ParseNameVersion(raw)
		if err != nil {
			return nil, err
		}
		if version != "" {
			constraints[name] = version
		}
	}
	return constraints, nil
}

// Targets returns the skill directories an install should act on. Without
// global it returns the project's per-tool skills dirs; with global it returns
// the configured global dirs.
func Targets(cfg config.Config, global bool, forTool tool.ID) ([]string, error) {
	if !global {
		rskDir, err := manifest.ProjectFolderPath()
		if err != nil {
			return nil, err
		}
		m, err := manifest.ReadMod(rskDir)
		if err != nil {
			return nil, err
		}
		return ProjectSkillsDirs(filepath.Dir(rskDir), m), nil
	}

	if forTool != "" {
		dir, ok := cfg.GlobalTargets[string(forTool)]
		if !ok {
			return nil, fmt.Errorf(
				"--for: tool %q is not configured — configured tools: %s",
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

// ProjectSkillsDirs returns the deduplicated project-local skills dirs for the
// tools listed in m.
func ProjectSkillsDirs(projectRoot string, m manifest.Mod) []string {
	seen := make(map[string]bool, len(m.Tools))
	dirs := make([]string, 0, len(m.Tools))
	for _, id := range m.Tools {
		t, ok := tool.Get(id)
		if !ok {
			continue
		}
		dir := t.ProjectSkillsDir(projectRoot)
		if !seen[dir] {
			seen[dir] = true
			dirs = append(dirs, dir)
		}
	}
	return dirs
}

// SyncPinnedAllTools rewrites each configured tool's project config from m's
// pinned list.
func SyncPinnedAllTools(rskDir string, m manifest.Mod) error {
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

func joinKeys(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}
