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
	// VersionPrefix is the prefix in SKILL.md frontmatter for the version field.
	VersionPrefix = "version:"
)

// Walk discovers all skills under root by looking for SKILL.md files.
// Any directory containing SKILL.md is a skill; recursion stops there.
func Walk(root string, source Source) ([]Skill, error) {
	var skills []Skill
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return err
		}
		if _, statErr := os.Stat(filepath.Join(path, SkillFileName)); statErr != nil {
			return nil
		}
		s, parseErr := parseSkill(root, path, source)
		if parseErr != nil {
			return parseErr
		}
		skills = append(skills, s)
		return fs.SkipDir
	})
	return skills, err
}

func parseSkill(root, dir string, source Source) (Skill, error) {
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
	}, nil
}

// ReadVersion reads the version field from SKILL.md in dir.
func ReadVersion(dir string) (string, error) {
	return readVersionFromFrontmatter(filepath.Join(dir, SkillFileName))
}

// readVersionFromFrontmatter reads the "version:" value from YAML frontmatter in path.
func readVersionFromFrontmatter(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

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
