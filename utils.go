package main

import (
	"strconv"

	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/expr/binary"
	"github.com/z7zmey/php-parser/node/scalar"
)

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

func constIntValue(x node.Node) (int, bool) {
	switch x := x.(type) {
	case *expr.UnaryMinus:
		v, ok := constIntValue(x.Expr)
		return -v, ok
	case *scalar.Lnumber:
		v, err := strconv.Atoi(x.Value)
		return v, err == nil
	case *binary.Plus:
		a, ok1 := constIntValue(x.Left)
		b, ok2 := constIntValue(x.Right)
		return a + b, ok1 && ok2
	default:
		return 0, false
	}
}
