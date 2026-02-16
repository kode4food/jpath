package jpath_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kode4food/jpath"
)

func TestFilterPathValues(t *testing.T) {
	doc := map[string]any{
		"x": []any{float64(10), float64(20)},
		"items": []any{
			map[string]any{"x": []any{float64(1), float64(2)}},
			map[string]any{"x": []any{float64(3), float64(4)}},
		},
	}
	items := doc["items"].([]any)

	reg := jpath.NewRegistry()
	got, err := reg.Query("$.items[?@.x[0] == 1]", doc)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []any{items[0]}, got)

	got, err = reg.Query("$.items[?$.x[0] == 10]", doc)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, items, got)
}

func TestFilterBuiltinFunctions(t *testing.T) {
	doc := []any{
		map[string]any{
			"name": "abcd",
			"vals": []any{float64(1), float64(2)},
		},
		map[string]any{
			"name": "xy",
			"vals": []any{float64(9)},
		},
	}

	reg := jpath.NewRegistry()

	got, err := reg.Query("$[?length(@.name) == 4]", doc)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []any{doc[0]}, got)

	got, err = reg.Query("$[?count(@.vals[*]) == 2]", doc)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []any{doc[0]}, got)

	got, err = reg.Query("$[?value(@.name) == 'xy']", doc)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []any{doc[1]}, got)
}

func TestRawFilterPathValueEval(t *testing.T) {
	doc := []any{
		map[string]any{"x": []any{float64(1)}, "id": float64(1)},
		map[string]any{"id": float64(2)},
	}
	rel := &jpath.PathValueExpr{
		Path: &jpath.PathExpr{
			Segments: []*jpath.SegmentExpr{{
				Selectors: []*jpath.SelectorExpr{{
					Kind: jpath.SelectorName,
					Name: "x",
				}},
			}},
		},
	}
	got := runRawFilter(rel, doc)
	assert.Equal(t, []any{doc[0]}, got)

	abs := &jpath.PathValueExpr{
		Absolute: true,
		Path: &jpath.PathExpr{
			Segments: []*jpath.SegmentExpr{{
				Selectors: []*jpath.SelectorExpr{{
					Kind:  jpath.SelectorIndex,
					Index: 0,
				}},
			}},
		},
	}
	got = runRawFilter(abs, doc)
	assert.Equal(t, doc, got)
}

func TestRawFilterPathValueEvalInvalidPath(t *testing.T) {
	doc := []any{
		map[string]any{"x": []any{float64(1)}},
	}
	bad := &jpath.PathValueExpr{
		Path: &jpath.PathExpr{
			Segments: []*jpath.SegmentExpr{{
				Selectors: []*jpath.SelectorExpr{{
					Kind: jpath.SelectorKind(255),
				}},
			}},
		},
	}
	got := runRawFilter(bad, doc)
	assert.Empty(t, got)
}

func TestRawFilterFuncEvalBranches(t *testing.T) {
	doc := []any{
		map[string]any{"x": []any{float64(1)}},
	}

	got := runRawFilter(&jpath.FuncExpr{Name: "length"}, doc)
	assert.Empty(t, got)

	got = runRawFilter(&jpath.FuncExpr{Name: "count"}, doc)
	assert.Empty(t, got)

	got = runRawFilter(&jpath.FuncExpr{Name: "value"}, doc)
	assert.Empty(t, got)

	got = runRawFilter(
		&jpath.FuncExpr{
			Name: "count",
			Args: []jpath.FilterExpr{
				&jpath.LiteralExpr{Value: float64(1)},
			},
		},
		doc,
	)
	assert.Empty(t, got)

	got = runRawFilter(
		&jpath.FuncExpr{
			Name: "value",
			Args: []jpath.FilterExpr{
				&jpath.LiteralExpr{Value: true},
			},
		},
		doc,
	)
	assert.Equal(t, doc, got)

	got = runRawFilter(&jpath.FuncExpr{Name: "noSuchFunction"}, doc)
	assert.Empty(t, got)
}

func TestRawFilterUnaryUnknownOp(t *testing.T) {
	doc := []any{map[string]any{"x": float64(1)}}
	got := runRawFilter(
		&jpath.UnaryExpr{
			Op:   "~",
			Expr: &jpath.LiteralExpr{Value: true},
		},
		doc,
	)
	assert.Empty(t, got)
}

func runRawFilter(filter jpath.FilterExpr, doc any) []any {
	p := &jpath.Path{
		Code: []jpath.Instruction{
			{Op: jpath.OpSegmentStart, Arg: 2},
			{Op: jpath.OpSelectFilter, Arg: 0},
			{Op: jpath.OpSegmentEnd},
		},
		Constants: []any{filter},
	}
	return p.Query(doc)
}
