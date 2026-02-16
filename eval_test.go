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

func TestFilterCompareOperators(t *testing.T) {
	doc := []any{
		float64(1), float64(2), float64(3),
	}
	reg := jpath.NewRegistry()

	got, err := reg.Query("$[?@ < 2]", doc)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []any{float64(1)}, got)

	got, err = reg.Query("$[?@ <= 2]", doc)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []any{float64(1), float64(2)}, got)

	got, err = reg.Query("$[?@ > 2]", doc)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []any{float64(3)}, got)

	got, err = reg.Query("$[?@ >= 2]", doc)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []any{float64(2), float64(3)}, got)

	got, err = reg.Query("$[?@ != 2]", doc)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []any{float64(1), float64(3)}, got)
}

func TestFilterMatchSearchEdgeCases(t *testing.T) {
	doc := []any{
		map[string]any{
			"left":  []any{"aa", "bb"},
			"right": []any{"xx", "yy"},
		},
	}
	reg := jpath.NewRegistry()

	got, err := reg.Query("$[?match(@.left[*], 'aa')]", doc)
	if !assert.NoError(t, err) {
		return
	}
	assert.Empty(t, got)

	got, err = reg.Query("$[?match('aa', @.right[*])]", doc)
	if !assert.NoError(t, err) {
		return
	}
	assert.Empty(t, got)

	got, err = reg.Query("$[?search('abc', '[')]", doc)
	if !assert.NoError(t, err) {
		return
	}
	assert.Empty(t, got)
}
