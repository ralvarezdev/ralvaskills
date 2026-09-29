package cmdx

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/termkit"
)

// Values accepted by the --include flag.
const (
	IncludePersonal = "personal"
	IncludeOfficial = "official"
)

// Includes reports which optional skill scopes a command was asked to cover.
type Includes struct {
	Personal bool
	Official bool
}

// RegisterInclude registers --include on cmd, plus the legacy --personal
// (and, when official is among allowed, --official) bool flags as hidden
// aliases that keep working exactly as before. It also wires shell
// completion for the allowed values.
func RegisterInclude(cmd *cobra.Command, usage string, allowed ...string) {
	flags := cmd.Flags()
	flags.StringSlice(FlagInclude, nil, usage+" ("+strings.Join(allowed, "|")+", repeatable)")
	flags.Bool(FlagPersonal, false, "Alias for --include personal")
	legacy := []string{FlagPersonal}
	if slices.Contains(allowed, IncludeOfficial) {
		flags.Bool(FlagOfficial, false, "Alias for --include official")
		legacy = append(legacy, FlagOfficial)
	}
	for _, name := range legacy {
		flags.Lookup(name).Hidden = true
	}

	err := cmd.RegisterFlagCompletionFunc(FlagInclude,
		func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
			return allowed, cobra.ShellCompDirectiveNoFileComp
		})
	if err != nil {
		panic(fmt.Sprintf("register completion for --%s on %q: %v", FlagInclude, cmd.Name(), err))
	}
}

// ReadIncludes resolves --include and the hidden legacy bool aliases into one
// Includes value, rejecting unknown values and values not in allowed.
func ReadIncludes(cmd *cobra.Command, allowed ...string) (Includes, error) {
	var inc Includes

	values, err := cmd.Flags().GetStringSlice(FlagInclude)
	if err != nil {
		panic(fmt.Sprintf("flag %q not registered on %q: %v", FlagInclude, cmd.Name(), err))
	}
	for _, value := range values {
		switch v := strings.ToLower(strings.TrimSpace(value)); {
		case v == "":
		case !slices.Contains(allowed, v):
			return Includes{}, termkit.NewFieldError(FlagInclude, fmt.Errorf(
				"--%s: invalid value %q (allowed: %s)",
				FlagInclude,
				value,
				strings.Join(allowed, ", "),
			))
		case v == IncludePersonal:
			inc.Personal = true
		case v == IncludeOfficial:
			inc.Official = true
		}
	}

	inc.Personal = inc.Personal || Bool(cmd, FlagPersonal)
	if slices.Contains(allowed, IncludeOfficial) {
		inc.Official = inc.Official || Bool(cmd, FlagOfficial)
	}
	return inc, nil
}
