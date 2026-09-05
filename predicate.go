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
	lookup := makeSingularLookup(expr.Path)
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
	left, right := scalarOperand(expr.Left), scalarOperand(expr.Right)
	if left != nil && right != nil {
		return func(ctx *FilterCtx) bool {
			lv, rv := left(ctx), right(ctx)
			return compareValues(op, &lv, &rv)
		}, nil
	}
	if literal, ok := expr.Left.(*LiteralExpr); ok {
		right, err := compileFilter(expr.Right, reg)
		if err != nil {
			return nil, err
		}
		value := ScalarValue(literal.Value)
		return func(ctx *FilterCtx) bool {
			return compareValues(op, value, right(ctx))
		}, nil
	}
	leftFilter, err := compileFilter(expr.Left, reg)
	if err != nil {
		return nil, err
	}
	if literal, ok := expr.Right.(*LiteralExpr); ok {
		value := ScalarValue(literal.Value)
		return func(ctx *FilterCtx) bool {
			return compareValues(op, leftFilter(ctx), value)
		}, nil
	}
	rightFilter, err := compileFilter(expr.Right, reg)
	if err != nil {
		return nil, err
	}
	return func(ctx *FilterCtx) bool {
		return compareValues(op, leftFilter(ctx), rightFilter(ctx))
	}, nil
}

func scalarOperand(expr FilterExpr) func(*FilterCtx) Value {
	switch v := expr.(type) {
	case *LiteralExpr:
		value := v.Value
		return func(_ *FilterCtx) Value {
			return Value{Scalar: value}
		}

	case *PathValueExpr:
		if !isSingularPath(v.Path) {
			return nil
		}
		lookup := makeContextLookup(v)
		return func(ctx *FilterCtx) Value {
			value, ok := lookup(ctx)
			// An absent node stays an empty node list, distinct from null.
			return Value{Scalar: value, IsNodes: !ok}
		}

	default:
		return nil
	}
}

func compareValues(op string, left, right *Value) bool {
	switch op {
	case OpEq:
		return compareValuesEq(left, right)

	case OpNe:
		return compareValuesNe(left, right)

	case OpLt:
		return compareValuesLt(left, right)

	case OpLte:
		return compareValuesLe(left, right)

	case OpGt:
		return compareValuesGt(left, right)

	case OpGte:
		return compareValuesGe(left, right)

	default:
		return false
	}
}
