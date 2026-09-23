package main

import (
	"reflect"
	"testing"

	"github.com/ralvarezdev/ralvaskills/internal/skill"
)

func TestUpdatePairNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []updatePair
		want []string
	}{
		{name: "empty", in: nil, want: []string{}},
		{
			name: "several",
			in: []updatePair{
				{name: "go-architect", installed: "1.0.0", latest: "1.1.0", newSkill: skill.Skill{}},
				{name: "tdd", installed: "1.0.0", latest: "1.0.1", newSkill: skill.Skill{}},
			},
			want: []string{"go-architect", "tdd"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := updatePairNames(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("updatePairNames(%v) = %#v, want %#v", tt.in, got, tt.want)
			}
		})
	}
}
