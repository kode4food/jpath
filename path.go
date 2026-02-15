package jpath

import "sort"

// Query executes the program against a JSON document
func (p *Path) Query(document any) []any {
	if p == nil {
		return []any{}
	}
	current := []any{document}
	var out []any
	for pc := 0; pc < len(p.Code); {
		inst := p.Code[pc]
		switch inst.Op {
		case OpDescend:
			current = descendantsOf(current)
			pc++

		case OpSegmentStart:
			out = out[:0]
			for _, node := range current {
				for cur := pc + 1; cur < inst.Arg; cur++ {
					out = p.selectNode(out, node, document, p.Code[cur])
				}
			}
			current = append(current[:0], out...)
			pc = inst.Arg + 1

		default:
			pc++
		}
	}
	if current == nil {
		return []any{}
	}
	return current
}

func (p *Path) selectNode(out []any, node, root any, i Instruction) []any {
	switch i.Op {
	case OpSelectName:
		name := p.Constants[i.Arg].(string)
		if obj, ok := node.(map[string]any); ok {
			if val, ok := obj[name]; ok {
				out = append(out, val)
			}
		}
		return out

	case OpSelectIndex:
		arr, ok := node.([]any)
		if !ok {
			return out
		}
		idx := p.Constants[i.Arg].(int)
		pos := normalizeIndex(len(arr), idx)
		if pos >= 0 && pos < len(arr) {
			out = append(out, arr[pos])
		}
		return out

	case OpSelectWildcard:
		return appendWildcard(out, node)

	case OpSelectArrayAll:
		if arr, ok := node.([]any); ok {
			out = append(out, arr...)
		}
		return out

	case OpSelectSliceF00:
		arr, ok := node.([]any)
		if !ok || len(arr) == 0 {
			return out
		}
		plan := p.Constants[i.Arg].(SlicePlan)
		return appendSliceF00(out, arr, plan.Step)

	case OpSelectSliceF10P:
		arr, ok := node.([]any)
		if !ok || len(arr) == 0 {
			return out
		}
		plan := p.Constants[i.Arg].(SlicePlan)
		return appendSliceF10P(out, arr, plan.Start, plan.Step)

	case OpSelectSliceF10N:
		arr, ok := node.([]any)
		if !ok || len(arr) == 0 {
			return out
		}
		plan := p.Constants[i.Arg].(SlicePlan)
		return appendSliceF10N(out, arr, plan.Start, plan.Step)

	case OpSelectSliceF01P:
		arr, ok := node.([]any)
		if !ok || len(arr) == 0 {
			return out
		}
		plan := p.Constants[i.Arg].(SlicePlan)
		return appendSliceF01P(out, arr, plan.End, plan.Step)

	case OpSelectSliceF01N:
		arr, ok := node.([]any)
		if !ok || len(arr) == 0 {
			return out
		}
		plan := p.Constants[i.Arg].(SlicePlan)
		return appendSliceF01N(out, arr, plan.End, plan.Step)

	case OpSelectSliceF11PP:
		arr, ok := node.([]any)
		if !ok || len(arr) == 0 {
			return out
		}
		plan := p.Constants[i.Arg].(SlicePlan)
		return appendSliceF11PP(out, arr, plan.Start, plan.End, plan.Step)

	case OpSelectSliceF11PN:
		arr, ok := node.([]any)
		if !ok || len(arr) == 0 {
			return out
		}
		plan := p.Constants[i.Arg].(SlicePlan)
		return appendSliceF11PN(out, arr, plan.Start, plan.End, plan.Step)

	case OpSelectSliceF11NP:
		arr, ok := node.([]any)
		if !ok || len(arr) == 0 {
			return out
		}
		plan := p.Constants[i.Arg].(SlicePlan)
		return appendSliceF11NP(out, arr, plan.Start, plan.End, plan.Step)

	case OpSelectSliceF11NN:
		arr, ok := node.([]any)
		if !ok || len(arr) == 0 {
			return out
		}
		plan := p.Constants[i.Arg].(SlicePlan)
		return appendSliceF11NN(out, arr, plan.Start, plan.End, plan.Step)

	case OpSelectSliceB00:
		arr, ok := node.([]any)
		if !ok || len(arr) == 0 {
			return out
		}
		plan := p.Constants[i.Arg].(SlicePlan)
		return appendSliceB00(out, arr, plan.Step)

	case OpSelectSliceB10P:
		arr, ok := node.([]any)
		if !ok || len(arr) == 0 {
			return out
		}
		plan := p.Constants[i.Arg].(SlicePlan)
		return appendSliceB10P(out, arr, plan.Start, plan.Step)

	case OpSelectSliceB10N:
		arr, ok := node.([]any)
		if !ok || len(arr) == 0 {
			return out
		}
		plan := p.Constants[i.Arg].(SlicePlan)
		return appendSliceB10N(out, arr, plan.Start, plan.Step)

	case OpSelectSliceB01P:
		arr, ok := node.([]any)
		if !ok || len(arr) == 0 {
			return out
		}
		plan := p.Constants[i.Arg].(SlicePlan)
		return appendSliceB01P(out, arr, plan.End, plan.Step)

	case OpSelectSliceB01N:
		arr, ok := node.([]any)
		if !ok || len(arr) == 0 {
			return out
		}
		plan := p.Constants[i.Arg].(SlicePlan)
		return appendSliceB01N(out, arr, plan.End, plan.Step)

	case OpSelectSliceB11PP:
		arr, ok := node.([]any)
		if !ok || len(arr) == 0 {
			return out
		}
		plan := p.Constants[i.Arg].(SlicePlan)
		return appendSliceB11PP(out, arr, plan.Start, plan.End, plan.Step)

	case OpSelectSliceB11PN:
		arr, ok := node.([]any)
		if !ok || len(arr) == 0 {
			return out
		}
		plan := p.Constants[i.Arg].(SlicePlan)
		return appendSliceB11PN(out, arr, plan.Start, plan.End, plan.Step)

	case OpSelectSliceB11NP:
		arr, ok := node.([]any)
		if !ok || len(arr) == 0 {
			return out
		}
		plan := p.Constants[i.Arg].(SlicePlan)
		return appendSliceB11NP(out, arr, plan.Start, plan.End, plan.Step)

	case OpSelectSliceB11NN:
		arr, ok := node.([]any)
		if !ok || len(arr) == 0 {
			return out
		}
		plan := p.Constants[i.Arg].(SlicePlan)
		return appendSliceB11NN(out, arr, plan.Start, plan.End, plan.Step)

	case OpSelectSliceEmpty:
		return out

	case OpSelectFilter:
		flt := p.Constants[i.Arg].(FilterExpr)
		return appendFilter(out, node, root, flt)

	default:
		return out
	}
}

func descendantsOf(nodes []any) []any {
	res := make([]any, 0, len(nodes))
	for _, node := range nodes {
		walkDescendants(node, func(v any) {
			res = append(res, v)
		})
	}
	return res
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

func appendSliceF00(out, arr []any, step int) []any {
	for idx := 0; idx < len(arr); idx += step {
		out = append(out, arr[idx])
	}
	return out
}

func appendSliceF10P(out, arr []any, start, step int) []any {
	start = forwardStartPos(start, len(arr))
	for idx := start; idx < len(arr); idx += step {
		out = append(out, arr[idx])
	}
	return out
}

func appendSliceF10N(out, arr []any, start, step int) []any {
	start = forwardStartNeg(start, len(arr))
	for idx := start; idx < len(arr); idx += step {
		out = append(out, arr[idx])
	}
	return out
}

func appendSliceF01P(out, arr []any, end, step int) []any {
	end = forwardEndPos(end, len(arr))
	for idx := 0; idx < end; idx += step {
		out = append(out, arr[idx])
	}
	return out
}

func appendSliceF01N(out, arr []any, end, step int) []any {
	end = forwardEndNeg(end, len(arr))
	for idx := 0; idx < end; idx += step {
		out = append(out, arr[idx])
	}
	return out
}

func appendSliceF11PP(out, arr []any, start, end, step int) []any {
	size := len(arr)
	start = forwardStartPos(start, size)
	end = forwardEndPos(end, size)
	if start >= end {
		return out
	}
	for idx := start; idx < end; idx += step {
		out = append(out, arr[idx])
	}
	return out
}

func appendSliceF11PN(out, arr []any, start, end, step int) []any {
	size := len(arr)
	start = forwardStartPos(start, size)
	end = forwardEndNeg(end, size)
	if start >= end {
		return out
	}
	for idx := start; idx < end; idx += step {
		out = append(out, arr[idx])
	}
	return out
}

func appendSliceF11NP(out, arr []any, start, end, step int) []any {
	size := len(arr)
	start = forwardStartNeg(start, size)
	end = forwardEndPos(end, size)
	if start >= end {
		return out
	}
	for idx := start; idx < end; idx += step {
		out = append(out, arr[idx])
	}
	return out
}

func appendSliceF11NN(out, arr []any, start, end, step int) []any {
	size := len(arr)
	start = forwardStartNeg(start, size)
	end = forwardEndNeg(end, size)
	if start >= end {
		return out
	}
	for idx := start; idx < end; idx += step {
		out = append(out, arr[idx])
	}
	return out
}

func appendSliceB00(out, arr []any, step int) []any {
	for idx := len(arr) - 1; idx > -1; idx += step {
		out = append(out, arr[idx])
	}
	return out
}

func appendSliceB10P(out, arr []any, start, step int) []any {
	start = backwardStartPos(start, len(arr))
	for idx := start; idx > -1; idx += step {
		out = append(out, arr[idx])
	}
	return out
}

func appendSliceB10N(out, arr []any, start, step int) []any {
	start = backwardStartNeg(start, len(arr))
	for idx := start; idx > -1; idx += step {
		out = append(out, arr[idx])
	}
	return out
}

func appendSliceB01P(out, arr []any, end, step int) []any {
	end = backwardEndPos(end, len(arr))
	for idx := len(arr) - 1; idx > end; idx += step {
		out = append(out, arr[idx])
	}
	return out
}

func appendSliceB01N(out, arr []any, end, step int) []any {
	end = backwardEndNeg(end, len(arr))
	for idx := len(arr) - 1; idx > end; idx += step {
		out = append(out, arr[idx])
	}
	return out
}

func appendSliceB11PP(out, arr []any, start, end, step int) []any {
	start = backwardStartPos(start, len(arr))
	end = backwardEndPos(end, len(arr))
	for idx := start; idx > end; idx += step {
		out = append(out, arr[idx])
	}
	return out
}

func appendSliceB11PN(out, arr []any, start, end, step int) []any {
	start = backwardStartPos(start, len(arr))
	end = backwardEndNeg(end, len(arr))
	for idx := start; idx > end; idx += step {
		out = append(out, arr[idx])
	}
	return out
}

func appendSliceB11NP(out, arr []any, start, end, step int) []any {
	start = backwardStartNeg(start, len(arr))
	end = backwardEndPos(end, len(arr))
	for idx := start; idx > end; idx += step {
		out = append(out, arr[idx])
	}
	return out
}

func appendSliceB11NN(out, arr []any, start, end, step int) []any {
	start = backwardStartNeg(start, len(arr))
	end = backwardEndNeg(end, len(arr))
	for idx := start; idx > end; idx += step {
		out = append(out, arr[idx])
	}
	return out
}

func forwardStartPos(start, size int) int {
	if start < 0 {
		return 0
	}
	if start > size {
		return size
	}
	return start
}

func forwardStartNeg(start, size int) int {
	start += size
	if start < 0 {
		return 0
	}
	if start > size {
		return size
	}
	return start
}

func forwardEndPos(end, size int) int {
	if end < 0 {
		return 0
	}
	if end > size {
		return size
	}
	return end
}

func forwardEndNeg(end, size int) int {
	end += size
	if end < 0 {
		return 0
	}
	if end > size {
		return size
	}
	return end
}

func backwardStartPos(start, size int) int {
	if start < 0 {
		return -1
	}
	if start >= size {
		return size - 1
	}
	return start
}

func backwardStartNeg(start, size int) int {
	start += size
	if start >= size {
		return size - 1
	}
	return start
}

func backwardEndPos(end, size int) int {
	if end < -1 {
		return -1
	}
	if end >= size {
		return size - 1
	}
	return end
}

func backwardEndNeg(end, size int) int {
	end += size
	if end < -1 {
		return -1
	}
	if end >= size {
		return size - 1
	}
	return end
}

func appendFilter(out []any, node, root any, flt FilterExpr) []any {
	ctx := &evalCtx{root: root}
	switch v := node.(type) {
	case []any:
		for _, elem := range v {
			ctx.current = elem
			if toBool(flt.eval(ctx)) {
				out = append(out, elem)
			}
		}
		return out
	case map[string]any:
		for _, k := range sortedKeys(v) {
			elem := v[k]
			ctx.current = elem
			if toBool(flt.eval(ctx)) {
				out = append(out, elem)
			}
		}
		return out
	default:
		return out
	}
}

func walkDescendants(node any, visit func(any)) {
	visit(node)
	switch v := node.(type) {
	case []any:
		for _, elem := range v {
			walkDescendants(elem, visit)
		}
	case map[string]any:
		for _, key := range sortedKeys(v) {
			walkDescendants(v[key], visit)
		}
	}
}

func sortedKeys(v map[string]any) []string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func normalizeIndex(size, idx int) int {
	if idx < 0 {
		idx += size
	}
	return idx
}
