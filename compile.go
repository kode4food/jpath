package jpath

import "fmt"

// Compiler compiles parsed JSONPath syntax trees into runnable programs
type Compiler struct {
	registry *Registry
}

// NewCompiler creates a new Compiler
func NewCompiler() *Compiler {
	return &Compiler{}
}

// Compile compiles a parsed PathExpr into an executable Path
func (c *Compiler) Compile(path *PathExpr) (Path, error) {
	return compilePath(path, c.registry)
}

func compilePath(path *PathExpr, registry *Registry) (Path, error) {
	if err := validatePath(path, registry); err != nil {
		return nil, err
	}
	return makePath(path, registry)
}

func makePath(path *PathExpr, registry *Registry) (Path, error) {
	segments := make([]SegmentFunc, len(path.Segments))
	for idx, segment := range path.Segments {
		compiled, err := compileSegment(segment, registry)
		if err != nil {
			return nil, err
		}
		segments[idx] = compiled
	}
	return ComposePath(segments...), nil
}

func compileSegment(
	segment *SegmentExpr, registry *Registry,
) (SegmentFunc, error) {
	selectors := make([]SelectorFunc, len(segment.Selectors))
	for idx, selector := range segment.Selectors {
		compiled, err := compileSelector(selector, registry)
		if err != nil {
			return nil, err
		}
		selectors[idx] = compiled
	}
	if segment.Descendant {
		return DescendantSegment(selectors...), nil
	}
	return ChildSegment(selectors...), nil
}

func compileSelector(
	sel *SelectorExpr, registry *Registry,
) (SelectorFunc, error) {
	switch sel.Kind {
	case SelectorName:
		return SelectName(sel.Name), nil

	case SelectorIndex:
		return SelectIndex(sel.Index), nil

	case SelectorWildcard:
		return SelectWildcard(), nil

	case SelectorSlice:
		return SelectSlice(sel.Slice), nil

	case SelectorFilter:
		filter, err := compileFilter(sel.Filter, registry)
		if err != nil {
			return nil, err
		}
		return selectFilter(filter), nil

	default:
		return nil, fmt.Errorf("unknown selector kind")
	}
}

func compileFilter(
	expr FilterExpr, registry *Registry,
) (filterFunc, error) {
	switch v := expr.(type) {
	case *LiteralExpr:
		value := v.Value
		return func(_ *evalCtx) *evalValue {
			return scalarValue(value)
		}, nil

	case *PathValueExpr:
		path, err := makePath(v.Path, registry)
		if err != nil {
			return nil, err
		}
		absolute := v.Absolute
		return func(ctx *evalCtx) *evalValue {
			base := ctx.current
			if absolute {
				base = ctx.root
			}
			return nodesValue(path.Query(base))
		}, nil

	case *UnaryExpr:
		exprFunc, err := compileFilter(v.Expr, registry)
		if err != nil {
			return nil, err
		}
		if v.Op == "!" {
			return makeNot(exprFunc), nil
		}
		return nil, fmt.Errorf("unknown unary operator: %s", v.Op)

	case *BinaryExpr:
		leftFunc, err := compileFilter(v.Left, registry)
		if err != nil {
			return nil, err
		}
		rightFunc, err := compileFilter(v.Right, registry)
		if err != nil {
			return nil, err
		}
		switch v.Op {
		case "&&":
			return makeAnd(leftFunc, rightFunc), nil
		case "||":
			return makeOr(leftFunc, rightFunc), nil
		case "==":
			return makeEq(leftFunc, rightFunc), nil
		case "!=":
			return makeNe(leftFunc, rightFunc), nil
		case "<":
			return makeLt(leftFunc, rightFunc), nil
		case "<=":
			return makeLe(leftFunc, rightFunc), nil
		case ">":
			return makeGt(leftFunc, rightFunc), nil
		case ">=":
			return makeGe(leftFunc, rightFunc), nil
		default:
			return nil, fmt.Errorf("unknown operator: %s", v.Op)
		}

	case *FuncExpr:
		args := make([]filterFunc, len(v.Args))
		for idx, arg := range v.Args {
			compiled, err := compileFilter(arg, registry)
			if err != nil {
				return nil, err
			}
			args[idx] = compiled
		}
		def, ok := registry.function(v.Name)
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrUnknownFunc, v.Name)
		}
		return makeCall(def.Eval, args), nil

	default:
		return nil, fmt.Errorf("unknown filter expression")
	}
}

func makeNot(expr filterFunc) filterFunc {
	return func(ctx *evalCtx) *evalValue {
		return scalarValue(!toBool(expr(ctx)))
	}
}

func makeAnd(left, right filterFunc) filterFunc {
	return func(ctx *evalCtx) *evalValue {
		leftValue := left(ctx)
		if !toBool(leftValue) {
			return scalarValue(false)
		}
		return scalarValue(toBool(right(ctx)))
	}
}

func makeOr(left, right filterFunc) filterFunc {
	return func(ctx *evalCtx) *evalValue {
		leftValue := left(ctx)
		if toBool(leftValue) {
			return scalarValue(true)
		}
		return scalarValue(toBool(right(ctx)))
	}
}

func makeEq(left, right filterFunc) filterFunc {
	return func(ctx *evalCtx) *evalValue {
		return scalarValue(compareValuesEq(left(ctx), right(ctx)))
	}
}

func makeNe(left, right filterFunc) filterFunc {
	return func(ctx *evalCtx) *evalValue {
		return scalarValue(compareValuesNe(left(ctx), right(ctx)))
	}
}

func makeLt(left, right filterFunc) filterFunc {
	return func(ctx *evalCtx) *evalValue {
		return scalarValue(compareValuesLt(left(ctx), right(ctx)))
	}
}

func makeLe(left, right filterFunc) filterFunc {
	return func(ctx *evalCtx) *evalValue {
		return scalarValue(compareValuesLe(left(ctx), right(ctx)))
	}
}

func makeGt(left, right filterFunc) filterFunc {
	return func(ctx *evalCtx) *evalValue {
		return scalarValue(compareValuesGt(left(ctx), right(ctx)))
	}
}

func makeGe(left, right filterFunc) filterFunc {
	return func(ctx *evalCtx) *evalValue {
		return scalarValue(compareValuesGe(left(ctx), right(ctx)))
	}
}

func makeCall(
	evaluator FunctionEvaluator, args []filterFunc,
) filterFunc {
	return func(ctx *evalCtx) *evalValue {
		return fromFunctionValue(evaluator(evalFunctionArgs(args, ctx)))
	}
}
