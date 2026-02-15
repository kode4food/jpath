package jpath

import (
	"errors"
	"fmt"
)

// ErrInvalidPath is raised when a JSONPath query cannot be parsed
var ErrInvalidPath = errors.New("invalid JSONPath query")

func wrapPathError(query string, pos int, err error) error {
	return fmt.Errorf(
		"%w at offset %d in %q: %w", ErrInvalidPath, pos, query, err,
	)
}
