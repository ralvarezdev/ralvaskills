package mcp

// signalRule recognizes one project trait and proposes the skills it maps to.
// A rule's detect returns the relative path of the file that proves the signal;
// it must never return ok with an empty proof.
type signalRule struct {
	candidates []string
	kind       SignalKind
	detect     func(*fileIndex) (string, bool)
}

// signalRules is the lookup table of signal kinds to candidate skills: it does
// not score, rank, or order by priority, it only proposes. The slice order
// fixes the order of Profile.Signals and of first-seen candidates.
//
// quality:refactor skills (logic-cleaner, improve-codebase-architecture) are
// deliberately absent: they are never proposed by a project signal, only when
// the agent is already refactoring.
func signalRules() []signalRule {
	return []signalRule{
		{
			kind:       SignalLanguageGo,
			candidates: []string{"go-architect"},
			detect: func(idx *fileIndex) (string, bool) {
				return idx.has("go.mod")
			},
		},
		{
			kind:       SignalLanguagePython,
			candidates: []string{"python-architect"},
			detect: func(idx *fileIndex) (string, bool) {
				return idx.hasAny("pyproject.toml", "requirements.txt", "uv.lock")
			},
		},
		{
			kind: SignalLanguageTS,
			detect: func(idx *fileIndex) (string, bool) {
				return idx.hasAny("tsconfig.json", "package.json")
			},
		},
		{
			kind:       SignalFrameworkNextJS,
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
			kind:       SignalFrameworkAstro,
			candidates: []string{"astro-architect"},
			detect: func(idx *fileIndex) (string, bool) {
				return idx.firstRootPrefixed("astro.config.")
			},
		},
		{
			kind:       SignalProtocolGRPC,
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
			kind:       SignalProtocolREST,
			candidates: []string{"rest-api-architect"},
			detect: func(idx *fileIndex) (string, bool) {
				if proof, ok := idx.hasAny("openapi.yaml", "openapi.yml", "swagger.json"); ok {
					return proof, true
				}
				return idx.firstBySuffix(".http")
			},
		},
		{
			kind:       SignalInfraDocker,
			candidates: []string{"docker-architect"},
			detect: func(idx *fileIndex) (string, bool) {
				return idx.hasAny("Dockerfile", "docker-compose.yml", "docker-compose.yaml")
			},
		},
		{
			kind: SignalInfraK8s,
			detect: func(idx *fileIndex) (string, bool) {
				return idx.hasAny("Chart.yaml", "kustomization.yaml", "kustomization.yml")
			},
		},
		{
			kind:       SignalInfraCI,
			candidates: []string{"ci-cd-architect"},
			detect: func(idx *fileIndex) (string, bool) {
				return idx.dirHasFiles(".github/workflows")
			},
		},
		{
			kind:       SignalRepoTooling,
			candidates: []string{"repo-tooling-architect"},
			detect: func(idx *fileIndex) (string, bool) {
				return idx.hasAny(".golangci.yml", ".golangci.yaml", "mise.toml", "Taskfile.yml", "Taskfile.yaml")
			},
		},
		{
			kind:       SignalRepoCLI,
			candidates: []string{"cli-tool-architect"},
			detect: func(idx *fileIndex) (string, bool) {
				if proof, ok := idx.has("main.go"); ok {
					return proof, true
				}
				return idx.firstNamedUnder("cmd/", "main.go")
			},
		},
		{
			kind: SignalDataPostgres,
			detect: func(idx *fileIndex) (string, bool) {
				if !idx.goModHasDep("sqlx") && !idx.pythonHasDep("psycopg") {
					return "", false
				}
				return idx.firstBySuffix(".sql")
			},
		},
	}
}
