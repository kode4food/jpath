package jpath

import "fmt"

// Compiler compiles parsed JSONPath syntax trees into runnable programs
type Compiler struct {
	registry *Registry
}

// NewCompiler creates a new Compiler
func NewCompiler() *Compiler {
	return &Compiler{
		registry: NewRegistry(),
	}
}

// Compile compiles a parsed PathExpr into an executable Path
func (c *Compiler) Compile(path *PathExpr) (Path, error) {
	return compilePath(path, c.registry)
}

func compilePath(path *PathExpr, reg *Registry) (Path, error) {
	if err := validatePath(path, reg); err != nil {
		return nil, err
	}
	return makePath(path, reg)
}

func makePath(path *PathExpr, reg *Registry) (Path, error) {
	if isSingularPath(path) {
		return makeSingularPath(path), nil
	}
	segments := make([]SegmentFunc, 0, len(path.Segments))
	for i := 0; i < len(path.Segments); {
		end := i
		for end < len(path.Segments) && isSingularSegment(path.Segments[end]) {
			end++
		}
		if end-i > 1 {
			segments = append(
				segments, makeSingularSegment(path.Segments[i:end]),
			)
			i = end
			continue
		}
		compiled, err := compileSegment(path.Segments[i], reg)
		if err != nil {
			return nil, err
		}
		segments = append(segments, compiled)
		i++
	}
	return ComposePath(segments...), nil
}

func compileSegment(seg *SegmentExpr, reg *Registry) (SegmentFunc, error) {
	selectors := make([]SelectorFunc, len(seg.Selectors))
	for idx, selector := range seg.Selectors {
		compiled, err := compileSelector(selector, reg)
		if err != nil {
			return nil, err
		}
		selectors[idx] = compiled
	}
	if !seg.Descendant {
		return ChildSegment(selectors...), nil
	}
	if segmentFunctionFree(seg) {
		return makeDescendantSegment(selectors), nil
	}
	return DescendantSegment(selectors...), nil
}

func compileSelector(sel *SelectorExpr, reg *Registry) (SelectorFunc, error) {
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
		filter, err := compilePredicate(sel.Filter, reg)
		if err != nil {
			return nil, err
		}
		return selectPredicate(filter), nil

	default:
		return nil, fmt.Errorf("%w: %d", ErrUnknownSelector, sel.Kind)
	}
}

func compileFilter(expr FilterExpr, reg *Registry) (FilterFunc, error) {
	switch v := expr.(type) {
	case *LiteralExpr:
		return Literal(v.Value), nil

	case *PathValueExpr:
		path, err := makePath(v.Path, reg)
		if err != nil {
			return nil, err
		}
		if v.Absolute {
			return PathRoot(path), nil
		}
		return PathCurrent(path), nil

	case *UnaryExpr, *BinaryExpr:
		predicate, err := compilePredicate(expr, reg)
		if err != nil {
			return nil, err
		}
		return func(ctx *FilterCtx) any {
			return predicate(ctx)
		}, nil

	case *FuncExpr:
		builtin, err := compileBuiltin(v, reg)
		if err != nil {
			return nil, err
		}
		if builtin != nil {
			return builtin, nil
		}
		args := make([]FilterFunc, len(v.Args))
		for idx, arg := range v.Args {
			compiled, err := compileFilter(arg, reg)
			if err != nil {
				return nil, err
			}
			args[idx] = compiled
		}
		def, ok := reg.function(v.Name)
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrUnknownFunc, v.Name)
		}
		return Call(def.Eval, args...), nil

	default:
		return nil, ErrUnknownExpr
	}
}
