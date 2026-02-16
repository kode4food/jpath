package jpath_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kode4food/jpath"
)

func TestTopLevelParseCompileQuery(t *testing.T) {
	ast, err := jpath.Parse("$[1]")
	if !assert.NoError(t, err) {
		return
	}
	path, err := jpath.Compile(ast)
	if !assert.NoError(t, err) {
		return
	}
	got := path.Query([]any{float64(10), float64(20), float64(30)})
	assert.Equal(t, []any{float64(20)}, got)
}

func TestTopLevelWrappers(t *testing.T) {
	ast, err := jpath.Parse("$[1]")
	if !assert.NoError(t, err) {
		return
	}
	path, err := jpath.Compile(ast)
	if !assert.NoError(t, err) {
		return
	}
	got := path.Query([]any{float64(1), float64(2)})
	assert.Equal(t, []any{float64(2)}, got)

	got, err = jpath.Query("$[0]", []any{float64(9)})
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []any{float64(9)}, got)

	_ = jpath.MustParse("$[0]")
	_ = jpath.MustCompile(ast)
	_ = jpath.MustQuery("$[0]", []any{float64(1)})
}

func TestTopLevelWrapperPanics(t *testing.T) {
	assert.Panics(t, func() {
		_ = jpath.MustParse("")
	})
	assert.Panics(t, func() {
		_ = jpath.MustQuery("", nil)
	})
}

func TestBadNumberError(t *testing.T) {
	_, err := jpath.Parse("$[?(@ == 01)]")
	assert.ErrorIs(t, err, jpath.ErrInvalidPath)
	assert.ErrorIs(t, err, jpath.ErrBadNumber)
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
