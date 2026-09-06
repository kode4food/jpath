package jpath

// Nodes represents a JSONPath node list
type Nodes []any

func valueCount(value any) int {
	if nodes, ok := value.(Nodes); ok {
		return len(nodes)
	}
	return 1
}

func singularValue(value any) (any, bool) {
	nodes, ok := value.(Nodes)
	if !ok {
		return value, true
	}
	if len(nodes) != 1 {
		return nil, false
	}
	return nodes[0], true
}
