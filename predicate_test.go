package jpath_test

import (
	"reflect"
	"testing"

	"github.com/kode4food/jpath"
)

func TestPredicateComparisons(t *testing.T) {
	reg := jpath.NewRegistry()
	values := []any{
		nil, false, true, 1, float64(1), "x", "y",
		[]any{1}, map[string]any{"a": 1},
	}
	// Include absence independently on each side, as well as JSON null
	var doc []any
	for i := -1; i < len(values); i++ {
		for j := -1; j < len(values); j++ {
			node := map[string]any{}
			if i >= 0 {
				node["a"] = values[i]
			}
			if j >= 0 {
				node["b"] = values[j]
			}
			doc = append(doc, node)
		}
	}
	left := jpath.PathCurrent(
		jpath.ComposePath(jpath.ChildSegment(jpath.SelectName("a"))),
	)
	right := jpath.PathCurrent(
		jpath.ComposePath(jpath.ChildSegment(jpath.SelectName("b"))),
	)
	for _, op := range []struct {
		name string
		make func(jpath.FilterFunc, jpath.FilterFunc) jpath.FilterFunc
	}{
		{name: jpath.OpEq, make: jpath.Eq},
		{name: jpath.OpNe, make: jpath.Ne},
		{name: jpath.OpLt, make: jpath.Lt},
		{name: jpath.OpLte, make: jpath.Le},
		{name: jpath.OpGt, make: jpath.Gt},
		{name: jpath.OpGte, make: jpath.Ge},
	} {
		query := "$[?@.a " + op.name + " @.b]"
		reference := jpath.ComposePath(
			jpath.ChildSegment(jpath.SelectFilter(op.make(left, right))),
		)
		got := reg.MustQuery(query, doc)
		want := reference(doc)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s: optimized and public evaluators differ", query)
		}
	}
}

func TestPredicatePreservesSharedFunctionResults(t *testing.T) {
	reg := jpath.NewRegistry()
	shared := jpath.Nodes{nil}
	var calls []string
	reg.MustRegisterDefinition("left", &jpath.FunctionDefinition{
		Eval: func([]any) any {
			calls = append(calls, "left")
			shared[0] = 1
			return shared
		},
	})
	reg.MustRegisterDefinition("right", &jpath.FunctionDefinition{
		Eval: func([]any) any {
			calls = append(calls, "right")
			shared[0] = 2
			return 2
		},
	})
	got := reg.MustQuery("$[?left() == right()]", []any{nil})
	if len(got) != 1 || !reflect.DeepEqual(calls, []string{"left", "right"}) {
		t.Fatalf("shared result or call order changed: %v, %v", got, calls)
	}
}

func TestPredicateArgumentsRemainIndependentBooleans(t *testing.T) {
	reg := jpath.NewRegistry()
	reg.MustRegisterDefinition("check", &jpath.FunctionDefinition{
		Eval: func(args []any) any {
			valid := len(args) == 2 && args[0] == true && args[1] == true
			args[0] = false
			return valid && args[1] == true
		},
	})
	query := "$[?check(!@.missing, @.a == null && !!@.a)]"
	path := reg.MustCompile(reg.MustParse(query))
	doc := []any{map[string]any{"a": nil}}
	for range 2 {
		if got := path(doc); len(got) != 1 {
			t.Fatalf("predicate argument value was shared or changed: %v", got)
		}
	}
}

func TestPredicateExistence(t *testing.T) {
	reg := jpath.NewRegistry()
	doc := []any{
		map[string]any{}, map[string]any{"a": nil},
		map[string]any{"a": false}, map[string]any{"a": []any{}},
	}
	for _, tc := range []struct {
		query string
		want  []any
	}{
		{"$[?@.a]", doc[1:]},
		{"$[?!@.a]", doc[:1]},
		{"$[?!!@.a]", doc[1:]},
		{"$[?@.a[*]]", []any{}},
	} {
		got := reg.MustQuery(tc.query, doc)
		if !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("%s: got %v, want %v", tc.query, got, tc.want)
		}
	}
}
