package jpath_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kode4food/jpath"
)

func TestPathSelectorMisses(t *testing.T) {
	consts := []any{
		"name",
		0,
		&jpath.SlicePlan{Start: 0, End: 1, Step: 1},
		&jpath.LiteralExpr{Value: true},
	}
	run := func(op jpath.Opcode, arg int, node any) []any {
		p := &jpath.Path{
			Code: []jpath.Instruction{
				{Op: jpath.OpSegmentStart, Arg: 2},
				{Op: op, Arg: arg},
				{Op: jpath.OpSegmentEnd},
			},
			Constants: consts,
		}
		return p.Query(node)
	}

	assert.Equal(t, []any{}, run(jpath.OpSelectName, 0, 1))
	assert.Equal(t, []any{}, run(jpath.OpSelectName, 0, map[string]any{}))
	assert.Equal(t, []any{}, run(jpath.OpSelectIndex, 1, map[string]any{}))
	assert.Equal(t, []any{}, run(jpath.OpSelectArrayAll, 0, map[string]any{}))

	sliceOps := []jpath.Opcode{
		jpath.OpSelectSliceF00,
		jpath.OpSelectSliceF10P,
		jpath.OpSelectSliceF10N,
		jpath.OpSelectSliceF01P,
		jpath.OpSelectSliceF01N,
		jpath.OpSelectSliceF11PP,
		jpath.OpSelectSliceF11PN,
		jpath.OpSelectSliceF11NP,
		jpath.OpSelectSliceF11NN,
		jpath.OpSelectSliceB00,
		jpath.OpSelectSliceB10P,
		jpath.OpSelectSliceB10N,
		jpath.OpSelectSliceB01P,
		jpath.OpSelectSliceB01N,
		jpath.OpSelectSliceB11PP,
		jpath.OpSelectSliceB11PN,
		jpath.OpSelectSliceB11NP,
		jpath.OpSelectSliceB11NN,
	}
	for _, op := range sliceOps {
		assert.Equal(t, []any{}, run(op, 2, map[string]any{}))
	}

	assert.Equal(t, []any{}, run(jpath.OpSelectFilter, 3, 1))
	assert.Panics(t, func() { run(jpath.Opcode(255), 0, 1) })
}

func TestPathInvalidProgramPanics(t *testing.T) {
	consts := []any{
		"name",
		0,
		&jpath.SlicePlan{Start: 0, End: 1, Step: 1},
		&jpath.LiteralExpr{Value: true},
	}
	ops := []jpath.Opcode{
		jpath.OpSegmentEnd,
		jpath.OpSelectName,
		jpath.OpSelectIndex,
		jpath.OpSelectWildcard,
		jpath.OpSelectArrayAll,
		jpath.OpSelectSliceF00,
		jpath.OpSelectSliceF10P,
		jpath.OpSelectSliceF10N,
		jpath.OpSelectSliceF01P,
		jpath.OpSelectSliceF01N,
		jpath.OpSelectSliceF11PP,
		jpath.OpSelectSliceF11PN,
		jpath.OpSelectSliceF11NP,
		jpath.OpSelectSliceF11NN,
		jpath.OpSelectSliceB00,
		jpath.OpSelectSliceB10P,
		jpath.OpSelectSliceB10N,
		jpath.OpSelectSliceB01P,
		jpath.OpSelectSliceB01N,
		jpath.OpSelectSliceB11PP,
		jpath.OpSelectSliceB11PN,
		jpath.OpSelectSliceB11NP,
		jpath.OpSelectSliceB11NN,
		jpath.OpSelectSliceEmpty,
		jpath.OpSelectFilter,
	}
	for _, op := range ops {
		p := &jpath.Path{
			Code:      []jpath.Instruction{{Op: op, Arg: 0}},
			Constants: consts,
		}
		assert.Panics(t, func() {
			p.Query(map[string]any{"name": float64(1)})
		})
	}
}

func TestPathSliceBounds(t *testing.T) {
	reg := jpath.NewRegistry()
	doc := []any{
		float64(0),
		float64(1),
		float64(2),
		float64(3),
		float64(4),
	}

	assertPathQuery(t, reg, "$[-9:2]", doc, []any{float64(0), float64(1)})
	assertPathQuery(t, reg, "$[9:]", doc, []any{})
	assertPathQuery(
		t, reg, "$[2:9]", doc,
		[]any{float64(2), float64(3), float64(4)},
	)
	assertPathQuery(t, reg, "$[:-9]", doc, []any{})
	assertPathQuery(t, reg, "$[-2::-1]", doc, []any{
		float64(3),
		float64(2),
		float64(1),
		float64(0),
	})
	assertPathQuery(t, reg, "$[-9::-1]", doc, []any{})
	assertPathQuery(t, reg, "$[:9:-1]", doc, []any{})
	assertPathQuery(t, reg, "$[9:1:-1]", doc, []any{
		float64(4),
		float64(3),
		float64(2),
	})
}

func assertPathQuery(
	t *testing.T, reg *jpath.Registry, query string, doc any, want []any,
) {
	t.Helper()
	got, err := reg.Query(query, doc)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, want, got)
}
