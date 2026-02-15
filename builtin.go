package jpath

import (
	"fmt"
	"maps"
	"regexp"
)

var defaultFunctions = map[string]FunctionDefinition{
	"length": {
		Validate: validateLengthFunction,
		Eval:     evalLength,
	},
	"count": {
		Validate: validateCountFunction,
		Eval:     evalCount,
	},
	"value": {
		Validate: validateValueFunction,
		Eval:     evalValueFn,
	},
	"match": {
		Validate: validateMatchSearchFunction,
		Eval: func(args []FunctionValue) FunctionValue {
			return evalMatch(args, true)
		},
	},
	"search": {
		Validate: validateMatchSearchFunction,
		Eval: func(args []FunctionValue) FunctionValue {
			return evalMatch(args, false)
		},
	},
}

func registerDefaultFunctions(r *Registry) {
	maps.Copy(r.functions, defaultFunctions)
}

func builtinFunction(name string) (FunctionDefinition, bool) {
	def, ok := defaultFunctions[name]
	return def, ok
}

func validateLengthFunction(
	args []FilterExpr, use FunctionUse, inComparison bool,
) error {
	if len(args) != 1 {
		return fmt.Errorf("invalid function arity")
	}
	if !inComparison && use == FunctionUseLogical {
		return fmt.Errorf("function result must be compared")
	}
	if pv, ok := args[0].(PathValueExpr); ok {
		if !isSingularPath(pv.Path) {
			return fmt.Errorf("function requires singular query")
		}
	}
	return nil
}

func validateCountFunction(
	args []FilterExpr, use FunctionUse, inComparison bool,
) error {
	if len(args) != 1 {
		return fmt.Errorf("invalid function arity")
	}
	if !inComparison && use == FunctionUseLogical {
		return fmt.Errorf("function result must be compared")
	}
	if _, ok := args[0].(PathValueExpr); !ok {
		return fmt.Errorf("count requires query argument")
	}
	return nil
}

func validateValueFunction(
	args []FilterExpr, use FunctionUse, inComparison bool,
) error {
	if len(args) != 1 {
		return fmt.Errorf("invalid function arity")
	}
	if !inComparison && use == FunctionUseLogical {
		return fmt.Errorf("function result must be compared")
	}
	if _, ok := args[0].(PathValueExpr); !ok {
		return fmt.Errorf("value requires query argument")
	}
	return nil
}

func validateMatchSearchFunction(
	args []FilterExpr, _ FunctionUse, inComparison bool,
) error {
	if len(args) != 2 {
		return fmt.Errorf("invalid function arity")
	}
	if inComparison {
		return fmt.Errorf("function result must not be compared")
	}
	return nil
}

func evalLength(args []FunctionValue) FunctionValue {
	if len(args) != 1 {
		return scalarFunctionValue(nothing)
	}
	v := args[0]
	if v.IsNodes {
		if len(v.Nodes) != 1 {
			return scalarFunctionValue(nothing)
		}
		return evalLengthValue(v.Nodes[0])
	}
	return evalLengthValue(v.Scalar)
}

func evalLengthValue(value any) FunctionValue {
	switch raw := value.(type) {
	case string:
		return scalarFunctionValue(float64(len([]rune(raw))))
	case []any:
		return scalarFunctionValue(float64(len(raw)))
	case map[string]any:
		return scalarFunctionValue(float64(len(raw)))
	default:
		return scalarFunctionValue(nothing)
	}
}

func evalCount(args []FunctionValue) FunctionValue {
	if len(args) != 1 {
		return scalarFunctionValue(nothing)
	}
	v := args[0]
	if v.IsNodes {
		return scalarFunctionValue(float64(len(v.Nodes)))
	}
	return scalarFunctionValue(nothing)
}

func evalValueFn(args []FunctionValue) FunctionValue {
	if len(args) != 1 {
		return scalarFunctionValue(nothing)
	}
	v := args[0]
	if v.IsNodes {
		if len(v.Nodes) != 1 {
			return scalarFunctionValue(nothing)
		}
		return scalarFunctionValue(v.Nodes[0])
	}
	return v
}

func evalMatch(args []FunctionValue, full bool) FunctionValue {
	if len(args) != 2 {
		return scalarFunctionValue(nothing)
	}
	lhs, ok := singularFunctionValue(args[0])
	if !ok {
		return scalarFunctionValue(nothing)
	}
	rhs, ok := singularFunctionValue(args[1])
	if !ok {
		return scalarFunctionValue(nothing)
	}
	left, ok := lhs.(string)
	if !ok {
		return scalarFunctionValue(nothing)
	}
	right, ok := rhs.(string)
	if !ok {
		return scalarFunctionValue(nothing)
	}
	if full {
		right = "^(?:" + right + ")$"
	}
	right = normalizeDotPattern(right)
	re, err := regexp.Compile(right)
	if err != nil {
		return scalarFunctionValue(nothing)
	}
	return scalarFunctionValue(re.MatchString(left))
}

func singularFunctionValue(v FunctionValue) (any, bool) {
	if v.IsNodes {
		if len(v.Nodes) != 1 {
			return nil, false
		}
		return v.Nodes[0], true
	}
	return v.Scalar, true
}

func scalarFunctionValue(value any) FunctionValue {
	return FunctionValue{Scalar: value}
}
