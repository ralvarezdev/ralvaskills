package main

import "testing"

func TestConstraintsFromArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want map[string]string
	}{
		{
			name: "no version constraints",
			args: []string{"go-grpc", "docs"},
			want: map[string]string{},
		},
		{
			name: "single skill with version",
			args: []string{"go-architect@1.0.0"},
			want: map[string]string{"go-architect": "1.0.0"},
		},
		{
			name: "mixed bundle and pinned skill",
			args: []string{"go-grpc", "go-architect@1.0.0"},
			want: map[string]string{"go-architect": "1.0.0"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := constraintsFromArgs(tc.args)
			if err != nil {
				t.Fatalf("constraintsFromArgs(%v): unexpected error: %v", tc.args, err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("constraintsFromArgs(%v) = %v, want %v", tc.args, got, tc.want)
			}
			for k, v := range tc.want {
				if got[k] != v {
					t.Errorf("constraintsFromArgs(%v)[%q] = %q, want %q", tc.args, k, got[k], v)
				}
			}
		})
	}

	if _, err := constraintsFromArgs([]string{"@1.0.0"}); err == nil {
		t.Error("constraintsFromArgs with invalid name@version: expected error, got nil")
	}
}
