package jpath

import (
	"maps"
	"regexp"

	"github.com/kode4food/lru"
)

const regexCacheLimit = 4096

var (
	defaultFunctions = map[string]*FunctionDefinition{
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
			Eval:     evalFullMatch,
		},
		"search": {
			Validate: validateMatchSearchFunction,
			Eval:     evalPartialMatch,
		},
	}

	regexCache = lru.NewCache[*regexp.Regexp](regexCacheLimit)
)

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

func evalFullMatch(args []*Value) *Value {
	left, pattern, ok := evalMatchArguments(args)
	if !ok {
		return ScalarValue(nothing)
	}
	return evalPatternMatch(left, "^(?:"+pattern+")$")
}

func evalPartialMatch(args []*Value) *Value {
	left, pattern, ok := evalMatchArguments(args)
	if !ok {
		return ScalarValue(nothing)
	}
	return evalPatternMatch(left, pattern)
}

func evalMatchArguments(args []*Value) (string, string, bool) {
	if len(args) != 2 {
		return "", "", false
	}
	lhs, ok := singularValue(args[0])
	if !ok {
		return "", "", false
	}
	rhs, ok := singularValue(args[1])
	if !ok {
		return "", "", false
	}
	left, ok := lhs.(string)
	if !ok {
		return "", "", false
	}
	pattern, ok := rhs.(string)
	if !ok {
		return "", "", false
	}
	return left, pattern, true
}

func evalPatternMatch(left, pattern string) *Value {
	pattern = normalizeDotPattern(pattern)
	re, ok := compileMatchPattern(pattern)
	if !ok {
		return ScalarValue(nothing)
	}
	return ScalarValue(re.MatchString(left))
}

func compileMatchPattern(pattern string) (*regexp.Regexp, bool) {
	re, err := regexCache.Get(pattern, func() (*regexp.Regexp, error) {
		return regexp.Compile(pattern)
	})
	return re, err == nil && re != nil
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
