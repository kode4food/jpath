package jpath

import (
	"reflect"
	"strings"
)

type (
	evalValue struct {
		kind   evalKind
		scalar any
		nodes  []any
	}

	evalCtx struct {
		root    any
		current any
	}

	nothingType struct{}

	evalKind uint8

	compiledPathValueExpr struct {
		absolute bool
		run      Runnable
	}

	compiledFuncExpr struct {
		evaluator FunctionEvaluator
		args      []FilterExpr
	}
)

const (
	evalScalar evalKind = iota
	evalNodes
)

var nothing nothingType

func (l LiteralExpr) eval(_ *evalCtx) evalValue {
	return scalarValue(l.Value)
}

func (p PathValueExpr) eval(ctx *evalCtx) evalValue {
	run, err := makeRunnable(p.Path, NewRegistry())
	if err != nil {
		return nodesValue(nil)
	}
	base := ctx.current
	if p.Absolute {
		base = ctx.root
	}
	res := run.Run(base)
	return nodesValue(res)
}

func (p compiledPathValueExpr) eval(ctx *evalCtx) evalValue {
	base := ctx.current
	if p.absolute {
		base = ctx.root
	}
	res := p.run.Run(base)
	return nodesValue(res)
}

func (u UnaryExpr) eval(ctx *evalCtx) evalValue {
	v := u.Expr.eval(ctx)
	if u.Op == "!" {
		return scalarValue(!toBool(v))
	}
	return scalarValue(nil)
}

func (b BinaryExpr) eval(ctx *evalCtx) evalValue {
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

func (f FuncExpr) eval(ctx *evalCtx) evalValue {
	def, ok := builtinFunction(f.Name)
	if !ok {
		return scalarValue(nil)
	}
	return fromFunctionValue(def.Eval(evalFunctionArgs(f.Args, ctx)))
}

func (f compiledFuncExpr) eval(ctx *evalCtx) evalValue {
	return fromFunctionValue(f.evaluator(evalFunctionArgs(f.args, ctx)))
}

func scalarValue(v any) evalValue {
	return evalValue{kind: evalScalar, scalar: v}
}

func nodesValue(v []any) evalValue {
	return evalValue{kind: evalNodes, nodes: v}
}

func evalFunctionArgs(args []FilterExpr, ctx *evalCtx) []FunctionValue {
	res := make([]FunctionValue, len(args))
	for idx, arg := range args {
		res[idx] = toFunctionValue(arg.eval(ctx))
	}
	return res
}

func toFunctionValue(v evalValue) FunctionValue {
	if v.kind == evalNodes {
		return FunctionValue{
			IsNodes: true,
			Nodes:   v.nodes,
		}
	}
	return FunctionValue{Scalar: v.scalar}
}

func fromFunctionValue(v FunctionValue) evalValue {
	if v.IsNodes {
		return nodesValue(v.Nodes)
	}
	return scalarValue(v.Scalar)
}

func compareValues(left evalValue, right evalValue, op string) bool {
	lc := expandCandidates(left)
	rc := expandCandidates(right)
	if len(lc) == 0 || len(rc) == 0 {
		switch op {
		case "==":
			if len(lc) == 0 && len(rc) == 0 {
				return true
			}
			if len(lc) == 0 && isNothing(rc) {
				return true
			}
			if len(rc) == 0 && isNothing(lc) {
				return true
			}
			return false
		case "!=":
			if len(lc) == 0 && len(rc) == 0 {
				return false
			}
			if len(lc) == 0 && isNothing(rc) {
				return false
			}
			if len(rc) == 0 && isNothing(lc) {
				return false
			}
			return true
		default:
			return false
		}
	}
	for _, lv := range lc {
		for _, rv := range rc {
			ok, matched := comparePair(lv, rv, op)
			if ok && matched {
				return true
			}
		}
	}
	return false
}

func comparePair(left any, right any, op string) (bool, bool) {
	switch op {
	case "==":
		return true, reflect.DeepEqual(left, right)
	case "!=":
		return true, !reflect.DeepEqual(left, right)
	case "<=", ">=":
		if reflect.DeepEqual(left, right) {
			return true, true
		}
	}
	lf, lok := asNumber(left)
	rf, rok := asNumber(right)
	if lok && rok {
		switch op {
		case "<":
			return true, lf < rf
		case "<=":
			return true, lf <= rf
		case ">":
			return true, lf > rf
		case ">=":
			return true, lf >= rf
		}
	}
	ls, lok := left.(string)
	rs, rok := right.(string)
	if lok && rok {
		switch op {
		case "<":
			return true, ls < rs
		case "<=":
			return true, ls <= rs
		case ">":
			return true, ls > rs
		case ">=":
			return true, ls >= rs
		}
	}
	return false, false
}

func expandCandidates(v evalValue) []any {
	if v.kind == evalNodes {
		return v.nodes
	}
	return []any{v.scalar}
}

func toBool(v evalValue) bool {
	if v.kind == evalNodes {
		return len(v.nodes) > 0
	}
	switch raw := v.scalar.(type) {
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

func asNumber(v any) (float64, bool) {
	if n, ok := v.(float64); ok {
		return n, true
	}
	if n, ok := v.(int); ok {
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
