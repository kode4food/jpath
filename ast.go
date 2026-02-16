package jpath

type (
	// PathExpr is a parsed JSONPath expression
	PathExpr struct {
		Segments []*SegmentExpr
	}

	// SegmentExpr is one child or descendant traversal segment
	SegmentExpr struct {
		Descendant bool
		Selectors  []*SelectorExpr
	}

	// SelectorExpr is one bracket or dot selector operation
	SelectorExpr struct {
		Kind   SelectorKind
		Name   string
		Index  int
		Slice  *SliceExpr
		Filter FilterExpr
	}

	// SliceExpr stores parsed slice bounds and step
	SliceExpr struct {
		HasStart bool
		Start    int
		HasEnd   bool
		End      int
		Step     int
	}

	// FilterExpr is the marker interface for filter AST nodes
	FilterExpr interface {
		filterExpr()
	}

	// LiteralExpr is a scalar literal in a filter expression
	LiteralExpr struct {
		Value any
	}

	// PathValueExpr is a root or current-node relative path in a filter
	PathValueExpr struct {
		Absolute bool
		Path     *PathExpr
	}

	// UnaryExpr is a unary filter expression
	UnaryExpr struct {
		Op   string
		Expr FilterExpr
	}

	// BinaryExpr is a binary filter expression
	BinaryExpr struct {
		Op          string
		Left, Right FilterExpr
	}

	// FuncExpr is a filter function call expression
	FuncExpr struct {
		Name string
		Args []FilterExpr
	}

	// SelectorKind identifies the selector variant in SelectorExpr
	SelectorKind uint8
)

const (
	SelectorName     SelectorKind = iota // object member by name
	SelectorIndex                        // array index
	SelectorWildcard                     // all direct child values
	SelectorSlice                        // array elements by slice bounds
	SelectorFilter                       // child values by filter predicate
)

func (l *LiteralExpr) filterExpr()   {}
func (p *PathValueExpr) filterExpr() {}
func (u *UnaryExpr) filterExpr()     {}
func (b *BinaryExpr) filterExpr()    {}
func (f *FuncExpr) filterExpr()      {}
