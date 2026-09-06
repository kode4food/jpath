package jpath_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kode4food/jpath"
)

func TestSingularRuns(t *testing.T) {
	reg := jpath.NewRegistry()
	doc := map[string]any{"items": []any{
		map[string]any{"author": map[string]any{"name": "a"}},
		map[string]any{"author": map[string]any{"name": nil}},
		map[string]any{"author": map[string]any{}},
	}}
	ast := reg.MustParse("$.items[*].author.name")
	path := reg.MustCompile(ast)
	ast.Segments[0].Selectors[0].Name = "changed"
	ast.Segments[2].Selectors[0].Name = "changed"
	assert.Equal(t, []any{"a", nil}, path(doc))
}

func TestDescendantTraversal(t *testing.T) {
	reg := jpath.NewRegistry()
	doc := map[string]any{
		"name":  "root",
		"child": map[string]any{"name": "child"},
	}
	reference := jpath.ComposePath(jpath.DescendantSegment(
		jpath.SelectName("name"),
	))
	assert.Equal(t, reference(doc), reg.MustQuery("$..name", doc))
}

func TestDescendantMutationSnapshot(t *testing.T) {
	reg := jpath.NewRegistry()
	child := map[string]any{"keep": true}
	children := []any{child}
	doc := map[string]any{"a": children}
	calls := 0
	reg.MustRegisterFunction("mutate", 1, func(args ...any) (any, bool) {
		calls++
		args[0].(map[string]any)["a"] = nil
		return true, true
	})
	assert.Equal(t,
		[]any{children, child, true},
		reg.MustQuery("$..[?mutate($)]", doc),
	)
	assert.Equal(t, 3, calls)
}

func BenchmarkTraversal(b *testing.B) {
	reg := jpath.NewRegistry()
	books := make([]any, 128)
	for i := range books {
		books[i] = map[string]any{
			"author": map[string]any{"name": "author"},
		}
	}
	doc := map[string]any{"store": map[string]any{"books": books}}
	for _, tc := range []struct {
		name  string
		query string
	}{
		{name: "singular-runs", query: "$.store.books[*].author.name"},
		{name: "descendant-name", query: "$..name"},
		{name: "descendant-wildcard", query: "$..*"},
	} {
		b.Run(tc.name, func(b *testing.B) {
			path := reg.MustCompile(reg.MustParse(tc.query))
			b.ReportAllocs()
			for b.Loop() {
				benchmarkComplianceSink = path(doc)
			}
		})
	}
}
