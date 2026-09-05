package jpath_test

import (
	"reflect"
	"testing"

	"github.com/kode4food/jpath"
)

func TestPredicateComparisons(t *testing.T) {
	reg := jpath.NewRegistry()
	nothing := jpath.WrapFunction(func(...any) (any, bool) {
		return nil, false
	})(nil).Scalar
	values := []any{
		nil, nothing, false, true, 1, float64(1), "x", "y",
		[]any{1}, map[string]any{"a": 1},
	}
	// Include absence independently on each side, as well as JSON null.
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
		{jpath.OpEq, jpath.Eq}, {jpath.OpNe, jpath.Ne},
		{jpath.OpLt, jpath.Lt}, {jpath.OpLte, jpath.Le},
		{jpath.OpGt, jpath.Gt}, {jpath.OpGte, jpath.Ge},
	} {
		query := "$[?@.a " + op.name + " @.b]"
		reference := jpath.ComposePath(
			jpath.ChildSegment(jpath.SelectFilter(op.make(left, right))),
		)
		got, want := reg.MustQuery(query, doc), reference(doc)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s: optimized and public evaluators differ", query)
		}
	}
}

func TestPredicatePreservesSharedFunctionResults(t *testing.T) {
	reg := jpath.NewRegistry()
	shared := &jpath.Value{}
	var calls []string
	reg.MustRegisterDefinition("left", &jpath.FunctionDefinition{
		Eval: func([]*jpath.Value) *jpath.Value {
			calls = append(calls, "left")
			shared.Scalar = 1
			return shared
		},
	})
	reg.MustRegisterDefinition("right", &jpath.FunctionDefinition{
		Eval: func([]*jpath.Value) *jpath.Value {
			calls = append(calls, "right")
			shared.Scalar = 2
			return jpath.ScalarValue(2)
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
		Eval: func(args []*jpath.Value) *jpath.Value {
			valid := len(args) == 2 && !args[0].IsNodes && !args[1].IsNodes
			valid = valid && args[0].Scalar == true && args[1].Scalar == true
			args[0].Scalar = false
			valid = valid && args[1].Scalar == true
			return jpath.ScalarValue(valid)
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
