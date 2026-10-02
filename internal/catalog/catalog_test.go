package catalog_test

import (
	"context"
	"slices"
	"testing"

	"github.com/ralvarezdev/ralvaskills/v3/internal/catalog"
	"github.com/ralvarezdev/ralvaskills/v3/internal/skill"
	"github.com/ralvarezdev/ralvaskills/v3/internal/source"
)

func TestFromSkillsAndFromIndexAgree(t *testing.T) {
	t.Parallel()

	skills := []skill.Skill{
		{Name: "tdd", Version: "1.0.0", Description: "Red-green-refactor."},
		{Name: "go-architect", Version: "1.2.0", Description: "Go 1.26 standards."},
	}
	index := map[string]*source.IndexEntry{
		"tdd":          {Name: "tdd", Latest: "1.0.0", Description: "Red-green-refactor."},
		"go-architect": {Name: "go-architect", Latest: "1.2.0", Description: "Go 1.26 standards."},
	}

	fromSkills := catalog.FromSkills(skills)
	fromIndex := catalog.FromIndex(index)
	if !slices.Equal(fromSkills, fromIndex) {
		t.Errorf("FromSkills = %+v, FromIndex = %+v, want equal", fromSkills, fromIndex)
	}
}

func TestBuildSortsByName(t *testing.T) {
	t.Parallel()

	src := stubResolver{skills: []skill.Skill{
		{Name: "zeta", Version: "1.0.0"},
		{Name: "alpha", Version: "2.0.0"},
	}}
	entries, err := catalog.Build(t.Context(), src)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	want := []catalog.Entry{
		{Name: "alpha", Latest: "2.0.0"},
		{Name: "zeta", Latest: "1.0.0"},
	}
	if !slices.Equal(entries, want) {
		t.Errorf("Build = %+v, want %+v", entries, want)
	}
}

func TestExcludePersonal(t *testing.T) {
	t.Parallel()

	entries := []catalog.Entry{
		{Name: "public"},
		{Name: "private", Personal: true},
	}
	got := catalog.ExcludePersonal(entries)
	want := []catalog.Entry{{Name: "public"}}
	if !slices.Equal(got, want) {
		t.Errorf("ExcludePersonal = %+v, want %+v", got, want)
	}
}

type stubResolver struct {
	skills []skill.Skill
}

func (s stubResolver) All(context.Context) ([]skill.Skill, error) {
	return s.skills, nil
}

func (s stubResolver) Find(context.Context, string) (skill.Skill, error) {
	return skill.Skill{}, nil
}
