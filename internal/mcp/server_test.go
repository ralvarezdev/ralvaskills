package mcp_test

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ralvarezdev/ralvaskills/v3/internal/config"
	"github.com/ralvarezdev/ralvaskills/v3/internal/fsperm"
	"github.com/ralvarezdev/ralvaskills/v3/internal/mcp"
)

func TestServerToolsAndResource(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	writeSkill(t, filepath.Join(repo, "skills", "tdd"), "tdd", "1.0.0", "Red-green-refactor workflow.")
	writeSkill(
		t,
		filepath.Join(repo, "skills", "languages", "go-architect"),
		"go-architect",
		"2.0.0",
		"Go 1.26 architectural standards.",
	)

	session := newSession(t, mcp.Deps{
		Cfg:     config.Config{RepoPath: repo},
		Logger:  slog.New(slog.DiscardHandler),
		Version: "test",
	})

	// tools/list: the read tools are registered, read-only, and schemaful.
	tools, err := session.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	for _, name := range []string{"project_profile", "search_skills"} {
		tool := findTool(t, tools.Tools, name)
		if tool.InputSchema == nil {
			t.Errorf("%s: inputSchema is empty", name)
		}
		if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint {
			t.Errorf("%s: want readOnlyHint true", name)
		}
	}

	installTool := findTool(t, tools.Tools, "install_skills")
	if installTool.InputSchema == nil {
		t.Error("install_skills: inputSchema is empty")
	}
	if installTool.Annotations == nil ||
		installTool.Annotations.DestructiveHint == nil ||
		!*installTool.Annotations.DestructiveHint {
		t.Error("install_skills: want destructiveHint true")
	}
	if v, ok := installTool.Meta["anthropic/requiresUserInteraction"].(bool); !ok || !v {
		t.Error("install_skills: want anthropic/requiresUserInteraction true")
	}

	// project_profile on a Go project.
	project := t.TempDir()
	if writeErr := os.WriteFile(
		filepath.Join(project, "go.mod"),
		[]byte("module example.com/demo\n"),
		fsperm.File,
	); writeErr != nil {
		t.Fatalf("write go.mod: %v", writeErr)
	}
	res := callTool(t, session, "project_profile", map[string]any{"path": project})
	if res.IsError {
		t.Fatalf("project_profile reported an error: %+v", res.Content)
	}
	var profile mcp.ProjectProfileOut
	decode(t, res.StructuredContent, &profile)
	if !slices.Contains(signalKinds(profile.Signals), "language:go") {
		t.Errorf("signals = %+v, want language:go", profile.Signals)
	}
	candidate, ok := findCandidate(profile.Candidates, "go-architect")
	if !ok {
		t.Fatalf("candidates = %+v, want go-architect", profile.Candidates)
	}
	if candidate.Latest != "2.0.0" {
		t.Errorf("go-architect latest = %q, want 2.0.0", candidate.Latest)
	}

	// search_skills by description.
	res = callTool(t, session, "search_skills", map[string]any{"query": "architectural standards"})
	var searchOut mcp.SearchSkillsOut
	decode(t, res.StructuredContent, &searchOut)
	if len(searchOut.Candidates) == 0 || searchOut.Candidates[0].Skill != "go-architect" {
		t.Errorf("search candidates = %+v, want go-architect first", searchOut.Candidates)
	}

	// Read the catalog resource.
	read, err := session.ReadResource(t.Context(), &sdk.ReadResourceParams{URI: "rsk://catalog"})
	if err != nil {
		t.Fatalf("ReadResource: %v", err)
	}
	if len(read.Contents) != 1 || !strings.Contains(read.Contents[0].Text, "go-architect") {
		t.Errorf("catalog = %+v, want a go-architect entry", read.Contents)
	}
}

func TestProjectProfileMissingPathIsToolError(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	session := newSession(t, mcp.Deps{
		Cfg:     config.Config{RepoPath: repo},
		Logger:  slog.New(slog.DiscardHandler),
		Version: "test",
	})

	missing := filepath.Join(t.TempDir(), "absent")
	res := callTool(t, session, "project_profile", map[string]any{"path": missing})
	if !res.IsError {
		t.Error("project_profile on a missing path: want isError true, got false")
	}
}

func newSession(t *testing.T, deps mcp.Deps) *sdk.ClientSession {
	t.Helper()

	srv := mcp.NewServer(deps)
	serverTransport, clientTransport := sdk.NewInMemoryTransports()
	go func() { _ = srv.Run(t.Context(), serverTransport) }()

	client := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "0.0.0"}, nil)
	session, err := client.Connect(t.Context(), clientTransport, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func callTool(t *testing.T, session *sdk.ClientSession, name string, args map[string]any) *sdk.CallToolResult {
	t.Helper()

	res, err := session.CallTool(t.Context(), &sdk.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("CallTool(%s): %v", name, err)
	}
	return res
}

func findTool(t *testing.T, tools []*sdk.Tool, name string) *sdk.Tool {
	t.Helper()

	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}
	t.Fatalf("tool %q not registered", name)
	return nil
}

func findCandidate(candidates []mcp.Candidate, skill string) (mcp.Candidate, bool) {
	for _, c := range candidates {
		if c.Skill == skill {
			return c, true
		}
	}
	return mcp.Candidate{}, false
}

func decode(t *testing.T, value, out any) {
	t.Helper()

	data, marshalErr := json.Marshal(value)
	if marshalErr != nil {
		t.Fatalf("marshal: %v", marshalErr)
	}
	if unmarshalErr := json.Unmarshal(data, out); unmarshalErr != nil {
		t.Fatalf("unmarshal %T: %v", out, unmarshalErr)
	}
}

func writeSkill(t *testing.T, dir, name, version, description string) {
	t.Helper()

	if err := os.MkdirAll(dir, fsperm.Dir); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := "---\nname: " + name + "\nversion: " + version + "\ndescription: " + description + "\n---\n\n# " + name + "\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(body), fsperm.File); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}
}

func TestInstallSkillsGlobal(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	writeSkill(t, filepath.Join(repo, "skills", "tdd"), "tdd", "1.0.0", "Red-green-refactor workflow.")
	globalDir := t.TempDir()

	session := newSession(t, mcp.Deps{
		Cfg: config.Config{
			RepoPath:           repo,
			GlobalTargets:      map[string]string{"claude-code": globalDir},
			DefaultTargetScope: "claude-code",
		},
		Logger:  slog.New(slog.DiscardHandler),
		Version: "test",
	})

	res := callTool(t, session, "install_skills", map[string]any{
		"names": []string{"tdd"},
		"scope": "global",
	})
	if res.IsError {
		t.Fatalf("install_skills reported an error: %+v", res.Content)
	}
	var out mcp.InstallSkillsOut
	decode(t, res.StructuredContent, &out)
	if len(out.Installed) != 1 || out.Installed[0].Name != "tdd" {
		t.Fatalf("installed = %+v, want tdd", out.Installed)
	}
	if _, err := os.Lstat(filepath.Join(globalDir, "tdd")); err != nil {
		t.Errorf("tdd not linked into the global dir: %v", err)
	}
}

func TestInstallSkillsUnknownNameIsToolError(t *testing.T) {
	t.Parallel()

	session := newSession(t, mcp.Deps{
		Cfg: config.Config{
			RepoPath:           t.TempDir(),
			GlobalTargets:      map[string]string{"claude-code": t.TempDir()},
			DefaultTargetScope: "claude-code",
		},
		Logger:  slog.New(slog.DiscardHandler),
		Version: "test",
	})

	res := callTool(t, session, "install_skills", map[string]any{
		"names": []string{"does-not-exist"},
		"scope": "global",
	})
	if !res.IsError {
		t.Error("install_skills with an unknown name: want isError true")
	}
}
