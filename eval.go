package jpath

import (
	"reflect"
	"strings"
)

type (
	FilterFunc func(*FilterCtx) *Value

	FilterCtx struct {
		Root    any
		Current any
	}

	matchFunc func(left, right any) bool
)

func evalFunctionArgs(args []FilterFunc, ctx *FilterCtx) []*Value {
	res := make([]*Value, len(args))
	for idx, arg := range args {
		res[idx] = arg(ctx)
	}
	return res
}

func compareEmptyEq(left, right *Value) bool {
	leftCount := left.Count()
	rightCount := right.Count()
	if leftCount == 0 && rightCount == 0 {
		return true
	}
	if leftCount == 0 && right.IsNothing() {
		return true
	}
	if rightCount == 0 && left.IsNothing() {
		return true
	}
	return false
}

func compareEmptyNe(left, right *Value) bool {
	leftCount := left.Count()
	rightCount := right.Count()
	if leftCount == 0 && rightCount == 0 {
		return false
	}
	if leftCount == 0 && right.IsNothing() {
		return false
	}
	if rightCount == 0 && left.IsNothing() {
		return false
	}
	return true
}

func compareValuesEq(left, right *Value) bool {
	if left.Count() == 0 || right.Count() == 0 {
		return compareEmptyEq(left, right)
	}
	return matchAny(left, right, reflect.DeepEqual)
}

func compareValuesNe(left, right *Value) bool {
	if left.Count() == 0 || right.Count() == 0 {
		return compareEmptyNe(left, right)
	}
	return matchAny(left, right, notDeepEqual)
}

func compareValuesLt(left, right *Value) bool {
	if left.Count() == 0 || right.Count() == 0 {
		return false
	}
	return matchAny(left, right, lessThanMatch)
}

func compareValuesLe(left, right *Value) bool {
	if left.Count() == 0 || right.Count() == 0 {
		return false
	}
	return matchAny(left, right, lessEqualMatch)
}

func compareValuesGt(left, right *Value) bool {
	if left.Count() == 0 || right.Count() == 0 {
		return false
	}
	return matchAny(left, right, greaterThanMatch)
}

func compareValuesGe(left, right *Value) bool {
	if left.Count() == 0 || right.Count() == 0 {
		return false
	}
	return matchAny(left, right, greaterEqualMatch)
}

func matchAny(left, right *Value, match matchFunc) bool {
	if left.IsNodes {
		if right.IsNodes {
			for _, lv := range left.Nodes {
				for _, rv := range right.Nodes {
					if match(lv, rv) {
						return true
					}
				}
			}
			return false
		}
		for _, lv := range left.Nodes {
			if match(lv, right.Scalar) {
				return true
			}
		}
		return false
	}
	if right.IsNodes {
		for _, rv := range right.Nodes {
			if match(left.Scalar, rv) {
				return true
			}
		}
		return false
	}
	return match(left.Scalar, right.Scalar)
}

func lessThan(left, right any) (bool, bool) {
	lf, lok := asNumber(left)
	rf, rok := asNumber(right)
	if lok && rok {
		return lf < rf, true
	}
	ls, lok := left.(string)
	rs, rok := right.(string)
	if lok && rok {
		return ls < rs, true
	}
	return false, false
}

func greaterThan(left, right any) (bool, bool) {
	lf, lok := asNumber(left)
	rf, rok := asNumber(right)
	if lok && rok {
		return lf > rf, true
	}
	ls, lok := left.(string)
	rs, rok := right.(string)
	if lok && rok {
		return ls > rs, true
	}
	return false, false
}

func notDeepEqual(left, right any) bool {
	return !reflect.DeepEqual(left, right)
}

func lessThanMatch(left, right any) bool {
	matched, ok := lessThan(left, right)
	return ok && matched
}

func lessEqualMatch(left, right any) bool {
	if reflect.DeepEqual(left, right) {
		return true
	}
	return lessThanMatch(left, right)
}

func greaterThanMatch(left, right any) bool {
	matched, ok := greaterThan(left, right)
	return ok && matched
}

func greaterEqualMatch(left, right any) bool {
	if reflect.DeepEqual(left, right) {
		return true
	}
	return greaterThanMatch(left, right)
}

func toBool(v *Value) bool {
	if v.IsNodes {
		return len(v.Nodes) > 0
	}
	switch raw := v.Scalar.(type) {
	case nil:
		return false
	case nothingType:
		return false
	case bool:
		return raw
	default:
		return true
	}
}

func asNumber(value any) (float64, bool) {
	if n, ok := value.(float64); ok {
		return n, true
	}
	if n, ok := value.(int); ok {
		return float64(n), true
	}
	return 0, false
}

func normalizeDotPattern(pattern string) string {
	var b strings.Builder
	escaped := false
	inClass := false
	for _, r := range pattern {
		switch {
		case escaped:
			b.WriteRune(r)
			escaped = false
		case r == '\\':
			b.WriteRune(r)
			escaped = true
		case r == '[':
			b.WriteRune(r)
			inClass = true
		case r == ']':
			b.WriteRune(r)
			inClass = false
		case r == '.' && !inClass:
			b.WriteString("[^\\r\\n]")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
