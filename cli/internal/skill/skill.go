// Package skill defines core types and operations for discovered skills on disk.
package skill

// Skill represents a skill discovered on disk.
type Skill struct {
	Name       string
	Version    string
	Path       string // absolute path to the skill folder
	Source     Source
	IsPersonal bool // true when path contains a "personal/" segment
}
