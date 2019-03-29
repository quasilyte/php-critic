package main

import (
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/state"
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/name"
	"github.com/z7zmey/php-parser/node/stmt"
	"github.com/z7zmey/php-parser/walker"
)

type metainfoExt struct {
	ctxt *linter.BlockContext

	st *meta.ClassParseState

	// TODO(quasilyte): change key type to *meta.ConstantInfo?
	// But how to get ConstantInfo by *stmt.Constant.ConstantName?
	// solver.GetConstant seem not to work.
	constValue map[string]node.Node
}

func (m *metainfoExt) AfterEnterNode(w walker.Walkable)  {}
func (m *metainfoExt) BeforeLeaveNode(w walker.Walkable) {}

func (m *metainfoExt) AfterLeaveNode(w walker.Walkable) {
	state.LeaveNode(m.st, w)
}

func (m *metainfoExt) BeforeEnterNode(w walker.Walkable) {
	state.EnterNode(m.st, w)

	switch n := w.(type) {
	case *expr.FunctionCall:
		fsym, ok := n.Function.(*name.Name)
		if !ok {
			return
		}
		if !meta.NameEquals(fsym, "define") {
			return
		}
		name := nodeToNameString(m.st, n.Arguments[0])
		m.constValue[name] = n.Arguments[1]
	case *stmt.Constant:
		name := nodeToNameString(m.st, n.ConstantName)
		m.constValue[name] = n.Expr
	}
}
