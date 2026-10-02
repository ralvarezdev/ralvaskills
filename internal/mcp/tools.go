package mcp

import (
	"cmp"
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ralvarezdev/mcpkit"

	"github.com/ralvarezdev/ralvaskills/v3/internal"
	"github.com/ralvarezdev/ralvaskills/v3/internal/catalog"
	"github.com/ralvarezdev/ralvaskills/v3/internal/manifest"
)

const (
	defaultSearchLimit = 10
	maxSearchLimit     = 25
	searchNameWeight   = 2

	// requiresUserInteractionKey is the Claude Code tool meta key that forces a
	// permission prompt on every call, regardless of auto-approval modes. The
	// MCP destructiveHint is advisory and does not trigger it.
	requiresUserInteractionKey = "anthropic/requiresUserInteraction"
)

type (
	// ProjectProfileIn is the project_profile tool input.
	ProjectProfileIn struct {
		Path string `json:"path,omitempty" jsonschema:"Ruta del proyecto; por defecto el cwd"`
	}

	// ProjectProfileOut is the project_profile tool output: the detected
	// signals, the skills they propose, and the project's install state.
	ProjectProfileOut struct {
		ProjectRoot string         `json:"project_root"`
		Signals     []Signal       `json:"signals"`
		Candidates  []Candidate    `json:"candidates"`
		Installed   []InstalledRef `json:"installed"`
		Manifest    *ManifestState `json:"manifest,omitempty"`
	}

	// InstalledRef is one skill already present in the project manifest.
	InstalledRef struct {
		Name     string `json:"name"`
		Version  string `json:"version"`
		Pinned   bool   `json:"pinned"`
		Personal bool   `json:"personal"`
	}

	// ManifestState summarizes the project's rsk.mod, when it exists.
	ManifestState struct {
		Exists bool `json:"exists"`
		Skills int  `json:"skills"`
		Pinned int  `json:"pinned"`
	}

	// SearchSkillsIn is the search_skills tool input.
	SearchSkillsIn struct {
		Query string `json:"query"`
		Limit int    `json:"limit,omitempty" jsonschema:"Por defecto 10, máximo 25"`
	}

	// SearchSkillsOut is the search_skills tool output. It is an object rather
	// than a bare array: some clients reject a top-level array structured result.
	SearchSkillsOut struct {
		Candidates []Candidate `json:"candidates"`
	}
)

func registerTools(srv *sdk.Server, deps Deps) {
	sdk.AddTool(srv, &sdk.Tool{
		Name:  "project_profile",
		Title: "Project profile",
		Description: "Detect this project's stack from its files and propose the " +
			"ralvaskills skills that apply. Call it before scaffolding or a " +
			"non-trivial change. Read-only.",
		Annotations: readOnlyAnnotations("Project profile"),
	}, profileHandler(deps))

	sdk.AddTool(srv, &sdk.Tool{
		Name:  "search_skills",
		Title: "Search skills",
		Description: "Search the ralvaskills catalog by keyword. Use it to find a " +
			"skill by name or intent when project_profile does not name one. " +
			"Read-only.",
		Annotations: readOnlyAnnotations("Search skills"),
	}, searchHandler(deps))

	installTool := &sdk.Tool{
		Name:  "install_skills",
		Title: "Install skills",
		Description: "Install one or more skills or bundles into this project by " +
			"name, resolving them through the ralvaskills catalog. Writes symlinks " +
			"and the project manifest. The client asks the user to approve.",
		Annotations: mutatingAnnotations("Install skills"),
	}
	installTool.Meta = sdk.Meta{requiresUserInteractionKey: true}
	sdk.AddTool(srv, installTool, installHandler(deps))
}

// mutatingAnnotations marks a tool that writes: the client asks for approval.
func mutatingAnnotations(title string) *sdk.ToolAnnotations {
	yes := true
	no := false
	return &sdk.ToolAnnotations{
		Title:           title,
		ReadOnlyHint:    false,
		DestructiveHint: &yes,
		IdempotentHint:  false,
		OpenWorldHint:   &no,
	}
}

// readOnlyAnnotations marks a tool that only reads.
func readOnlyAnnotations(title string) *sdk.ToolAnnotations {
	no := false
	return &sdk.ToolAnnotations{
		Title:           title,
		ReadOnlyHint:    true,
		DestructiveHint: &no,
		IdempotentHint:  true,
		OpenWorldHint:   &no,
	}
}

func profileHandler(deps Deps) sdk.ToolHandlerFor[ProjectProfileIn, ProjectProfileOut] {
	return func(
		ctx context.Context, _ *sdk.CallToolRequest, in ProjectProfileIn,
	) (*sdk.CallToolResult, ProjectProfileOut, error) {
		root := in.Path
		if root == "" {
			root = "."
		}

		prof, err := Scan(ctx, root)
		if err != nil {
			errMsg := mcpkit.ErrorMsg{Public: "could not profile the project"}
			return nil, ProjectProfileOut{}, errMsg.ToolError(deps.Logger, err)
		}
		return nil, buildProfile(ctx, deps, prof), nil
	}
}

func searchHandler(deps Deps) sdk.ToolHandlerFor[SearchSkillsIn, SearchSkillsOut] {
	return func(
		ctx context.Context, _ *sdk.CallToolRequest, in SearchSkillsIn,
	) (*sdk.CallToolResult, SearchSkillsOut, error) {
		entries, err := deps.catalog(ctx)
		if err != nil {
			errMsg := mcpkit.ErrorMsg{Public: "could not search the skill catalog"}
			return nil, SearchSkillsOut{}, errMsg.ToolError(deps.Logger, err)
		}
		out := SearchSkillsOut{Candidates: searchCatalog(entries, in.Query, clampLimit(in.Limit))}
		return nil, out, nil
	}
}

func clampLimit(limit int) int {
	switch {
	case limit <= 0:
		return defaultSearchLimit
	case limit > maxSearchLimit:
		return maxSearchLimit
	default:
		return limit
	}
}

// buildProfile enriches the deterministic scan with catalog and manifest state.
// A catalog failure is non-fatal: the signals still answer the question, only
// the latest/installed annotations are missing.
func buildProfile(ctx context.Context, deps Deps, prof Profile) ProjectProfileOut {
	out := ProjectProfileOut{
		ProjectRoot: prof.Root,
		Signals:     prof.Signals,
		Candidates:  prof.Candidates,
	}

	if entries, err := deps.catalog(ctx); err == nil {
		byName := make(map[string]catalog.Entry, len(entries))
		for _, e := range entries {
			byName[e.Name] = e
		}
		out.Candidates = enrichCandidates(prof.Candidates, byName)
	}

	mod, lock, state := readProjectState(prof.Root)
	out.Manifest = state
	out.Installed = installedRefs(lock, mod)

	installed := make(map[string]struct{}, len(lock.Skills))
	for _, le := range lock.Skills {
		installed[le.Name] = struct{}{}
	}
	for i := range out.Candidates {
		if _, ok := installed[out.Candidates[i].Skill]; ok {
			out.Candidates[i].Installed = true
		}
	}
	return out
}

// enrichCandidates drops personal skills and fills Latest from the catalog.
func enrichCandidates(candidates []Candidate, byName map[string]catalog.Entry) []Candidate {
	out := make([]Candidate, 0, len(candidates))
	for _, c := range candidates {
		entry, ok := byName[c.Skill]
		if ok && entry.Personal {
			continue
		}
		if ok {
			c.Latest = entry.Latest
		}
		out = append(out, c)
	}
	return out
}

// readProjectState reads the project's rsk.mod/rsk.lock if present. A missing
// manifest is not an error: it just means the project does not use rsk yet.
func readProjectState(projectRoot string) (manifest.Mod, manifest.Lock, *ManifestState) {
	rskDir := filepath.Join(projectRoot, internal.ProjectFolderName)
	if _, err := os.Stat(manifest.ModPath(rskDir)); err != nil {
		return manifest.Mod{}, manifest.Lock{}, &ManifestState{}
	}

	mod, err := manifest.ReadMod(rskDir)
	if err != nil {
		return manifest.Mod{}, manifest.Lock{}, &ManifestState{Exists: true}
	}
	lock, lockErr := manifest.ReadLock(rskDir)
	if lockErr != nil {
		// A malformed lock must not break profiling; report no installs.
		lock = manifest.Lock{}
	}
	return mod, lock, &ManifestState{
		Exists: true,
		Skills: len(mod.Skills),
		Pinned: len(mod.Pinned),
	}
}

func installedRefs(lock manifest.Lock, mod manifest.Mod) []InstalledRef {
	pinned := make(map[string]struct{}, len(mod.Pinned))
	for _, name := range mod.Pinned {
		pinned[name] = struct{}{}
	}

	refs := make([]InstalledRef, 0, len(lock.Skills))
	for _, le := range lock.Skills {
		_, isPinned := pinned[le.Name]
		refs = append(refs, InstalledRef{Name: le.Name, Version: le.Version, Pinned: isPinned})
	}
	return refs
}

// searchCatalog scores catalog entries against the query tokens, weighing a
// name match over a description match, and returns at most limit candidates.
func searchCatalog(entries []catalog.Entry, query string, limit int) []Candidate {
	out := make([]Candidate, 0)
	tokens := strings.Fields(fold(strings.ToLower(query)))
	if len(tokens) == 0 {
		return out
	}

	type hit struct {
		entry catalog.Entry
		score int
	}
	hits := make([]hit, 0, len(entries))
	for _, e := range entries {
		if e.Personal {
			continue
		}

		name := fold(strings.ToLower(e.Name))
		desc := fold(strings.ToLower(e.Description))
		score := 0
		for _, token := range tokens {
			if strings.Contains(name, token) {
				score += searchNameWeight
			}
			if strings.Contains(desc, token) {
				score++
			}
		}
		if score > 0 {
			hits = append(hits, hit{entry: e, score: score})
		}
	}

	slices.SortFunc(hits, func(a, b hit) int {
		if a.score != b.score {
			return cmp.Compare(b.score, a.score)
		}
		return cmp.Compare(a.entry.Name, b.entry.Name)
	})

	for i, h := range hits {
		if i >= limit {
			break
		}
		out = append(out, Candidate{
			Skill:   h.entry.Name,
			Latest:  h.entry.Latest,
			Because: []SignalKind{},
		})
	}
	return out
}

// accentReplacer strips the Spanish accents rsk skill names and descriptions
// use, so a query typed without accents still matches.
var accentReplacer = strings.NewReplacer(
	"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n",
)

func fold(s string) string {
	return accentReplacer.Replace(s)
}
