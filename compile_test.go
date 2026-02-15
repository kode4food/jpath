package jpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type unknownFilterExpr struct{}

func TestNewCompiler(t *testing.T) {
	c := NewCompiler()
	assert.NotNil(t, c)
}

func TestCompileInternalErrorBranches(t *testing.T) {
	_, err := compileFilter(unknownFilterExpr{}, NewRegistry())
	assert.Error(t, err)

	_, err = compileSelector(
		SelectorExpr{Kind: SelectorKind(255)},
		&Path{},
		NewRegistry(),
	)
	assert.Error(t, err)
}

func (unknownFilterExpr) eval(_ *evalCtx) evalValue {
	return scalarValue(nil)
}
