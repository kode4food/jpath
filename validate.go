package jpath

import "fmt"

type (
	exprContext uint8

	unaryArgValidator func(name string, arg FilterExpr) error
)

const (
	contextLogical exprContext = iota
	contextComparisonOperand
	contextFunctionArg
)

func validatePath(path *PathExpr, registry *Registry) error {
	for _, sg := range path.Segments {
		for _, sel := range sg.Selectors {
			if sel.Kind != SelectorFilter {
				continue
			}
			if err := validateExpr(
				sel.Filter, contextLogical, false,
				registry,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateExpr(
	ex FilterExpr, ctx exprContext, inComparison bool, registry *Registry,
) error {
	switch v := ex.(type) {
	case *LiteralExpr:
		if ctx == contextLogical {
			return fmt.Errorf("%w", ErrLiteralMustBeCompared)
		}
		return nil

	case *PathValueExpr:
		if inComparison && !isSingularPath(v.Path) {
			return fmt.Errorf("%w", ErrCompRequiresSingularQuery)
		}
		return nil

	case *UnaryExpr:
		return validateExpr(v.Expr, contextLogical, false, registry)

	case *BinaryExpr:
		switch v.Op {
		case "&&", "||":
			if err := validateExpr(
				v.Left, contextLogical, false,
				registry,
			); err != nil {
				return err
			}
			return validateExpr(v.Right, contextLogical, false, registry)

		case "==", "!=", "<", "<=", ">", ">=":
			if err := validateExpr(
				v.Left, contextComparisonOperand, true, registry,
			); err != nil {
				return err
			}
			return validateExpr(
				v.Right, contextComparisonOperand, true,
				registry,
			)

		default:
			return fmt.Errorf("unknown operator: %s", v.Op)
		}

	case *FuncExpr:
		if err := validateFunction(v, ctx, inComparison, registry); err != nil {
			return err
		}
		for _, a := range v.Args {
			if err := validateExpr(
				a, contextFunctionArg, false,
				registry,
			); err != nil {
				return err
			}
		}
		return nil

	default:
		return fmt.Errorf("unknown expression")
	}
}

func validateFunction(
	f *FuncExpr, ctx exprContext, inComparison bool, registry *Registry,
) error {
	def, ok := registry.function(f.Name)
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownFunc, f.Name)
	}
	if def.Validate == nil {
		return nil
	}
	return def.Validate(f.Args, functionUse(ctx), inComparison)
}

func functionUse(ctx exprContext) FunctionUse {
	switch ctx {
	case contextLogical:
		return FunctionUseLogical

	case contextComparisonOperand:
		return FunctionUseComparisonOperand

	default:
		return FunctionUseArgument
	}
}

func isSingularPath(path *PathExpr) bool {
	for _, sg := range path.Segments {
		if sg.Descendant || len(sg.Selectors) != 1 {
			return false
		}
		sl := sg.Selectors[0]
		if sl.Kind != SelectorName && sl.Kind != SelectorIndex {
			return false
		}
	}
	return true
}

func validateLengthFunction(
	args []FilterExpr, use FunctionUse, inComparison bool,
) error {
	return validateUnaryComparedSingular("length", args, use, inComparison)
}

func validateCountFunction(
	args []FilterExpr, use FunctionUse, inComparison bool,
) error {
	return validateUnaryComparedReq("count", args, use, inComparison)
}

func validateValueFunction(
	args []FilterExpr, use FunctionUse, inComparison bool,
) error {
	return validateUnaryComparedReq("value", args, use, inComparison)
}

func validateMatchSearchFunction(
	args []FilterExpr, _ FunctionUse, inComparison bool,
) error {
	if err := validateFunctionArity("match/search", args, 2); err != nil {
		return err
	}
	if inComparison {
		return fmt.Errorf(
			"%w: match/search",
			ErrFuncResultMustNotBeCompared,
		)
	}
	return nil
}

func validateUnaryCompared(
	name string, args []FilterExpr, use FunctionUse, inComparison bool,
	argValidator unaryArgValidator,
) error {
	if err := validateFunctionArity(name, args, 1); err != nil {
		return err
	}
	if err := validateComparedUse(name, use, inComparison); err != nil {
		return err
	}
	return argValidator(name, args[0])
}

func validateUnaryComparedReq(
	name string, args []FilterExpr, use FunctionUse, inComparison bool,
) error {
	return validateUnaryCompared(
		name, args, use, inComparison, validateQueryArg,
	)
}

func validateUnaryComparedSingular(
	name string, args []FilterExpr, use FunctionUse, inComparison bool,
) error {
	return validateUnaryCompared(
		name, args, use, inComparison, validateSingularQueryArg,
	)
}

func validateFunctionArity(name string, args []FilterExpr, want int) error {
	if len(args) == want {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrInvalidFuncArity, name)
}

func validateComparedUse(
	name string, use FunctionUse, inComparison bool,
) error {
	if inComparison || use != FunctionUseLogical {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrFuncResultMustBeCompared, name)
}

func validateQueryArg(name string, arg FilterExpr) error {
	if _, ok := arg.(*PathValueExpr); ok {
		return nil
	}
	return fmt.Errorf(
		"%w: %s requires query argument",
		ErrFuncRequiresQueryArgument,
		name,
	)
}

func validateSingularQueryArg(name string, arg FilterExpr) error {
	pv, ok := arg.(*PathValueExpr)
	if !ok {
		return nil
	}
	if isSingularPath(pv.Path) {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrFuncRequiresSingularQuery, name)
}
