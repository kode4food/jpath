package jpath

// Literal builds a filter function that returns a scalar literal value
func Literal(value any) FilterFunc {
	return func(_ *FilterCtx) *Value {
		return ScalarValue(value)
	}
}

// PathCurrent builds a filter function that queries from the current node
func PathCurrent(path Path) FilterFunc {
	return func(ctx *FilterCtx) *Value {
		return NodesValue(path(ctx.Current))
	}
}

// PathRoot builds a filter function that queries from the root node
func PathRoot(path Path) FilterFunc {
	return func(ctx *FilterCtx) *Value {
		return NodesValue(path(ctx.Root))
	}
}

// Not builds a negation filter function
func Not(expr FilterFunc) FilterFunc {
	return func(ctx *FilterCtx) *Value {
		return ScalarValue(!toBool(expr(ctx)))
	}
}

// And builds a logical-AND filter function with short-circuit behavior
func And(left, right FilterFunc) FilterFunc {
	return func(ctx *FilterCtx) *Value {
		leftValue := left(ctx)
		if !toBool(leftValue) {
			return ScalarValue(false)
		}
		return ScalarValue(toBool(right(ctx)))
	}
}

// Or builds a logical-OR filter function with short-circuit behavior
func Or(left, right FilterFunc) FilterFunc {
	return func(ctx *FilterCtx) *Value {
		leftValue := left(ctx)
		if toBool(leftValue) {
			return ScalarValue(true)
		}
		return ScalarValue(toBool(right(ctx)))
	}
}

// Eq builds an equality comparison filter function
func Eq(left, right FilterFunc) FilterFunc {
	return func(ctx *FilterCtx) *Value {
		return ScalarValue(compareValuesEq(left(ctx), right(ctx)))
	}
}

// Ne builds an inequality comparison filter function
func Ne(left, right FilterFunc) FilterFunc {
	return func(ctx *FilterCtx) *Value {
		return ScalarValue(compareValuesNe(left(ctx), right(ctx)))
	}
}

// Lt builds a less-than comparison filter function
func Lt(left, right FilterFunc) FilterFunc {
	return func(ctx *FilterCtx) *Value {
		return ScalarValue(compareValuesLt(left(ctx), right(ctx)))
	}
}

// Le builds a less-than-or-equal comparison filter function
func Le(left, right FilterFunc) FilterFunc {
	return func(ctx *FilterCtx) *Value {
		return ScalarValue(compareValuesLe(left(ctx), right(ctx)))
	}
}

// Gt builds a greater-than comparison filter function
func Gt(left, right FilterFunc) FilterFunc {
	return func(ctx *FilterCtx) *Value {
		return ScalarValue(compareValuesGt(left(ctx), right(ctx)))
	}
}

// Ge builds a greater-than-or-equal comparison filter function
func Ge(left, right FilterFunc) FilterFunc {
	return func(ctx *FilterCtx) *Value {
		return ScalarValue(compareValuesGe(left(ctx), right(ctx)))
	}
}

// Call builds a filter function from a function evaluator and arguments
func Call(evaluator Evaluator, args ...FilterFunc) FilterFunc {
	return func(ctx *FilterCtx) *Value {
		return evaluator(evalFunctionArgs(args, ctx))
	}
}
