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
		return SelectFilter(filter), nil

	default:
		return nil, fmt.Errorf("unknown selector kind")
	}
}

func compileFilter(expr FilterExpr, registry *Registry) (FilterFunc, error) {
	switch v := expr.(type) {
	case *LiteralExpr:
		return Literal(v.Value), nil

	case *PathValueExpr:
		path, err := makePath(v.Path, registry)
		if err != nil {
			return nil, err
		}
		if v.Absolute {
			return PathRoot(path), nil
		}
		return PathCurrent(path), nil

	case *UnaryExpr:
		exprFunc, err := compileFilter(v.Expr, registry)
		if err != nil {
			return nil, err
		}
		if v.Op == "!" {
			return Not(exprFunc), nil
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
			return And(leftFunc, rightFunc), nil
		case "||":
			return Or(leftFunc, rightFunc), nil
		case "==":
			return Eq(leftFunc, rightFunc), nil
		case "!=":
			return Ne(leftFunc, rightFunc), nil
		case "<":
			return Lt(leftFunc, rightFunc), nil
		case "<=":
			return Le(leftFunc, rightFunc), nil
		case ">":
			return Gt(leftFunc, rightFunc), nil
		case ">=":
			return Ge(leftFunc, rightFunc), nil
		default:
			return nil, fmt.Errorf("unknown operator: %s", v.Op)
		}

	case *FuncExpr:
		args := make([]FilterFunc, len(v.Args))
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
		return Call(def.Eval, args...), nil

	default:
		return nil, fmt.Errorf("unknown filter expression")
	}
}
