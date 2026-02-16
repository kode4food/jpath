package jpath_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kode4food/jpath"
)

func TestRegistryParseCompileQuery(t *testing.T) {
	reg := jpath.NewRegistry()
	ast, err := reg.Parse("$[1]")
	if !assert.NoError(t, err) {
		return
	}
	path, err := reg.Compile(ast)
	if !assert.NoError(t, err) {
		return
	}
	got := path.Query([]any{float64(10), float64(20), float64(30)})
	assert.Equal(t, []any{float64(20)}, got)
}

func TestRegistryMustHelpers(t *testing.T) {
	reg := jpath.NewRegistry()
	_ = reg.MustParse("$[0]")

	ast := reg.MustParse("$[0]")
	path := reg.MustCompile(ast)
	got := path.Query([]any{float64(1)})
	assert.Equal(t, []any{float64(1)}, got)

	got = reg.MustQuery("$[0]", []any{float64(7)})
	assert.Equal(t, []any{float64(7)}, got)
}

func TestRegistryMustHelpersPanic(t *testing.T) {
	reg := jpath.NewRegistry()

	assert.Panics(t, func() {
		_ = reg.MustParse("")
	})
	assert.Panics(t, func() {
		_ = reg.MustQuery("", nil)
	})
	assert.Panics(t, func() {
		_ = reg.MustCompile(&jpath.PathExpr{
			Segments: []*jpath.SegmentExpr{{
				Selectors: []*jpath.SelectorExpr{{
					Kind:   jpath.SelectorFilter,
					Filter: &jpath.LiteralExpr{Value: true},
				}},
			}},
		})
	})
}

func TestRegistryQueryErrorWrapping(t *testing.T) {
	reg := jpath.NewRegistry()
	_, err := reg.Query("", nil)
	assert.ErrorIs(t, err, jpath.ErrInvalidPath)
	assert.ErrorIs(t, err, jpath.ErrExpectedRoot)
}

func TestRegistrySliceSpecializations(t *testing.T) {
	reg := jpath.NewRegistry()
	doc := []any{
		float64(0), float64(1), float64(2), float64(3), float64(4),
	}

	assertRegistryQuery(t, reg, "$[:-1]", doc, []any{
		float64(0), float64(1), float64(2), float64(3),
	})
	assertRegistryQuery(t, reg, "$[-2::-1]", doc, []any{
		float64(3), float64(2), float64(1), float64(0),
	})
	assertRegistryQuery(t, reg, "$[:-3:-1]", doc, []any{
		float64(4), float64(3),
	})
}

func TestRegistryExtensionFunction(t *testing.T) {
	reg := jpath.NewRegistry()
	err := reg.RegisterFunction("alwaysTrue", &jpath.FunctionDefinition{
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
	if !assert.NoError(t, err) {
		return
	}

	got, err := reg.Query(
		"$[?alwaysTrue(@)]",
		[]any{float64(1), float64(2)},
	)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []any{float64(1), float64(2)}, got)
}

func TestRegistryExtensionFunctionNodes(t *testing.T) {
	reg := jpath.NewRegistry()
	err := reg.RegisterFunction("nodeTruthy", &jpath.FunctionDefinition{
		Eval: func(_ []*jpath.FunctionValue) *jpath.FunctionValue {
			return &jpath.FunctionValue{
				IsNodes: true,
				Nodes:   []any{float64(1)},
			}
		},
	})
	if !assert.NoError(t, err) {
		return
	}
	got, err := reg.Query("$[?nodeTruthy()]", []any{float64(3), float64(4)})
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []any{float64(3), float64(4)}, got)
}

func TestRegistryFunctionIsolation(t *testing.T) {
	base := jpath.NewRegistry()
	sandbox := base.Clone()
	err := sandbox.RegisterFunction("alwaysFalse", &jpath.FunctionDefinition{
		Validate: func(
			_ []jpath.FilterExpr, _ jpath.FunctionUse, _ bool,
		) error {
			return nil
		},
		Eval: func(_ []*jpath.FunctionValue) *jpath.FunctionValue {
			return &jpath.FunctionValue{Scalar: false}
		},
	})
	if !assert.NoError(t, err) {
		return
	}

	_, err = base.Query("$[?alwaysFalse(@)]", nil)
	assert.ErrorIs(t, err, jpath.ErrUnknownFunc)
	assert.ErrorIs(t, err, jpath.ErrInvalidPath)

	_, err = sandbox.Query("$[?alwaysFalse(@)]", nil)
	assert.NoError(t, err)
}

func TestRegistryRegisterFunctionErrors(t *testing.T) {
	reg := jpath.NewRegistry()
	err := reg.RegisterFunction("1bad", &jpath.FunctionDefinition{
		Eval: func(_ []*jpath.FunctionValue) *jpath.FunctionValue {
			return &jpath.FunctionValue{}
		},
	})
	assert.ErrorIs(t, err, jpath.ErrBadFuncName)

	err = reg.RegisterFunction("x", &jpath.FunctionDefinition{})
	assert.ErrorIs(t, err, jpath.ErrBadFuncDefinition)

	err = reg.RegisterFunction("dup", &jpath.FunctionDefinition{
		Eval: func(_ []*jpath.FunctionValue) *jpath.FunctionValue {
			return &jpath.FunctionValue{}
		},
	})
	if !assert.NoError(t, err) {
		return
	}

	err = reg.RegisterFunction("dup", &jpath.FunctionDefinition{
		Eval: func(_ []*jpath.FunctionValue) *jpath.FunctionValue {
			return &jpath.FunctionValue{}
		},
	})
	assert.ErrorIs(t, err, jpath.ErrFuncExists)
}

func TestMustRegisterFunction(t *testing.T) {
	reg := jpath.NewRegistry()

	assert.NotPanics(t, func() {
		reg.MustRegisterFunction("alwaysTrue", &jpath.FunctionDefinition{
			Eval: func(_ []*jpath.FunctionValue) *jpath.FunctionValue {
				return &jpath.FunctionValue{Scalar: true}
			},
		})
	})

	assert.Panics(t, func() {
		reg.MustRegisterFunction("alwaysTrue", &jpath.FunctionDefinition{
			Eval: func(_ []*jpath.FunctionValue) *jpath.FunctionValue {
				return &jpath.FunctionValue{Scalar: true}
			},
		})
	})
}

func assertRegistryQuery(
	t *testing.T, reg *jpath.Registry, query string, doc any, want []any,
) {
	t.Helper()
	got, err := reg.Query(query, doc)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, want, got)
}
