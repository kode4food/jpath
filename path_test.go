package jpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathBoundaryHelpers(t *testing.T) {
	assert.Equal(t, 0, forwardStartPos(-1, 3))
	assert.Equal(t, 3, forwardStartPos(9, 3))
	assert.Equal(t, 0, forwardStartNeg(-9, 3))
	assert.Equal(t, 3, forwardStartNeg(9, 3))

	assert.Equal(t, 0, forwardEndPos(-1, 3))
	assert.Equal(t, 3, forwardEndPos(9, 3))
	assert.Equal(t, 0, forwardEndNeg(-9, 3))
	assert.Equal(t, 3, forwardEndNeg(9, 3))

	assert.Equal(t, -1, backwardStartPos(-1, 3))
	assert.Equal(t, 2, backwardStartPos(9, 3))
	assert.Equal(t, 2, backwardStartNeg(9, 3))

	assert.Equal(t, -1, backwardEndPos(-2, 3))
	assert.Equal(t, 2, backwardEndPos(9, 3))
	assert.Equal(t, -1, backwardEndNeg(-9, 3))
	assert.Equal(t, 2, backwardEndNeg(9, 3))
}

func TestSelectNodeFallbackCases(t *testing.T) {
	path := &Path{
		Constants: []any{
			"name",
			0,
			SlicePlan{Start: 0, End: 1, Step: 1},
			LiteralExpr{Value: true},
		},
	}
	base := []any{"x"}

	assert.Equal(
		t,
		base,
		path.selectNode(base, 1, nil, Instruction{Op: OpSelectName, Arg: 0}),
	)
	assert.Equal(
		t,
		base,
		path.selectNode(base, map[string]any{}, nil, Instruction{
			Op:  OpSelectName,
			Arg: 0,
		}),
	)
	assert.Equal(
		t,
		base,
		path.selectNode(base, map[string]any{}, nil, Instruction{
			Op:  OpSelectIndex,
			Arg: 1,
		}),
	)
	assert.Equal(
		t,
		base,
		path.selectNode(
			base,
			map[string]any{},
			nil,
			Instruction{Op: OpSelectArrayAll},
		),
	)

	sliceOps := []Opcode{
		OpSelectSliceF00,
		OpSelectSliceF10P,
		OpSelectSliceF10N,
		OpSelectSliceF01P,
		OpSelectSliceF01N,
		OpSelectSliceF11PP,
		OpSelectSliceF11PN,
		OpSelectSliceF11NP,
		OpSelectSliceF11NN,
		OpSelectSliceB00,
		OpSelectSliceB10P,
		OpSelectSliceB10N,
		OpSelectSliceB01P,
		OpSelectSliceB01N,
		OpSelectSliceB11PP,
		OpSelectSliceB11PN,
		OpSelectSliceB11NP,
		OpSelectSliceB11NN,
	}
	for _, op := range sliceOps {
		assert.Equal(
			t,
			base,
			path.selectNode(base, map[string]any{}, nil, Instruction{
				Op:  op,
				Arg: 2,
			}),
		)
	}

	assert.Equal(
		t,
		base,
		path.selectNode(base, 1, nil, Instruction{Op: OpSelectFilter, Arg: 3}),
	)
	assert.Equal(
		t,
		base,
		path.selectNode(base, 1, nil, Instruction{Op: Opcode(255), Arg: 0}),
	)
}
