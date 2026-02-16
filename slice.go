package jpath

type (
	// SlicePlan stores precompiled bounds for slice selection
	SlicePlan struct {
		Start int
		End   int
		Step  int
	}

	selectorCase uint8

	append0Func func(out, arr []any, step int) []any
	append1Func func(out, arr []any, bound, step int) []any
	append2Func func(out, arr []any, start, end, step int) []any

	selectorMakerFunc func(plan *SlicePlan) SelectorFunc
)

const (
	selectorCaseArrayAll selectorCase = iota
	selectorCaseF00
	selectorCaseF10P
	selectorCaseF10N
	selectorCaseF01P
	selectorCaseF01N
	selectorCaseF11PP
	selectorCaseF11PN
	selectorCaseF11NP
	selectorCaseF11NN
	selectorCaseB00
	selectorCaseB10P
	selectorCaseB10N
	selectorCaseB01P
	selectorCaseB01N
	selectorCaseB11PP
	selectorCaseB11PN
	selectorCaseB11NP
	selectorCaseB11NN
	selectorCaseEmpty
	selectorCaseCount
)

var selectorDispatch = [selectorCaseCount]selectorMakerFunc{
	selectorCaseArrayAll: MakeSelectorArrayAll,
	selectorCaseF00:      MakeSelectorF00,
	selectorCaseF10P:     MakeSelectorF10P,
	selectorCaseF10N:     MakeSelectorF10N,
	selectorCaseF01P:     MakeSelectorF01P,
	selectorCaseF01N:     MakeSelectorF01N,
	selectorCaseF11PP:    MakeSelectorF11PP,
	selectorCaseF11PN:    MakeSelectorF11PN,
	selectorCaseF11NP:    MakeSelectorF11NP,
	selectorCaseF11NN:    MakeSelectorF11NN,
	selectorCaseB00:      MakeSelectorB00,
	selectorCaseB10P:     MakeSelectorB10P,
	selectorCaseB10N:     MakeSelectorB10N,
	selectorCaseB01P:     MakeSelectorB01P,
	selectorCaseB01N:     MakeSelectorB01N,
	selectorCaseB11PP:    MakeSelectorB11PP,
	selectorCaseB11PN:    MakeSelectorB11PN,
	selectorCaseB11NP:    MakeSelectorB11NP,
	selectorCaseB11NN:    MakeSelectorB11NN,
}

// MakeSelectorArrayAll builds a selector for selectorCaseArrayAll
func MakeSelectorArrayAll(_ *SlicePlan) SelectorFunc {
	return func(out []any, node, _ any) []any {
		arr, ok := node.([]any)
		if !ok {
			return out
		}
		return append(out, arr...)
	}
}

// MakeSelectorF00 builds a selector for selectorCaseF00
func MakeSelectorF00(plan *SlicePlan) SelectorFunc {
	return makeSelector0(plan.Step, appendSliceF00)
}

// MakeSelectorF10P builds a selector for selectorCaseF10P
func MakeSelectorF10P(plan *SlicePlan) SelectorFunc {
	return makeSelector1(plan.Start, plan.Step, appendSliceF10P)
}

// MakeSelectorF10N builds a selector for selectorCaseF10N
func MakeSelectorF10N(plan *SlicePlan) SelectorFunc {
	return makeSelector1(plan.Start, plan.Step, appendSliceF10N)
}

// MakeSelectorF01P builds a selector for selectorCaseF01P
func MakeSelectorF01P(plan *SlicePlan) SelectorFunc {
	return makeSelector1(plan.End, plan.Step, appendSliceF01P)
}

// MakeSelectorF01N builds a selector for selectorCaseF01N
func MakeSelectorF01N(plan *SlicePlan) SelectorFunc {
	return makeSelector1(plan.End, plan.Step, appendSliceF01N)
}

// MakeSelectorF11PP builds a selector for selectorCaseF11PP
func MakeSelectorF11PP(plan *SlicePlan) SelectorFunc {
	return makeSelector2(
		plan.Start, plan.End, plan.Step, appendSliceF11PP,
	)
}

// MakeSelectorF11PN builds a selector for selectorCaseF11PN
func MakeSelectorF11PN(plan *SlicePlan) SelectorFunc {
	return makeSelector2(
		plan.Start, plan.End, plan.Step, appendSliceF11PN,
	)
}

// MakeSelectorF11NP builds a selector for selectorCaseF11NP
func MakeSelectorF11NP(plan *SlicePlan) SelectorFunc {
	return makeSelector2(
		plan.Start, plan.End, plan.Step, appendSliceF11NP,
	)
}

// MakeSelectorF11NN builds a selector for selectorCaseF11NN
func MakeSelectorF11NN(plan *SlicePlan) SelectorFunc {
	return makeSelector2(
		plan.Start, plan.End, plan.Step, appendSliceF11NN,
	)
}

// MakeSelectorB00 builds a selector for selectorCaseB00
func MakeSelectorB00(plan *SlicePlan) SelectorFunc {
	return makeSelector0(plan.Step, appendSliceB00)
}

// MakeSelectorB10P builds a selector for selectorCaseB10P
func MakeSelectorB10P(plan *SlicePlan) SelectorFunc {
	return makeSelector1(plan.Start, plan.Step, appendSliceB10P)
}

// MakeSelectorB10N builds a selector for selectorCaseB10N
func MakeSelectorB10N(plan *SlicePlan) SelectorFunc {
	return makeSelector1(plan.Start, plan.Step, appendSliceB10N)
}

// MakeSelectorB01P builds a selector for selectorCaseB01P
func MakeSelectorB01P(plan *SlicePlan) SelectorFunc {
	return makeSelector1(plan.End, plan.Step, appendSliceB01P)
}

// MakeSelectorB01N builds a selector for selectorCaseB01N
func MakeSelectorB01N(plan *SlicePlan) SelectorFunc {
	return makeSelector1(plan.End, plan.Step, appendSliceB01N)
}

// MakeSelectorB11PP builds a selector for selectorCaseB11PP
func MakeSelectorB11PP(plan *SlicePlan) SelectorFunc {
	return makeSelector2(
		plan.Start, plan.End, plan.Step, appendSliceB11PP,
	)
}

// MakeSelectorB11PN builds a selector for selectorCaseB11PN
func MakeSelectorB11PN(plan *SlicePlan) SelectorFunc {
	return makeSelector2(
		plan.Start, plan.End, plan.Step, appendSliceB11PN,
	)
}

// MakeSelectorB11NP builds a selector for selectorCaseB11NP
func MakeSelectorB11NP(plan *SlicePlan) SelectorFunc {
	return makeSelector2(
		plan.Start, plan.End, plan.Step, appendSliceB11NP,
	)
}

// MakeSelectorB11NN builds a selector for selectorCaseB11NN
func MakeSelectorB11NN(plan *SlicePlan) SelectorFunc {
	return makeSelector2(
		plan.Start, plan.End, plan.Step, appendSliceB11NN,
	)
}

func selectorCaseFor(s *SliceExpr) selectorCase {
	if s.Step == 0 {
		return selectorCaseEmpty
	}
	if s.Step == 1 && !s.HasStart && !s.HasEnd {
		return selectorCaseArrayAll
	}
	if s.Step > 0 {
		switch {
		case !s.HasStart && !s.HasEnd:
			return selectorCaseF00
		case s.HasStart && !s.HasEnd:
			if s.Start >= 0 {
				return selectorCaseF10P
			}
			return selectorCaseF10N
		case !s.HasStart && s.HasEnd:
			if s.End >= 0 {
				return selectorCaseF01P
			}
			return selectorCaseF01N
		default:
			switch {
			case s.Start >= 0 && s.End >= 0:
				return selectorCaseF11PP
			case s.Start >= 0 && s.End < 0:
				return selectorCaseF11PN
			case s.Start < 0 && s.End >= 0:
				return selectorCaseF11NP
			default:
				return selectorCaseF11NN
			}
		}
	}

	switch {
	case !s.HasStart && !s.HasEnd:
		return selectorCaseB00
	case s.HasStart && !s.HasEnd:
		if s.Start >= 0 {
			return selectorCaseB10P
		}
		return selectorCaseB10N
	case !s.HasStart && s.HasEnd:
		if s.End >= 0 {
			return selectorCaseB01P
		}
		return selectorCaseB01N
	default:
		switch {
		case s.Start >= 0 && s.End >= 0:
			return selectorCaseB11PP
		case s.Start >= 0 && s.End < 0:
			return selectorCaseB11PN
		case s.Start < 0 && s.End >= 0:
			return selectorCaseB11NP
		default:
			return selectorCaseB11NN
		}
	}
}

func makeSelector0(step int, appendFunc append0Func) SelectorFunc {
	return func(out []any, node, _ any) []any {
		arr, ok := node.([]any)
		if !ok || len(arr) == 0 {
			return out
		}
		return appendFunc(out, arr, step)
	}
}

func makeSelector1(bound, step int, appendFunc append1Func) SelectorFunc {
	return func(out []any, node, _ any) []any {
		arr, ok := node.([]any)
		if !ok || len(arr) == 0 {
			return out
		}
		return appendFunc(out, arr, bound, step)
	}
}

func makeSelector2(start, end, step int, appendFunc append2Func) SelectorFunc {
	return func(out []any, node, _ any) []any {
		arr, ok := node.([]any)
		if !ok || len(arr) == 0 {
			return out
		}
		return appendFunc(out, arr, start, end, step)
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
