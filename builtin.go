package jpath

import (
	"regexp"
	"unicode/utf8"

	"github.com/kode4food/lru"
)

type (
	regexMatcher struct {
		cache *lru.Cache[*regexp.Regexp]
	}

	regexInput struct {
		text    string
		pattern string
	}
)

const regexCacheLimit = 4096

func (r *regexMatcher) fullMatch(args []any) any {
	return r.evalMatch(args, true)
}

func (r *regexMatcher) partialMatch(args []any) any {
	return r.evalMatch(args, false)
}

func (r *regexMatcher) evalMatch(args []any, full bool) any {
	match, ok := evalMatchArguments(args)
	if !ok {
		return Nodes(nil)
	}
	if full {
		match.pattern = "^(?:" + match.pattern + ")$"
	}
	return r.match(match)
}

func (r *regexMatcher) match(args regexInput) any {
	re, err := r.compile(args.pattern)
	if err != nil {
		return Nodes(nil)
	}
	return re.MatchString(args.text)
}

func (r *regexMatcher) compile(pattern string) (*regexp.Regexp, error) {
	return r.cache.Get(pattern, func() (*regexp.Regexp, error) {
		return regexp.Compile(normalizeDotPattern(pattern))
	})
}

func defaultFunctions(matcher *regexMatcher) map[string]*FunctionDefinition {
	return map[string]*FunctionDefinition{
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

func evalLength(args []any) any {
	value, ok := singularValue(args[0])
	if !ok {
		return Nodes(nil)
	}
	return evalLengthValue(value)
}

func evalLengthValue(value any) any {
	switch raw := value.(type) {
	case string:
		return float64(utf8.RuneCountInString(raw))

	case []any:
		return float64(len(raw))

	case map[string]any:
		return float64(len(raw))

	default:
		return Nodes(nil)
	}
}

func evalCount(args []any) any {
	if nodes, ok := args[0].(Nodes); ok {
		return float64(len(nodes))
	}
	return Nodes(nil)
}

func evalValueFunc(args []any) any {
	if value, ok := singularValue(args[0]); ok {
		return value
	}
	return Nodes(nil)
}

func evalMatchArguments(args []any) (regexInput, bool) {
	lhs, ok := singularValue(args[0])
	if !ok {
		return regexInput{}, false
	}
	rhs, ok := singularValue(args[1])
	if !ok {
		return regexInput{}, false
	}
	left, ok := lhs.(string)
	if !ok {
		return regexInput{}, false
	}
	pattern, ok := rhs.(string)
	if !ok {
		return regexInput{}, false
	}
	return regexInput{text: left, pattern: pattern}, true
}
