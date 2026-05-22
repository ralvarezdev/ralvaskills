// count-tokens scans every SKILL.md under skills/, estimates token counts
// (body + description-only + side files), and writes docs/TOKEN_COUNTS.md.
//
// Usage:
//
//	task tokens
package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// bytesPerToken is a rough estimate for Claude's tokenizer on English markdown.
// Actual range is 3.5–4.5 bytes/token; 4 is a safe midpoint.
const bytesPerToken = 4

type skill struct {
	name          string
	category      string
	bodyBytes     int64
	bodyTokens    int64
	descTokens    int64
	stackTokens   int64
	recipesTokens int64
}

func main() {
	root, err := findRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: could not find repo root (no skills/ directory found)")
		os.Exit(1)
	}

	skills, err := scan(filepath.Join(root, "skills"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error scanning skills:", err)
		os.Exit(1)
	}

	sort.Slice(skills, func(i, j int) bool {
		return skills[i].bodyTokens > skills[j].bodyTokens
	})

	out := render(skills)

	outPath := filepath.Join(root, "docs", "TOKEN_COUNTS.md")
	if err := os.WriteFile(outPath, []byte(out), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "error writing", outPath, ":", err)
		os.Exit(1)
	}
	fmt.Println("wrote", outPath)
}

// findRoot walks up from the working directory until it finds a skills/ folder.
func findRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "skills")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not found")
		}
		dir = parent
	}
}

// scan walks skillsDir and collects a skill entry for each SKILL.md found,
// also picking up sibling STACK.md and RECIPES.md when present.
func scan(skillsDir string) ([]skill, error) {
	var skills []skill
	err := filepath.WalkDir(skillsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() != "SKILL.md" {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}

		info, statErr := d.Info()
		if statErr != nil {
			return nil
		}

		skillDir := filepath.Dir(path)
		name := filepath.Base(skillDir)
		category := filepath.Base(filepath.Dir(skillDir))

		desc := extractDescription(data)

		s := skill{
			name:       name,
			category:   category,
			bodyBytes:  info.Size(),
			bodyTokens: info.Size() / bytesPerToken,
			descTokens: int64(len(desc)) / bytesPerToken,
		}

		if si, err := os.Stat(filepath.Join(skillDir, "STACK.md")); err == nil {
			s.stackTokens = si.Size() / bytesPerToken
		}
		if ri, err := os.Stat(filepath.Join(skillDir, "RECIPES.md")); err == nil {
			s.recipesTokens = ri.Size() / bytesPerToken
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

		if value == ">" || value == "|" {
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

func render(skills []skill) string {
	var buf bytes.Buffer

	var totalBodyBytes, totalBodyTokens, totalDescTokens int64
	var totalStackTokens, totalRecipesTokens int64
	for _, s := range skills {
		totalBodyBytes += s.bodyBytes
		totalBodyTokens += s.bodyTokens
		totalDescTokens += s.descTokens
		totalStackTokens += s.stackTokens
		totalRecipesTokens += s.recipesTokens
	}
	totalSideTokens := totalStackTokens + totalRecipesTokens

	buf.WriteString("# SKILL.md Token Estimates\n\n")
	buf.WriteString("> Auto-generated. **Do not edit by hand.** Run `task tokens` to refresh.\n")
	buf.WriteString(">\n")
	buf.WriteString("> Estimate: ~4 bytes/token (Claude tokenizer, English markdown). Actual range ±15%.\n\n")
	fmt.Fprintf(&buf, "_Last updated: %s · %d skills_\n\n", time.Now().UTC().Format("2006-01-02 15:04 UTC"), len(skills))

	buf.WriteString("## Load model\n\n")
	buf.WriteString("| What | When loaded | Estimated tokens |\n")
	buf.WriteString("|---|---|---:|\n")
	fmt.Fprintf(&buf, "| All `description:` fields (skill index) | **Every turn** | ~%d |\n", totalDescTokens)
	fmt.Fprintf(&buf, "| All `SKILL.md` bodies | Only when skill is invoked | ~%d |\n", totalBodyTokens)
	fmt.Fprintf(&buf, "| All side files (`STACK` + `RECIPES`) | On-demand only | ~%d |\n", totalSideTokens)
	fmt.Fprintf(&buf, "| Everything combined | Absolute maximum | ~%d |\n\n", totalBodyTokens+totalSideTokens)

	buf.WriteString("## Per skill\n\n")
	buf.WriteString("| # | Skill | Category | Body bytes | ~Body tkns | ~Desc tkns | ~Stack tkns | ~Recipes tkns |\n")
	buf.WriteString("|---|---|---|---:|---:|---:|---:|---:|\n")
	for i, s := range skills {
		fmt.Fprintf(&buf, "| %d | `%s` | %s | %d | ~%d | ~%d | ~%d | ~%d |\n",
			i+1, s.name, s.category, s.bodyBytes, s.bodyTokens, s.descTokens, s.stackTokens, s.recipesTokens)
	}

	fmt.Fprintf(&buf, "\n**Totals:** %d body bytes · ~%d body tokens · ~%d desc tokens · ~%d side tokens\n\n",
		totalBodyBytes, totalBodyTokens, totalDescTokens, totalSideTokens)

	buf.WriteString("## Notes\n\n")
	buf.WriteString("- **Body tokens** cost only when the skill is invoked in a session.\n")
	buf.WriteString("- **Description tokens** cost on every single turn — keep descriptions short.\n")
	buf.WriteString("- `STACK.md` and `RECIPES.md` are never auto-loaded; they cost 0 per turn.\n")
	buf.WriteString("- For exact counts: run each file through [`tiktoken`](https://github.com/openai/tiktoken) with `cl100k_base`.\n")

	return buf.String()
}
