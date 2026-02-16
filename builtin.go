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
		Eval: func(args []*FunctionValue) *FunctionValue {
			return evalMatch(args, true)
		},
	},
	"search": {
		Validate: validateMatchSearchFunction,
		Eval: func(args []*FunctionValue) *FunctionValue {
			return evalMatch(args, false)
		},
	},
}

func registerDefaultFunctions(r *Registry) {
	maps.Copy(r.functions, defaultFunctions)
}

func evalLength(args []*FunctionValue) *FunctionValue {
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

func evalLengthValue(value any) *FunctionValue {
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

func evalCount(args []*FunctionValue) *FunctionValue {
	if len(args) != 1 {
		return scalarFunctionValue(nothing)
	}
	v := args[0]
	if v.IsNodes {
		return scalarFunctionValue(float64(len(v.Nodes)))
	}
	return scalarFunctionValue(nothing)
}

func evalValueFunc(args []*FunctionValue) *FunctionValue {
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

func evalMatch(args []*FunctionValue, full bool) *FunctionValue {
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

func singularFunctionValue(v *FunctionValue) (any, bool) {
	if v.IsNodes {
		if len(v.Nodes) != 1 {
			return nil, false
		}
		return v.Nodes[0], true
	}
	return v.Scalar, true
}

func scalarFunctionValue(value any) *FunctionValue {
	return &FunctionValue{Scalar: value}
}
