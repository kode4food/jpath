package jpath

import "reflect"

func makeSingularPath(path *PathExpr) Path {
	lookup := makeSingularLookup(path.Segments)
	return func(node any) []any {
		value, ok := lookup(node)
		if !ok {
			return []any{}
		}
		return []any{value}
	}
}

func makeSingularSegment(segments []*SegmentExpr) SegmentFunc {
	lookup := makeSingularLookup(segments)
	return func(in []any, _ any) []any {
		out := make([]any, 0, len(in))
		for _, node := range in {
			if value, ok := lookup(node); ok {
				out = append(out, value)
			}
		}
		return out
	}
}

func makeSingularLookup(segments []*SegmentExpr) func(any) (any, bool) {
	// Copy selectors so later AST edits do not change the compiled path
	selectors := make([]SelectorExpr, len(segments))
	for i, seg := range segments {
		selectors[i] = *seg.Selectors[0]
	}
	return func(node any) (any, bool) {
		for _, selector := range selectors {
			switch selector.Kind {
			case SelectorName:
				obj, ok := node.(map[string]any)
				if !ok {
					return nil, false
				}
				node, ok = obj[selector.Name]
				if !ok {
					return nil, false
				}

			case SelectorIndex:
				arr, ok := node.([]any)
				if !ok {
					return nil, false
				}
				i := normalizeIndex(len(arr), selector.Index)
				if i < 0 || i >= len(arr) {
					return nil, false
				}
				node = arr[i]
			}
		}
		return node, true
	}
}

func compileLogical(expr *BinaryExpr, reg *Registry) (predicateFunc, error) {
	expr = factorLogical(expr)
	args := logicalArgs(nil, expr, expr.Op)
	compiled := make([]predicateFunc, len(args))
	for i, arg := range args {
		fn, err := compilePredicate(arg, reg)
		if err != nil {
			return nil, err
		}
		compiled[i] = fn
	}
	and := expr.Op == OpAnd
	return func(ctx *FilterCtx) bool {
		for _, fn := range compiled {
			if fn(ctx) != and {
				return !and
			}
		}
		return and
	}, nil
}

func logicalArgs(args []FilterExpr, expr FilterExpr, op string) []FilterExpr {
	b, ok := expr.(*BinaryExpr)
	if !ok || b.Op != op {
		return append(args, expr)
	}
	// Factor each subtree before flattening can erase its grouping
	if factored := factorLogical(b); factored != b {
		return append(args, factored)
	}
	args = logicalArgs(args, b.Left, op)
	return logicalArgs(args, b.Right, op)
}

func factorLogical(expr *BinaryExpr) *BinaryExpr {
	left, ok := expr.Left.(*BinaryExpr)
	if !ok {
		return expr
	}
	right, ok := expr.Right.(*BinaryExpr)
	if !ok || left.Op != right.Op {
		return expr
	}
	inner := OpAnd
	if expr.Op == OpAnd {
		inner = OpOr
	}
	if left.Op != inner || !reflect.DeepEqual(left.Left, right.Left) {
		return expr
	}
	// A branch can mutate data read by the repeated factor. Exclude all
	// function calls, including calls inside nested path filters
	if !functionFree(expr) {
		return expr
	}
	return &BinaryExpr{
		Op:    inner,
		Left:  left.Left,
		Right: &BinaryExpr{Op: expr.Op, Left: left.Right, Right: right.Right},
	}
}

func functionFree(expr FilterExpr) bool {
	switch v := expr.(type) {
	case *LiteralExpr:
		return true

	case *PathValueExpr:
		for _, seg := range v.Path.Segments {
			if !segmentFunctionFree(seg) {
				return false
			}
		}
		return true

	case *UnaryExpr:
		return functionFree(v.Expr)

	case *BinaryExpr:
		return functionFree(v.Left) && functionFree(v.Right)

	default:
		return false
	}
}

func segmentFunctionFree(seg *SegmentExpr) bool {
	for _, selector := range seg.Selectors {
		if selector.Filter != nil && !functionFree(selector.Filter) {
			return false
		}
	}
	return true
}
