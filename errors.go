package jpath

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidPath is raised when a JSONPath query cannot be parsed
	ErrInvalidPath = errors.New("invalid JSONPath query")

	// ErrLiteralMustBeCompared is raised for bare literals in logical context
	ErrLiteralMustBeCompared = errors.New("literal must be compared")

	// ErrCompRequiresSingularQuery is raised for non-singular comparisons
	ErrCompRequiresSingularQuery = errors.New(
		"comparison requires singular query",
	)

	// ErrInvalidFuncArity is raised when function arity is invalid
	ErrInvalidFuncArity = errors.New("invalid function arity")

	// ErrFuncResultMustBeCompared is raised for logical use without compare
	ErrFuncResultMustBeCompared = errors.New(
		"function result must be compared",
	)

	// ErrFuncResultMustNotBeCompared is raised for prohibited comparisons
	ErrFuncResultMustNotBeCompared = errors.New(
		"function result must not be compared",
	)

	// ErrFuncRequiresSingularQuery is raised for singular-path requirements
	ErrFuncRequiresSingularQuery = errors.New(
		"function requires singular query",
	)

	// ErrFuncRequiresQueryArgument is raised for query-arg requirements
	ErrFuncRequiresQueryArgument = errors.New(
		"function requires query argument",
	)
)

func wrapPathError(query string, pos int, err error) error {
	return fmt.Errorf(
		"%w at offset %d in %q: %w", ErrInvalidPath, pos, query, err,
	)
}
