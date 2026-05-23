package skill

import (
	"path/filepath"
	"slices"
	"strings"
)

// IsPersonalPath reports whether path contains a "personal" path segment.
func IsPersonalPath(path string) bool {
	clean := filepath.ToSlash(filepath.Clean(path))
	return slices.Contains(strings.Split(clean, "/"), PersonalFolderName)
}

// FindByName returns the first skill in skills whose Name equals name, and
// whether one was found.
func FindByName(skills []Skill, name string) (Skill, bool) {
	for _, s := range skills {
		if s.Name == name {
			return s, true
		}
	}
	return Skill{}, false
}
