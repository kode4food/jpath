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

	nothingType struct{}
)

var nothing nothingType

// ScalarValue constructs a scalar filter value
func ScalarValue(value any) *Value {
	return &Value{Scalar: value}
}

// NodesValue constructs a node-list filter value
func NodesValue(v []any) *Value {
	return &Value{IsNodes: true, Nodes: v}
}

func evalFunctionArgs(args []FilterFunc, ctx *FilterCtx) []*Value {
	res := make([]*Value, len(args))
	for idx, arg := range args {
		res[idx] = arg(ctx)
	}
	return res
}

func compareEmptyEq(left, right []any) bool {
	if len(left) == 0 && len(right) == 0 {
		return true
	}
	if len(left) == 0 && isNothing(right) {
		return true
	}
	if len(right) == 0 && isNothing(left) {
		return true
	}
	return false
}

func compareEmptyNe(left, right []any) bool {
	if len(left) == 0 && len(right) == 0 {
		return false
	}
	if len(left) == 0 && isNothing(right) {
		return false
	}
	if len(right) == 0 && isNothing(left) {
		return false
	}
	return true
}

func compareValuesEq(left, right *Value) bool {
	lc := expandCandidates(left)
	rc := expandCandidates(right)
	if len(lc) == 0 || len(rc) == 0 {
		return compareEmptyEq(lc, rc)
	}
	for _, lv := range lc {
		for _, rv := range rc {
			if reflect.DeepEqual(lv, rv) {
				return true
			}
		}
	}
	return false
}

func compareValuesNe(left, right *Value) bool {
	lc := expandCandidates(left)
	rc := expandCandidates(right)
	if len(lc) == 0 || len(rc) == 0 {
		return compareEmptyNe(lc, rc)
	}
	for _, lv := range lc {
		for _, rv := range rc {
			if !reflect.DeepEqual(lv, rv) {
				return true
			}
		}
	}
	return false
}

func compareValuesLt(left, right *Value) bool {
	lc := expandCandidates(left)
	rc := expandCandidates(right)
	if len(lc) == 0 || len(rc) == 0 {
		return false
	}
	for _, lv := range lc {
		for _, rv := range rc {
			if matched, ok := lessThan(lv, rv); ok && matched {
				return true
			}
		}
	}
	return false
}

func compareValuesLe(left, right *Value) bool {
	lc := expandCandidates(left)
	rc := expandCandidates(right)
	if len(lc) == 0 || len(rc) == 0 {
		return false
	}
	for _, lv := range lc {
		for _, rv := range rc {
			if reflect.DeepEqual(lv, rv) {
				return true
			}
			if matched, ok := lessThan(lv, rv); ok && matched {
				return true
			}
		}
	}
	return false
}

func compareValuesGt(left, right *Value) bool {
	lc := expandCandidates(left)
	rc := expandCandidates(right)
	if len(lc) == 0 || len(rc) == 0 {
		return false
	}
	for _, lv := range lc {
		for _, rv := range rc {
			if matched, ok := greaterThan(lv, rv); ok && matched {
				return true
			}
		}
	}
	return false
}

func compareValuesGe(left, right *Value) bool {
	lc := expandCandidates(left)
	rc := expandCandidates(right)
	if len(lc) == 0 || len(rc) == 0 {
		return false
	}
	for _, lv := range lc {
		for _, rv := range rc {
			if reflect.DeepEqual(lv, rv) {
				return true
			}
			if matched, ok := greaterThan(lv, rv); ok && matched {
				return true
			}
		}
	}
	return false
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

func expandCandidates(v *Value) []any {
	if v.IsNodes {
		return v.Nodes
	}
	return []any{v.Scalar}
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

func isNothing(v []any) bool {
	if len(v) != 1 {
		return false
	}
	_, ok := v[0].(nothingType)
	return ok
}
