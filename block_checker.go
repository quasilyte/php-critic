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
	"github.com/z7zmey/php-parser/node/stmt"
	"github.com/z7zmey/php-parser/walker"
)

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
	case *binary.Div:
		c.handleDupSubExpr(n, n.Left, n.Right, "/")
	case *binary.Mod:
		c.handleDupSubExpr(n, n.Left, n.Right, "%")
	case *binary.Minus:
		c.handleDupSubExpr(n, n.Left, n.Right, "-")
	case *binary.NotIdentical:
		c.handleDupSubExpr(n, n.Left, n.Right, "!==")
	case *binary.NotEqual:
		c.handleDupSubExpr(n, n.Left, n.Right, "!=")
	case *binary.Identical:
		c.handleDupSubExpr(n, n.Left, n.Right, "===")
	case *binary.Equal:
		c.handleDupSubExpr(n, n.Left, n.Right, "==")
	case *binary.Smaller:
		c.handleDupSubExpr(n, n.Left, n.Right, "<")
	case *binary.SmallerOrEqual:
		c.handleDupSubExpr(n, n.Left, n.Right, "<=")
	case *binary.GreaterOrEqual:
		c.handleDupSubExpr(n, n.Left, n.Right, ">=")
	case *binary.Greater:
		c.handleDupSubExpr(n, n.Left, n.Right, ">")
	case *binary.BooleanAnd:
		c.handleBooleanAnd(n)
	case *binary.BooleanOr:
		c.handleBooleanOr(n)
	case *stmt.If:
		c.handleIf(n)
	case *stmt.While:
		c.handleWhile(n)
	case *stmt.Do:
		c.handleDoWhile(n)
	}
}

func (c *blockChecker) checkBadCond(cond node.Node) bool {
	cv, ok := constFold(c.mi, cond).(BoolConst)
	if !ok {
		return false
	}
	if cv {
		c.ctxt.Report(cond, linter.LevelWarning, "badCond", "always true condition")
	} else {
		c.ctxt.Report(cond, linter.LevelWarning, "badCond", "always false condition")
	}
	return true
}

func (c *blockChecker) handleIf(ifstmt *stmt.If) {
	// Recognize `if (false) {}` and skip it.
	ifFalse := false
	if fetch, ok := ifstmt.Cond.(*expr.ConstFetch); ok {
		ifFalse = meta.NameNodeToString(fetch.Constant) == "false"
	}
	if !ifFalse {
		c.checkBadCond(ifstmt.Cond)
	}
}

func (c *blockChecker) handleWhile(while *stmt.While) {
	// Recognize `while (true) {}` and skip it.
	whileTrue := false
	if fetch, ok := while.Cond.(*expr.ConstFetch); ok {
		whileTrue = meta.NameNodeToString(fetch.Constant) == "true"
	}
	if !whileTrue {
		c.checkBadCond(while.Cond)
	}
}

func (c *blockChecker) handleDoWhile(while *stmt.Do) {
	c.checkBadCond(while.Cond)
}

func (c *blockChecker) handleDupSubExpr(n node.Node, lhs, rhs node.Node, op string) {
	if sameSimpleExpr(lhs, rhs) {
		c.ctxt.Report(n, linter.LevelWarning, "dupSubExpr", "suspiciously duplicated LHS and RHS of '%s'", op)
	}
}

func (c *blockChecker) handleBooleanOr(cond *binary.BooleanOr) {
	if !c.checkBadCond(cond) {
		c.handleBooleanOrNeqNeq(cond)
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
	if !c.checkBadCond(cond) {
		c.handleBooleanAndLtGt(cond)
		c.handleBooleanAndEqEq(cond)
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
