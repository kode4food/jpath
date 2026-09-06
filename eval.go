package jpath

import (
	"reflect"
	"strings"
)

type (
	// FilterFunc evaluates a filter to a scalar value or Nodes
	FilterFunc func(*FilterCtx) any

	// FilterCtx provides the document root and current node to a filter
	FilterCtx struct {
		Root    any
		Current any
	}

	comparison struct {
		left  any
		right any
	}

	matchFunc func(left, right any) bool
)

func compareValues(op string, values comparison) bool {
	leftEmpty := valueCount(values.left) == 0
	rightEmpty := valueCount(values.right) == 0
	if leftEmpty || rightEmpty {
		return op == OpEq && leftEmpty == rightEmpty ||
			op == OpNe && leftEmpty != rightEmpty
	}
	var match matchFunc
	switch op {
	case OpEq:
		match = reflect.DeepEqual

	case OpNe:
		match = notDeepEqual

	case OpLt:
		match = lessThanMatch

	case OpLte:
		match = lessEqualMatch

	case OpGt:
		match = greaterThanMatch

	case OpGte:
		match = greaterEqualMatch

	default:
		return false
	}
	return matchAny(values, match)
}

func matchAny(values comparison, match matchFunc) bool {
	if left, ok := values.left.(Nodes); ok {
		if right, ok := values.right.(Nodes); ok {
			for _, lv := range left {
				for _, rv := range right {
					if match(lv, rv) {
						return true
					}
				}
			}
			return false
		}
		for _, lv := range left {
			if match(lv, values.right) {
				return true
			}
		}
		return false
	}
	if right, ok := values.right.(Nodes); ok {
		for _, rv := range right {
			if match(values.left, rv) {
				return true
			}
		}
		return false
	}
	return match(values.left, values.right)
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

func toBool(value any) bool {
	switch raw := value.(type) {
	case Nodes:
		return len(raw) > 0

	case nil:
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
