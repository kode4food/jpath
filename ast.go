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
		eval(*evalCtx) *evalValue
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

func (l *LiteralExpr) eval(_ *evalCtx) *evalValue {
	return scalarValue(l.Value)
}

func (p *PathValueExpr) eval(ctx *evalCtx) *evalValue {
	path, err := makePath(p.Path, NewRegistry())
	if err != nil {
		return nodesValue(nil)
	}
	base := ctx.current
	if p.Absolute {
		base = ctx.root
	}
	return nodesValue(path.Query(base))
}

func (u *UnaryExpr) eval(ctx *evalCtx) *evalValue {
	v := u.Expr.eval(ctx)
	if u.Op == "!" {
		return scalarValue(!toBool(v))
	}
	return scalarValue(nil)
}

func (b *BinaryExpr) eval(ctx *evalCtx) *evalValue {
	if b.Op == "&&" {
		left := b.Left.eval(ctx)
		if !toBool(left) {
			return scalarValue(false)
		}
		return scalarValue(toBool(b.Right.eval(ctx)))
	}
	if b.Op == "||" {
		left := b.Left.eval(ctx)
		if toBool(left) {
			return scalarValue(true)
		}
		return scalarValue(toBool(b.Right.eval(ctx)))
	}
	left := b.Left.eval(ctx)
	right := b.Right.eval(ctx)
	return scalarValue(compareValues(left, right, b.Op))
}

func (f *FuncExpr) eval(ctx *evalCtx) *evalValue {
	def, ok := builtinFunction(f.Name)
	if !ok {
		return scalarValue(nil)
	}
	return fromFunctionValue(def.Eval(evalFunctionArgs(f.Args, ctx)))
}
