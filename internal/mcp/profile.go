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
	"strings"
)

type (
	// SignalKind identifies a detected project trait. It is a string type so it
	// marshals to its wire value unchanged.
	SignalKind string

	// Signal is one project trait detected on disk. Proof is the path of the
	// file that evidences it, relative to the project root — a signal without a
	// proof is a bug, not a hit.
	Signal struct {
		Kind  SignalKind
		Proof string
	}

	// Candidate is a skill or bundle proposed because of one or more signals.
	// Because lists the signal kinds that proposed it, in first-seen order.
	Candidate struct {
		Because []SignalKind
		Skill   string
	}

	// Profile is the deterministic result of scanning a project directory: the
	// traits found and the skills they propose.
	Profile struct {
		Signals    []Signal
		Candidates []Candidate
		Root       string
	}

	// candidateSet accumulates candidates in first-seen order and records, per
	// skill, the signal kinds that proposed it.
	candidateSet struct {
		order   []string
		because map[string][]SignalKind
	}

	// fileIndex is a snapshot of the project's files, with slash-separated paths
	// relative to the root. Rules query it instead of touching the filesystem,
	// so detection stays cheap and consistent across rules.
	fileIndex struct {
		list  []string
		root  string
		files map[string]struct{}
	}
)

// Signal kinds, one per recognizable project trait.
const (
	SignalLanguageGo      SignalKind = "language:go"
	SignalLanguagePython  SignalKind = "language:python"
	SignalLanguageTS      SignalKind = "language:typescript"
	SignalFrameworkNextJS SignalKind = "framework:nextjs"
	SignalFrameworkAstro  SignalKind = "framework:astro"
	SignalProtocolGRPC    SignalKind = "protocol:grpc"
	SignalProtocolREST    SignalKind = "protocol:rest"
	SignalInfraDocker     SignalKind = "infra:docker"
	SignalInfraK8s        SignalKind = "infra:k8s"
	SignalInfraCI         SignalKind = "infra:ci"
	SignalRepoTooling     SignalKind = "repo:tooling"
	SignalRepoCLI         SignalKind = "repo:cli"
	SignalDataPostgres    SignalKind = "data:postgres"
)

// String returns the wire value of the kind.
func (k SignalKind) String() string {
	return string(k)
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
	candidates := newCandidateSet()
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

func newCandidateSet() *candidateSet {
	return &candidateSet{because: make(map[string][]SignalKind)}
}

func (c *candidateSet) add(skill string, kind SignalKind) {
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

func newFileIndex(root string) *fileIndex {
	return &fileIndex{root: root, files: make(map[string]struct{})}
}

// skipDir reports whether a scan must not descend into a directory with the
// given name: version control, rsk's own manifest, dependency/build caches, and
// test data. A fixture's .proto or Dockerfile says nothing about the project's
// stack and would only add noise.
func skipDir(name string) bool {
	switch name {
	case ".git", ".rsk", "node_modules", "vendor", ".venv", "venv", ".next", ".cache", "testdata":
		return true
	default:
		return false
	}
}

func indexFiles(ctx context.Context, root string) (*fileIndex, error) {
	idx := newFileIndex(root)

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
			if skipDir(d.Name()) {
				return fs.SkipDir
			}
			return nil
		}

		idx.files[rel] = struct{}{}
		idx.list = append(idx.list, rel)
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("walk %s: %w", root, walkErr)
	}

	slices.Sort(idx.list)
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

// read returns the contents of a project file, or ok=false when it cannot be
// read. It is used only to test for a dependency, so a read failure is a
// no-match rather than an error to propagate.
func (idx *fileIndex) read(rel string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(idx.root, filepath.FromSlash(rel)))
	if err != nil {
		return "", false
	}
	return string(data), true
}

// goModHasDep reports whether go.mod requires a module whose path contains dep.
func (idx *fileIndex) goModHasDep(dep string) bool {
	text, ok := idx.read("go.mod")
	if !ok {
		return false
	}
	return strings.Contains(text, dep)
}

// pkgJSONHasDep reports whether package.json lists dep in dependencies or
// devDependencies. A malformed file is treated as no match.
func (idx *fileIndex) pkgJSONHasDep(dep string) bool {
	text, ok := idx.read("package.json")
	if !ok {
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
		if text, ok := idx.read(name); ok && strings.Contains(text, dep) {
			return true
		}
	}
	return false
}
