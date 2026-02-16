package jpath_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kode4food/jpath"
)

func TestFunctionValidationPaths(t *testing.T) {
	reg := jpath.NewRegistry()
	reg.MustRegisterFunction("novalidate", &jpath.FunctionDefinition{
		Eval: func(_ []*jpath.FunctionValue) *jpath.FunctionValue {
			return &jpath.FunctionValue{Scalar: true}
		},
	})

	got, err := reg.Query("$[?novalidate()]", []any{float64(1), float64(2)})
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []any{float64(1), float64(2)}, got)

	_, err = reg.Query("$[?missing()]", []any{float64(1)})
	assert.ErrorIs(t, err, jpath.ErrInvalidPath)
	assert.ErrorIs(t, err, jpath.ErrUnknownFunc)
}

func TestFunctionValidatorError(t *testing.T) {
	reg := jpath.NewRegistry()
	reg.MustRegisterFunction("needsArg", &jpath.FunctionDefinition{
		Validate: func(
			args []jpath.FilterExpr, _ jpath.FunctionUse, _ bool,
		) error {
			if len(args) != 1 {
				return errors.New("invalid function arity")
			}
			return nil
		},
		Eval: func(_ []*jpath.FunctionValue) *jpath.FunctionValue {
			return &jpath.FunctionValue{Scalar: true}
		},
	})

	_, err := reg.Query("$[?needsArg()]", []any{float64(1)})
	assert.Error(t, err)
	assert.ErrorContains(t, err, "invalid function arity")
}

func TestValueRequiresQueryArgument(t *testing.T) {
	reg := jpath.NewRegistry()
	_, err := reg.Query("$[?value(1) == 1]", []any{float64(1)})
	assert.Error(t, err)
	assert.ErrorContains(t, err, "value requires query argument")
}
