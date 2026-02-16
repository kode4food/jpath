package jpath

type (
	// Value stores either a scalar or node list argument/result
	Value struct {
		IsNodes bool
		Scalar  any
		Nodes   []any
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

// Count reports the number of candidates this value contributes to filter
// comparisons. For node values, this is len(Nodes). For scalar values, this
// is always 1
func (v *Value) Count() int {
	if v.IsNodes {
		return len(v.Nodes)
	}
	return 1
}

// IsNothing reports whether this value is the internal "nothing" sentinel. A
// node-list value is nothing only when it contains exactly one sentinel
func (v *Value) IsNothing() bool {
	if v.IsNodes {
		if len(v.Nodes) != 1 {
			return false
		}
		_, ok := v.Nodes[0].(nothingType)
		return ok
	}
	_, ok := v.Scalar.(nothingType)
	return ok
}

func (v *Value) singularValue() (any, bool) {
	if v.IsNodes {
		if len(v.Nodes) != 1 {
			return nil, false
		}
		return v.Nodes[0], true
	}
	return v.Scalar, true
}
