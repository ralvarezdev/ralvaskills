package config

// SourceID identifies where a skill originates.
type SourceID string

const (
	SourceLocal    SourceID = "local"    // symlinked from the local ralvaskills repo
	SourceOfficial SourceID = "official" // fetched from anthropics/skills
)

// SkillRef is a reference to a named skill from a specific source.
type SkillRef struct {
	Name   string
	Source SourceID
}

// Bundle is a named, ordered set of skill references.
type Bundle struct {
	Name        string
	Description string
	Skills      []SkillRef
}

func local(name string) SkillRef    { return SkillRef{Name: name, Source: SourceLocal} }
func official(name string) SkillRef { return SkillRef{Name: name, Source: SourceOfficial} }

// Catalog returns the full bundle catalog in canonical display order.
func Catalog() []Bundle {
	return []Bundle{
		{
			Name:        "global",
			Description: "Universal skills — install machine-wide; apply to every project",
			Skills: []SkillRef{
				local("commit-author"),
				local("logic-cleaner"),
				local("code-design-refactor"),
				local("tdd"),
				local("grill-with-docs"),
				local("caveman"),
				local("feature-planner"),
				local("security-reviewer"),
				local("ddd-architect"),
				local("hexagonal-arch"),
				local("design-patterns"),
				local("improve-codebase-architecture"),
				local("rsk-guide"),
				local("repo-tooling-architect"),
				local("skill-builder"),
			},
		},
		{
			Name:        "docs",
			Description: "Document creation and technical reference skills (official)",
			Skills: []SkillRef{
				official("docx"),
				official("xlsx"),
				official("pdf"),
				official("pptx"),
				official("find-docs"),
			},
		},
		{
			Name:        "design",
			Description: "Frontend and UI/UX design skills",
			Skills: []SkillRef{
				official("frontend-design"),
				local("react-architect"),
				local("nextjs-architect"),
				local("ui-ux-architect"),
			},
		},
		{
			Name:        "go-grpc",
			Description: "Go service exposing a gRPC API",
			Skills: []SkillRef{
				local("go-architect"),
				local("grpc-architect"),
				local("protobuf-architect"),
				local("docker-architect"),
				local("sql-architect"),
			},
		},
		{
			Name:        "gin",
			Description: "Go service exposing a REST API via Gin",
			Skills: []SkillRef{
				local("go-architect"),
				local("gin-architect"),
				local("rest-api-architect"),
				local("docker-architect"),
				local("sql-architect"),
			},
		},
		{
			Name:        "nethttp",
			Description: "Go service exposing a REST API via stdlib net/http",
			Skills: []SkillRef{
				local("go-architect"),
				local("nethttp-architect"),
				local("rest-api-architect"),
				local("docker-architect"),
				local("sql-architect"),
			},
		},
		{
			Name:        "go-cli",
			Description: "Go command-line tool",
			Skills: []SkillRef{
				local("go-architect"),
				local("cli-tool-architect"),
				local("docker-architect"),
			},
		},
		{
			Name:        "fastapi",
			Description: "Python service exposing a REST API with FastAPI",
			Skills: []SkillRef{
				local("python-architect"),
				local("fastapi-architect"),
				local("rest-api-architect"),
				local("docker-architect"),
				local("sql-architect"),
			},
		},
		{
			Name:        "llm-app",
			Description: "Python-based LLM application or RAG pipeline",
			Skills: []SkillRef{
				local("python-architect"),
				local("fastapi-architect"),
				local("llm-app-architect"),
				local("agent-architect"),
				local("hexagonal-arch"),
				local("docker-architect"),
			},
		},
		{
			Name:        "ros2",
			Description: "ROS2 robotics application",
			Skills: []SkillRef{
				local("python-architect"),
				local("ros2-architect"),
				local("docker-architect"),
			},
		},
	}
}

// FindBundle returns the bundle with the given name and whether it was found.
func FindBundle(name string) (Bundle, bool) {
	for _, b := range Catalog() {
		if b.Name == name {
			return b, true
		}
	}
	return Bundle{}, false
}

// BundleNames returns all bundle names in catalog order.
func BundleNames() []string {
	all := Catalog()
	names := make([]string, len(all))
	for i, b := range all {
		names[i] = b.Name
	}
	return names
}
