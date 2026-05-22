package skill

import "testing"

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input   string
		want    Version
		wantErr bool
	}{
		{"1.2.3", Version{1, 2, 3}, false},
		{"v1.2.3", Version{1, 2, 3}, false},
		{"0.0.0", Version{0, 0, 0}, false},
		{"10.20.30", Version{10, 20, 30}, false},
		{"1.2", Version{}, true},
		{"abc", Version{}, true},
		{"1.x.3", Version{}, true},
		{"", Version{}, true},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, err := ParseVersion(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ParseVersion(%q): expected error, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseVersion(%q): unexpected error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("ParseVersion(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestVersion_String(t *testing.T) {
	v := Version{1, 2, 3}
	if got := v.String(); got != "1.2.3" {
		t.Errorf("String() = %q, want %q", got, "1.2.3")
	}
}

func TestVersion_Compare(t *testing.T) {
	tests := []struct {
		a, b Version
		want int
	}{
		{Version{1, 0, 0}, Version{2, 0, 0}, -1},
		{Version{2, 0, 0}, Version{1, 0, 0}, 1},
		{Version{1, 0, 0}, Version{1, 0, 0}, 0},
		{Version{1, 1, 0}, Version{1, 2, 0}, -1},
		{Version{1, 0, 1}, Version{1, 0, 2}, -1},
		{Version{1, 2, 3}, Version{1, 2, 3}, 0},
	}
	for _, tc := range tests {
		got := tc.a.Compare(tc.b)
		if got != tc.want {
			t.Errorf("%v.Compare(%v) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestVersion_BeforeAfter(t *testing.T) {
	old := Version{1, 0, 0}
	new := Version{2, 0, 0}

	if !old.Before(new) {
		t.Errorf("expected %v.Before(%v)", old, new)
	}
	if old.After(new) {
		t.Errorf("unexpected %v.After(%v)", old, new)
	}
	if new.Before(old) {
		t.Errorf("unexpected %v.Before(%v)", new, old)
	}
	if !new.After(old) {
		t.Errorf("expected %v.After(%v)", new, old)
	}
	if old.Before(old) || old.After(old) {
		t.Errorf("equal version should be neither Before nor After itself")
	}
}
