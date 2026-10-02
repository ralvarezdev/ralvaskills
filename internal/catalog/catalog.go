// Package catalog presents skills from any rsk source in a single shape, so
// consumers such as the MCP server can list and search them without caring
// whether a skill came from a local clone or the hosted registry.
package catalog

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"github.com/ralvarezdev/ralvaskills/v3/internal/skill"
	"github.com/ralvarezdev/ralvaskills/v3/internal/source"
)

// Entry is one skill as the catalog exposes it. Latest is the version a fresh
// install would resolve; for a local source it is the version on disk.
type Entry struct {
	Name        string
	Description string
	Latest      string
	Personal    bool
}

// Build lists src and returns its catalog, sorted by name. It works for any
// source.Resolver — local, official and registry skills all carry a description
// once resolved.
func Build(ctx context.Context, src source.Resolver) ([]Entry, error) {
	skills, err := src.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list catalog: %w", err)
	}
	return FromSkills(skills), nil
}

// FromSkills maps resolved skills to catalog entries, sorted by name.
func FromSkills(skills []skill.Skill) []Entry {
	entries := make([]Entry, 0, len(skills))
	for _, s := range skills {
		entries = append(entries, Entry{
			Name:        s.Name,
			Description: s.Description,
			Latest:      s.Version,
			Personal:    s.IsPersonal,
		})
	}
	return sortByName(entries)
}

// FromIndex maps a registry index to catalog entries, sorted by name. It is the
// offline path: the caller supplies an index it already holds (for example from
// the disk cache in internal/mcp) instead of hitting the network.
func FromIndex(index map[string]*source.IndexEntry) []Entry {
	entries := make([]Entry, 0, len(index))
	for _, e := range index {
		entries = append(entries, Entry{
			Name:        e.Name,
			Description: e.Description,
			Latest:      e.Latest,
			Personal:    e.Personal,
		})
	}
	return sortByName(entries)
}

// ExcludePersonal returns the entries that are not personal. Personal skills
// are private to their author and are never proposed to other developers.
func ExcludePersonal(entries []Entry) []Entry {
	kept := make([]Entry, 0, len(entries))
	for _, e := range entries {
		if !e.Personal {
			kept = append(kept, e)
		}
	}
	return kept
}

func sortByName(entries []Entry) []Entry {
	slices.SortFunc(entries, func(a, b Entry) int {
		return cmp.Compare(a.Name, b.Name)
	})
	return entries
}
