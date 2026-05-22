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
