package jpath

import (
	"errors"
	"fmt"
)

type (
	exprContext uint8

	unaryArgValidator func(name string, arg FilterExpr) error
)

const (
	contextLogical exprContext = iota
	contextComparisonOperand
	contextFunctionArg
)

var (
	// ErrUnknownSelector indicates an unsupported AST selector kind
	ErrUnknownSelector = errors.New("unknown selector kind")

	// ErrUnknownOperator indicates an unsupported AST operator
	ErrUnknownOperator = errors.New("unknown operator")

	// ErrUnknownExpr indicates an unsupported filter AST expression
	ErrUnknownExpr = errors.New("unknown filter expression")

	// ErrLiteralMustBeCompared is raised for bare literals in logical context
	ErrLiteralMustBeCompared = errors.New("literal must be compared")

	// ErrCompRequiresSingularQuery is raised for non-singular comparisons
	ErrCompRequiresSingularQuery = errors.New(
		"comparison requires singular query",
	)

	// ErrInvalidFuncArity is raised when function arity is invalid
	ErrInvalidFuncArity = errors.New("invalid function arity")

	// ErrFuncResultMustBeCompared is raised for logical use without compare
	ErrFuncResultMustBeCompared = errors.New("function result must be compared")

	// ErrFuncResultMustNotBeCompared is raised for prohibited comparisons
	ErrFuncResultMustNotBeCompared = errors.New(
		"function result must not be compared",
	)

	// ErrFuncRequiresSingularQuery is raised for singular-path requirements
	ErrFuncRequiresSingularQuery = errors.New(
		"function requires singular query",
	)

	// ErrFuncRequiresQueryArgument is raised for query-arg requirements
	ErrFuncRequiresQueryArgument = errors.New(
		"function requires query argument",
	)
)

func validatePath(path *PathExpr, reg *Registry) error {
	for _, seg := range path.Segments {
		for _, sel := range seg.Selectors {
			if sel.Kind != SelectorFilter {
				continue
			}
			if err := validateExpr(
				sel.Filter, contextLogical, false, reg,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateExpr(
	ex FilterExpr, ctx exprContext, inComparison bool, reg *Registry,
) error {
	switch v := ex.(type) {
	case *LiteralExpr:
		if ctx == contextLogical {
			return ErrLiteralMustBeCompared
		}
		return nil

	case *PathValueExpr:
		if inComparison && !isSingularPath(v.Path) {
			return ErrCompRequiresSingularQuery
		}
		return nil

	case *UnaryExpr:
		return validateExpr(v.Expr, contextLogical, false, reg)

	case *BinaryExpr:
		switch v.Op {
		case OpAnd, OpOr:
			if err := validateExpr(
				v.Left, contextLogical, false, reg,
			); err != nil {
				return err
			}
			return validateExpr(v.Right, contextLogical, false, reg)

		case OpEq, OpNe, OpLt, OpLte, OpGt, OpGte:
			if err := validateExpr(
				v.Left, contextComparisonOperand, true, reg,
			); err != nil {
				return err
			}
			return validateExpr(
				v.Right, contextComparisonOperand, true, reg,
			)

		default:
			return fmt.Errorf("%w: %s", ErrUnknownOperator, v.Op)
		}

	case *FuncExpr:
		if err := validateFunction(v, ctx, inComparison, reg); err != nil {
			return err
		}
		for _, a := range v.Args {
			if err := validateExpr(
				a, contextFunctionArg, false, reg,
			); err != nil {
				return err
			}
		}
		return nil

	default:
		return ErrUnknownExpr
	}
}

func validateFunction(
	f *FuncExpr, ctx exprContext, inComparison bool, reg *Registry,
) error {
	def, ok := reg.function(f.Name)
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
	for _, seg := range path.Segments {
		if seg.Descendant || len(seg.Selectors) != 1 {
			return false
		}
		sl := seg.Selectors[0]
		if sl.Kind != SelectorName && sl.Kind != SelectorIndex {
			return false
		}
	}
	return true
}

func validateMatchSearchFunction(
	args []FilterExpr, _ FunctionUse, inComparison bool,
) error {
	if err := validateFunctionArity("match/search", args, 2); err != nil {
		return err
	}
	if inComparison {
		return fmt.Errorf("%w: match/search", ErrFuncResultMustNotBeCompared)
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
		"%w: %s requires query argument", ErrFuncRequiresQueryArgument, name,
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
