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
func (c *Compiler) Compile(path PathExpr) (Path, error) {
	return compilePath(path, c.registry)
}

func compilePath(path PathExpr, registry *Registry) (Path, error) {
	if err := validatePath(path, registry); err != nil {
		return Path{}, err
	}
	return makePath(path, registry)
}

func makePath(path PathExpr, registry *Registry) (Path, error) {
	res := Path{}
	for _, sg := range path.Segments {
		if sg.Descendant {
			res.Code = append(res.Code, Instruction{Op: OpDescend})
		}
		segStart := len(res.Code)
		res.Code = append(res.Code, Instruction{Op: OpSegmentStart})
		for _, sl := range sg.Selectors {
			inst, err := compileSelector(sl, &res, registry)
			if err != nil {
				return Path{}, err
			}
			res.Code = append(res.Code, inst)
		}
		segEnd := len(res.Code)
		res.Code[segStart].Arg = segEnd
		res.Code = append(res.Code, Instruction{Op: OpSegmentEnd})
	}
	return res, nil
}

func compileSelector(
	sel SelectorExpr, run *Path, registry *Registry,
) (Instruction, error) {
	switch sel.Kind {
	case SelectorName:
		return Instruction{Op: OpSelectName, Arg: run.addConst(sel.Name)}, nil
	case SelectorIndex:
		return Instruction{Op: OpSelectIndex, Arg: run.addConst(sel.Index)}, nil
	case SelectorWildcard:
		return Instruction{Op: OpSelectWildcard}, nil
	case SelectorSlice:
		return compileSlice(sel.Slice, run), nil
	case SelectorFilter:
		flt, err := compileFilter(sel.Filter, registry)
		if err != nil {
			return Instruction{}, err
		}
		return Instruction{
			Op:  OpSelectFilter,
			Arg: run.addConst(flt),
		}, nil
	default:
		return Instruction{}, fmt.Errorf("unknown selector kind")
	}
}

func compileSlice(s SliceExpr, run *Path) Instruction {
	if s.Step == 0 {
		return Instruction{Op: OpSelectSliceEmpty}
	}
	if s.Step == 1 && !s.HasStart && !s.HasEnd {
		return Instruction{Op: OpSelectArrayAll}
	}

	plan := SlicePlan{Step: s.Step}
	if s.HasStart {
		plan.Start = s.Start
	}
	if s.HasEnd {
		plan.End = s.End
	}
	return Instruction{
		Op:  sliceOpcode(s),
		Arg: run.addConst(plan),
	}
}

func sliceOpcode(s SliceExpr) Opcode {
	if s.Step > 0 {
		return sliceForwardOpcode(s)
	}
	return sliceBackwardOpcode(s)
}

func sliceForwardOpcode(s SliceExpr) Opcode {
	switch {
	case !s.HasStart && !s.HasEnd:
		return OpSelectSliceF00
	case s.HasStart && !s.HasEnd:
		return sliceForwardStartOpcode(s.Start)
	case !s.HasStart && s.HasEnd:
		return sliceForwardEndOpcode(s.End)
	default:
		return sliceForwardRangeOpcode(s.Start, s.End)
	}
}

func sliceBackwardOpcode(s SliceExpr) Opcode {
	switch {
	case !s.HasStart && !s.HasEnd:
		return OpSelectSliceB00
	case s.HasStart && !s.HasEnd:
		return sliceBackwardStartOpcode(s.Start)
	case !s.HasStart && s.HasEnd:
		return sliceBackwardEndOpcode(s.End)
	default:
		return sliceBackwardRangeOpcode(s.Start, s.End)
	}
}

func sliceForwardStartOpcode(start int) Opcode {
	if start >= 0 {
		return OpSelectSliceF10P
	}
	return OpSelectSliceF10N
}

func sliceForwardEndOpcode(end int) Opcode {
	if end >= 0 {
		return OpSelectSliceF01P
	}
	return OpSelectSliceF01N
}

func sliceForwardRangeOpcode(start, end int) Opcode {
	switch {
	case start >= 0 && end >= 0:
		return OpSelectSliceF11PP
	case start >= 0:
		return OpSelectSliceF11PN
	case end >= 0:
		return OpSelectSliceF11NP
	default:
		return OpSelectSliceF11NN
	}
}

func sliceBackwardStartOpcode(start int) Opcode {
	if start >= 0 {
		return OpSelectSliceB10P
	}
	return OpSelectSliceB10N
}

func sliceBackwardEndOpcode(end int) Opcode {
	if end >= 0 {
		return OpSelectSliceB01P
	}
	return OpSelectSliceB01N
}

func sliceBackwardRangeOpcode(start, end int) Opcode {
	switch {
	case start >= 0 && end >= 0:
		return OpSelectSliceB11PP
	case start >= 0:
		return OpSelectSliceB11PN
	case end >= 0:
		return OpSelectSliceB11NP
	default:
		return OpSelectSliceB11NN
	}
}

func compileFilter(expr FilterExpr, registry *Registry) (FilterExpr, error) {
	switch v := expr.(type) {
	case LiteralExpr:
		return v, nil
	case PathValueExpr:
		path, err := makePath(v.Path, registry)
		if err != nil {
			return nil, err
		}
		return compiledPathValueExpr{absolute: v.Absolute, path: path}, nil
	case UnaryExpr:
		ex, err := compileFilter(v.Expr, registry)
		if err != nil {
			return nil, err
		}
		v.Expr = ex
		return v, nil
	case BinaryExpr:
		left, err := compileFilter(v.Left, registry)
		if err != nil {
			return nil, err
		}
		right, err := compileFilter(v.Right, registry)
		if err != nil {
			return nil, err
		}
		v.Left = left
		v.Right = right
		return v, nil
	case FuncExpr:
		args := make([]FilterExpr, len(v.Args))
		for idx, arg := range v.Args {
			compiled, err := compileFilter(arg, registry)
			if err != nil {
				return nil, err
			}
			args[idx] = compiled
		}
		def, ok := registry.function(v.Name)
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrUnknownFunction, v.Name)
		}
		return compiledFuncExpr{
			evaluator: def.Eval,
			args:      args,
		}, nil
	default:
		return nil, fmt.Errorf("unknown filter expression")
	}
}
