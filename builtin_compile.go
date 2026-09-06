package jpath

func compileBuiltin(expr *FuncExpr, reg *Registry) (FilterFunc, error) {
	switch expr.Name {
	case "match", "search":
		return compileMatch(expr, reg)

	case "count":
		return compileCount(expr.Args[0].(*PathValueExpr), reg)

	case "length", "value":
		arg, err := compileOperand(expr.Args[0], reg)
		if err != nil {
			return nil, err
		}
		isLength := expr.Name == "length"
		return func(ctx *FilterCtx) any {
			value, ok := singularValue(arg(ctx))
			if !ok {
				return Nodes(nil)
			}
			if isLength {
				return evalLengthValue(value)
			}
			return value
		}, nil

	default:
		return nil, nil
	}
}

func compileCount(expr *PathValueExpr, reg *Registry) (FilterFunc, error) {
	if isSingularPath(expr.Path) {
		lookup := makeContextLookup(expr)
		return func(ctx *FilterCtx) any {
			if _, ok := lookup(ctx); ok {
				return float64(1)
			}
			return float64(0)
		}, nil
	}
	path, err := makePath(expr.Path, reg)
	if err != nil {
		return nil, err
	}
	absolute := expr.Absolute
	return func(ctx *FilterCtx) any {
		node := ctx.Current
		if absolute {
			node = ctx.Root
		}
		return float64(len(path(node)))
	}, nil
}

func compileMatch(expr *FuncExpr, reg *Registry) (FilterFunc, error) {
	if literal, ok := expr.Args[1].(*LiteralExpr); ok {
		if pattern, ok := literal.Value.(string); ok {
			return compileLiteralMatch(pattern, expr, reg)
		}
	}
	left, err := compileOperand(expr.Args[0], reg)
	if err != nil {
		return nil, err
	}
	right, err := compileOperand(expr.Args[1], reg)
	if err != nil {
		return nil, err
	}
	matcher := reg.matcher
	fullMatch := expr.Name == "match"
	return func(ctx *FilterCtx) any {
		// Keep callback results live until both arguments have run
		args := [2]any{left(ctx), right(ctx)}
		return matcher.evalMatch(args[:], fullMatch)
	}, nil
}

func compileLiteralMatch(
	pattern string, expr *FuncExpr, reg *Registry,
) (FilterFunc, error) {
	text, err := compileOperand(expr.Args[0], reg)
	if err != nil {
		return nil, err
	}
	if expr.Name == "match" {
		pattern = "^(?:" + pattern + ")$"
	}
	re, err := reg.matcher.compile(pattern)
	if err != nil {
		return func(ctx *FilterCtx) any {
			// Invalid patterns still evaluate their text argument for effects
			text(ctx)
			return Nodes(nil)
		}, nil
	}
	return func(ctx *FilterCtx) any {
		value, ok := singularValue(text(ctx))
		if !ok {
			return Nodes(nil)
		}
		raw, ok := value.(string)
		if !ok {
			return Nodes(nil)
		}
		return re.MatchString(raw)
	}, nil
}
