package jpath_test

import (
	"errors"
	"strings"
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
	got := path([]any{float64(10), float64(20), float64(30)})
	assert.Equal(t, []any{float64(20)}, got)
}

func TestRegistryMustHelpers(t *testing.T) {
	reg := jpath.NewRegistry()
	_ = reg.MustParse("$[0]")

	ast := reg.MustParse("$[0]")
	path := reg.MustCompile(ast)
	got := path([]any{float64(1)})
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
	err := reg.RegisterDefinition("alwaysTrue", &jpath.FunctionDefinition{
		Validate: func(
			args []jpath.FilterExpr, _ jpath.FunctionUse, _ bool,
		) error {
			if len(args) != 1 {
				return errors.New("invalid function arity")
			}
			return nil
		},
		Eval: func(_ []*jpath.Value) *jpath.Value {
			return &jpath.Value{Scalar: true}
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

func TestRegistryRegisterFunction(t *testing.T) {
	reg := jpath.NewRegistry()
	err := reg.RegisterFunction(
		"startsWith", 2, func(args ...any) (any, bool) {
			left, ok := args[0].(string)
			if !ok {
				return nil, false
			}
			right, ok := args[1].(string)
			if !ok {
				return nil, false
			}
			return strings.HasPrefix(left, right), true
		},
	)
	if !assert.NoError(t, err) {
		return
	}

	doc := []any{
		map[string]any{"name": "alpha"},
		map[string]any{"name": "beta"},
	}
	got, err := reg.Query("$[?startsWith(@.name, 'al')]", doc)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []any{doc[0]}, got)
}

func TestRegistryRegisterFunctionArityValidation(t *testing.T) {
	reg := jpath.NewRegistry()
	err := reg.RegisterFunction(
		"startsWith", 2, func(args ...any) (any, bool) {
			return true, true
		},
	)
	if !assert.NoError(t, err) {
		return
	}

	_, err = reg.Query("$[?startsWith(@.name)]", []any{
		map[string]any{"name": "alpha"},
	})
	assert.ErrorIs(t, err, jpath.ErrInvalidPath)
	assert.ErrorIs(t, err, jpath.ErrInvalidFuncArity)
}

func TestRegistryRegisterFunctionSingularCoercion(t *testing.T) {
	reg := jpath.NewRegistry()
	err := reg.RegisterFunction(
		"isPositive", 1, func(args ...any) (any, bool) {
			n, ok := args[0].(float64)
			if !ok {
				return nil, false
			}
			return n > 0, true
		},
	)
	if !assert.NoError(t, err) {
		return
	}

	doc := []any{
		map[string]any{"vals": []any{float64(1), float64(2)}},
		map[string]any{"vals": []any{float64(-1), float64(-2)}},
	}
	got, err := reg.Query("$[?isPositive(@.vals[0])]", doc)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []any{doc[0]}, got)

	got, err = reg.Query("$[?isPositive(@.vals[*])]", doc)
	if !assert.NoError(t, err) {
		return
	}
	assert.Empty(t, got)
}

func TestRegistryExtensionFunctionNodes(t *testing.T) {
	reg := jpath.NewRegistry()
	err := reg.RegisterDefinition("nodeTruthy", &jpath.FunctionDefinition{
		Eval: func(_ []*jpath.Value) *jpath.Value {
			return &jpath.Value{
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
	err := sandbox.RegisterDefinition("alwaysFalse", &jpath.FunctionDefinition{
		Validate: func(
			_ []jpath.FilterExpr, _ jpath.FunctionUse, _ bool,
		) error {
			return nil
		},
		Eval: func(_ []*jpath.Value) *jpath.Value {
			return &jpath.Value{Scalar: false}
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
	err := reg.RegisterDefinition("1bad", &jpath.FunctionDefinition{
		Eval: func(_ []*jpath.Value) *jpath.Value {
			return &jpath.Value{}
		},
	})
	assert.ErrorIs(t, err, jpath.ErrBadFuncName)

	err = reg.RegisterDefinition("x", &jpath.FunctionDefinition{})
	assert.ErrorIs(t, err, jpath.ErrBadFuncDefinition)

	err = reg.RegisterDefinition("dup", &jpath.FunctionDefinition{
		Eval: func(_ []*jpath.Value) *jpath.Value {
			return &jpath.Value{}
		},
	})
	if !assert.NoError(t, err) {
		return
	}

	err = reg.RegisterDefinition("dup", &jpath.FunctionDefinition{
		Eval: func(_ []*jpath.Value) *jpath.Value {
			return &jpath.Value{}
		},
	})
	assert.ErrorIs(t, err, jpath.ErrFuncExists)
}

func TestMustRegisterDefinition(t *testing.T) {
	reg := jpath.NewRegistry()

	assert.NotPanics(t, func() {
		reg.MustRegisterDefinition("alwaysTrue", &jpath.FunctionDefinition{
			Eval: func(_ []*jpath.Value) *jpath.Value {
				return &jpath.Value{Scalar: true}
			},
		})
	})

	assert.Panics(t, func() {
		reg.MustRegisterDefinition("alwaysTrue", &jpath.FunctionDefinition{
			Eval: func(_ []*jpath.Value) *jpath.Value {
				return &jpath.Value{Scalar: true}
			},
		})
	})
}

func TestMustRegisterFunction(t *testing.T) {
	reg := jpath.NewRegistry()

	assert.NotPanics(t, func() {
		reg.MustRegisterFunction(
			"alwaysTrue", 0, func(args ...any) (any, bool) {
				return true, true
			},
		)
	})

	assert.Panics(t, func() {
		reg.MustRegisterFunction(
			"alwaysTrue", 0, func(args ...any) (any, bool) {
				return true, true
			},
		)
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
