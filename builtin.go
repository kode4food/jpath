package jpath

import (
	"maps"
	"regexp"
)

var defaultFunctions = map[string]*FunctionDefinition{
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
		Eval:     evalValueFunc,
	},
	"match": {
		Validate: validateMatchSearchFunction,
		Eval:     evalMatchFunction,
	},
	"search": {
		Validate: validateMatchSearchFunction,
		Eval:     evalSearchFunction,
	},
}

func registerDefaultFunctions(r *Registry) {
	maps.Copy(r.functions, defaultFunctions)
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

func evalMatchFunction(args []*Value) *Value {
	return evalMatch(args, true)
}

func evalSearchFunction(args []*Value) *Value {
	return evalMatch(args, false)
}

func evalLength(args []*Value) *Value {
	if len(args) != 1 {
		return ScalarValue(nothing)
	}
	v := args[0]
	if v.IsNodes {
		if len(v.Nodes) != 1 {
			return ScalarValue(nothing)
		}
		return evalLengthValue(v.Nodes[0])
	}
	return evalLengthValue(v.Scalar)
}

func evalLengthValue(value any) *Value {
	switch raw := value.(type) {
	case string:
		return ScalarValue(float64(len([]rune(raw))))
	case []any:
		return ScalarValue(float64(len(raw)))
	case map[string]any:
		return ScalarValue(float64(len(raw)))
	default:
		return ScalarValue(nothing)
	}
}

func evalCount(args []*Value) *Value {
	if len(args) != 1 {
		return ScalarValue(nothing)
	}
	v := args[0]
	if v.IsNodes {
		return ScalarValue(float64(len(v.Nodes)))
	}
	return ScalarValue(nothing)
}

func evalValueFunc(args []*Value) *Value {
	if len(args) != 1 {
		return ScalarValue(nothing)
	}
	v := args[0]
	if v.IsNodes {
		if len(v.Nodes) != 1 {
			return ScalarValue(nothing)
		}
		return ScalarValue(v.Nodes[0])
	}
	return v
}

func evalMatch(args []*Value, full bool) *Value {
	if len(args) != 2 {
		return ScalarValue(nothing)
	}
	lhs, ok := singularValue(args[0])
	if !ok {
		return ScalarValue(nothing)
	}
	rhs, ok := singularValue(args[1])
	if !ok {
		return ScalarValue(nothing)
	}
	left, ok := lhs.(string)
	if !ok {
		return ScalarValue(nothing)
	}
	right, ok := rhs.(string)
	if !ok {
		return ScalarValue(nothing)
	}
	if full {
		right = "^(?:" + right + ")$"
	}
	right = normalizeDotPattern(right)
	re, err := regexp.Compile(right)
	if err != nil {
		return ScalarValue(nothing)
	}
	return ScalarValue(re.MatchString(left))
}

func singularValue(v *Value) (any, bool) {
	if v.IsNodes {
		if len(v.Nodes) != 1 {
			return nil, false
		}
		return v.Nodes[0], true
	}
	return v.Scalar, true
}
