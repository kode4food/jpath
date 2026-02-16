package jpath_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kode4food/jpath"
)

func TestPathSelectorMisses(t *testing.T) {
	reg := jpath.NewRegistry()

	assertPathQuery(t, reg, "$.name", 1, []any{})
	assertPathQuery(t, reg, "$.name", map[string]any{}, []any{})
	assertPathQuery(t, reg, "$[0]", map[string]any{}, []any{})
	assertPathQuery(t, reg, "$[*]", map[string]any{}, []any{})
	assertPathQuery(t, reg, "$[?@ == 1]", 1, []any{})
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
