package jpath

import "sort"

// SelectName builds a selector for object-member lookup by name
func SelectName(name string) SelectorFunc {
	return func(out []any, node, _ any) []any {
		obj, ok := node.(map[string]any)
		if !ok {
			return out
		}
		value, ok := obj[name]
		if !ok {
			return out
		}
		return append(out, value)
	}
}

// SelectIndex builds a selector for array element lookup by index
func SelectIndex(index int) SelectorFunc {
	return func(out []any, node, _ any) []any {
		arr, ok := node.([]any)
		if !ok {
			return out
		}
		pos := normalizeIndex(len(arr), index)
		if pos >= 0 && pos < len(arr) {
			return append(out, arr[pos])
		}
		return out
	}
}

// SelectWildcard builds a selector for wildcard child selection
func SelectWildcard() SelectorFunc {
	return func(out []any, node, _ any) []any {
		return appendWildcard(out, node)
	}
}

// SelectSlice builds a selector for array slice selection
func SelectSlice(s *SliceExpr) SelectorFunc {
	op := selectorCaseFor(s)
	if op == selectorCaseEmpty {
		return selectorIdentity
	}

	plan := &SlicePlan{Step: s.Step}
	if s.HasStart {
		plan.Start = s.Start
	}
	if s.HasEnd {
		plan.End = s.End
	}

	maker := selectorDispatch[op]
	if maker == nil {
		return selectorIdentity
	}
	return maker(plan)
}

// SelectFilter builds a selector for filter-based child selection
func SelectFilter(filter FilterFunc) SelectorFunc {
	return func(out []any, node, root any) []any {
		return appendFilter(out, node, root, filter)
	}
}

func appendWildcard(out []any, node any) []any {
	switch v := node.(type) {
	case []any:
		return append(out, v...)
	case map[string]any:
		for _, key := range sortedKeys(v) {
			out = append(out, v[key])
		}
		return out
	default:
		return out
	}
}

func appendFilter(out []any, node, root any, flt FilterFunc) []any {
	ctx := &FilterCtx{Root: root}
	switch v := node.(type) {
	case []any:
		for _, elem := range v {
			ctx.Current = elem
			if toBool(flt(ctx)) {
				out = append(out, elem)
			}
		}
		return out
	case map[string]any:
		for _, k := range sortedKeys(v) {
			elem := v[k]
			ctx.Current = elem
			if toBool(flt(ctx)) {
				out = append(out, elem)
			}
		}
		return out
	default:
		return out
	}
}

func normalizeIndex(size, idx int) int {
	if idx < 0 {
		idx += size
	}
	return idx
}

func sortedKeys(v map[string]any) []string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
