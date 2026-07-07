package skill

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	// VersionPrefix is the YAML frontmatter key used to store a skill's version.
	VersionPrefix = "version:"
)

// Walk discovers all skills under root. A directory that contains SKILL.md is
// a skill root; Walk does not descend into it further. The returned skills all
// have Source set to source.
func Walk(root string, source Source) ([]Skill, error) {
	var skills []Skill
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return err
		}
		if _, statErr := os.Stat(filepath.Join(path, SkillFileName)); statErr == nil {
			s := parseSkill(root, path, source)
			skills = append(skills, s)
			return fs.SkipDir
		}
		return nil
	})
	return skills, err
}

// ReadVersion reads the version field from SKILL.md in dir.
func ReadVersion(dir string) (string, error) {
	return readVersionFromFrontmatter(filepath.Join(dir, SkillFileName))
}

func parseSkill(root, dir string, source Source) Skill {
	version, err := readVersionFromFrontmatter(filepath.Join(dir, SkillFileName))
	if err != nil {
		version = ""
	}

	rel, err := filepath.Rel(root, dir)
	if err != nil {
		rel = dir
	}

	return Skill{
		Name:       filepath.Base(dir),
		Version:    version,
		Path:       dir,
		Source:     source,
		IsPersonal: IsPersonalPath(rel),
	}
}

// readVersionFromFrontmatter reads the "version:" value from YAML frontmatter
// in path. Returns an error if the field is absent.
func readVersionFromFrontmatter(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	inFrontmatter := false
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			}
			break
		}
		if inFrontmatter && strings.HasPrefix(line, VersionPrefix) {
			val := strings.TrimSpace(strings.TrimPrefix(line, VersionPrefix))
			return strings.Trim(val, `"'`), nil
		}
	}
	return "", fmt.Errorf("version field not found in %s", path)
}
