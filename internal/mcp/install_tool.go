package mcp

import (
	"context"
	"errors"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ralvarezdev/mcpkit"

	"github.com/ralvarezdev/ralvaskills/v3/internal/config"
	"github.com/ralvarezdev/ralvaskills/v3/internal/install"
	"github.com/ralvarezdev/ralvaskills/v3/internal/skill"
	"github.com/ralvarezdev/ralvaskills/v3/internal/source"
	"github.com/ralvarezdev/ralvaskills/v3/internal/tool"
)

type (
	// InstallSkillsIn is the install_skills tool input.
	InstallSkillsIn struct {
		Names []string `json:"names"`
		Scope string   `json:"scope,omitempty" jsonschema:"project (por defecto) | global"`
		Pin   bool     `json:"pin,omitempty"`
		For   string   `json:"for,omitempty"   jsonschema:"claude-code|opencode"`
	}

	// InstallSkillsOut is the install_skills tool output.
	InstallSkillsOut struct {
		Installed []InstallResultOut `json:"installed"`
		Failed    []InstallResultOut `json:"failed"`
	}

	// InstallResultOut is the outcome for one skill.
	InstallResultOut struct {
		FilesChanged []string `json:"files_changed,omitempty"`
		Name         string   `json:"name"`
		Version      string   `json:"version"`
		Scope        string   `json:"scope"`
		Error        string   `json:"error,omitempty"`
	}
)

func installHandler(deps Deps) sdk.ToolHandlerFor[InstallSkillsIn, InstallSkillsOut] {
	return func(
		ctx context.Context, _ *sdk.CallToolRequest, in InstallSkillsIn,
	) (*sdk.CallToolResult, InstallSkillsOut, error) {
		scope, forTool, err := validateInstall(in)
		if err != nil {
			return nil, InstallSkillsOut{}, err
		}

		skills, err := resolveForInstall(ctx, deps, in.Names)
		if err != nil {
			return nil, InstallSkillsOut{}, err
		}
		if len(skills) == 0 {
			return nil, InstallSkillsOut{}, nil
		}

		targets, err := install.Targets(deps.Cfg, scope == install.ScopeGlobal, forTool)
		if err != nil {
			return nil, InstallSkillsOut{}, err
		}
		return applyInstall(deps, skills, scope, targets, in.Pin)
	}
}

func validateInstall(in InstallSkillsIn) (install.Scope, tool.ID, error) {
	if len(in.Names) == 0 {
		return "", "", errors.New("names must not be empty")
	}

	scope, err := parseScope(in.Scope)
	if err != nil {
		return "", "", err
	}
	forTool, err := parseForTool(in.For)
	if err != nil {
		return "", "", err
	}
	if scope == install.ScopeGlobal && in.Pin {
		return "", "", errors.New("pin only applies to project scope")
	}
	return scope, forTool, nil
}

// resolveForInstall expands names into skills. Resolution errors are agent-safe
// and actionable (an unknown name, a planned skill), so they are returned as-is
// rather than hidden behind a generic message.
func resolveForInstall(ctx context.Context, deps Deps, names []string) ([]skill.Skill, error) {
	bundles, catalogErr := config.LoadCatalog("")
	if catalogErr != nil && deps.Logger != nil {
		deps.Logger.Warn("user catalog", "error", catalogErr)
	}

	skills, _, err := install.Resolve(
		ctx, names, bundles,
		install.LocalSource(deps.Cfg),
		source.NewOfficial(deps.Cfg.OfficialCache),
	)
	if err != nil {
		return nil, err
	}
	return skills, rejectPersonal(skills)
}

func applyInstall(
	deps Deps, skills []skill.Skill, scope install.Scope, targets []string, pin bool,
) (*sdk.CallToolResult, InstallSkillsOut, error) {
	results, applyErr := install.Apply(skills, install.Options{
		Scope:   scope,
		Targets: targets,
		Pin:     pin,
	})
	out := toInstallOut(results)
	if applyErr != nil {
		errMsg := mcpkit.ErrorMsg{Public: "could not write the project manifest"}
		return nil, out, errMsg.ToolError(deps.Logger, applyErr)
	}
	return nil, out, nil
}

func parseScope(raw string) (install.Scope, error) {
	switch raw {
	case "", string(install.ScopeProject):
		return install.ScopeProject, nil
	case string(install.ScopeGlobal):
		return install.ScopeGlobal, nil
	default:
		return "", fmt.Errorf(
			"scope must be %q or %q (got %q)",
			install.ScopeProject, install.ScopeGlobal, raw,
		)
	}
}

func parseForTool(raw string) (tool.ID, error) {
	if raw == "" {
		return "", nil
	}
	return tool.ParseID(raw)
}

// rejectPersonal refuses to install personal skills: they are private to their
// author and the MCP input has no opt-in for them.
func rejectPersonal(skills []skill.Skill) error {
	for _, s := range skills {
		if s.IsPersonal {
			return fmt.Errorf("skill %q is in personal/ and cannot be installed through the MCP server", s.Name)
		}
	}
	return nil
}

func toInstallOut(results []install.Result) InstallSkillsOut {
	out := InstallSkillsOut{
		Installed: make([]InstallResultOut, 0, len(results)),
		Failed:    make([]InstallResultOut, 0),
	}
	for _, r := range results {
		item := InstallResultOut{
			Name:         r.Name,
			Version:      r.Version,
			Scope:        string(r.Scope),
			FilesChanged: r.Files,
		}
		if r.Err != nil {
			item.Error = r.Err.Error()
			out.Failed = append(out.Failed, item)
			continue
		}
		out.Installed = append(out.Installed, item)
	}
	return out
}
