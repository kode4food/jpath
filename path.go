package jpath

import "sort"

type (
	// Path is an executable VM program
	Path struct {
		// Code stores VM instructions
		Code []Instruction
		// Constants stores literal values referenced by instructions
		Constants []any
	}

	// Instruction is a single VM instruction
	Instruction struct {
		// Op identifies the operation
		Op Opcode
		// Arg stores an operand index or immediate value for Op
		Arg int
	}

	// SlicePlan stores precompiled bounds for slice selection
	SlicePlan struct {
		// Start is the raw start bound from the parsed selector
		Start int
		// End is the raw end bound from the parsed selector
		End int
		// Step is the raw step bound from the parsed selector
		Step int
	}

	segmentFrame struct {
		input []any
		out   []any
		idx   int
		start int
	}
)

// Query executes the program against a JSON document
func (p *Path) Query(document any) []any {
	var FRAMES []segmentFrame
	var NODE any

	ROOT := document
	CODE := p.Code
	CUR := []any{ROOT}

	for PC := 0; PC < len(CODE); {
		INST := CODE[PC]
		switch INST.Op {
		case OpDescend:
			CUR = descendantsOf(CUR)
			PC++

		case OpSegmentStart:
			if len(CUR) == 0 {
				PC = INST.Arg + 1
				continue
			}
			FRAMES = append(FRAMES, segmentFrame{
				input: CUR,
				out:   make([]any, 0),
				idx:   0,
				start: PC + 1,
			})
			NODE = CUR[0]
			PC++

		case OpSegmentEnd:
			top := &FRAMES[len(FRAMES)-1]
			top.idx++
			if top.idx < len(top.input) {
				NODE = top.input[top.idx]
				PC = top.start
				continue
			}
			CUR = top.out
			FRAMES = FRAMES[:len(FRAMES)-1]
			PC++

		case OpSelectName:
			top := &FRAMES[len(FRAMES)-1]
			name := p.Constants[INST.Arg].(string)
			if obj, ok := NODE.(map[string]any); ok {
				if val, ok := obj[name]; ok {
					top.out = append(top.out, val)
				}
			}
			PC++

		case OpSelectIndex:
			top := &FRAMES[len(FRAMES)-1]
			arr, ok := NODE.([]any)
			if !ok {
				PC++
				continue
			}
			idx := p.Constants[INST.Arg].(int)
			pos := normalizeIndex(len(arr), idx)
			if pos >= 0 && pos < len(arr) {
				top.out = append(top.out, arr[pos])
			}
			PC++

		case OpSelectWildcard:
			top := &FRAMES[len(FRAMES)-1]
			top.out = appendWildcard(top.out, NODE)
			PC++

		case OpSelectArrayAll:
			top := &FRAMES[len(FRAMES)-1]
			if arr, ok := NODE.([]any); ok {
				top.out = append(top.out, arr...)
			}
			PC++

		case OpSelectSliceF00:
			top := &FRAMES[len(FRAMES)-1]
			arr, ok := NODE.([]any)
			if !ok || len(arr) == 0 {
				PC++
				continue
			}
			plan := p.Constants[INST.Arg].(SlicePlan)
			top.out = appendSliceF00(top.out, arr, plan.Step)
			PC++

		case OpSelectSliceF10P:
			top := &FRAMES[len(FRAMES)-1]
			arr, ok := NODE.([]any)
			if !ok || len(arr) == 0 {
				PC++
				continue
			}
			plan := p.Constants[INST.Arg].(SlicePlan)
			top.out = appendSliceF10P(top.out, arr, plan.Start, plan.Step)
			PC++

		case OpSelectSliceF10N:
			top := &FRAMES[len(FRAMES)-1]
			arr, ok := NODE.([]any)
			if !ok || len(arr) == 0 {
				PC++
				continue
			}
			plan := p.Constants[INST.Arg].(SlicePlan)
			top.out = appendSliceF10N(top.out, arr, plan.Start, plan.Step)
			PC++

		case OpSelectSliceF01P:
			top := &FRAMES[len(FRAMES)-1]
			arr, ok := NODE.([]any)
			if !ok || len(arr) == 0 {
				PC++
				continue
			}
			plan := p.Constants[INST.Arg].(SlicePlan)
			top.out = appendSliceF01P(top.out, arr, plan.End, plan.Step)
			PC++

		case OpSelectSliceF01N:
			top := &FRAMES[len(FRAMES)-1]
			arr, ok := NODE.([]any)
			if !ok || len(arr) == 0 {
				PC++
				continue
			}
			plan := p.Constants[INST.Arg].(SlicePlan)
			top.out = appendSliceF01N(top.out, arr, plan.End, plan.Step)
			PC++

		case OpSelectSliceF11PP:
			top := &FRAMES[len(FRAMES)-1]
			arr, ok := NODE.([]any)
			if !ok || len(arr) == 0 {
				PC++
				continue
			}
			plan := p.Constants[INST.Arg].(SlicePlan)
			top.out = appendSliceF11PP(
				top.out, arr, plan.Start, plan.End, plan.Step,
			)
			PC++

		case OpSelectSliceF11PN:
			top := &FRAMES[len(FRAMES)-1]
			arr, ok := NODE.([]any)
			if !ok || len(arr) == 0 {
				PC++
				continue
			}
			plan := p.Constants[INST.Arg].(SlicePlan)
			top.out = appendSliceF11PN(
				top.out, arr, plan.Start, plan.End, plan.Step,
			)
			PC++

		case OpSelectSliceF11NP:
			top := &FRAMES[len(FRAMES)-1]
			arr, ok := NODE.([]any)
			if !ok || len(arr) == 0 {
				PC++
				continue
			}
			plan := p.Constants[INST.Arg].(SlicePlan)
			top.out = appendSliceF11NP(
				top.out, arr, plan.Start, plan.End, plan.Step,
			)
			PC++

		case OpSelectSliceF11NN:
			top := &FRAMES[len(FRAMES)-1]
			arr, ok := NODE.([]any)
			if !ok || len(arr) == 0 {
				PC++
				continue
			}
			plan := p.Constants[INST.Arg].(SlicePlan)
			top.out = appendSliceF11NN(
				top.out, arr, plan.Start, plan.End, plan.Step,
			)
			PC++

		case OpSelectSliceB00:
			top := &FRAMES[len(FRAMES)-1]
			arr, ok := NODE.([]any)
			if !ok || len(arr) == 0 {
				PC++
				continue
			}
			plan := p.Constants[INST.Arg].(SlicePlan)
			top.out = appendSliceB00(top.out, arr, plan.Step)
			PC++

		case OpSelectSliceB10P:
			top := &FRAMES[len(FRAMES)-1]
			arr, ok := NODE.([]any)
			if !ok || len(arr) == 0 {
				PC++
				continue
			}
			plan := p.Constants[INST.Arg].(SlicePlan)
			top.out = appendSliceB10P(top.out, arr, plan.Start, plan.Step)
			PC++

		case OpSelectSliceB10N:
			top := &FRAMES[len(FRAMES)-1]
			arr, ok := NODE.([]any)
			if !ok || len(arr) == 0 {
				PC++
				continue
			}
			plan := p.Constants[INST.Arg].(SlicePlan)
			top.out = appendSliceB10N(top.out, arr, plan.Start, plan.Step)
			PC++

		case OpSelectSliceB01P:
			top := &FRAMES[len(FRAMES)-1]
			arr, ok := NODE.([]any)
			if !ok || len(arr) == 0 {
				PC++
				continue
			}
			plan := p.Constants[INST.Arg].(SlicePlan)
			top.out = appendSliceB01P(top.out, arr, plan.End, plan.Step)
			PC++

		case OpSelectSliceB01N:
			top := &FRAMES[len(FRAMES)-1]
			arr, ok := NODE.([]any)
			if !ok || len(arr) == 0 {
				PC++
				continue
			}
			plan := p.Constants[INST.Arg].(SlicePlan)
			top.out = appendSliceB01N(top.out, arr, plan.End, plan.Step)
			PC++

		case OpSelectSliceB11PP:
			top := &FRAMES[len(FRAMES)-1]
			arr, ok := NODE.([]any)
			if !ok || len(arr) == 0 {
				PC++
				continue
			}
			plan := p.Constants[INST.Arg].(SlicePlan)
			top.out = appendSliceB11PP(
				top.out, arr, plan.Start, plan.End, plan.Step,
			)
			PC++

		case OpSelectSliceB11PN:
			top := &FRAMES[len(FRAMES)-1]
			arr, ok := NODE.([]any)
			if !ok || len(arr) == 0 {
				PC++
				continue
			}
			plan := p.Constants[INST.Arg].(SlicePlan)
			top.out = appendSliceB11PN(
				top.out, arr, plan.Start, plan.End, plan.Step,
			)
			PC++

		case OpSelectSliceB11NP:
			top := &FRAMES[len(FRAMES)-1]
			arr, ok := NODE.([]any)
			if !ok || len(arr) == 0 {
				PC++
				continue
			}
			plan := p.Constants[INST.Arg].(SlicePlan)
			top.out = appendSliceB11NP(
				top.out, arr, plan.Start, plan.End, plan.Step,
			)
			PC++

		case OpSelectSliceB11NN:
			top := &FRAMES[len(FRAMES)-1]
			arr, ok := NODE.([]any)
			if !ok || len(arr) == 0 {
				PC++
				continue
			}
			plan := p.Constants[INST.Arg].(SlicePlan)
			top.out = appendSliceB11NN(
				top.out, arr, plan.Start, plan.End, plan.Step,
			)
			PC++

		case OpSelectSliceEmpty:
			_ = FRAMES[len(FRAMES)-1]
			PC++

		case OpSelectFilter:
			top := &FRAMES[len(FRAMES)-1]
			flt := p.Constants[INST.Arg].(FilterExpr)
			top.out = appendFilter(top.out, NODE, ROOT, flt)
			PC++

		default:
			panic("unknown opcode")
		}
	}
	return CUR
}

func (p *Path) addConst(value any) int {
	p.Constants = append(p.Constants, value)
	return len(p.Constants) - 1
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
	return start
}

func forwardEndPos(end, size int) int {
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
	return end
}

func backwardStartPos(start, size int) int {
	if start >= size {
		return size - 1
	}
	return start
}

func backwardStartNeg(start, size int) int {
	start += size
	if start < -1 {
		return -1
	}
	return start
}

func backwardEndPos(end, size int) int {
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
