package source

import "errors"

// ErrNotFound is returned when a skill cannot be located in a source.
var ErrNotFound = errors.New("skill not found")
