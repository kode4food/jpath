package jpath

import (
	"regexp"

	"github.com/kode4food/lru"
)

type regexMatcher struct {
	cache *lru.Cache[*regexp.Regexp]
}

const regexCacheLimit = 4096

func (r *regexMatcher) fullMatch(args []*Value) *Value {
	match, ok := evalMatchArguments(args)
	if !ok {
		return ScalarValue(nothingType{})
	}
	match.pattern = "^(?:" + match.pattern + ")$"
	return r.match(match)
}

func (r *regexMatcher) partialMatch(args []*Value) *Value {
	match, ok := evalMatchArguments(args)
	if !ok {
		return ScalarValue(nothingType{})
	}
	return r.match(match)
}

type matchArgs struct {
	text    string
	pattern string
}

func (r *regexMatcher) match(args matchArgs) *Value {
	pattern := normalizeDotPattern(args.pattern)
	re, err := r.cache.Get(pattern, func() (*regexp.Regexp, error) {
		return regexp.Compile(pattern)
	})
	if err != nil || re == nil {
		return ScalarValue(nothingType{})
	}
	return ScalarValue(re.MatchString(args.text))
}

func registerDefaultFunctions(reg *Registry) {
	matcher := &regexMatcher{
		cache: lru.NewCache[*regexp.Regexp](regexCacheLimit),
	}
	reg.functions = map[string]*FunctionDefinition{
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
			Eval:     matcher.fullMatch,
		},
		"search": {
			Validate: validateMatchSearchFunction,
			Eval:     matcher.partialMatch,
		},
	}
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
		return ScalarValue(nothingType{})
	}
	value, ok := args[0].singularValue()
	if !ok {
		return ScalarValue(nothingType{})
	}
	return evalLengthValue(value)
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
		return ScalarValue(nothingType{})
	}
}

func evalCount(args []*Value) *Value {
	if len(args) != 1 {
		return ScalarValue(nothingType{})
	}
	v := args[0]
	if v.IsNodes {
		return ScalarValue(float64(len(v.Nodes)))
	}
	return ScalarValue(nothingType{})
}

func evalValueFunc(args []*Value) *Value {
	if len(args) != 1 {
		return ScalarValue(nothingType{})
	}
	v := args[0]
	if !v.IsNodes {
		return v
	}
	if len(v.Nodes) != 1 {
		return ScalarValue(nothingType{})
	}
	return ScalarValue(v.Nodes[0])
}

func evalMatchArguments(args []*Value) (matchArgs, bool) {
	if len(args) != 2 {
		return matchArgs{}, false
	}
	lhs, ok := args[0].singularValue()
	if !ok {
		return matchArgs{}, false
	}
	rhs, ok := args[1].singularValue()
	if !ok {
		return matchArgs{}, false
	}
	left, ok := lhs.(string)
	if !ok {
		return matchArgs{}, false
	}
	pattern, ok := rhs.(string)
	if !ok {
		return matchArgs{}, false
	}
	return matchArgs{text: left, pattern: pattern}, true
}
