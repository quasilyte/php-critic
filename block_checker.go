package main

import (
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/state"
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/expr/binary"
	"github.com/z7zmey/php-parser/node/name"
	"github.com/z7zmey/php-parser/node/scalar"
	"github.com/z7zmey/php-parser/walker"
)

// TODO(quasilyte): badCond:
// - find redundant expressions, like `a == b && a == b`.
// - solve Equal-vs-Identical question
// - handle non-int const exprs
// - handle consts like false, null, etc

type blockChecker struct {
	ctxt linter.BlockContext
	mi   *metainfoExt
}

func (c *blockChecker) AfterEnterNode(w walker.Walkable)  {}
func (c *blockChecker) BeforeLeaveNode(w walker.Walkable) {}

func (c *blockChecker) AfterLeaveNode(w walker.Walkable) {
	state.LeaveNode(c.mi.st, w)
}

func (c *blockChecker) BeforeEnterNode(w walker.Walkable) {
	state.EnterNode(c.mi.st, w)

	switch n := w.(type) {
	case *expr.FunctionCall:
		c.handleFunctionCall(n)
	case *binary.BooleanAnd:
		c.handleBooleanAnd(n)
	case *binary.BooleanOr:
		c.handleBooleanOr(n)
	}
}

func (c *blockChecker) handleBooleanOr(cond *binary.BooleanOr) {
	cv, ok := constFold(c.mi, cond).(BoolConst)
	if !ok {
		c.handleBooleanOrNeqNeq(cond)
		return
	}
	if cv {
		c.ctxt.Report(cond, linter.LevelWarning, "badCond", "always true condition")
	} else {
		c.ctxt.Report(cond, linter.LevelWarning, "badCond", "always false condition")
	}
}

func (c *blockChecker) handleBooleanOrNeqNeq(cond *binary.BooleanOr) {
	lhs, ok := cond.Left.(*binary.NotEqual)
	if !ok {
		return
	}
	rhs, ok := cond.Right.(*binary.NotEqual)
	if !ok {
		return
	}
	if !sameSimpleExpr(lhs.Left, rhs.Left) {
		return
	}

	x := constFold(c.mi, lhs.Right)
	y := constFold(c.mi, rhs.Right)
	res, ok := constEqual(x, y).(BoolConst)
	if ok && !bool(res) {
		c.ctxt.Report(cond, linter.LevelWarning, "badCond", "always true condition")
	}
}

func (c *blockChecker) handleBooleanAndLtGt(cond *binary.BooleanAnd) {
	lhs, ok := cond.Left.(*binary.Smaller)
	if !ok {
		return
	}
	rhs, ok := cond.Right.(*binary.Greater)
	if !ok {
		return
	}
	if !sameSimpleExpr(lhs.Left, rhs.Left) {
		return
	}

	x := constFold(c.mi, lhs.Right)
	y := constFold(c.mi, rhs.Right)
	res, ok := constLessThan(x, y).(BoolConst)
	if ok && bool(res) {
		c.ctxt.Report(cond, linter.LevelWarning, "badCond", "always false condition")
	}
}

func (c *blockChecker) handleBooleanAndEqEq(cond *binary.BooleanAnd) {
	lhs, ok := cond.Left.(*binary.Equal)
	if !ok {
		return
	}
	rhs, ok := cond.Right.(*binary.Equal)
	if !ok {
		return
	}
	if !sameSimpleExpr(lhs.Left, rhs.Left) {
		return
	}

	x := constFold(c.mi, lhs.Right)
	y := constFold(c.mi, rhs.Right)
	res, ok := constEqual(x, y).(BoolConst)
	if ok && !bool(res) {
		c.ctxt.Report(cond, linter.LevelWarning, "badCond", "always false condition")
	}
}

func (c *blockChecker) handleBooleanAnd(cond *binary.BooleanAnd) {
	cv, ok := constFold(c.mi, cond).(BoolConst)
	if !ok {
		c.handleBooleanAndLtGt(cond)
		c.handleBooleanAndEqEq(cond)
		return
	}
	if cv {
		c.ctxt.Report(cond, linter.LevelWarning, "badCond", "always true condition")
	} else {
		c.ctxt.Report(cond, linter.LevelWarning, "badCond", "always false condition")
	}
}

func (c *blockChecker) handleFunctionCall(call *expr.FunctionCall) {
	name, ok := call.Function.(*name.Name)
	if !ok {
		return
	}

	switch {
	case meta.NameEquals(name, "define"):
		c.checkDefine(call)
	case meta.NameEquals(name, "strpos"):
		c.checkStrpos(call)
	}
}

func (c *blockChecker) checkDefine(define *expr.FunctionCall) {
	if len(define.Arguments) != 2 {
		c.ctxt.Report(define.Arguments[2], linter.LevelWarning, "sloppyArg", "don't use case_insensitive argument")
	}
}

func (c *blockChecker) checkStrpos(call *expr.FunctionCall) {
	if len(call.Arguments) != 2 {
		return
	}
	str := call.Arguments[0].(*node.Argument).Expr
	substr := call.Arguments[1].(*node.Argument).Expr
	if c.isStringLit(str) && !c.isStringLit(substr) {
		c.ctxt.Report(call, linter.LevelWarning, "argOrder", "suspicious args order")
	}
}

func (c *blockChecker) isStringLit(n node.Node) bool {
	_, ok := n.(*scalar.String)
	return ok
}
