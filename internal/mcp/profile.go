package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

// Signal is one project trait detected on disk. Proof is the path of the file
// that evidences it, relative to the project root — a signal without a proof is
// a bug, not a hit.
type Signal struct {
	Kind  string
	Proof string
}

// Candidate is a skill or bundle proposed because of one or more signals.
// Because lists the signal kinds that proposed it, in first-seen order.
type Candidate struct {
	Skill   string
	Because []string
}

// Profile is the deterministic result of scanning a project directory: the
// traits found and the skills they propose.
type Profile struct {
	Root       string
	Signals    []Signal
	Candidates []Candidate
}

// Scan detects project signals under root and maps them to candidate skills.
//
// It reads only the filesystem: no network, no catalog, no MCP transport. The
// root must be a directory; a missing root is an error, but a project with no
// recognizable traits yields an empty Profile and a nil error.
func Scan(ctx context.Context, root string) (Profile, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return Profile{}, fmt.Errorf("resolve project root: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return Profile{}, fmt.Errorf("scan %s: %w", abs, err)
	}
	if !info.IsDir() {
		return Profile{}, fmt.Errorf("scan %s: not a directory", abs)
	}

	idx, err := indexFiles(ctx, abs)
	if err != nil {
		return Profile{}, err
	}

	signals := make([]Signal, 0, len(idx.list))
	var candidates candidateSet
	for _, rule := range signalRules() {
		proof, ok := rule.detect(idx)
		if !ok || proof == "" {
			continue
		}
		signals = append(signals, Signal{Kind: rule.kind, Proof: proof})
		for _, skill := range rule.candidates {
			candidates.add(skill, rule.kind)
		}
	}

	return Profile{Root: abs, Signals: signals, Candidates: candidates.list()}, nil
}

// candidateSet accumulates candidates in first-seen order and records, per
// skill, the signal kinds that proposed it.
type candidateSet struct {
	because map[string][]string
	order   []string
}

func (c *candidateSet) add(skill, kind string) {
	if c.because == nil {
		c.because = make(map[string][]string)
	}
	if _, seen := c.because[skill]; !seen {
		c.order = append(c.order, skill)
	}
	if !slices.Contains(c.because[skill], kind) {
		c.because[skill] = append(c.because[skill], kind)
	}
}

func (c *candidateSet) list() []Candidate {
	out := make([]Candidate, 0, len(c.order))
	for _, skill := range c.order {
		out = append(out, Candidate{Skill: skill, Because: c.because[skill]})
	}
	return out
}

// fileIndex is a snapshot of the project's files, with slash-separated paths
// relative to the root. Rules query it instead of touching the filesystem, so
// detection stays cheap and consistent across rules.
type fileIndex struct {
	root  string
	files map[string]struct{}
	dirs  map[string]struct{}
	list  []string
}

// skippedDirs are directories a scan never descends into: version control, rsk's
// own manifest, dependency/build caches, and test data. A fixture's .proto or
// Dockerfile says nothing about the project's stack and would only add noise.
var skippedDirs = map[string]struct{}{
	".git":         {},
	".rsk":         {},
	"node_modules": {},
	"vendor":       {},
	".venv":        {},
	"venv":         {},
	".next":        {},
	".cache":       {},
	"testdata":     {},
}

func indexFiles(ctx context.Context, root string) (*fileIndex, error) {
	idx := &fileIndex{
		root:  root,
		files: make(map[string]struct{}),
		dirs:  make(map[string]struct{}),
	}

	walkErr := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		if p == root {
			return nil
		}

		rel, relErr := filepath.Rel(root, p)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)

		if d.IsDir() {
			if _, skip := skippedDirs[d.Name()]; skip {
				return fs.SkipDir
			}
			idx.dirs[rel] = struct{}{}
			return nil
		}

		idx.files[rel] = struct{}{}
		idx.list = append(idx.list, rel)
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("walk %s: %w", root, walkErr)
	}

	sort.Strings(idx.list)
	return idx, nil
}

// has returns rel if the index holds that file.
func (idx *fileIndex) has(rel string) (string, bool) {
	if _, ok := idx.files[rel]; ok {
		return rel, true
	}
	return "", false
}

// hasAny returns the first of names present, in argument order.
func (idx *fileIndex) hasAny(names ...string) (string, bool) {
	for _, name := range names {
		if _, ok := idx.files[name]; ok {
			return name, true
		}
	}
	return "", false
}

// firstBySuffix returns the lexicographically first file ending in suffix, at
// any depth (e.g. ".proto").
func (idx *fileIndex) firstBySuffix(suffix string) (string, bool) {
	for _, rel := range idx.list {
		if strings.HasSuffix(rel, suffix) {
			return rel, true
		}
	}
	return "", false
}

// firstRootPrefixed returns the lexicographically first file at the root whose
// name starts with prefix (e.g. "next.config.").
func (idx *fileIndex) firstRootPrefixed(prefix string) (string, bool) {
	for _, rel := range idx.list {
		if strings.Contains(rel, "/") {
			continue
		}
		if strings.HasPrefix(rel, prefix) {
			return rel, true
		}
	}
	return "", false
}

// firstUnder returns the lexicographically first file under the dir prefix.
func (idx *fileIndex) firstUnder(prefix string) (string, bool) {
	for _, rel := range idx.list {
		if strings.HasPrefix(rel, prefix) {
			return rel, true
		}
	}
	return "", false
}

// firstNamedUnder returns the lexicographically first file under the dir prefix
// whose base name equals name (e.g. "cmd/", "main.go").
func (idx *fileIndex) firstNamedUnder(prefix, name string) (string, bool) {
	for _, rel := range idx.list {
		if strings.HasPrefix(rel, prefix) && path.Base(rel) == name {
			return rel, true
		}
	}
	return "", false
}

// dirHasFiles reports whether dir contains at least one regular file.
func (idx *fileIndex) dirHasFiles(dir string) (string, bool) {
	return idx.firstUnder(dir + "/")
}

func (idx *fileIndex) read(rel string) (string, error) {
	data, err := os.ReadFile(filepath.Join(idx.root, filepath.FromSlash(rel)))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// goModHasDep reports whether go.mod requires a module whose path contains dep.
func (idx *fileIndex) goModHasDep(dep string) bool {
	rel, ok := idx.has("go.mod")
	if !ok {
		return false
	}
	text, err := idx.read(rel)
	if err != nil {
		return false
	}
	return strings.Contains(text, dep)
}

// pkgJSONHasDep reports whether package.json lists dep in dependencies or
// devDependencies. A malformed file is treated as no match.
func (idx *fileIndex) pkgJSONHasDep(dep string) bool {
	rel, ok := idx.has("package.json")
	if !ok {
		return false
	}
	text, err := idx.read(rel)
	if err != nil {
		return false
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if unmarshalErr := json.Unmarshal([]byte(text), &pkg); unmarshalErr != nil {
		return false
	}
	_, inDeps := pkg.Dependencies[dep]
	_, inDev := pkg.DevDependencies[dep]
	return inDeps || inDev
}

// pythonHasDep reports whether any Python manifest lists dep.
func (idx *fileIndex) pythonHasDep(dep string) bool {
	for _, name := range []string{"pyproject.toml", "requirements.txt", "uv.lock"} {
		rel, ok := idx.has(name)
		if !ok {
			continue
		}
		if text, err := idx.read(rel); err == nil && strings.Contains(text, dep) {
			return true
		}
	}
	return false
}
