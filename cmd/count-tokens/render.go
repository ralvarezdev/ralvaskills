package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

func render(skills []skill, bundles []bundle) string {
	var buf bytes.Buffer

	type categoryTotals struct {
		count      int
		bodyTokens int64
		descTokens int64
		sideTokens int64
	}
	catMap := map[string]*categoryTotals{}

	skillByName := make(map[string]*skill, len(skills))
	for i := range skills {
		skillByName[skills[i].name] = &skills[i]
	}

	var totalBodyBytes, totalBodyTokens, totalDescTokens int64
	var totalStackTokens, totalRecipesTokens, totalOtherTokens int64
	for _, s := range skills {
		totalBodyBytes += s.bodyBytes
		totalBodyTokens += s.bodyTokens
		totalDescTokens += s.descTokens
		totalStackTokens += s.stackTokens
		totalRecipesTokens += s.recipesTokens
		totalOtherTokens += s.otherTokens

		ct := catMap[s.category]
		if ct == nil {
			ct = &categoryTotals{}
			catMap[s.category] = ct
		}
		ct.count++
		ct.bodyTokens += s.bodyTokens
		ct.descTokens += s.descTokens
		ct.sideTokens += s.stackTokens + s.recipesTokens + s.otherTokens
	}
	totalSideTokens := totalStackTokens + totalRecipesTokens + totalOtherTokens

	catNames := make([]string, 0, len(catMap))
	for c := range catMap {
		catNames = append(catNames, c)
	}
	sort.Slice(catNames, func(i, j int) bool {
		return catMap[catNames[i]].bodyTokens > catMap[catNames[j]].bodyTokens
	})

	// Per-bundle totals. Skills referenced by a bundle but not present in the
	// local scan are categorized: official-source skills (docx, xlsx, etc.)
	// live outside this repo and are intentionally external; local-source
	// skills not found in the scan are genuinely missing.
	type bundleTotals struct {
		known      int
		external   int // source = "official", not in this repo by design
		missing    int // source = "local" but not found in skills/ tree
		bodyTokens int64
		descTokens int64
	}
	bundledNames := make(map[string]struct{})
	bundleTotalsMap := make(map[string]*bundleTotals, len(bundles))
	for _, b := range bundles {
		t := &bundleTotals{}
		for _, ref := range b.Skills {
			bundledNames[ref.Name] = struct{}{}
			s, ok := skillByName[ref.Name]
			if !ok {
				if ref.Source == "official" {
					t.external++
				} else {
					t.missing++
				}
				continue
			}
			t.known++
			t.bodyTokens += s.bodyTokens
			t.descTokens += s.descTokens
		}
		bundleTotalsMap[b.Name] = t
	}
	globalTotals := bundleTotalsMap[globalBundleName]

	var unbundledSkills []skill
	for _, s := range skills {
		if _, inBundle := bundledNames[s.name]; !inBundle {
			unbundledSkills = append(unbundledSkills, s)
		}
	}
	sort.Slice(unbundledSkills, func(i, j int) bool {
		return unbundledSkills[i].descTokens > unbundledSkills[j].descTokens
	})

	buf.WriteString("# SKILL.md Token Estimates\n\n")
	buf.WriteString("> Auto-generated. **Do not edit by hand.** Run `task tokens` to refresh.\n")
	buf.WriteString(">\n")
	buf.WriteString(
		"> Estimate: ~4 bytes/token for bodies, ~3 bytes/token for descriptions" +
			" (Claude tokenizer). Actual range ±15%.\n\n",
	)
	ts := time.Now().UTC().Format("2006-01-02")
	fmt.Fprintf(&buf, "_Last updated: %s · %d skills · %d bundles_\n\n", ts, len(skills), len(bundles))

	buf.WriteString("## Load model\n\n")
	buf.WriteString(
		"Description tokens hit **every turn** for installed skills." +
			" Body tokens are paid only when a skill is actually invoked." +
			" Side files are never auto-loaded.\n\n",
	)
	buf.WriteString("| What | When loaded | Estimated tokens |\n")
	buf.WriteString("|---|---|---:|\n")
	fmt.Fprintf(&buf, "| All `SKILL.md` bodies | Only when invoked | ~%d |\n", totalBodyTokens)
	fmt.Fprintf(
		&buf, "| All side files (`STACK` + `RECIPES` + topic files) | On-demand only | ~%d |\n",
		totalSideTokens,
	)
	fmt.Fprintf(
		&buf, "| All `description:` fields (every skill) | Every turn, if all installed | ~%d |\n\n",
		totalDescTokens,
	)

	// Session profiles: global baseline, plus what each project bundle adds.
	if globalTotals != nil {
		buf.WriteString("## Session profiles\n\n")
		buf.WriteString("Budgets: **global ≤ 1500** desc tokens (paid by every project), " +
			"**session ≤ 2200** desc tokens (global + one project bundle). " +
			"Status: `ok` < 90% · `warn` ≥ 90% · `OVER` ≥ 100%.\n\n")

		globalStatus := descBudgetStatus(globalTotals.descTokens, globalDescBudget)
		fmt.Fprintf(&buf,
			"**Global baseline** (`rsk install global --global`): %d skills · "+
				"~%d / %d desc tokens every turn (`%s`) · "+
				"~%d body tokens if all invoked.\n\n",
			globalTotals.known,
			globalTotals.descTokens, globalDescBudget, globalStatus,
			globalTotals.bodyTokens)

		buf.WriteString("Per-project bundle additions on top of the global baseline:\n\n")
		buf.WriteString("| Bundle | Skills | + Desc tkns / turn | Session total | Budget | Status |\n")
		buf.WriteString("|---|---:|---:|---:|---:|---:|\n")

		type bundleRow struct {
			name string
			t    *bundleTotals
		}
		var rows []bundleRow
		for _, b := range bundles {
			if b.Name == globalBundleName {
				continue
			}
			rows = append(rows, bundleRow{name: b.Name, t: bundleTotalsMap[b.Name]})
		}
		sort.Slice(rows, func(i, j int) bool { return rows[i].t.descTokens < rows[j].t.descTokens })
		for _, r := range rows {
			sessionTotal := globalTotals.descTokens + r.t.descTokens
			status := descBudgetStatus(sessionTotal, sessionDescBudget)
			var notes []string
			if r.t.external > 0 {
				notes = append(notes, fmt.Sprintf("+%d external", r.t.external))
			}
			if r.t.missing > 0 {
				notes = append(notes, fmt.Sprintf("+%d missing", r.t.missing))
			}
			note := ""
			if len(notes) > 0 {
				note = " *(" + strings.Join(notes, ", ") + ")*"
			}
			fmt.Fprintf(&buf, "| `%s`%s | %d | +%d | ~%d | %d | `%s` |\n",
				r.name, note, r.t.known, r.t.descTokens,
				sessionTotal, sessionDescBudget, status)
		}
		buf.WriteString("\n")
	}

	if len(unbundledSkills) > 0 {
		buf.WriteString("## Personal / unbundled skills\n\n")
		buf.WriteString("These skills are not part of any bundle. " +
			"They are installed individually with `rsk install <name> --personal` " +
			"and their desc tokens are only paid when explicitly installed.\n\n")
		buf.WriteString("| Skill | Category | ~Body tkns | ~Desc tkns | ~Side tkns |\n")
		buf.WriteString("|---|---|---:|---:|---:|\n")
		for _, s := range unbundledSkills {
			sideTokens := s.stackTokens + s.recipesTokens + s.otherTokens
			fmt.Fprintf(&buf, "| `%s` | %s | ~%d | ~%d | ~%d |\n",
				s.name, s.category, s.bodyTokens, s.descTokens, sideTokens)
		}
		buf.WriteString("\n")
	}

	buf.WriteString("## By category\n\n")
	buf.WriteString("| Category | Skills | ~Body tkns | ~Desc tkns | ~Side tkns |\n")
	buf.WriteString("|---|---:|---:|---:|---:|\n")
	for _, c := range catNames {
		ct := catMap[c]
		fmt.Fprintf(&buf, "| %s | %d | ~%d | ~%d | ~%d |\n",
			c, ct.count, ct.bodyTokens, ct.descTokens, ct.sideTokens)
	}
	buf.WriteString("\n")

	if len(bundles) > 0 {
		buf.WriteString("## By bundle\n\n")
		buf.WriteString("Catalog from `internal/config/catalog.toml`. " +
			"`External` are `source = \"official\"` skills (e.g. `docx`, `xlsx`) that live outside this repo by design. " +
			"`Missing` are `source = \"local\"` skills the bundle references that aren't in the local `skills/` tree.\n\n")
		buf.WriteString("| Bundle | Local | External | Missing | ~Body tkns (local) | ~Desc tkns (local) |\n")
		buf.WriteString("|---|---:|---:|---:|---:|---:|\n")
		for _, b := range bundles {
			t := bundleTotalsMap[b.Name]
			fmt.Fprintf(&buf, "| `%s` | %d | %d | %d | ~%d | ~%d |\n",
				b.Name, t.known, t.external, t.missing, t.bodyTokens, t.descTokens)
		}
		buf.WriteString("\n")
	}

	buf.WriteString("## Per skill\n\n")
	buf.WriteString(
		"| # | Skill | Category | Body bytes | ~Body tkns | ~Desc tkns" +
			" | ~Stack tkns | ~Recipes tkns | ~Topic tkns |\n",
	)
	buf.WriteString("|---|---|---|---:|---:|---:|---:|---:|---:|\n")
	for i, s := range skills {
		fmt.Fprintf(
			&buf, "| %d | `%s` | %s | %d | ~%d | ~%d | ~%d | ~%d | ~%d |\n",
			i+1, s.name, s.category, s.bodyBytes, s.bodyTokens, s.descTokens,
			s.stackTokens, s.recipesTokens, s.otherTokens,
		)
	}

	fmt.Fprintf(&buf, "\n**Totals:** %d body bytes · ~%d body tokens · ~%d desc tokens · ~%d side tokens\n\n",
		totalBodyBytes, totalBodyTokens, totalDescTokens, totalSideTokens)

	var withOther []skill
	for _, s := range skills {
		if len(s.otherFiles) > 0 {
			withOther = append(withOther, s)
		}
	}
	if len(withOther) > 0 {
		sort.Slice(withOther, func(i, j int) bool {
			return withOther[i].otherTokens > withOther[j].otherTokens
		})
		buf.WriteString("## Topic files\n\n")
		buf.WriteString("Side files in a skill directory other than `STACK.md` and `RECIPES.md` — " +
			"typically topic-specific reference docs referenced by name from `SKILL.md` " +
			"(e.g. `PAYLOADS.md`, `AUTH_PATTERNS.md`). These contribute to the `~Topic tkns` column above " +
			"and are loaded only when the parent skill chooses to read them.\n\n")
		for _, s := range withOther {
			files := append([]otherFile(nil), s.otherFiles...)
			sort.Slice(files, func(i, j int) bool { return files[i].tokens > files[j].tokens })
			fmt.Fprintf(&buf, "- `%s` (~%d tokens total):\n", s.name, s.otherTokens)
			for _, f := range files {
				fmt.Fprintf(&buf, "  - `%s` — ~%d tokens\n", f.name, f.tokens)
			}
		}
		buf.WriteString("\n")
	}

	var heavyBody []skill
	for _, s := range skills {
		if s.bodyTokens > bodyThreshold {
			heavyBody = append(heavyBody, s)
		}
	}
	byDesc := append([]skill(nil), skills...)
	sort.Slice(byDesc, func(i, j int) bool { return byDesc[i].descTokens > byDesc[j].descTokens })
	if len(byDesc) > topDescCount {
		byDesc = byDesc[:topDescCount]
	}

	buf.WriteString("## Skills to consider trimming\n\n")
	if len(heavyBody) > 0 {
		fmt.Fprintf(
			&buf, "Body > %d tokens — consider moving examples to `RECIPES.md` or topic files:\n\n",
			bodyThreshold,
		)
		sort.Slice(heavyBody, func(i, j int) bool { return heavyBody[i].bodyTokens > heavyBody[j].bodyTokens })
		for _, s := range heavyBody {
			fmt.Fprintf(&buf, "- `%s` (~%d body tokens)\n", s.name, s.bodyTokens)
		}
		buf.WriteString("\n")
	}
	if len(byDesc) > 0 {
		fmt.Fprintf(
			&buf,
			"Heaviest %d descriptions — each desc token is paid every turn for any session that installs the skill:\n\n",
			len(byDesc),
		)
		for _, s := range byDesc {
			fmt.Fprintf(&buf, "- `%s` (~%d desc tokens)\n", s.name, s.descTokens)
		}
		buf.WriteString("\n")
	}

	buf.WriteString("## Notes\n\n")
	buf.WriteString("- **Body tokens** cost only when the skill is invoked in a session.\n")
	buf.WriteString(
		"- **Description tokens** cost every turn for any session that has the skill installed." +
			" The cost depends on which bundles are installed, not on the corpus total.\n",
	)
	buf.WriteString(
		"- Side files (`STACK.md`, `RECIPES.md`, topic files) are never auto-loaded; they cost 0 per turn.\n",
	)
	buf.WriteString(
		"- For exact counts: run each file through" +
			" [`tiktoken`](https://github.com/openai/tiktoken) with `cl100k_base`.\n",
	)

	return buf.String()
}

func renderJSON(skills []skill) ([]byte, error) {
	type otherFileJSON struct {
		Name   string `json:"name"`
		Tokens int64  `json:"tokens"`
	}
	type skillJSON struct {
		Name          string          `json:"name"`
		Category      string          `json:"category"`
		BodyBytes     int64           `json:"body_bytes"`
		BodyTokens    int64           `json:"body_tokens"`
		DescTokens    int64           `json:"desc_tokens"`
		StackTokens   int64           `json:"stack_tokens"`
		RecipesTokens int64           `json:"recipes_tokens"`
		OtherTokens   int64           `json:"other_tokens"`
		OtherFiles    []otherFileJSON `json:"other_files,omitempty"`
	}
	type totalsJSON struct {
		BodyBytes  int64 `json:"body_bytes"`
		BodyTokens int64 `json:"body_tokens"`
		DescTokens int64 `json:"desc_tokens"`
		SideTokens int64 `json:"side_tokens"`
		SkillCount int   `json:"skill_count"`
	}
	type output struct {
		GeneratedAt string      `json:"generated_at"`
		Totals      totalsJSON  `json:"totals"`
		Skills      []skillJSON `json:"skills"`
	}

	var totalBodyBytes, totalBodyTokens, totalDescTokens, totalSideTokens int64
	rows := make([]skillJSON, len(skills))
	for i, s := range skills {
		var others []otherFileJSON
		for _, f := range s.otherFiles {
			others = append(others, otherFileJSON{Name: f.name, Tokens: f.tokens})
		}
		rows[i] = skillJSON{
			Name:          s.name,
			Category:      s.category,
			BodyBytes:     s.bodyBytes,
			BodyTokens:    s.bodyTokens,
			DescTokens:    s.descTokens,
			StackTokens:   s.stackTokens,
			RecipesTokens: s.recipesTokens,
			OtherTokens:   s.otherTokens,
			OtherFiles:    others,
		}
		totalBodyBytes += s.bodyBytes
		totalBodyTokens += s.bodyTokens
		totalDescTokens += s.descTokens
		totalSideTokens += s.stackTokens + s.recipesTokens + s.otherTokens
	}

	out := output{
		GeneratedAt: time.Now().UTC().Format("2006-01-02"),
		Totals: totalsJSON{
			BodyBytes:  totalBodyBytes,
			BodyTokens: totalBodyTokens,
			DescTokens: totalDescTokens,
			SideTokens: totalSideTokens,
			SkillCount: len(skills),
		},
		Skills: rows,
	}
	return json.MarshalIndent(out, "", "  ")
}
