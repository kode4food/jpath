package jpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCompiler(t *testing.T) {
	c := NewCompiler()
	assert.NotNil(t, c)
}

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

func TestPathValueExprEval(t *testing.T) {
	ctx := &evalCtx{
		root: map[string]any{
			"x": []any{float64(10), float64(20)},
		},
		current: map[string]any{
			"x": []any{float64(1), float64(2)},
		},
	}
	path := PathExpr{
		Segments: []SegmentExpr{
			{
				Selectors: []SelectorExpr{{
					Kind: SelectorName,
					Name: "x",
				}},
			},
			{
				Selectors: []SelectorExpr{{
					Kind:  SelectorIndex,
					Index: 0,
				}},
			},
		},
	}

	relative := PathValueExpr{Path: path}
	relativeRes := relative.eval(ctx)
	assert.Equal(t, evalNodes, relativeRes.kind)
	assert.Equal(t, []any{float64(1)}, relativeRes.nodes)

	absolute := PathValueExpr{Absolute: true, Path: path}
	absoluteRes := absolute.eval(ctx)
	assert.Equal(t, evalNodes, absoluteRes.kind)
	assert.Equal(t, []any{float64(10)}, absoluteRes.nodes)
}

func TestPathValueExprEvalInvalidPath(t *testing.T) {
	ctx := &evalCtx{}
	bad := PathValueExpr{
		Path: PathExpr{
			Segments: []SegmentExpr{{
				Selectors: []SelectorExpr{{
					Kind:   SelectorFilter,
					Filter: LiteralExpr{Value: true},
				}},
			}},
		},
	}

	res := bad.eval(ctx)
	assert.Equal(t, evalNodes, res.kind)
	assert.Empty(t, res.nodes)
}

func TestFuncExprEval(t *testing.T) {
	ctx := &evalCtx{}

	known := FuncExpr{
		Name: "length",
		Args: []FilterExpr{
			LiteralExpr{Value: "abcd"},
		},
	}
	knownRes := known.eval(ctx)
	assert.Equal(t, evalScalar, knownRes.kind)
	assert.Equal(t, float64(4), knownRes.scalar)

	unknown := FuncExpr{Name: "notFound"}
	unknownRes := unknown.eval(ctx)
	assert.Equal(t, evalScalar, unknownRes.kind)
	assert.Nil(t, unknownRes.scalar)
}

func TestCompiledFuncExprEvalNodes(t *testing.T) {
	ex := compiledFuncExpr{
		evaluator: func(_ []FunctionValue) FunctionValue {
			return FunctionValue{
				IsNodes: true,
				Nodes:   []any{float64(7)},
			}
		},
	}

	res := ex.eval(&evalCtx{})
	assert.Equal(t, evalNodes, res.kind)
	assert.Equal(t, []any{float64(7)}, res.nodes)
}
