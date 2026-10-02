package skill

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Frontmatter is the subset of a SKILL.md YAML frontmatter that rsk reads.
type Frontmatter struct {
	Version     string
	Description string
}

const (
	// VersionPrefix is the YAML frontmatter key used to store a skill's version.
	VersionPrefix = "version:"

	// DescriptionPrefix is the YAML frontmatter key used to store a skill's
	// description.
	DescriptionPrefix = "description:"
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

// ReadFrontmatter parses the version and description from the SKILL.md in dir.
// A missing description is not an error; a missing version is.
func ReadFrontmatter(dir string) (Frontmatter, error) {
	return readFrontmatter(filepath.Join(dir, SkillFileName))
}

// ReadVersion reads the version field from SKILL.md in dir.
func ReadVersion(dir string) (string, error) {
	fm, err := ReadFrontmatter(dir)
	if err != nil {
		return "", err
	}
	return fm.Version, nil
}

func parseSkill(root, dir string, source Source) Skill {
	fm, err := readFrontmatter(filepath.Join(dir, SkillFileName))
	if err != nil {
		fm = Frontmatter{}
	}

	rel, relErr := filepath.Rel(root, dir)
	if relErr != nil {
		rel = dir
	}

	return Skill{
		Name:        filepath.Base(dir),
		Version:     fm.Version,
		Description: fm.Description,
		Path:        dir,
		Source:      source,
		IsPersonal:  IsPersonalPath(rel),
	}
}

// readFrontmatter reads the "version:" and "description:" values from YAML
// frontmatter in path. Returns an error if the version field is absent.
func readFrontmatter(path string) (Frontmatter, error) {
	f, err := os.Open(path)
	if err != nil {
		return Frontmatter{}, err
	}
	defer func() { _ = f.Close() }()

	var fm Frontmatter
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
		if !inFrontmatter {
			continue
		}
		if v, ok := strings.CutPrefix(line, VersionPrefix); ok {
			fm.Version = strings.Trim(strings.TrimSpace(v), `"'`)
		}
		if v, ok := strings.CutPrefix(line, DescriptionPrefix); ok {
			fm.Description = strings.Trim(strings.TrimSpace(v), `"'`)
		}
	}
	if fm.Version == "" {
		return Frontmatter{}, fmt.Errorf("version field not found in %s", path)
	}
	return fm, nil
}
