// Package update computes and persists whether skill updates are available.
// It never applies anything itself — that stays the job of `rsk update`
// (cmd/rsk/update.go); this package's only job is detection and caching so a
// bare `rsk` invocation can flag drift without ever touching the network.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ralvarezdev/ralvaskills/v3/internal/config"
	"github.com/ralvarezdev/ralvaskills/v3/internal/fsperm"
	"github.com/ralvarezdev/ralvaskills/v3/internal/fsx"
	rskgit "github.com/ralvarezdev/ralvaskills/v3/internal/git"
	"github.com/ralvarezdev/ralvaskills/v3/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/v3/internal/source"
	"github.com/ralvarezdev/ralvaskills/v3/internal/tool"
)

const (
	cacheFileName = "update-check.json"
	tempPattern   = ".rsk-update-check-*.tmp"

	// ModeLocal marks a Result computed against a local-repo-clone config.
	ModeLocal = "local"
	// ModeRegistry marks a Result computed against a hosted-registry config.
	ModeRegistry = "registry"

	// StaleAfter is how long a cached Result stays trustworthy. Older than
	// this, a bare invocation should kick off a background refresh instead
	// of showing (or trusting the absence of) a notice.
	StaleAfter = 24 * time.Hour
)

// Result reports which skills — or, in local mode, a synthetic entry
// describing local-clone drift — have updates available.
type Result struct {
	Mode     string   `json:"mode"`
	Outdated []string `json:"outdated"`
}

// cacheFile is the on-disk shape written to <VersionsCache>/update-check.json.
type cacheFile struct {
	CheckedAt time.Time `json:"checked_at"`
	Mode      string    `json:"mode"`
	Outdated  []string  `json:"outdated"`
}

func cachePath(cfg config.Config) string {
	return filepath.Join(cfg.VersionsCache, cacheFileName)
}

// Load reads the persisted cache for cfg. ok is false whenever there's
// nothing usable — missing file, unreadable, or malformed — which callers
// should treat the same as "no data yet", not as an error worth surfacing.
func Load(cfg config.Config) (result Result, checkedAt time.Time, ok bool) {
	data, err := os.ReadFile(cachePath(cfg))
	if err != nil {
		return Result{}, time.Time{}, false
	}

	var c cacheFile
	if err = json.Unmarshal(data, &c); err != nil {
		return Result{}, time.Time{}, false
	}
	return Result{Mode: c.Mode, Outdated: c.Outdated}, c.CheckedAt, true
}

// Stale reports whether a cache last checked at checkedAt is old enough that
// its data should no longer be trusted without a refresh.
func Stale(checkedAt time.Time) bool {
	return time.Since(checkedAt) > StaleAfter
}

// Save persists result to cfg's cache file atomically, creating the cache
// directory as needed.
func Save(cfg config.Config, result Result) error {
	if err := os.MkdirAll(cfg.VersionsCache, fsperm.Dir); err != nil {
		return fmt.Errorf("create versions cache dir: %w", err)
	}

	c := cacheFile{CheckedAt: time.Now().UTC(), Mode: result.Mode, Outdated: result.Outdated}
	return fsx.WriteAtomic(cachePath(cfg), tempPattern, func(w io.Writer) error {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(c); err != nil {
			return fmt.Errorf("marshal update-check cache: %w", err)
		}
		return nil
	})
}

// Check computes the current update-availability state for cfg — it never
// mutates anything, including the cache; callers decide whether/when to
// persist the result via Save. This is the only place in the codebase that
// performs network I/O for update-checking purposes, and it's always
// explicit: called from `rsk update` itself, or from the detached background
// worker spawned by a bare `rsk` invocation, never inline on every run.
func Check(ctx context.Context, cfg config.Config) (Result, error) {
	if cfg.LocalMode() {
		return checkLocal(ctx, cfg)
	}
	return checkRegistry(ctx, cfg)
}

// checkLocal reports how far a local-repo-clone config is behind its remote.
// Local mode has no per-skill version concept — a plain `git pull` is the
// whole update story — so "outdated" here means "the clone itself has
// upstream commits it hasn't pulled yet".
func checkLocal(ctx context.Context, cfg config.Config) (Result, error) {
	n, err := rskgit.Behind(ctx, cfg.RepoPath)
	if err != nil {
		return Result{Mode: ModeLocal}, fmt.Errorf("check local repo drift: %w", err)
	}

	var outdated []string
	if n > 0 {
		outdated = []string{fmt.Sprintf("local repo is %d commit(s) behind", n)}
	}
	return Result{Mode: ModeLocal, Outdated: outdated}, nil
}

// checkRegistry compares every installed skill's version against the
// registry's published latest, mirroring the exact comparison
// cmd/rsk/update.go's runUpdateRegistry performs — see InstalledVersionFromTargets.
func checkRegistry(ctx context.Context, cfg config.Config) (Result, error) {
	reg := source.NewRegistry(cfg.RegistryURL, cfg.RegistryCache())

	index, err := reg.Index(ctx)
	if err != nil {
		return Result{Mode: ModeRegistry}, fmt.Errorf("fetch registry index: %w", err)
	}

	targets := mergedTargetDirs(cfg)

	seen := make(map[string]bool)
	var names []string
	for _, target := range targets {
		entries, readErr := os.ReadDir(target)
		if readErr != nil {
			continue
		}
		for _, e := range entries {
			if !seen[e.Name()] {
				seen[e.Name()] = true
				names = append(names, e.Name())
			}
		}
	}

	var outdated []string
	for _, name := range names {
		entry, ok := index[name]
		if !ok {
			continue
		}
		installed := InstalledVersionFromTargets(name, targets, cfg.RegistryCache())
		if installed != entry.Latest {
			outdated = append(outdated, name)
		}
	}
	return Result{Mode: ModeRegistry, Outdated: outdated}, nil
}

// InstalledVersionFromTargets reads the symlink target for name in each
// target dir and extracts the version segment from a registry cache path.
// Exported so cmd/rsk/update.go's runUpdateRegistry and this package's
// checkRegistry share the exact same version-comparison logic instead of
// each maintaining their own copy.
func InstalledVersionFromTargets(name string, targets []string, registryCacheDir string) string {
	for _, target := range targets {
		linkPath := filepath.Join(target, name)
		dest, err := os.Readlink(linkPath)
		if err != nil {
			continue
		}
		// Registry cache layout: <registryCacheDir>/<name>/<version>/
		prefix := filepath.Join(registryCacheDir, name) + string(filepath.Separator)
		if len(dest) > len(prefix) && dest[:len(prefix)] == prefix {
			rest := dest[len(prefix):]
			// rest is "<version>" or "<version>/<something>"
			for i, c := range rest {
				if c == filepath.Separator {
					return rest[:i]
				}
			}
			return rest
		}
	}
	return ""
}

// mergedTargetDirs returns every configured global target directory, plus —
// when the current working directory is an rsk project — its project-local
// skill directories too. This mirrors cmd/rsk/resolve.go's allTargetDirs,
// duplicated in miniature here rather than imported: resolve.go lives in
// package main (cmd/rsk), which this package cannot import, and reshaping
// resolve.go itself is out of scope for this change. Errors resolving the
// project (e.g. the cwd isn't one) are swallowed — global-only coverage is a
// fine fallback for a passive, best-effort check.
func mergedTargetDirs(cfg config.Config) []string {
	dirs := make([]string, 0, len(cfg.GlobalTargets)+1)
	seen := make(map[string]bool, len(cfg.GlobalTargets))
	for _, dir := range cfg.GlobalTargets {
		if !seen[dir] {
			seen[dir] = true
			dirs = append(dirs, dir)
		}
	}

	rskDir, err := manifest.ProjectFolderPath()
	if err != nil {
		return dirs
	}
	m, err := manifest.ReadMod(rskDir)
	if err != nil {
		return dirs
	}

	projectRoot := filepath.Dir(rskDir)
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
