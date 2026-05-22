package source

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ralvarezdev/ralvaskills/cli/internal/skill"
)

const defaultHTTPTimeout = 30 * time.Second

// indexSkillEntry mirrors the SkillEntry shape in the published index.json.
type indexSkillEntry struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Personal    bool                      `json:"personal"`
	Latest      string                    `json:"latest"`
	Versions    map[string]*indexVersion  `json:"versions"`
}

type indexVersion struct {
	Version    string `json:"version"`
	ArchiveURL string `json:"archive_url"`
}

type registryIndex struct {
	Skills map[string]*indexSkillEntry `json:"skills"`
}

// Registry resolves skills from the hosted registry at baseURL.
// Skills are downloaded as tarballs and cached in cacheDir.
type Registry struct {
	baseURL  string
	cacheDir string
	client   *http.Client
}

// NewRegistry returns a Registry backed by baseURL, caching extractions in cacheDir.
func NewRegistry(baseURL, cacheDir string) *Registry {
	return &Registry{
		baseURL:  baseURL,
		cacheDir: cacheDir,
		client:   &http.Client{Timeout: defaultHTTPTimeout},
	}
}

// Index fetches and returns the registry index.
func (r *Registry) Index(ctx context.Context) (map[string]*indexSkillEntry, error) {
	url := r.baseURL + "/index.json"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build index request: %w", err)
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch index: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch index: HTTP %d from %s", resp.StatusCode, url)
	}
	var idx registryIndex
	if err := json.NewDecoder(resp.Body).Decode(&idx); err != nil {
		return nil, fmt.Errorf("decode index: %w", err)
	}
	return idx.Skills, nil
}

// All fetches the index and returns every non-personal skill at its latest version.
func (r *Registry) All(ctx context.Context) ([]skill.Skill, error) {
	index, err := r.Index(ctx)
	if err != nil {
		return nil, err
	}
	skills := make([]skill.Skill, 0, len(index))
	for _, entry := range index {
		skills = append(skills, skill.Skill{
			Name:       entry.Name,
			Version:    entry.Latest,
			Source:     skill.SourceRegistry,
			IsPersonal: entry.Personal,
		})
	}
	return skills, nil
}

// Find returns the skill at its latest version, downloading and caching the
// tarball if it is not already present in cacheDir.
func (r *Registry) Find(ctx context.Context, name string) (skill.Skill, error) {
	return r.FindVersion(ctx, name, "")
}

// FindVersion returns the skill at a specific version (empty string = latest).
func (r *Registry) FindVersion(ctx context.Context, name, version string) (skill.Skill, error) {
	index, err := r.Index(ctx)
	if err != nil {
		return skill.Skill{}, err
	}
	entry, ok := index[name]
	if !ok {
		return skill.Skill{}, fmt.Errorf("%w: registry skill %q not found", ErrNotFound, name)
	}
	if version == "" {
		version = entry.Latest
	}
	ver, ok := entry.Versions[version]
	if !ok {
		return skill.Skill{}, fmt.Errorf("%w: registry skill %q has no version %s", ErrNotFound, name, version)
	}

	skillDir, err := r.ensureCached(ctx, name, version, ver.ArchiveURL)
	if err != nil {
		return skill.Skill{}, fmt.Errorf("cache skill %q@%s: %w", name, version, err)
	}

	return skill.Skill{
		Name:       name,
		Version:    version,
		Path:       skillDir,
		Source:     skill.SourceRegistry,
		IsPersonal: entry.Personal,
	}, nil
}

// ensureCached downloads and extracts the tarball for name@version if not
// already present in cacheDir. Returns the path to the extracted skill directory.
func (r *Registry) ensureCached(ctx context.Context, name, version, archiveURL string) (string, error) {
	skillDir := filepath.Join(r.cacheDir, name, version)
	if _, err := os.Stat(skillDir); err == nil {
		return skillDir, nil // already cached
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, archiveURL, nil)
	if err != nil {
		return "", fmt.Errorf("build download request: %w", err)
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download %s: %w", archiveURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download %s: HTTP %d", archiveURL, resp.StatusCode)
	}

	if err := os.MkdirAll(skillDir, 0o750); err != nil {
		return "", fmt.Errorf("create cache dir: %w", err)
	}

	if err := extractTarball(resp.Body, skillDir, name); err != nil {
		_ = os.RemoveAll(skillDir) // clean up partial extraction
		return "", fmt.Errorf("extract tarball: %w", err)
	}

	return skillDir, nil
}

// extractTarball extracts a .tar.gz stream into destDir.
// Entries rooted at skillName/ are stripped of that prefix so the skill files
// land directly in destDir.
//
// Only regular files and directories are extracted. Symlinks and hard links are
// rejected to prevent link-based escape attacks. All resolved paths are
// validated to remain inside destDir (zip-slip protection).
func extractTarball(r io.Reader, destDir, skillName string) error {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gr.Close()

	// destDir must be absolute and clean so the HasPrefix check below is reliable.
	destDir = filepath.Clean(destDir)
	safeRoot := destDir + string(filepath.Separator)

	prefix := skillName + "/"
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch hdr.Typeflag {
		case tar.TypeDir, tar.TypeReg:
			// allowed
		default:
			// Reject symlinks, hard links, devices, etc.
			return fmt.Errorf("unsupported tar entry type %d for %q", hdr.Typeflag, hdr.Name)
		}

		// Strip the leading skillName/ prefix from archive paths.
		rel := hdr.Name
		if len(rel) > len(prefix) && rel[:len(prefix)] == prefix {
			rel = rel[len(prefix):]
		}
		if rel == "" || rel == "." {
			continue
		}

		target := filepath.Join(destDir, filepath.FromSlash(rel))

		// Zip-slip guard: resolved path must stay inside destDir.
		if target != destDir && !strings.HasPrefix(target, safeRoot) {
			return fmt.Errorf("archive entry %q escapes destination directory", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o750); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode)&0o777)
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(f, tr)
			closeErr := f.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		}
	}
	return nil
}
