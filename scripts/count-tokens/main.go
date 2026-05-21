// count-tokens scans every SKILL.md under skills/, estimates token counts
// (body + description-only), and writes docs/TOKEN_COUNTS.md.
//
// Usage:
//
//	go run ./scripts/count-tokens
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

// tokensPerByte is a rough estimate for Claude's tokenizer on English markdown.
// Actual range is 3.5–4.5; use 4 as a safe midpoint.
const tokensPerByte = 4

type skill struct {
	name        string
	category    string
	bodyBytes   int64
	descBytes   int64
	bodyTokens  int64
	descTokens  int64
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

// scan walks skillsDir and collects a skill entry for each SKILL.md found.
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

		// derive skill name and category from directory structure:
		// skills/<category>/<skill-name>/SKILL.md
		skillDir := filepath.Dir(path)
		name := filepath.Base(skillDir)
		category := filepath.Base(filepath.Dir(skillDir))

		desc := extractDescription(data)
		descBytes := int64(len(desc))
		bodyBytes := info.Size()

		skills = append(skills, skill{
			name:       name,
			category:   category,
			bodyBytes:  bodyBytes,
			descBytes:  descBytes,
			bodyTokens: bodyBytes / tokensPerByte,
			descTokens: descBytes / tokensPerByte,
		})
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

	// find closing ---
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

		// block scalar (> or |): collect indented continuation lines
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

	var totalBodyBytes, totalBodyTokens, totalDescBytes, totalDescTokens int64
	for _, s := range skills {
		totalBodyBytes += s.bodyBytes
		totalBodyTokens += s.bodyTokens
		totalDescBytes += s.descBytes
		totalDescTokens += s.descTokens
	}

	buf.WriteString("# SKILL.md Token Estimates\n\n")
	buf.WriteString("> Auto-generated. **Do not edit by hand.** Run `go run ./scripts/count-tokens` to refresh.\n")
	buf.WriteString(">\n")
	buf.WriteString("> Estimate: ~4 chars/token (Claude tokenizer, English markdown). Actual range ±15%.\n\n")
	fmt.Fprintf(&buf, "_Last updated: %s · %d skills_\n\n", time.Now().UTC().Format("2006-01-02 15:04 UTC"), len(skills))

	buf.WriteString("## Load model\n\n")
	buf.WriteString("| What | When loaded | Estimated tokens |\n")
	buf.WriteString("|---|---|---:|\n")
	fmt.Fprintf(&buf, "| All `description:` fields (skill index) | **Every turn** | ~%d |\n", totalDescTokens)
	fmt.Fprintf(&buf, "| All `SKILL.md` bodies | Only when skill is invoked | ~%d |\n", totalBodyTokens)
	fmt.Fprintf(&buf, "| Everything combined | If all skills active at once | ~%d |\n\n", totalBodyTokens)

	buf.WriteString("## Per skill\n\n")
	buf.WriteString("| # | Skill | Category | Body bytes | ~Body tokens | ~Desc tokens |\n")
	buf.WriteString("|---|---|---|---:|---:|---:|\n")
	for i, s := range skills {
		fmt.Fprintf(&buf, "| %d | `%s` | %s | %d | ~%d | ~%d |\n",
			i+1, s.name, s.category, s.bodyBytes, s.bodyTokens, s.descTokens)
	}

	fmt.Fprintf(&buf, "\n**Totals:** %d bytes · ~%d body tokens · ~%d description tokens (always-loaded)\n\n",
		totalBodyBytes, totalBodyTokens, totalDescTokens)

	buf.WriteString("## Notes\n\n")
	buf.WriteString("- **Body tokens** cost only when the skill is invoked in a session.\n")
	buf.WriteString("- **Description tokens** cost on every single turn — keep descriptions short.\n")
	buf.WriteString("- `RECIPES.md` and other side files are never auto-loaded; they cost 0 per turn.\n")
	buf.WriteString("- For exact counts: run each file through [`tiktoken`](https://github.com/openai/tiktoken) with `cl100k_base`.\n")

	return buf.String()
}
