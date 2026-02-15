package jpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustRegisterFunction(t *testing.T) {
	reg := NewRegistry()

	assert.NotPanics(t, func() {
		reg.MustRegisterFunction("alwaysTrue", FunctionDefinition{
			Eval: func(_ []FunctionValue) FunctionValue {
				return FunctionValue{Scalar: true}
			},
		})
	})

	assert.Panics(t, func() {
		reg.MustRegisterFunction("alwaysTrue", FunctionDefinition{
			Eval: func(_ []FunctionValue) FunctionValue {
				return FunctionValue{Scalar: true}
			},
		})
	})
}
