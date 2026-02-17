package jpath_test

import (
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
	got := path([]any{float64(10), float64(20), float64(30)})
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
	got := path([]any{float64(1), float64(2)})
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

func TestTopLevelFilterShorthand(t *testing.T) {
	explicit := `$[?($.product_info.name == "Professional Laptop")]`
	shorthand := `$.product_info.name == "Professional Laptop"`

	explicitAST, err := jpath.Parse(explicit)
	if !assert.NoError(t, err) {
		return
	}
	shorthandAST, err := jpath.Parse(shorthand)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, explicitAST, shorthandAST)

	docMatch := map[string]any{
		"product_info": map[string]any{
			"name": "Professional Laptop",
		},
	}
	gotExplicit, err := jpath.Query(explicit, docMatch)
	if !assert.NoError(t, err) {
		return
	}
	gotShorthand, err := jpath.Query(shorthand, docMatch)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, gotExplicit, gotShorthand)

	docNoMatch := map[string]any{
		"product_info": map[string]any{
			"name": "Budget Laptop",
		},
	}
	gotExplicit, err = jpath.Query(explicit, docNoMatch)
	if !assert.NoError(t, err) {
		return
	}
	gotShorthand, err = jpath.Query(shorthand, docNoMatch)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, gotExplicit, gotShorthand)
	assert.Empty(t, gotShorthand)
}

func TestTopLevelFilterShorthandCurrentNode(t *testing.T) {
	explicit := `$[?@.name == "Professional Laptop"]`
	shorthand := `@.name == "Professional Laptop"`

	explicitAST, err := jpath.Parse(explicit)
	if !assert.NoError(t, err) {
		return
	}
	shorthandAST, err := jpath.Parse(shorthand)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, explicitAST, shorthandAST)

	doc := []any{
		map[string]any{"name": "Professional Laptop"},
		map[string]any{"name": "Budget Laptop"},
	}
	gotExplicit, err := jpath.Query(explicit, doc)
	if !assert.NoError(t, err) {
		return
	}
	gotShorthand, err := jpath.Query(shorthand, doc)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, gotExplicit, gotShorthand)
}

func TestTopLevelFilterShorthandKeepsPathGrammar(t *testing.T) {
	ast, err := jpath.Parse("$.product_info.name")
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Len(t, ast.Segments, 2) {
		return
	}
	assert.Equal(t, jpath.SelectorName, ast.Segments[0].Selectors[0].Kind)
	assert.Equal(t, jpath.SelectorName, ast.Segments[1].Selectors[0].Kind)
}

func TestBadNumberError(t *testing.T) {
	_, err := jpath.Parse("$[?(@ == 01)]")
	assert.ErrorIs(t, err, jpath.ErrInvalidPath)
	assert.ErrorIs(t, err, jpath.ErrBadNumber)
}
