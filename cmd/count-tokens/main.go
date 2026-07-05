// count-tokens scans every SKILL.md under skills/, estimates token counts
// (body + description-only + side files), and writes docs/TOKEN_COUNTS.md.
//
// Usage:
//
//	task tokens
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	// proseBytesPerToken is a rough estimate for English markdown prose.
	// Actual range is 3.5–4.5 bytes/token; 4 is a safe midpoint.
	proseBytesPerToken = 4

	// codeBytesPerToken applies inside fenced code blocks: identifiers,
	// punctuation, and lack of whitespace padding tokenize closer to ~3 bytes/token.
	codeBytesPerToken = 3

	// descBytesPerToken applies to description fields: short dense noun phrases
	// with few spaces tokenize closer to ~3 bytes/token.
	descBytesPerToken = 3

	// bytesPerToken is kept for side-file estimation (STACK.md, RECIPES.md) where
	// we only have file size and no content to inspect.
	bytesPerToken = proseBytesPerToken

	// Per-profile description token budgets. The global budget is tight because
	// every project pays it; the session budget is looser because it's project-specific.
	globalDescBudget  int64 = 1500
	sessionDescBudget int64 = 2200

	// File and directory names.
	skillMarkdownFile   = "SKILL.md"
	stackMarkdownFile   = "STACK.md"
	recipesMarkdownFile = "RECIPES.md"
	gitDir              = ".git"
	skillsDir           = "skills"
	catalogFilePath     = "internal/config/catalog.toml"
	docsDir             = "docs"
	tokenCountsFile     = "TOKEN_COUNTS.md"
	tokenCountsJSONFile = "TOKEN_COUNTS.json"

	// Global bundle name.
	globalBundleName = "global"

	// Output file permissions.
	filePermission = 0o644

	// Token estimation thresholds for trimming suggestions.
	bodyThreshold = 2500
	topDescCount  = 10

	// Percentage thresholds for description budget status.
	budgetFullPercent = 100
	budgetWarnPercent = 90
	percentageBase    = 100
)

type (
	skill struct {
		name          string
		category      string
		bodyBytes     int64
		bodyTokens    int64
		descTokens    int64
		stackTokens   int64
		recipesTokens int64
		otherTokens   int64
		otherFiles    []otherFile
	}

	otherFile struct {
		name   string
		tokens int64
	}

	// bundle mirrors a [[bundle]] entry in internal/config/catalog.toml.
	bundle struct {
		Name        string        `toml:"name"`
		Description string        `toml:"description"`
		Skills      []bundleSkill `toml:"skills"`
	}

	bundleSkill struct {
		Name   string `toml:"name"`
		Source string `toml:"source"`
	}

	catalogFile struct {
		Bundle []bundle `toml:"bundle"`
	}
)

func descBudgetStatus(tokens, budget int64) string {
	pct := tokens * percentageBase / budget
	switch {
	case pct >= budgetFullPercent:
		return "OVER"
	case pct >= budgetWarnPercent:
		return "warn"
	default:
		return "ok"
	}
}

func main() {
	format := flag.String("format", "markdown", "output format: markdown or json")
	sortBy := flag.String("sort", "body", "sort column: body, desc, name, category")
	flag.Parse()

	root, err := findRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: could not find repo root:", err)
		os.Exit(1)
	}

	skills, err := scan(filepath.Join(root, skillsDir))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error scanning skills:", err)
		os.Exit(1)
	}

	bundles, err := loadCatalog(filepath.Join(root, catalogFilePath))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error loading catalog:", err)
		os.Exit(1)
	}

	switch *sortBy {
	case "body":
		sort.Slice(skills, func(i, j int) bool { return skills[i].bodyTokens > skills[j].bodyTokens })
	case "desc":
		sort.Slice(skills, func(i, j int) bool { return skills[i].descTokens > skills[j].descTokens })
	case "name":
		sort.Slice(skills, func(i, j int) bool { return skills[i].name < skills[j].name })
	case "category":
		sort.Slice(skills, func(i, j int) bool {
			if skills[i].category != skills[j].category {
				return skills[i].category < skills[j].category
			}
			return skills[i].name < skills[j].name
		})
	default:
		fmt.Fprintf(os.Stderr, "error: unknown sort %q — use body, desc, name, or category\n", *sortBy)
		os.Exit(1)
	}

	switch *format {
	case "json":
		out, jsonErr := renderJSON(skills)
		if jsonErr != nil {
			fmt.Fprintln(os.Stderr, "error encoding json:", jsonErr)
			os.Exit(1)
		}
		outPath := filepath.Join(root, docsDir, tokenCountsJSONFile)
		if writeErr := os.WriteFile(outPath, out, filePermission); writeErr != nil {
			fmt.Fprintln(os.Stderr, "error writing", outPath, ":", writeErr)
			os.Exit(1)
		}
		fmt.Println("wrote", outPath)
	case "markdown":
		out := render(skills, bundles)
		outPath := filepath.Join(root, docsDir, tokenCountsFile)
		if writeErr := os.WriteFile(outPath, []byte(out), filePermission); writeErr != nil {
			fmt.Fprintln(os.Stderr, "error writing", outPath, ":", writeErr)
			os.Exit(1)
		}
		fmt.Println("wrote", outPath)
	default:
		fmt.Fprintf(os.Stderr, "error: unknown format %q — use markdown or json\n", *format)
		os.Exit(1)
	}
}

// loadCatalog reads the embedded bundle catalog used by the rsk CLI.
func loadCatalog(path string) ([]bundle, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var cat catalogFile
	if _, err = toml.Decode(string(raw), &cat); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return cat.Bundle, nil
}

// findRoot walks up from the working directory until it finds a skills/ folder.
func findRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, gitDir)); statErr == nil {
			if _, skillsErr := os.Stat(filepath.Join(dir, skillsDir)); skillsErr != nil {
				return "", fmt.Errorf("found repo root at %s but no skills/ directory", dir)
			}
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("no .git directory found — run this from inside the repo")
		}
		dir = parent
	}
}

// scan walks skillsDir and collects a skill entry for each SKILL.md found,
// also picking up sibling STACK.md and RECIPES.md when present.
func scan(skillsDir string) ([]skill, error) {
	var skills []skill
	err := filepath.WalkDir(skillsDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			fmt.Fprintf(os.Stderr, "warn: skipping %s: %v\n", path, walkErr)
			return nil
		}
		if d.IsDir() || d.Name() != skillMarkdownFile {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "warn: skipping %s: %v\n", path, readErr)
			return nil
		}

		skillDir := filepath.Dir(path)
		name := filepath.Base(skillDir)
		category := filepath.Base(filepath.Dir(skillDir))

		desc := extractDescription(data)
		body := stripFrontmatter(data)

		s := skill{
			name:       name,
			category:   category,
			bodyBytes:  int64(len(body)),
			bodyTokens: estimateTokens(body),
			descTokens: roundDiv(int64(len(desc)), descBytesPerToken),
		}

		siblings, globErr := filepath.Glob(filepath.Join(skillDir, "*.md"))
		if globErr != nil {
			siblings = nil
		}
		for _, sib := range siblings {
			if filepath.Base(sib) == skillMarkdownFile {
				continue
			}
			si, siErr := os.Stat(sib)
			if siErr != nil {
				fmt.Fprintf(os.Stderr, "warn: skipping %s: %v\n", sib, siErr)
				continue
			}
			tokens := roundDiv(si.Size(), bytesPerToken)
			base := filepath.Base(sib)
			switch base {
			case stackMarkdownFile:
				s.stackTokens = tokens
			case recipesMarkdownFile:
				s.recipesTokens = tokens
			default:
				s.otherTokens += tokens
				s.otherFiles = append(s.otherFiles, otherFile{name: base, tokens: tokens})
			}
		}

		skills = append(skills, s)
		return nil
	})
	return skills, err
}

// extractDescription parses the YAML frontmatter and returns the raw value of
// the description field. Handles both single-line and block-scalar (> or |)
// forms, collapsing the latter to a single line for length estimation.
func extractDescription(data []byte) string {
	content := string(data)
	if !strings.HasPrefix(content, "---") {
		return ""
	}

	rest := content[3:]
	frontmatter, _, ok := strings.Cut(rest, "\n---")
	if !ok {
		return ""
	}

	lines := strings.Split(frontmatter, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "description:") {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(trimmed, "description:"))

		if strings.HasPrefix(value, ">") || strings.HasPrefix(value, "|") {
			var parts []string
			for _, cont := range lines[i+1:] {
				if cont == "" || cont[0] == ' ' || cont[0] == '\t' {
					parts = append(parts, strings.TrimSpace(cont))
				} else {
					break
				}
			}
			return strings.Join(parts, " ")
		}
		return value
	}
	return ""
}

// roundDiv divides n by d rounding to nearest integer instead of truncating.
func roundDiv(n, d int64) int64 {
	return (n + d/2) / d
}

// estimateTokens splits content into fenced-code regions and prose regions,
// applying different bytes/token ratios to each for a more accurate estimate.
func estimateTokens(content []byte) int64 {
	var proseBytes, codeBytes int64
	inCode := false
	for line := range strings.SplitSeq(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCode = !inCode
			// count the fence line itself as prose
			proseBytes += int64(len(line)) + 1
			continue
		}
		if inCode {
			codeBytes += int64(len(line)) + 1
		} else {
			proseBytes += int64(len(line)) + 1
		}
	}
	return roundDiv(proseBytes, proseBytesPerToken) + roundDiv(codeBytes, codeBytesPerToken)
}

// stripFrontmatter returns the file content after the closing --- of the YAML
// frontmatter block. If no frontmatter is present, the full content is returned.
func stripFrontmatter(data []byte) []byte {
	content := string(data)
	if !strings.HasPrefix(content, "---") {
		return data
	}
	rest := content[3:]
	_, after, ok := strings.Cut(rest, "\n---")
	if !ok {
		return data
	}
	// skip the newline immediately after the closing ---
	after = strings.TrimPrefix(after, "\n")
	return []byte(after)
}
