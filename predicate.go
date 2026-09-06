package jpath

import "fmt"

type predicateFunc func(*FilterCtx) bool

func selectPredicate(predicate predicateFunc) SelectorFunc {
	return func(out []any, node, root any) []any {
		return appendFilter(out, node, root, predicate)
	}
}

func compilePredicate(expr FilterExpr, reg *Registry) (predicateFunc, error) {
	switch v := expr.(type) {
	case *PathValueExpr:
		if isSingularPath(v.Path) {
			lookup := makeContextLookup(v)
			return func(ctx *FilterCtx) bool {
				_, ok := lookup(ctx)
				return ok
			}, nil
		}

	case *UnaryExpr:
		if v.Op != OpNot {
			return nil, fmt.Errorf("%w: unary %s", ErrUnknownOperator, v.Op)
		}
		predicate, err := compilePredicate(v.Expr, reg)
		if err != nil {
			return nil, err
		}
		return func(ctx *FilterCtx) bool {
			return !predicate(ctx)
		}, nil

	case *BinaryExpr:
		switch v.Op {
		case OpAnd, OpOr:
			return compileLogical(v, reg)

		case OpEq, OpNe, OpLt, OpLte, OpGt, OpGte:
			return compileComparison(v, reg)

		default:
			return nil, fmt.Errorf("%w: %s", ErrUnknownOperator, v.Op)
		}
	}
	filter, err := compileFilter(expr, reg)
	if err != nil {
		return nil, err
	}
	return func(ctx *FilterCtx) bool {
		return toBool(filter(ctx))
	}, nil
}

func makeContextLookup(expr *PathValueExpr) func(*FilterCtx) (any, bool) {
	lookup := makeSingularLookup(expr.Path.Segments)
	absolute := expr.Absolute
	return func(ctx *FilterCtx) (any, bool) {
		node := ctx.Current
		if absolute {
			node = ctx.Root
		}
		return lookup(node)
	}
}

func compileComparison(expr *BinaryExpr, reg *Registry) (predicateFunc, error) {
	op := expr.Op
	left, err := compileOperand(expr.Left, reg)
	if err != nil {
		return nil, err
	}
	right, err := compileOperand(expr.Right, reg)
	if err != nil {
		return nil, err
	}
	return func(ctx *FilterCtx) bool {
		return compareValues(op, comparison{left: left(ctx), right: right(ctx)})
	}, nil
}

func compileOperand(expr FilterExpr, reg *Registry) (FilterFunc, error) {
	switch v := expr.(type) {
	case *PathValueExpr:
		if isSingularPath(v.Path) {
			lookup := makeContextLookup(v)
			return func(ctx *FilterCtx) any {
				if value, ok := lookup(ctx); ok {
					return value
				}
				return Nodes(nil)
			}, nil
		}
	}
	return compileFilter(expr, reg)
}
