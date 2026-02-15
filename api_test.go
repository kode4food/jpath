package jpath_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kode4food/jpath"
)

func TestTopLevelParseCompileQuery(t *testing.T) {
	reg := jpath.NewRegistry()

	ast, err := reg.Parse("$[1]")
	if !assert.NoError(t, err) {
		return
	}
	run, err := reg.Compile(ast)
	if !assert.NoError(t, err) {
		return
	}
	got := run.Run([]any{float64(10), float64(20), float64(30)})
	assert.Equal(t, []any{float64(20)}, got)
}

func TestTopLevelWrappers(t *testing.T) {
	ast, err := jpath.Parse("$[1]")
	if !assert.NoError(t, err) {
		return
	}
	run, err := jpath.Compile(ast)
	if !assert.NoError(t, err) {
		return
	}
	got := run.Run([]any{float64(1), float64(2)})
	assert.Equal(t, []any{float64(2)}, got)

	path, err := jpath.Query("$[0]")
	if !assert.NoError(t, err) {
		return
	}
	got = path.Query([]any{float64(9)})
	assert.Equal(t, []any{float64(9)}, got)

	_ = jpath.MustParse("$[0]")
	_ = jpath.MustCompile(ast)
	_ = jpath.MustQuery("$[0]")
}

func TestTopLevelWrapperPanics(t *testing.T) {
	assert.Panics(t, func() {
		_ = jpath.MustParse("")
	})
	assert.Panics(t, func() {
		_ = jpath.MustQuery("")
	})
}

func TestMustHelpers(t *testing.T) {
	reg := jpath.NewRegistry()
	_ = reg.MustParse("$[0]")

	ast := reg.MustParse("$[0]")
	run := reg.MustCompile(ast)
	got := run.Run([]any{float64(1)})
	assert.Equal(t, []any{float64(1)}, got)

	path := reg.MustQuery("$[0]")
	got = path.Query([]any{float64(7)})
	assert.Equal(t, []any{float64(7)}, got)
}

func TestMustHelpersPanic(t *testing.T) {
	reg := jpath.NewRegistry()

	assert.Panics(t, func() {
		_ = reg.MustParse("")
	})
	assert.Panics(t, func() {
		_ = reg.MustQuery("")
	})
	assert.Panics(t, func() {
		_ = reg.MustCompile(jpath.PathExpr{
			Segments: []jpath.SegmentExpr{{
				Selectors: []jpath.SelectorExpr{{
					Kind:   jpath.SelectorFilter,
					Filter: jpath.LiteralExpr{Value: true},
				}},
			}},
		})
	})
}

func TestQueryErrorWrapping(t *testing.T) {
	reg := jpath.NewRegistry()
	_, err := reg.Query("")
	assert.ErrorIs(t, err, jpath.ErrInvalidPath)
	assert.ErrorIs(t, err, jpath.ErrExpectedRoot)
}

func TestBadNumberError(t *testing.T) {
	_, err := jpath.Parse("$[?(@ == 01)]")
	assert.ErrorIs(t, err, jpath.ErrInvalidPath)
	assert.ErrorIs(t, err, jpath.ErrBadNumber)
}

func TestSliceSpecializations(t *testing.T) {
	reg := jpath.NewRegistry()
	doc := []any{
		float64(0), float64(1), float64(2), float64(3), float64(4),
	}

	assertQuery(t, reg, "$[:-1]", doc, []any{
		float64(0), float64(1), float64(2), float64(3),
	})
	assertQuery(t, reg, "$[-2::-1]", doc, []any{
		float64(3), float64(2), float64(1), float64(0),
	})
	assertQuery(t, reg, "$[:-3:-1]", doc, []any{
		float64(4), float64(3),
	})
}

func TestStringHelpers(t *testing.T) {
	reg := jpath.NewRegistry()
	ast := reg.MustParse("$[0]")
	assert.Equal(t, "path(segments=1)", ast.String())
	assert.True(t,
		strings.HasPrefix(
			jpath.OpSelectSliceB11PN.String(),
			"sel/slice/b11pn",
		),
	)
	assert.True(t, strings.HasPrefix(jpath.Opcode(255).String(), "Opcode("))
}

func TestRegistryExtensionFunction(t *testing.T) {
	reg := jpath.NewRegistry()
	err := reg.RegisterFunction("alwaysTrue", jpath.FunctionDefinition{
		Validate: func(
			args []jpath.FilterExpr, _ jpath.FunctionUse, _ bool,
		) error {
			if len(args) != 1 {
				return errors.New("invalid function arity")
			}
			return nil
		},
		Eval: func(_ []jpath.FunctionValue) jpath.FunctionValue {
			return jpath.FunctionValue{Scalar: true}
		},
	})
	if !assert.NoError(t, err) {
		return
	}

	path, err := reg.Query("$[?alwaysTrue(@)]")
	if !assert.NoError(t, err) {
		return
	}
	got := path.Query([]any{float64(1), float64(2)})
	assert.Equal(t, []any{float64(1), float64(2)}, got)
}

func TestRegistryFunctionIsolation(t *testing.T) {
	base := jpath.NewRegistry()
	sandbox := base.Clone()
	err := sandbox.RegisterFunction("alwaysFalse", jpath.FunctionDefinition{
		Validate: func(
			_ []jpath.FilterExpr, _ jpath.FunctionUse, _ bool,
		) error {
			return nil
		},
		Eval: func(_ []jpath.FunctionValue) jpath.FunctionValue {
			return jpath.FunctionValue{Scalar: false}
		},
	})
	if !assert.NoError(t, err) {
		return
	}

	_, err = base.Query("$[?alwaysFalse(@)]")
	assert.ErrorIs(t, err, jpath.ErrUnknownFunction)
	assert.ErrorIs(t, err, jpath.ErrInvalidPath)

	_, err = sandbox.Query("$[?alwaysFalse(@)]")
	assert.NoError(t, err)
}

func TestRegisterFunctionErrors(t *testing.T) {
	reg := jpath.NewRegistry()
	err := reg.RegisterFunction("1bad", jpath.FunctionDefinition{
		Eval: func(_ []jpath.FunctionValue) jpath.FunctionValue {
			return jpath.FunctionValue{}
		},
	})
	assert.ErrorIs(t, err, jpath.ErrBadFunctionName)

	err = reg.RegisterFunction("x", jpath.FunctionDefinition{})
	assert.ErrorIs(t, err, jpath.ErrBadFunctionDefinition)

	err = reg.RegisterFunction("dup", jpath.FunctionDefinition{
		Eval: func(_ []jpath.FunctionValue) jpath.FunctionValue {
			return jpath.FunctionValue{}
		},
	})
	if !assert.NoError(t, err) {
		return
	}

	err = reg.RegisterFunction("dup", jpath.FunctionDefinition{
		Eval: func(_ []jpath.FunctionValue) jpath.FunctionValue {
			return jpath.FunctionValue{}
		},
	})
	assert.ErrorIs(t, err, jpath.ErrFunctionExists)
}

func assertQuery(
	t *testing.T, reg *jpath.Registry, query string, doc any, want []any,
) {
	t.Helper()
	path, err := reg.Query(query)
	if !assert.NoError(t, err) {
		return
	}
	got := path.Query(doc)
	assert.Equal(t, want, got)
}
