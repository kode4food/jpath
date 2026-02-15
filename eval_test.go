package jpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestBuiltinEvalEdgeCases(t *testing.T) {
	res := evalCount(nil)
	_, ok := res.Scalar.(nothingType)
	assert.True(t, ok)

	res = evalCount([]FunctionValue{
		{Scalar: float64(3)},
	})
	_, ok = res.Scalar.(nothingType)
	assert.True(t, ok)

	res = evalCount([]FunctionValue{
		{IsNodes: true, Nodes: []any{float64(1), float64(2)}},
	})
	assert.Equal(t, float64(2), res.Scalar)

	res = evalValueFn(nil)
	_, ok = res.Scalar.(nothingType)
	assert.True(t, ok)

	res = evalValueFn([]FunctionValue{
		{IsNodes: true, Nodes: []any{float64(1), float64(2)}},
	})
	_, ok = res.Scalar.(nothingType)
	assert.True(t, ok)

	res = evalValueFn([]FunctionValue{
		{IsNodes: true, Nodes: []any{float64(9)}},
	})
	assert.Equal(t, float64(9), res.Scalar)

	_, ok = singularFunctionValue(FunctionValue{
		IsNodes: true,
		Nodes:   nil,
	})
	assert.False(t, ok)
}

func TestCompareValuesEmptyCases(t *testing.T) {
	empty := evalValue{kind: evalNodes, nodes: nil}
	nothingNode := evalValue{
		kind:  evalNodes,
		nodes: []any{nothingType{}},
	}
	valueNode := evalValue{
		kind:  evalNodes,
		nodes: []any{float64(1)},
	}

	assert.True(t, compareValues(empty, empty, "=="))
	assert.True(t, compareValues(empty, nothingNode, "=="))
	assert.False(t, compareValues(empty, valueNode, "=="))
	assert.True(t, compareValues(empty, valueNode, "!="))
	assert.False(t, compareValues(empty, valueNode, "<"))
}
