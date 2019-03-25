package main

import (
	"strconv"
	"strings"

	"github.com/VKCOM/noverify/src/meta"
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/expr/binary"
	"github.com/z7zmey/php-parser/node/name"
	"github.com/z7zmey/php-parser/node/scalar"
)

func isDynamicString(lit *scalar.String) bool {
	if !strings.HasPrefix(lit.Value, `"`) {
		return false
	}
	dollars := strings.Count(lit.Value, "$")
	escapedDollars := strings.Count(lit.Value, `\$`)
	return dollars != escapedDollars
}

func constFold(mi *metainfoExt, e node.Node) Constant {
	switch e := e.(type) {
	case *binary.Smaller:
		return constLessThan(constFold(mi, e.Left), constFold(mi, e.Right))
	case *binary.Greater:
		return constGreaterThan(constFold(mi, e.Left), constFold(mi, e.Right))
	case *binary.BooleanAnd:
		return constAnd(constFold(mi, e.Left), constFold(mi, e.Right))
	case *binary.BooleanOr:
		return constOr(constFold(mi, e.Left), constFold(mi, e.Right))
	case *binary.Equal:
		return constEqual(constFold(mi, e.Left), constFold(mi, e.Right))
	case *binary.Identical:
		return constIdentical(constFold(mi, e.Left), constFold(mi, e.Right))
	case *binary.Plus:
		return constAdd(constFold(mi, e.Left), constFold(mi, e.Right))
	case *expr.UnaryMinus:
		return constNegate(constFold(mi, e.Expr))
	case *expr.ConstFetch:
		name := nodeToNameString(mi.st, e.Constant)
		return constFold(mi, mi.constValue[name])
	case *scalar.String:
		if isDynamicString(e) {
			return UnknownConst{}
		}
		unquoted := e.Value[1 : len(e.Value)-1]
		return StringConst(unquoted)
	case *scalar.Dnumber:
		v, err := strconv.ParseFloat(e.Value, 64)
		if err == nil {
			return FloatConst(v)
		}
	case *scalar.Lnumber:
		v, err := strconv.ParseInt(e.Value, 10, 64)
		if err == nil {
			return IntConst(v)
		}
	}
	return UnknownConst{}
}

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
