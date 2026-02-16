package jpath

import "fmt"

type exprContext uint8

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
				sel.Filter,
				contextLogical,
				false,
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
			return fmt.Errorf("literal must be compared")
		}
		return nil
	case *PathValueExpr:
		if inComparison && !isSingularPath(v.Path) {
			return fmt.Errorf("comparison requires singular query")
		}
		return nil
	case *UnaryExpr:
		return validateExpr(v.Expr, contextLogical, false, registry)
	case *BinaryExpr:
		switch v.Op {
		case "&&", "||":
			if err := validateExpr(
				v.Left,
				contextLogical,
				false,
				registry,
			); err != nil {
				return err
			}
			return validateExpr(v.Right, contextLogical, false, registry)
		case "==", "!=", "<", "<=", ">", ">=":
			if err := validateExpr(
				v.Left,
				contextComparisonOperand,
				true,
				registry,
			); err != nil {
				return err
			}
			return validateExpr(
				v.Right,
				contextComparisonOperand,
				true,
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
				a,
				contextFunctionArg,
				false,
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
		return fmt.Errorf("%w: %s", ErrUnknownFunction, f.Name)
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
