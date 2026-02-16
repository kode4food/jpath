package jpath_test

import (
	"testing"

	"github.com/kode4food/jpath"
)

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
