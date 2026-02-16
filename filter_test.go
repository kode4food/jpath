package jpath_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kode4food/jpath"
)

func TestComposedFilterPathCurrent(t *testing.T) {
	doc := map[string]any{
		"items": []any{
			map[string]any{"price": float64(1)},
			map[string]any{"price": float64(3)},
			map[string]any{"price": float64(5)},
		},
	}
	items := doc["items"].([]any)

	reg := jpath.NewRegistry()
	price := reg.MustCompile(reg.MustParse("$.price"))
	filter := jpath.Gt(jpath.PathCurrent(price), jpath.Literal(float64(2)))
	path := jpath.ComposePath(
		jpath.ChildSegment(jpath.SelectName("items")),
		jpath.ChildSegment(jpath.SelectFilter(filter)),
	)

	assert.Equal(t, []any{items[1], items[2]}, path(doc))
}

func TestComposedFilterPathRoot(t *testing.T) {
	doc := map[string]any{
		"limit": float64(4),
		"items": []any{
			map[string]any{"price": float64(1)},
			map[string]any{"price": float64(5)},
		},
	}
	items := doc["items"].([]any)

	reg := jpath.NewRegistry()
	price := reg.MustCompile(reg.MustParse("$.price"))
	limit := reg.MustCompile(reg.MustParse("$.limit"))
	filter := jpath.Le(jpath.PathCurrent(price), jpath.PathRoot(limit))
	path := jpath.ComposePath(
		jpath.ChildSegment(jpath.SelectName("items")),
		jpath.ChildSegment(jpath.SelectFilter(filter)),
	)

	assert.Equal(t, []any{items[0]}, path(doc))
}

func TestComposedFilterUsesContext(t *testing.T) {
	doc := map[string]any{
		"kind": "book",
		"items": []any{
			map[string]any{"kind": "book"},
			map[string]any{"kind": "video"},
		},
	}
	items := doc["items"].([]any)

	filter := func(ctx *jpath.FilterCtx) *jpath.Value {
		root := ctx.Root.(map[string]any)
		current := ctx.Current.(map[string]any)
		return jpath.ScalarValue(root["kind"] == current["kind"])
	}
	path := jpath.ComposePath(
		jpath.ChildSegment(jpath.SelectName("items")),
		jpath.ChildSegment(jpath.SelectFilter(filter)),
	)

	assert.Equal(t, []any{items[0]}, path(doc))
}

func TestComposedFilterCall(t *testing.T) {
	doc := map[string]any{
		"items": []any{
			map[string]any{"tag": "keep"},
			map[string]any{"tag": "drop"},
		},
	}
	items := doc["items"].([]any)

	reg := jpath.NewRegistry()
	tagPath := reg.MustCompile(reg.MustParse("$.tag"))
	filter := jpath.Call(
		func(args []*jpath.Value) *jpath.Value {
			if len(args) != 1 || !args[0].IsNodes {
				return &jpath.Value{Scalar: false}
			}
			if len(args[0].Nodes) != 1 {
				return &jpath.Value{Scalar: false}
			}
			tag, ok := args[0].Nodes[0].(string)
			if !ok {
				return &jpath.Value{Scalar: false}
			}
			return &jpath.Value{Scalar: tag == "keep"}
		},
		jpath.PathCurrent(tagPath),
	)
	path := jpath.ComposePath(
		jpath.ChildSegment(jpath.SelectName("items")),
		jpath.ChildSegment(jpath.SelectFilter(filter)),
	)

	assert.Equal(t, []any{items[0]}, path(doc))
}

func TestComposedFilterWrappedFunction(t *testing.T) {
	doc := map[string]any{
		"items": []any{
			map[string]any{"tag": "keep"},
			map[string]any{"tag": "drop"},
		},
	}
	items := doc["items"].([]any)

	reg := jpath.NewRegistry()
	tagPath := reg.MustCompile(reg.MustParse("$.tag"))
	filter := jpath.Call(
		jpath.WrapFunction(func(args ...any) (any, bool) {
			tag, ok := args[0].(string)
			if !ok {
				return nil, false
			}
			return tag == "keep", true
		}),
		jpath.PathCurrent(tagPath),
	)
	path := jpath.ComposePath(
		jpath.ChildSegment(jpath.SelectName("items")),
		jpath.ChildSegment(jpath.SelectFilter(filter)),
	)

	assert.Equal(t, []any{items[0]}, path(doc))
}
