package main

import (
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/VKCOM/noverify/src/meta"
	"github.com/quasilyte/php-critic/internal/constant"
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/expr/binary"
	"github.com/z7zmey/php-parser/node/name"
	"github.com/z7zmey/php-parser/node/scalar"
)

// FIXME: is *scalar.String actually ever contain unescaped $ signs?
func isDynamicString(lit *scalar.String) bool {
	if !strings.HasPrefix(lit.Value, `"`) {
		return false
	}
	dollars := strings.Count(lit.Value, "$")
	escapedDollars := strings.Count(lit.Value, `\$`)
	return dollars != escapedDollars
}

func constFold(mi *metainfoExt, e node.Node) constant.Value {
	switch e := e.(type) {
	case *node.Argument:
		return constFold(mi, e.Expr)

	case *binary.Plus:
		return constant.Add(constFold(mi, e.Left), constFold(mi, e.Right))
	case *binary.Minus:
		return constant.Sub(constFold(mi, e.Left), constFold(mi, e.Right))
	case *expr.UnaryMinus:
		return constant.Neg(constFold(mi, e.Expr))

	case *binary.Smaller:
		return constant.LessThan(constFold(mi, e.Left), constFold(mi, e.Right))
	case *binary.Greater:
		return constant.GreaterThan(constFold(mi, e.Left), constFold(mi, e.Right))
	case *expr.BooleanNot:
		return constant.Not(constFold(mi, e.Expr))
	case *binary.BooleanAnd:
		return constant.And(constFold(mi, e.Left), constFold(mi, e.Right))
	case *binary.BooleanOr:
		return constant.Or(constFold(mi, e.Left), constFold(mi, e.Right))

	case *binary.Equal:
		return constant.Equal(constFold(mi, e.Left), constFold(mi, e.Right))
	case *binary.Identical:
		return constant.Identical(constFold(mi, e.Left), constFold(mi, e.Right))

	case *expr.ConstFetch:
		name := nodeToNameString(mi.st, e.Constant)
		return constFold(mi, mi.constValue[name])

	case *binary.Concat:
		return constant.Concat(constFold(mi, e.Left), constFold(mi, e.Right))
	case *scalar.String:
		if isDynamicString(e) {
			return constant.UnknownValue{}
		}
		unquoted := e.Value[1 : len(e.Value)-1]
		s, ok := interpretString(unquoted, e.Value[0])
		if !ok {
			return constant.UnknownValue{}
		}
		return constant.StringValue(s)
	case *scalar.Dnumber:
		v, err := strconv.ParseFloat(e.Value, 64)
		if err == nil {
			return constant.FloatValue(v)
		}
	case *scalar.Lnumber:
		v, err := strconv.ParseInt(e.Value, 10, 64)
		if err == nil {
			return constant.IntValue(v)
		}
	}
	return constant.UnknownValue{}
}

func interpretString(s string, quote byte) (string, bool) {
	switch quote {
	case '\'', '"':
		// OK
	default:
		return "", false
	}

	if !strings.Contains(s, `\`) {
		// Fast path.
		return s, true
	}

	var out strings.Builder
	i := 0
	for i < len(s) {
		ch := s[i]
		switch {
		case ch == '\\':
			if len(s) < i+1 {
				return "", false
			}
			switch s[i+1] {
			case '\'':
				if quote == '"' {
					out.WriteString(`\'`)
				} else {
					out.WriteByte('\'')
				}
				i += 2
			case '"':
				if quote == '"' {
					out.WriteByte('"')
				} else {
					out.WriteString(`\"`)
				}
				i += 2
			case '$':
				if quote == '"' {
					out.WriteByte('$')
				} else {
					out.WriteString(`\$`)
				}
				i += 2
			case 'n':
				if quote == '"' {
					out.WriteByte('\n')
				} else {
					out.WriteString(`\n`)
				}
				i += 2
			case 'r':
				if quote == '"' {
					out.WriteByte('\r')
				} else {
					out.WriteString(`\r`)
				}
				i += 2
			case 't':
				if quote == '"' {
					out.WriteByte('\r')
				} else {
					out.WriteString(`\r`)
				}
				i += 2
			case '\\':
				out.WriteByte(s[i+1])
				i += 2
			case 'x':
				if quote == '"' {
					if len(s) < i+3 {
						return "", false
					}
					v, err := strconv.ParseInt(s[i+2:i+4], 16, 64)
					if err != nil || v > 255 {
						return "", false
					}
					out.WriteByte(byte(v))
					i += 4
				} else {
					out.WriteString(`\x`)
					i += 2
				}
			default:
				return "", false
			}
		case ch <= unicode.MaxASCII:
			out.WriteByte(ch)
			i++
		default:
			r, n := utf8.DecodeRuneInString(s[i:])
			out.WriteRune(r)
			i += n
		}
	}
	return out.String(), true
}

func nodeToNameString(st *meta.ClassParseState, n node.Node) string {
	switch n := n.(type) {
	case *node.Identifier:
		return st.Namespace + `\` + n.Value
	case *name.Name:
		return st.Namespace + `\` + meta.NameToString(n)
	case *node.Argument:
		return nodeToNameString(st, n.Expr)
	case *scalar.String:
		return `\` + n.Value[1:len(n.Value)-1] // Unquoted
	default:
		return ""
	}
}

func sameNode(a, b node.Node) bool {
	return sameSimpleExpr(a, b) || reflect.DeepEqual(a, b)
}

func sameSimpleExpr(a, b node.Node) bool {
	// TODO(quasilyte): handle const exprs?
	switch a := a.(type) {
	case *node.Argument:
		b, ok := b.(*node.Argument)
		return ok && sameSimpleExpr(a.Expr, b.Expr)
	case *expr.ArrayDimFetch:
		b, ok := b.(*expr.ArrayDimFetch)
		return ok &&
			sameSimpleExpr(a.Variable, b.Variable) &&
			sameSimpleExpr(a.Dim, b.Dim)
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
	// TODO: handle constants?
	return false
}

func intMax(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func intMin(x, y int) int {
	if x < y {
		return x
	}
	return y
}
