package internal

// OutputFormat represents the --output flag value for the list command.
type OutputFormat string

const (
	OutputText OutputFormat = "text"
	OutputJSON OutputFormat = "json"
)

// String returns the string representation of the output format.
func (o OutputFormat) String() string { return string(o) }

// Valid reports whether o is a known output format.
func (o OutputFormat) Valid() bool {
	switch o {
	case OutputText, OutputJSON:
		return true
	default:
		return false
	}
}
