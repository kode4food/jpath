package jpath

import (
	"errors"
	"fmt"
)

// Path is a compiled JSONPath query
type Path struct {
	source   string
	runnable Runnable
}

// ErrInvalidPath is raised when a JSONPath query cannot be parsed
var ErrInvalidPath = errors.New("invalid JSONPath query")

// Query executes a compiled path against a JSON document
func (p *Path) Query(document any) []any {
	return p.runnable.Run(document)
}

func wrapPathError(query string, pos int, err error) error {
	return fmt.Errorf(
		"%w at offset %d in %q: %w",
		ErrInvalidPath,
		pos,
		query,
		err,
	)
}
