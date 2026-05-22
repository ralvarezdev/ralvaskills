// generate-registry walks the skills/ directory, creates versioned tarballs for
// any skill whose version is not yet in the existing index, and writes an
// updated index.json plus a new-versions.json for the CI publish step.
package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ---- index schema -------------------------------------------------------

type Index struct {
	Version     int                     `json:"version"`
	GeneratedAt string                  `json:"generated_at"`
	Skills      map[string]*SkillEntry  `json:"skills"`
}

type SkillEntry struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Personal    bool                      `json:"personal,omitempty"`
	Latest      string                    `json:"latest"`
	Versions    map[string]*VersionEntry  `json:"versions"`
}

type VersionEntry struct {
	Version     string `json:"version"`
	PublishedAt string `json:"published_at"`
	ArchiveURL  string `json:"archive_url"`
}

// NewVersion is written to new-versions.json so CI knows what to publish.
type NewVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Archive string `json:"archive"` // filename inside output-dir
}

// ---- skill discovery ----------------------------------------------------

type skillInfo struct {
	Name        string
	Version     string
	Description string
	Personal    bool
	Path        string
}

func walkSkills(root string) ([]skillInfo, error) {
	var skills []skillInfo
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		skillMD := filepath.Join(path, "SKILL.md")
		if _, statErr := os.Stat(skillMD); os.IsNotExist(statErr) {
			return nil
		}
		version, description, parseErr := readFrontmatter(skillMD)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "warn: skip %s — %v\n", path, parseErr)
			return fs.SkipDir
		}
		rel, _ := filepath.Rel(root, path)
		personal := strings.Contains(filepath.ToSlash(rel), "/personal/") ||
			strings.Contains(filepath.ToSlash(rel), "personal/")
		skills = append(skills, skillInfo{
			Name:        filepath.Base(path),
			Version:     version,
			Description: description,
			Personal:    personal,
			Path:        path,
		})
		return fs.SkipDir // recursion stops at first SKILL.md
	})
	return skills, err
}

func readFrontmatter(path string) (version, description string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", "", err
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
		if !inFrontmatter {
			continue
		}
		if v, ok := strings.CutPrefix(line, "version:"); ok {
			version = strings.Trim(strings.TrimSpace(v), `"'`)
		}
		if v, ok := strings.CutPrefix(line, "description:"); ok {
			description = strings.Trim(strings.TrimSpace(v), `"'`)
		}
	}
	if version == "" {
		return "", "", fmt.Errorf("version field not found in %s", path)
	}
	return version, description, nil
}

// ---- tarball creation ---------------------------------------------------

// createTarball packs skillPath into a .tar.gz at dest.
// Archive entries are rooted at skillName/ so extracting creates a single directory.
func createTarball(skillPath, skillName, dest string) error {
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	return filepath.WalkDir(skillPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(skillPath, path)
		if err != nil {
			return err
		}
		// Archive path: skillName/rel (forward slashes)
		archivePath := skillName + "/" + filepath.ToSlash(rel)

		info, err := d.Info()
		if err != nil {
			return err
		}
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = archivePath
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()
		_, err = io.Copy(tw, src)
		return err
	})
}

// ---- index helpers ------------------------------------------------------

func loadIndex(path string) *Index {
	if path == "" {
		return &Index{Version: 1, Skills: make(map[string]*SkillEntry)}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return &Index{Version: 1, Skills: make(map[string]*SkillEntry)}
	}
	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return &Index{Version: 1, Skills: make(map[string]*SkillEntry)}
	}
	if idx.Skills == nil {
		idx.Skills = make(map[string]*SkillEntry)
	}
	return &idx
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// ---- main ---------------------------------------------------------------

func main() {
	skillsDir := flag.String("skills-dir", "skills", "Path to skills directory")
	outputDir := flag.String("output-dir", "dist", "Output directory for tarballs and index")
	existingIndex := flag.String("existing-index", "", "Path to existing index.json to merge with")
	githubRepo := flag.String("github-repo", "ralvarezdev/ralvaskills", "GitHub repo (owner/name) used to build release asset URLs")
	flag.Parse()

	index := loadIndex(*existingIndex)

	skills, err := walkSkills(*skillsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error walking skills: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(*outputDir, 0o750); err != nil {
		fmt.Fprintf(os.Stderr, "error creating output dir: %v\n", err)
		os.Exit(1)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	var newVersions []NewVersion

	for _, s := range skills {
		entry, exists := index.Skills[s.Name]
		if !exists {
			entry = &SkillEntry{
				Name:     s.Name,
				Personal: s.Personal,
				Versions: make(map[string]*VersionEntry),
			}
			index.Skills[s.Name] = entry
		}

		// Already published — nothing to do.
		if _, published := entry.Versions[s.Version]; published {
			continue
		}

		archiveFile := fmt.Sprintf("%s-v%s.tar.gz", s.Name, s.Version)
		archivePath := filepath.Join(*outputDir, archiveFile)

		if err := createTarball(s.Path, s.Name, archivePath); err != nil {
			fmt.Fprintf(os.Stderr, "error creating tarball for %s: %v\n", s.Name, err)
			os.Exit(1)
		}

		tag := url.PathEscape(fmt.Sprintf("%s@v%s", s.Name, s.Version))
		entry.Description = s.Description
		entry.Latest = s.Version
		entry.Versions[s.Version] = &VersionEntry{
			Version:     s.Version,
			PublishedAt: now,
			ArchiveURL:  fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", *githubRepo, tag, archiveFile),
		}

		newVersions = append(newVersions, NewVersion{
			Name:    s.Name,
			Version: s.Version,
			Archive: archiveFile,
		})
		fmt.Printf("+ %s@v%s\n", s.Name, s.Version)
	}

	index.GeneratedAt = now

	if err := writeJSON(filepath.Join(*outputDir, "index.json"), index); err != nil {
		fmt.Fprintf(os.Stderr, "error writing index.json: %v\n", err)
		os.Exit(1)
	}

	if err := writeJSON(filepath.Join(*outputDir, "new-versions.json"), newVersions); err != nil {
		fmt.Fprintf(os.Stderr, "error writing new-versions.json: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("index: %d skills total, %d new versions\n", len(index.Skills), len(newVersions))
}
