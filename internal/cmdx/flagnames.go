package cmdx

// Flag name constants used when registering and reading cobra flags.
const (
	FlagGlobal   = "global"
	FlagDryRun   = "dry-run"
	FlagFor      = "for"
	FlagPersonal = "personal"
	FlagVersion  = "version"
	FlagSource   = "source"
	FlagOutput   = "output"
	FlagBundle   = "bundle"
	FlagBundles  = "bundles"
	FlagStack    = "stack"
	FlagRefresh  = "refresh"
	FlagOfficial = "official"
	FlagPin      = "pin"
	FlagForce    = "force"
)

// TargetScope is the typed value of the --for flag and the
// default_target_scope config field. A scope is either a specific tool ID or
// ScopeAll.
type TargetScope string

// ScopeAll selects every configured tool when used as the --for flag value or
// as default_target_scope in config.json.
const ScopeAll TargetScope = "all"

// ForAll is the string value of ScopeAll kept for backward-compatible
// string comparisons. Prefer ScopeAll when the type matters.
const ForAll = string(ScopeAll)
