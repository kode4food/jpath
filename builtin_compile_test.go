package jpath_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kode4food/jpath"
)

func TestLiteralRegexPreparation(t *testing.T) {
	reg := jpath.NewRegistry()
	for _, tc := range []struct {
		name    string
		text    string
		pattern string
		match   bool
		search  bool
	}{
		{
			name:    "unicode",
			text:    "a😀b",
			pattern: "a.b",
			match:   true,
			search:  true,
		},
		{name: "cr", text: "a\rb", pattern: "a.b"},
		{name: "lf", text: "a\nb", pattern: "a.b"},
		{
			name:    "escaped",
			text:    "a.b",
			pattern: `a\.b`,
			match:   true,
			search:  true,
		},
		{
			name:    "class",
			text:    "a.b",
			pattern: "a[.]b",
			match:   true,
			search:  true,
		},
		{name: "partial", text: "xabx", pattern: "ab", search: true},
		{name: "invalid", text: "ab", pattern: "["},
	} {
		t.Run(tc.name, func(t *testing.T) {
			doc := []any{map[string]any{
				"text": tc.text, "pattern": tc.pattern,
			}}
			for _, name := range []string{"match", "search"} {
				want := tc.match
				if name == "search" {
					want = tc.search
				}
				prefix := "$[?" + name + "(@.text, "
				query := prefix + strconv.Quote(tc.pattern) + ")]"
				ast := reg.MustParse(query)
				path := reg.MustCompile(ast)
				call := ast.Segments[0].Selectors[0].Filter.(*jpath.FuncExpr)
				call.Args[1].(*jpath.LiteralExpr).Value = "changed"
				assert.Equal(t, want, len(path(doc)) == 1)
				assert.Equal(t,
					path(doc), reg.MustQuery(prefix+"@.pattern)]", doc),
				)
			}
		})
	}
}

func TestInvalidLiteralRegexPreservesEffects(t *testing.T) {
	reg := jpath.NewRegistry()
	calls := 0
	reg.MustRegisterFunction("text", 0, func(...any) (any, bool) {
		calls++
		return "ab", true
	})
	path := reg.MustCompile(reg.MustParse("$[?match(text(), '[')]"))
	assert.Empty(t, path([]any{nil}))
	assert.Equal(t, 1, calls)
}

func BenchmarkBuiltinEvaluation(b *testing.B) {
	reg := jpath.NewRegistry()
	doc := make([]any, 128)
	for i := range doc {
		doc[i] = map[string]any{
			"name": "alpha", "pattern": "a.*",
			"items": []any{float64(i), "beta"},
		}
	}
	for _, tc := range []struct {
		name  string
		query string
	}{
		{name: "length", query: "$[?length(@.name) == 5]"},
		{name: "count-singular", query: "$[?count(@.name) == 1]"},
		{name: "count-many", query: "$[?count(@.items[*]) == 2]"},
		{name: "value", query: "$[?value(@.name) == 'alpha']"},
		{name: "match", query: "$[?match(@.name, 'a.*')]"},
		{name: "search-literal", query: "$[?search(@.name, 'a.*')]"},
		{name: "search", query: "$[?search(@.name, @.pattern)]"},
		{name: "nested", query: "$[?length(value(@.name)) == 5]"},
	} {
		b.Run(tc.name, func(b *testing.B) {
			path := reg.MustCompile(reg.MustParse(tc.query))
			input := any(doc)
			b.ReportAllocs()
			for b.Loop() {
				benchmarkComplianceSink = path(input)
			}
		})
	}
}
