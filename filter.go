package jpath

// Literal builds a filter function that returns a scalar literal value
func Literal(value any) FilterFunc {
	return func(_ *FilterCtx) any {
		return value
	}
}

// PathCurrent builds a filter function that queries from the current node
func PathCurrent(path Path) FilterFunc {
	return func(ctx *FilterCtx) any {
		return Nodes(path(ctx.Current))
	}
}

// PathRoot builds a filter function that queries from the root node
func PathRoot(path Path) FilterFunc {
	return func(ctx *FilterCtx) any {
		return Nodes(path(ctx.Root))
	}
}

// Not builds a negation filter function
func Not(expr FilterFunc) FilterFunc {
	return func(ctx *FilterCtx) any {
		return !toBool(expr(ctx))
	}
}

// And builds a logical-AND filter function with short-circuit behavior
func And(left, right FilterFunc) FilterFunc {
	return func(ctx *FilterCtx) any {
		leftValue := left(ctx)
		if !toBool(leftValue) {
			return false
		}
		return toBool(right(ctx))
	}
}

// Or builds a logical-OR filter function with short-circuit behavior
func Or(left, right FilterFunc) FilterFunc {
	return func(ctx *FilterCtx) any {
		leftValue := left(ctx)
		if toBool(leftValue) {
			return true
		}
		return toBool(right(ctx))
	}
}

// Eq builds an equality comparison filter function
func Eq(left, right FilterFunc) FilterFunc {
	return func(ctx *FilterCtx) any {
		return compareValues(
			OpEq, comparison{left: left(ctx), right: right(ctx)},
		)
	}
}

// Ne builds an inequality comparison filter function
func Ne(left, right FilterFunc) FilterFunc {
	return func(ctx *FilterCtx) any {
		return compareValues(
			OpNe, comparison{left: left(ctx), right: right(ctx)},
		)
	}
}

// Lt builds a less-than comparison filter function
func Lt(left, right FilterFunc) FilterFunc {
	return func(ctx *FilterCtx) any {
		return compareValues(
			OpLt, comparison{left: left(ctx), right: right(ctx)},
		)
	}
}

// Le builds a less-than-or-equal comparison filter function
func Le(left, right FilterFunc) FilterFunc {
	return func(ctx *FilterCtx) any {
		return compareValues(
			OpLte, comparison{left: left(ctx), right: right(ctx)},
		)
	}
}

// Gt builds a greater-than comparison filter function
func Gt(left, right FilterFunc) FilterFunc {
	return func(ctx *FilterCtx) any {
		return compareValues(
			OpGt, comparison{left: left(ctx), right: right(ctx)},
		)
	}
}

// Ge builds a greater-than-or-equal comparison filter function
func Ge(left, right FilterFunc) FilterFunc {
	return func(ctx *FilterCtx) any {
		return compareValues(
			OpGte, comparison{left: left(ctx), right: right(ctx)},
		)
	}
}

// Call builds a filter function from a function evaluator and arguments
func Call(evaluator Evaluator, args ...FilterFunc) FilterFunc {
	return func(ctx *FilterCtx) any {
		values := make([]any, len(args))
		for i, arg := range args {
			values[i] = arg(ctx)
		}
		return evaluator(values)
	}
}
