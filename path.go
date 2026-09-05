package jpath

import "slices"

type (
	// Path is a compiled query function chain. Calling it executes the query
	// against a JSON document
	Path func(document any) []any

	// SegmentFunc processes input nodes and returns output nodes for one
	// segment step
	SegmentFunc func(in []any, root any) []any

	// SelectorFunc appends selected output nodes for one input node
	SelectorFunc func(out []any, node, root any) []any
)

// ComposePath composes segment functions into an executable Path
func ComposePath(segments ...SegmentFunc) Path {
	chain := composeSegments(segments)
	return func(document any) []any {
		return chain([]any{document}, document)
	}
}

// ChildSegment builds a non-descendant segment from selector functions
func ChildSegment(selectors ...SelectorFunc) SegmentFunc {
	return composeSegment(selectors, false)
}

// DescendantSegment builds a descendant segment from selector functions
func DescendantSegment(selectors ...SelectorFunc) SegmentFunc {
	return composeSegment(selectors, true)
}

func composeSegments(segments []SegmentFunc) SegmentFunc {
	if len(segments) == 1 {
		return segments[0]
	}
	segments = slices.Clone(segments)
	return func(in []any, root any) []any {
		for _, seg := range segments {
			in = seg(in, root)
		}
		return in
	}
}

func composeSegment(selectors []SelectorFunc, descendant bool) SegmentFunc {
	chain := composeSelectors(selectors)
	if descendant {
		return func(in []any, root any) []any {
			desc := descendantsOf(in)
			out := make([]any, 0)
			for _, node := range desc {
				out = chain(out, node, root)
			}
			return out
		}
	}
	return func(in []any, root any) []any {
		out := make([]any, 0)
		for _, node := range in {
			out = chain(out, node, root)
		}
		return out
	}
}

func composeSelectors(selectors []SelectorFunc) SelectorFunc {
	if len(selectors) == 1 {
		return selectors[0]
	}
	selectors = slices.Clone(selectors)
	return func(out []any, node, root any) []any {
		for _, selector := range selectors {
			out = selector(out, node, root)
		}
		return out
	}
}

func selectorIdentity(out []any, _, _ any) []any {
	return out
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
