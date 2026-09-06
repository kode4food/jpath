package jpath_test

import (
	"reflect"
	"testing"

	"github.com/kode4food/jpath"
)

func TestLogicalOptimization(t *testing.T) {
	reg := jpath.NewRegistry()
	queries := []string{
		"$[?(@.a && @.b) || (@.a && @.c)]",
		"$[?(@.a || @.b) && (@.a || @.c)]",
		"$[?@.a && (@.b && @.c)]",
		"$[?(@.a || @.b) || @.c]",
	}
	for bits := range 8 {
		a, b, c := bits&1 != 0, bits&2 != 0, bits&4 != 0
		node := map[string]any{}
		for key, exists := range map[string]bool{"a": a, "b": b, "c": c} {
			if exists {
				node[key] = nil
			}
		}
		want := []bool{a && (b || c), a || (b && c), a && b && c, a || b || c}
		for i, query := range queries {
			ast := reg.MustParse(query)
			before := reg.MustParse(query)
			path := reg.MustCompile(ast)
			if !reflect.DeepEqual(ast, before) {
				t.Fatalf("compile mutated AST: %s", query)
			}
			got := path([]any{node})
			if (len(got) == 1) != want[i] {
				t.Fatalf("%s, bits %d: got %v", query, bits, got)
			}
		}
	}
}

func TestLogicalOptimizationPreservesEffects(t *testing.T) {
	reg := jpath.NewRegistry()
	calls := 0
	reg.MustRegisterFunction("flip", 0, func(...any) (any, bool) {
		calls++
		return calls%2 == 0, true
	})
	query := "$[?(flip() && @.b) || (flip() && @.c)]"
	got := reg.MustQuery(query, []any{map[string]any{"c": true}})
	if len(got) != 1 || calls != 2 {
		t.Fatalf("stateful factor: got %v, calls %d", got, calls)
	}

	reg.MustRegisterFunction("mutate", 1, func(args ...any) (any, bool) {
		delete(args[0].(map[string]any), "a")
		return false, true
	})
	query = "$[?(@.a && mutate(@)) || (@.a && @.c)]"
	node := map[string]any{"a": true, "c": true}
	if got := reg.MustQuery(query, []any{node}); len(got) != 0 {
		t.Fatalf("mutating branch: got %v", got)
	}

	calls = 0
	query = "$[?(@.items[?flip()] && @.b) || (@.items[?flip()] && @.c)]"
	node = map[string]any{"items": []any{1}, "c": true}
	got = reg.MustQuery(query, []any{node})
	if len(got) != 1 || calls != 2 {
		t.Fatalf("nested stateful factor: got %v, calls %d", got, calls)
	}
}

func TestLogicalOptimizationShortCircuit(t *testing.T) {
	for _, tc := range []struct {
		query string
		value bool
		want  []string
	}{
		{"$[?a() && (b() && c())]", false, []string{"a"}},
		{"$[?(a() && b()) && c()]", true, []string{"a", "b", "c"}},
		{"$[?a() || (b() || c())]", true, []string{"a"}},
		{"$[?(a() || b()) || c()]", false, []string{"a", "b", "c"}},
	} {
		reg := jpath.NewRegistry()
		var calls []string
		for _, name := range []string{"a", "b", "c"} {
			reg.MustRegisterFunction(name, 0, func(...any) (any, bool) {
				calls = append(calls, name)
				return tc.value, true
			})
		}
		reg.MustQuery(tc.query, []any{1})
		if !reflect.DeepEqual(calls, tc.want) {
			t.Fatalf("%s: calls %v, want %v", tc.query, calls, tc.want)
		}
	}
}

func TestCompositionCapturesFunctions(t *testing.T) {
	selectors := []jpath.SelectorFunc{
		jpath.SelectName("a"), jpath.SelectName("b"),
	}
	seg := jpath.ChildSegment(selectors...)
	selectors[0] = jpath.SelectName("missing")
	segments := []jpath.SegmentFunc{
		seg, func(in []any, _ any) []any { return in },
	}
	path := jpath.ComposePath(segments...)
	segments[1] = jpath.ChildSegment()
	node := map[string]any{"a": 1, "b": 2}
	if got := path(node); !reflect.DeepEqual(got, []any{1, 2}) {
		t.Fatalf("segment slice mutation changed path: %v", got)
	}
	got := jpath.ComposePath(seg)(node)
	if !reflect.DeepEqual(got, []any{1, 2}) {
		t.Fatalf("selector order or snapshot changed: %v", got)
	}
	if got := jpath.ComposePath()(node); !reflect.DeepEqual(got, []any{node}) {
		t.Fatalf("empty composition changed: %v", got)
	}
}

func TestLogicalOptimizationBooleanArgument(t *testing.T) {
	reg := jpath.NewRegistry()
	reg.MustRegisterFunction("boolean", 1, func(args ...any) (any, bool) {
		_, ok := args[0].(bool)
		return ok, true
	})
	query := "$[?boolean((@.a && @.b) || (@.a && @.c))]"
	for _, node := range []any{nil, map[string]any{"a": 1, "b": 2}} {
		if got := reg.MustQuery(query, []any{node}); len(got) != 1 {
			t.Fatalf("logical argument lost boolean type: %v", got)
		}
	}
	_, err := reg.Query("$[?@.a && unknown()]", []any{nil})
	if err == nil {
		t.Fatal("unreachable function was not validated")
	}
}

func BenchmarkLogicalOptimization(b *testing.B) {
	reg := jpath.NewRegistry()
	node := map[string]any{
		"a": map[string]any{"b": map[string]any{"c": true}},
		"z": true,
	}
	doc := []any{node}
	for _, query := range []string{
		"$[?(@.a.b.c && @.x) || (@.a.b.c && @.z)]",
		"$[?@.a && @.a.b && @.a.b.c && @.z]",
		"$.a.b.c",
		"$[?@.a.b.c == true]",
		"$[?((@.a.b.c && @.x) || (@.a.b.c && @.z)) || @.d]",
	} {
		b.Run(query, func(b *testing.B) {
			path := reg.MustCompile(reg.MustParse(query))
			input := any(doc)
			if query == "$.a.b.c" {
				input = node
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchmarkComplianceSink = path(input)
			}
		})
	}
}

func TestConstantFilterOptimization(t *testing.T) {
	reg := jpath.NewRegistry()
	doc := map[string]any{
		"n":    float64(1),
		"tags": []any{"a", "b"},
	}
	all := []any{float64(1), []any{"a", "b"}}
	for _, query := range []struct {
		text string
		want []any
	}{
		{"$[?$.n == 1]", all},
		{"$[?$.n == 2]", []any{}},
		{`$[?$.tags[?@ == "a"]]`, all},
		{`$[?$.tags[?@ == "z"]]`, []any{}},
		{`$[?$.n == 1 && $.tags[?@ == "b"]]`, all},
		{`$.tags[?@ == "a"]`, []any{"a"}},
		{"$.tags[?$.n == 1]", []any{"a", "b"}},
		{"$.n[?$.n == 1]", []any{}},
	} {
		got := reg.MustQuery(query.text, doc)
		if !reflect.DeepEqual(got, query.want) {
			t.Fatalf("%s: got %v, want %v", query.text, got, query.want)
		}
	}
}

func TestConstantFilterPreservesEffects(t *testing.T) {
	reg := jpath.NewRegistry()
	calls := 0
	reg.MustRegisterFunction("seen", 0, func(...any) (any, bool) {
		calls++
		return true, true
	})
	doc := map[string]any{"a": []any{1}, "x": 1, "y": 2, "z": 3}
	got := reg.MustQuery("$[?$.a[?seen()]]", doc)
	if len(got) != 4 || calls != 4 {
		t.Fatalf("nested function collapsed: got %v, calls %d", got, calls)
	}
}

func BenchmarkConstantFilter(b *testing.B) {
	reg := jpath.NewRegistry()
	doc := map[string]any{
		"tags":     []any{"domain:payments", "tier:gold", "market:europe"},
		"type":     "service",
		"handling": "standard",
	}
	for _, query := range []string{
		`$.tags[?@ == "tier:gold"]`,
		`$.tags[?@ == "tier:gold"] && $.tags[?@ == "domain:payments"]`,
		`$.type == "service" && $.handling == "standard"`,
	} {
		b.Run(query, func(b *testing.B) {
			path := reg.MustCompile(reg.MustParse(query))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchmarkComplianceSink = path(doc)
			}
		})
	}
}

func TestSingularPathOptimization(t *testing.T) {
	reg := jpath.NewRegistry()
	for _, tc := range []struct {
		query string
		doc   any
		want  []any
	}{
		{"$", nil, []any{nil}},
		{"$.a", map[string]any{"a": nil}, []any{nil}},
		{"$.a", map[string]any{}, []any{}},
		{"$.a.b", map[string]any{"a": nil}, []any{}},
		{"$.a", []any{1}, []any{}},
		{"$[0]", map[string]any{}, []any{}},
		{"$[0]", []any{}, []any{}},
		{"$[-1]", []any{1, nil}, []any{nil}},
		{"$[-3]", []any{1, 2}, []any{}},
		{"$[2]", []any{1, 2}, []any{}},
		{
			"$.a[-1].b",
			map[string]any{"a": []any{map[string]any{"b": 3}}},
			[]any{3},
		},
	} {
		got := reg.MustQuery(tc.query, tc.doc)
		if !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("%s: got %v, want %v", tc.query, got, tc.want)
		}
	}
	ast := reg.MustParse("$.a[0]")
	path := reg.MustCompile(ast)
	ast.Segments[0].Selectors[0].Name = "b"
	ast.Segments[1].Selectors[0].Index = 1
	doc := map[string]any{"a": []any{1, 2}, "b": []any{3, 4}}
	if got := path(doc); !reflect.DeepEqual(got, []any{1}) {
		t.Fatalf("AST mutation changed singular path: %v", got)
	}

	doc = map[string]any{"items": []any{map[string]any{"x": nil}}}
	query := "$.items[?count(@.x) == 1 && count(@.missing) == 0]"
	if got := reg.MustQuery(query, doc); len(got) != 1 {
		t.Fatalf("singular function arguments lost node-list type: %v", got)
	}
	query = "$.items[?$.items[0].x == null]"
	if got := reg.MustQuery(query, doc); len(got) != 1 {
		t.Fatalf("absolute singular lookup lost root: %v", got)
	}
}

func TestLiteralComparisonOptimization(t *testing.T) {
	reg := jpath.NewRegistry()
	doc := []any{map[string]any{}}
	for _, value := range []any{
		nil, false, true, 2, float64(2), float64(3), "x", "y",
		[]any{}, map[string]any{},
	} {
		doc = append(doc, map[string]any{"x": value})
	}
	operand := jpath.PathCurrent(
		jpath.ComposePath(jpath.ChildSegment(jpath.SelectName("x"))),
	)
	for _, op := range []struct {
		name string
		make func(jpath.FilterFunc, jpath.FilterFunc) jpath.FilterFunc
	}{
		{jpath.OpEq, jpath.Eq}, {jpath.OpNe, jpath.Ne},
		{jpath.OpLt, jpath.Lt}, {jpath.OpLte, jpath.Le},
		{jpath.OpGt, jpath.Gt}, {jpath.OpGte, jpath.Ge},
	} {
		for _, literal := range []struct {
			text  string
			value any
		}{
			{"null", nil}, {"2", float64(2)}, {`"x"`, "x"}, {"true", true},
		} {
			for _, left := range []bool{false, true} {
				query := "$[?@.x " + op.name + " " + literal.text + "]"
				lhs, rhs := operand, jpath.Literal(literal.value)
				if left {
					query = "$[?" + literal.text + " " + op.name + " @.x]"
					lhs, rhs = rhs, lhs
				}
				reference := jpath.ComposePath(
					jpath.ChildSegment(jpath.SelectFilter(op.make(lhs, rhs))),
				)
				got, want := reg.MustQuery(query, doc), reference(doc)
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("%s: got %v, want %v", query, got, want)
				}
			}
		}
	}
	calls := 0
	reg.MustRegisterFunction("missing", 0, func(...any) (any, bool) {
		calls++
		return nil, false
	})
	for _, query := range []string{
		"$[?missing() != null]", "$[?null != missing()]",
	} {
		if got := reg.MustQuery(query, []any{1}); len(got) != 1 {
			t.Fatalf("Nothing compared as null: %s", query)
		}
	}
	if calls != 2 {
		t.Fatalf("literal comparison changed calls: %d", calls)
	}
}

func TestNestedFactoringOptimization(t *testing.T) {
	reg := jpath.NewRegistry()
	for _, pair := range [][2]string{
		{
			"$[?((@.a && @.b) || (@.a && @.c)) || @.d]",
			"$[?( @.a && (@.b || @.c)) || @.d]",
		},
		{
			"$[?@.d || ((@.a && @.b) || (@.a && @.c))]",
			"$[?@.d || (@.a && (@.b || @.c))]",
		},
		{
			"$[?((@.a || @.b) && (@.a || @.c)) && @.d]",
			"$[?(@.a || (@.b && @.c)) && @.d]",
		},
	} {
		ast := reg.MustParse(pair[0])
		optimized := reg.MustCompile(ast)
		reference := reg.MustCompile(reg.MustParse(pair[1]))
		if !reflect.DeepEqual(ast, reg.MustParse(pair[0])) {
			t.Fatal("nested factoring mutated the AST")
		}
		for bits := range 16 {
			node := map[string]any{}
			for i, key := range []string{"a", "b", "c", "d"} {
				if bits&(1<<i) != 0 {
					node[key] = nil
				}
			}
			doc := []any{node}
			if !reflect.DeepEqual(optimized(doc), reference(doc)) {
				t.Fatalf("%s, bits %d: results differ", pair[0], bits)
			}
		}
		doc := any([]any{map[string]any{"a": nil, "c": nil, "d": nil}})
		got := testing.AllocsPerRun(100, func() {
			benchmarkComplianceSink = optimized(doc)
		})
		want := testing.AllocsPerRun(100, func() {
			benchmarkComplianceSink = reference(doc)
		})
		if got != want {
			t.Fatalf("%s: allocations %g, factored %g", pair[0], got, want)
		}
	}
}
