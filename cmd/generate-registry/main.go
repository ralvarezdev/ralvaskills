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

	"github.com/ralvarezdev/ralvaskills/internal/fsperm"
	"github.com/ralvarezdev/ralvaskills/internal/fsx"
	"github.com/ralvarezdev/ralvaskills/internal/skill"
)

const (
	descriptionField    = "description:"
	frontmatterDelim    = "---"
	archiveSuffix       = ".tar.gz"
	tempFilePattern     = ".rsk-gen-*.tmp"
	releaseURLPath      = "releases/download"
	indexFileName       = "index.json"
	newVersionsFileName = "new-versions.json"
	gitHubRepoDefault   = "ralvarezdev/ralvaskills"
)

type (
	// Index is the top-level shape of the published index.json.
	Index struct {
		Version     int                    `json:"version"`
		GeneratedAt string                 `json:"generated_at"`
		Skills      map[string]*SkillEntry `json:"skills"`
	}

	// SkillEntry records a single skill's metadata and version history in the
	// published index.
	SkillEntry struct {
		Name        string                   `json:"name"`
		Description string                   `json:"description"`
		Personal    bool                     `json:"personal,omitempty"`
		Latest      string                   `json:"latest"`
		Versions    map[string]*VersionEntry `json:"versions"`
	}

	// VersionEntry records a single published version of a skill.
	VersionEntry struct {
		Version     string `json:"version"`
		PublishedAt string `json:"published_at"`
		ArchiveURL  string `json:"archive_url"`
	}

	// NewVersion is written to new-versions.json so the CI publish step knows
	// which archive files to upload as GitHub release assets.
	NewVersion struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Archive string `json:"archive"`
	}

	skillInfo struct {
		Name        string
		Version     string
		Description string
		Personal    bool
		Path        string
	}
)

func walkSkills(root string) ([]skillInfo, error) {
	var skills []skillInfo
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return err
		}

		skillMD := filepath.Join(path, skill.SkillFileName)
		if _, statErr := os.Stat(skillMD); os.IsNotExist(statErr) {
			return nil
		}

		version, description, parseErr := readFrontmatter(skillMD)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "warn: skip %s — %v\n", path, parseErr)
			return fs.SkipDir
		}

		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			rel = path
		}
		skills = append(skills, skillInfo{
			Name:        filepath.Base(path),
			Version:     version,
			Description: description,
			Personal:    skill.IsPersonalPath(rel),
			Path:        path,
		})
		return fs.SkipDir
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
		if line == frontmatterDelim {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			}
			break
		}
		if !inFrontmatter {
			continue
		}
		if v, ok := strings.CutPrefix(line, skill.VersionPrefix); ok {
			version = strings.Trim(strings.TrimSpace(v), `"'`)
		}
		if v, ok := strings.CutPrefix(line, descriptionField); ok {
			description = strings.Trim(strings.TrimSpace(v), `"'`)
		}
	}
	if version == "" {
		return "", "", fmt.Errorf("version field not found in %s", path)
	}
	return version, description, nil
}

// createTarball packs skillPath into a .tar.gz at dest. Archive entries are
// rooted at skillName/ so extracting creates a single directory. On failure
// the partial file at dest is removed.
func createTarball(skillPath, skillName, dest string) (retErr error) {
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
		if retErr != nil {
			os.Remove(dest)
		}
	}()

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	walkErr := filepath.WalkDir(skillPath, func(path string, d fs.DirEntry, walkFnErr error) error {
		if walkFnErr != nil {
			return walkFnErr
		}
		rel, relErr := filepath.Rel(skillPath, path)
		if relErr != nil {
			return relErr
		}
		archivePath := skillName + "/" + filepath.ToSlash(rel)

		info, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}
		hdr, hdrErr := tar.FileInfoHeader(info, "")
		if hdrErr != nil {
			return hdrErr
		}
		hdr.Name = archivePath
		if writeErr := tw.WriteHeader(hdr); writeErr != nil {
			return writeErr
		}
		if d.IsDir() {
			return nil
		}
		src, openErr := os.Open(path)
		if openErr != nil {
			return openErr
		}
		defer src.Close()
		_, copyErr := io.Copy(tw, src)
		return copyErr
	})
	if walkErr != nil {
		return walkErr
	}
	if err = tw.Close(); err != nil {
		return err
	}
	return gw.Close()
}

func loadIndex(path string) *Index {
	if path == "" {
		return &Index{Version: 1, Skills: make(map[string]*SkillEntry)}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return &Index{Version: 1, Skills: make(map[string]*SkillEntry)}
	}
	var idx Index
	if err = json.Unmarshal(data, &idx); err != nil {
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
	data = append(data, '\n')
	return fsx.WriteAtomic(path, tempFilePattern, func(w io.Writer) error {
		_, writeErr := w.Write(data)
		return writeErr
	})
}

func main() {
	skillsDir := flag.String("skills-dir", "skills", "Path to skills directory")
	outputDir := flag.String("output-dir", "dist", "Output directory for tarballs and index")
	existingIndex := flag.String("existing-index", "", "Path to existing index.json to merge with")
	githubRepo := flag.String(
		"github-repo", gitHubRepoDefault,
		"GitHub repo (owner/name) used to build release asset URLs",
	)
	flag.Parse()

	index := loadIndex(*existingIndex)

	skills, err := walkSkills(*skillsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error walking skills: %v\n", err)
		os.Exit(1)
	}

	if err = os.MkdirAll(*outputDir, fsperm.Dir); err != nil {
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

		if _, published := entry.Versions[s.Version]; published {
			continue
		}

		archiveFile := fmt.Sprintf("%s-v%s%s", s.Name, s.Version, archiveSuffix)
		archivePath := filepath.Join(*outputDir, archiveFile)

		if err = createTarball(s.Path, s.Name, archivePath); err != nil {
			fmt.Fprintf(os.Stderr, "error creating tarball for %s: %v\n", s.Name, err)
			os.Exit(1)
		}

		tag := url.PathEscape(fmt.Sprintf("%s@v%s", s.Name, s.Version))
		entry.Description = s.Description
		entry.Latest = s.Version
		entry.Versions[s.Version] = &VersionEntry{
			Version:     s.Version,
			PublishedAt: now,
			ArchiveURL:  fmt.Sprintf("https://github.com/%s/%s/%s/%s", *githubRepo, releaseURLPath, tag, archiveFile),
		}

		newVersions = append(newVersions, NewVersion{
			Name:    s.Name,
			Version: s.Version,
			Archive: archiveFile,
		})
		fmt.Printf("+ %s@v%s\n", s.Name, s.Version)
	}

	index.GeneratedAt = now

	if err = writeJSON(filepath.Join(*outputDir, indexFileName), index); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", indexFileName, err)
		os.Exit(1)
	}

	if err = writeJSON(filepath.Join(*outputDir, newVersionsFileName), newVersions); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", newVersionsFileName, err)
		os.Exit(1)
	}

	fmt.Printf("index: %d skills total, %d new versions\n", len(index.Skills), len(newVersions))
}
