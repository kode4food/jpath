package jpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateFunctionBranches(t *testing.T) {
	reg := NewRegistry()
	reg.MustRegisterFunction("novalidate", FunctionDefinition{
		Eval: func(_ []FunctionValue) FunctionValue {
			return FunctionValue{Scalar: true}
		},
	})

	err := validateFunction(
		FuncExpr{Name: "novalidate"},
		contextLogical,
		false,
		reg,
	)
	assert.NoError(t, err)

	err = validateFunction(
		FuncExpr{Name: "missing"},
		contextLogical,
		false,
		reg,
	)
	assert.ErrorIs(t, err, ErrUnknownFunction)
}
