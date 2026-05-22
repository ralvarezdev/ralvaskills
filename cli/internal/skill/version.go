package skill

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is a parsed semver triplet.
type Version struct {
	Major int
	Minor int
	Patch int
}

// ParseVersion parses a semver string of the form "MAJOR.MINOR.PATCH" or "vMAJOR.MINOR.PATCH".
func ParseVersion(s string) (Version, error) {
	s = strings.TrimPrefix(s, "v")
	parts := strings.SplitN(s, ".", 3)
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version %q: want MAJOR.MINOR.PATCH", s)
	}
	var v Version
	var err error
	if v.Major, err = strconv.Atoi(parts[0]); err != nil {
		return Version{}, fmt.Errorf("invalid major in %q: %w", s, err)
	}
	if v.Minor, err = strconv.Atoi(parts[1]); err != nil {
		return Version{}, fmt.Errorf("invalid minor in %q: %w", s, err)
	}
	if v.Patch, err = strconv.Atoi(parts[2]); err != nil {
		return Version{}, fmt.Errorf("invalid patch in %q: %w", s, err)
	}
	return v, nil
}

// String formats v as "MAJOR.MINOR.PATCH".
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare returns -1, 0, or 1 when v is less than, equal to, or greater than other.
func (v Version) Compare(other Version) int {
	if c := cmpInt(v.Major, other.Major); c != 0 {
		return c
	}
	if c := cmpInt(v.Minor, other.Minor); c != 0 {
		return c
	}
	return cmpInt(v.Patch, other.Patch)
}

// Before reports whether v is strictly older than other.
func (v Version) Before(other Version) bool { return v.Compare(other) < 0 }

// After reports whether v is strictly newer than other.
func (v Version) After(other Version) bool { return v.Compare(other) > 0 }

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
