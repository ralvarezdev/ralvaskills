package source

import "errors"

var (
	// ErrNotFound is returned when a skill cannot be located in a source.
	ErrNotFound = errors.New("skill not found")
)
