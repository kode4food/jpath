package jpath_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kode4food/jpath"
)

type (
	ctsSuite struct {
		Tests []ctsCase `json:"tests"`
	}

	ctsCase struct {
		Name            string `json:"name"`
		Selector        string `json:"selector"`
		InvalidSelector bool   `json:"invalid_selector"`
	}
)

func TestNewCompiler(t *testing.T) {
	c := jpath.NewCompiler()
	assert.NotNil(t, c)
}

func TestCompileBytecode(t *testing.T) {
	ctsPath := filepath.Join(
		"testdata", "jsonpath-compliance-test-suite", "cts.json",
	)
	buf, err := os.ReadFile(ctsPath)
	if err != nil {
		t.Skipf("compliance suite unavailable at %s: %v", ctsPath, err)
	}
	var suite ctsSuite
	if !assert.NoError(t, json.Unmarshal(buf, &suite)) {
		return
	}
	reg := jpath.NewRegistry()
	for _, tc := range suite.Tests {
		if tc.InvalidSelector {
			continue
		}
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			ast, err := reg.Parse(tc.Selector)
			if !assert.NoError(t, err, tc.Selector) {
				return
			}
			path, err := reg.Compile(ast)
			if !assert.NoError(t, err, tc.Selector) {
				return
			}
			assertPathShape(t, path)
		})
	}
}

func assertPathShape(t *testing.T, path jpath.Path) {
	t.Helper()
	end := -1
	for pc, inst := range path.Code {
		switch inst.Op {
		case jpath.OpDescend:
			assert.Equal(t, -1, end)
		case jpath.OpSegmentStart:
			assert.Equal(t, -1, end)
			if !assert.Greater(t, inst.Arg, pc) {
				return
			}
			if !assert.Less(t, inst.Arg, len(path.Code)) {
				return
			}
			assert.Equal(t, jpath.OpSegmentEnd, path.Code[inst.Arg].Op)
			end = inst.Arg
		case jpath.OpSegmentEnd:
			assert.Equal(t, end, pc)
			end = -1
		default:
			if !assert.NotEqual(t, -1, end) {
				return
			}
			assertSelectorInst(t, path, inst)
		}
	}
	assert.Equal(t, -1, end)
}

func assertSelectorInst(
	t *testing.T, path jpath.Path, inst jpath.Instruction,
) {
	t.Helper()
	switch inst.Op {
	case jpath.OpSelectName:
		assertConstType[string](t, path, inst.Arg)
	case jpath.OpSelectIndex:
		assertConstType[int](t, path, inst.Arg)
	case jpath.OpSelectFilter:
		assertFilterType(t, path, inst.Arg)
	case jpath.OpSelectSliceF00, jpath.OpSelectSliceF10P:
		assertConstType[jpath.SlicePlan](t, path, inst.Arg)
	case jpath.OpSelectSliceF10N, jpath.OpSelectSliceF01P:
		assertConstType[jpath.SlicePlan](t, path, inst.Arg)
	case jpath.OpSelectSliceF01N, jpath.OpSelectSliceF11PP:
		assertConstType[jpath.SlicePlan](t, path, inst.Arg)
	case jpath.OpSelectSliceF11PN, jpath.OpSelectSliceF11NP:
		assertConstType[jpath.SlicePlan](t, path, inst.Arg)
	case jpath.OpSelectSliceF11NN, jpath.OpSelectSliceB00:
		assertConstType[jpath.SlicePlan](t, path, inst.Arg)
	case jpath.OpSelectSliceB10P, jpath.OpSelectSliceB10N:
		assertConstType[jpath.SlicePlan](t, path, inst.Arg)
	case jpath.OpSelectSliceB01P, jpath.OpSelectSliceB01N:
		assertConstType[jpath.SlicePlan](t, path, inst.Arg)
	case jpath.OpSelectSliceB11PP, jpath.OpSelectSliceB11PN:
		assertConstType[jpath.SlicePlan](t, path, inst.Arg)
	case jpath.OpSelectSliceB11NP, jpath.OpSelectSliceB11NN:
		assertConstType[jpath.SlicePlan](t, path, inst.Arg)
	case jpath.OpSelectWildcard, jpath.OpSelectArrayAll:
		assert.Zero(t, inst.Arg)
	case jpath.OpSelectSliceEmpty:
		assert.Zero(t, inst.Arg)
	default:
		assert.Failf(t, "unexpected opcode", "%v", inst.Op)
	}
}

func assertFilterType(t *testing.T, path jpath.Path, arg int) {
	t.Helper()
	if !assert.GreaterOrEqual(t, arg, 0) {
		return
	}
	if !assert.Less(t, arg, len(path.Constants)) {
		return
	}
	_, ok := path.Constants[arg].(jpath.FilterExpr)
	assert.True(t, ok)
}

func assertConstType[T any](t *testing.T, path jpath.Path, arg int) {
	t.Helper()
	if !assert.GreaterOrEqual(t, arg, 0) {
		return
	}
	if !assert.Less(t, arg, len(path.Constants)) {
		return
	}
	_, ok := path.Constants[arg].(T)
	assert.True(t, ok)
}
