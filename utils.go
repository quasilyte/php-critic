package main

import (
	"strconv"

	"github.com/VKCOM/noverify/src/meta"
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/expr/binary"
	"github.com/z7zmey/php-parser/node/name"
	"github.com/z7zmey/php-parser/node/scalar"
)

func nodeToNameString(st *meta.ClassParseState, n node.Node) string {
	switch n := n.(type) {
	case *node.Identifier:
		return st.Namespace + `\` + n.Value
	case *name.Name:
		return st.Namespace + `\` + meta.NameToString(n)
	default:
		return ""
	}
}

func sameSimpleExpr(a, b node.Node) bool {
	// TODO(quasilyte): handle const exprs?

	switch a := a.(type) {
	case *expr.Variable:
		b, ok := b.(*expr.Variable)
		return ok && sameSimpleExpr(a.VarName, b.VarName)
	case *node.Identifier:
		b, ok := b.(*node.Identifier)
		return ok && a.Value == b.Value
	case *scalar.Lnumber:
		b, ok := b.(*scalar.Lnumber)
		return ok && a.Value == b.Value
	case *binary.NotEqual:
		b, ok := b.(*binary.NotEqual)
		return ok &&
			sameSimpleExpr(a.Left, b.Left) &&
			sameSimpleExpr(a.Right, b.Right)
	}
	return false
}

func constIntValue(mi *metainfoExt, x node.Node) (int, bool) {
	switch x := x.(type) {
	case *expr.UnaryMinus:
		v, ok := constIntValue(mi, x.Expr)
		return -v, ok
	case *scalar.Lnumber:
		v, err := strconv.Atoi(x.Value)
		return v, err == nil
	case *binary.Plus:
		a, ok1 := constIntValue(mi, x.Left)
		b, ok2 := constIntValue(mi, x.Right)
		return a + b, ok1 && ok2
	case *expr.ConstFetch:
		name := nodeToNameString(mi.st, x.Constant)
		return constIntValue(mi, mi.constValue[name])
	default:
		return 0, false
	}
}
