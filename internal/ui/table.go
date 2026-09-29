package ui

// SkillName returns a styled skill name for table rows.
func SkillName(name string) string {
	return BoldStyle.Render(name)
}

// SkillVersion returns a styled version string for table rows.
func SkillVersion(version string) string {
	if version == "" {
		return MutedStyle.Render("-")
	}
	return MutedStyle.Render(version)
}

// MutedPath returns a dimmed path string.
func MutedPath(p string) string {
	return MutedStyle.Render(p)
}
