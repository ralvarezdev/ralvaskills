package skill

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const skillFile = "SKILL.md"

// Walk discovers all skills under root by looking for SKILL.md files.
// Any directory containing SKILL.md is a skill; recursion stops there.
func Walk(root string, source Source) ([]Skill, error) {
	var skills []Skill
	if err := walkDir(root, root, source, &skills); err != nil {
		return nil, err
	}
	return skills, nil
}

func walkDir(root, dir string, source Source, out *[]Skill) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read dir %s: %w", dir, err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		subdir := filepath.Join(dir, e.Name())
		_, statErr := os.Stat(filepath.Join(subdir, skillFile))
		if statErr == nil {
			s, parseErr := parseSkill(root, subdir, source)
			if parseErr != nil {
				return parseErr
			}
			*out = append(*out, s)
			continue
		}
		if walkErr := walkDir(root, subdir, source, out); walkErr != nil {
			return walkErr
		}
	}
	return nil
}

func parseSkill(root, dir string, source Source) (Skill, error) {
	version, err := readVersionFromFrontmatter(filepath.Join(dir, skillFile))
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
	return readVersionFromFrontmatter(filepath.Join(dir, skillFile))
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
		if inFrontmatter && strings.HasPrefix(line, "version:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "version:"))
			return strings.Trim(val, `"'`), nil
		}
	}
	return "", fmt.Errorf("version field not found in %s", path)
}
