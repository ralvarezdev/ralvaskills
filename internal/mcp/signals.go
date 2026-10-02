package mcp

// signalRule recognizes one project trait and proposes the skills it maps to.
// A rule's detect returns the relative path of the file that proves the signal;
// it must never return ok with an empty proof.
type signalRule struct {
	detect     func(*fileIndex) (string, bool)
	kind       string
	candidates []string
}

// signalRules is the lookup table of SKILL and skill names: it does not score,
// rank, or order by priority, it only proposes. The slice order fixes the order
// of Profile.Signals and of first-seen candidates.
//
// quality:refactor skills (logic-cleaner, improve-codebase-architecture) are
// deliberately absent: they are never proposed by a project signal, only when
// the agent is already refactoring.
func signalRules() []signalRule {
	return []signalRule{
		{
			kind:       "language:go",
			candidates: []string{"go-architect"},
			detect: func(idx *fileIndex) (string, bool) {
				return idx.has("go.mod")
			},
		},
		{
			kind:       "language:python",
			candidates: []string{"python-architect"},
			detect: func(idx *fileIndex) (string, bool) {
				return idx.hasAny("pyproject.toml", "requirements.txt", "uv.lock")
			},
		},
		{
			kind: "language:typescript",
			detect: func(idx *fileIndex) (string, bool) {
				return idx.hasAny("tsconfig.json", "package.json")
			},
		},
		{
			kind:       "framework:nextjs",
			candidates: []string{"nextjs-architect", "react-architect"},
			detect: func(idx *fileIndex) (string, bool) {
				if proof, ok := idx.firstRootPrefixed("next.config."); ok {
					return proof, true
				}
				if idx.pkgJSONHasDep("next") {
					return "package.json", true
				}
				return "", false
			},
		},
		{
			kind:       "framework:astro",
			candidates: []string{"astro-architect"},
			detect: func(idx *fileIndex) (string, bool) {
				return idx.firstRootPrefixed("astro.config.")
			},
		},
		{
			kind:       "protocol:grpc",
			candidates: []string{"grpc-architect", "protobuf-architect"},
			detect: func(idx *fileIndex) (string, bool) {
				if proof, ok := idx.firstBySuffix(".proto"); ok {
					return proof, true
				}
				if idx.goModHasDep("google.golang.org/grpc") {
					return "go.mod", true
				}
				return "", false
			},
		},
		{
			kind:       "protocol:rest",
			candidates: []string{"rest-api-architect"},
			detect: func(idx *fileIndex) (string, bool) {
				if proof, ok := idx.hasAny("openapi.yaml", "openapi.yml", "swagger.json"); ok {
					return proof, true
				}
				return idx.firstBySuffix(".http")
			},
		},
		{
			kind:       "infra:docker",
			candidates: []string{"docker-architect"},
			detect: func(idx *fileIndex) (string, bool) {
				return idx.hasAny("Dockerfile", "docker-compose.yml", "docker-compose.yaml")
			},
		},
		{
			kind: "infra:k8s",
			detect: func(idx *fileIndex) (string, bool) {
				return idx.hasAny("Chart.yaml", "kustomization.yaml", "kustomization.yml")
			},
		},
		{
			kind:       "infra:ci",
			candidates: []string{"ci-cd-architect"},
			detect: func(idx *fileIndex) (string, bool) {
				return idx.dirHasFiles(".github/workflows")
			},
		},
		{
			kind:       "repo:tooling",
			candidates: []string{"repo-tooling-architect"},
			detect: func(idx *fileIndex) (string, bool) {
				return idx.hasAny(".golangci.yml", ".golangci.yaml", "mise.toml", "Taskfile.yml", "Taskfile.yaml")
			},
		},
		{
			kind:       "repo:cli",
			candidates: []string{"cli-tool-architect"},
			detect: func(idx *fileIndex) (string, bool) {
				if proof, ok := idx.has("main.go"); ok {
					return proof, true
				}
				return idx.firstNamedUnder("cmd/", "main.go")
			},
		},
		{
			kind: "data:postgres",
			detect: func(idx *fileIndex) (string, bool) {
				if !idx.goModHasDep("sqlx") && !idx.pythonHasDep("psycopg") {
					return "", false
				}
				return idx.firstBySuffix(".sql")
			},
		},
	}
}
