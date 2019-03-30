package main

import (
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/state"
	"github.com/quasilyte/php-critic/internal/constant"
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/expr/binary"
	"github.com/z7zmey/php-parser/node/name"
	"github.com/z7zmey/php-parser/node/scalar"
	"github.com/z7zmey/php-parser/node/stmt"
	"github.com/z7zmey/php-parser/walker"
)

type blockChecker struct {
	ctxt *linter.BlockContext
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
		c.checkBadCond(n)
		c.handleDupSubExpr(n, n.Left, n.Right, "!==")
	case *binary.NotEqual:
		c.checkBadCond(n)
		c.handleDupSubExpr(n, n.Left, n.Right, "!=")
	case *binary.Identical:
		c.checkBadCond(n)
		c.handleDupSubExpr(n, n.Left, n.Right, "===")
		c.handleIdentical(n)
	case *binary.Equal:
		c.checkBadCond(n)
		c.handleDupSubExpr(n, n.Left, n.Right, "==")
	case *binary.Smaller:
		c.checkBadCond(n)
		c.handleDupSubExpr(n, n.Left, n.Right, "<")
		c.handleSmaller(n)
	case *binary.SmallerOrEqual:
		c.checkBadCond(n)
		c.handleDupSubExpr(n, n.Left, n.Right, "<=")
	case *binary.GreaterOrEqual:
		c.checkBadCond(n)
		c.handleDupSubExpr(n, n.Left, n.Right, ">=")
	case *binary.Greater:
		c.checkBadCond(n)
		c.handleDupSubExpr(n, n.Left, n.Right, ">")
		c.handleGreater(n)
	case *binary.BooleanAnd:
		c.handleBooleanAnd(n)
	case *binary.BooleanOr:
		c.handleBooleanOr(n)
	case *stmt.If:
		c.handleIf(n)
	case *stmt.Do:
		c.handleDoWhile(n)
	case *stmt.Switch:
		c.handleSwitch(n)
	}
}

func (c *blockChecker) handleGreater(cmp *binary.Greater) {
	cv, ok := constFold(c.mi, cmp.Right).(constant.IntValue)
	if ok && cv == 0 {
		strcmp, ok := cmp.Left.(*expr.FunctionCall)
		if ok && meta.NameNodeToString(strcmp.Function) == "strcmp" {
			c.ctxt.Report(cmp, linter.LevelDoNotReject, "simplify",
				"can replace 'strcmp(s1, s2) > 0' with 's1 > s2'")
		}
	}
}

func (c *blockChecker) handleSmaller(cmp *binary.Smaller) {
	cv, ok := constFold(c.mi, cmp.Right).(constant.IntValue)
	if ok && cv == 0 {
		strcmp, ok := cmp.Left.(*expr.FunctionCall)
		if ok && meta.NameNodeToString(strcmp.Function) == "strcmp" {
			c.ctxt.Report(cmp, linter.LevelDoNotReject, "simplify",
				"can replace 'strcmp(s1, s2) < 0' with 's1 < s2'")
		}
	}
}

func (c *blockChecker) handleIdentical(eq *binary.Identical) {
	cv, ok := constFold(c.mi, eq.Right).(constant.IntValue)
	if ok && cv == 0 {
		// Handle `strcmp($s1, $s2) === 0`.
		strcmp, ok := eq.Left.(*expr.FunctionCall)
		if ok && meta.NameNodeToString(strcmp.Function) == "strcmp" {
			c.ctxt.Report(eq, linter.LevelDoNotReject, "simplify",
				"can replace 'strcmp(s1, s2) === 0' with 's1 === s2'")
		}
	}
}

func (c *blockChecker) checkBadCond(cond node.Node) bool {
	cv, ok := constFold(c.mi, cond).(constant.BoolValue)
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
	bodies := make([]node.Node, 0, 2+len(ifstmt.ElseIf))
	bodies = append(bodies, ifstmt.Stmt)
	for _, elseif := range ifstmt.ElseIf {
		bodies = append(bodies, elseif.(*stmt.ElseIf).Stmt)
	}
	if ifstmt.Else != nil {
		bodies = append(bodies, ifstmt.Else.(*stmt.Else).Stmt)
	}
	for i, b1 := range bodies {
		for j, b2 := range bodies[i+1:] {
			j += i + 1
			if sameNode(b1, b2) {
				c.ctxt.Report(ifstmt, linter.LevelWarning, "dupBranchBody",
					"duplicated <%d> and <%d> bodies", i, j)
			}
		}
	}
}

func (c *blockChecker) handleDoWhile(while *stmt.Do) {
	c.checkBadCond(while.Cond)
}

func (c *blockChecker) handleSwitch(swt *stmt.Switch) {
	for _, cas := range swt.Cases {
		cas, ok := cas.(*stmt.Case)
		if !ok {
			continue // Could be *stmt.Default
		}
		c.checkBadCond(cas.Cond)
	}
}

func (c *blockChecker) checkDupArg(call *expr.FunctionCall, i, j int) {
	if len(call.Arguments) <= intMax(i, j) {
		return
	}
	x := call.Arguments[i]
	y := call.Arguments[j]
	if sameSimpleExpr(x, y) {
		c.ctxt.Report(x, linter.LevelWarning, "dupArg", "suspiciously duplicated argument")
	}
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
	res, ok := constant.Equal(x, y).(constant.BoolValue)
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
	res, ok := constant.LessThan(x, y).(constant.BoolValue)
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
	res, ok := constant.Equal(x, y).(constant.BoolValue)
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

	switch meta.NameNodeToString(name) {
	case "define":
		c.checkDefine(call)
	case "strpos":
		c.checkStrpos(call)
		c.checkDupArg(call, 0, 1)
	case "strncmp":
		c.checkStrncmp(call)
		c.checkDupArg(call, 0, 1)
	case "strcmp", "min", "max", "str_replace":
		c.checkDupArg(call, 0, 1)
	}
}

func (c *blockChecker) checkStrncmp(strncmp *expr.FunctionCall) {
	if len(strncmp.Arguments) != 3 {
		return
	}
	cv1, ok1 := constFold(c.mi, strncmp.Arguments[0]).(constant.StringValue)
	cv2, ok2 := constFold(c.mi, strncmp.Arguments[1]).(constant.StringValue)
	validLen := 0
	switch {
	case ok1 && ok2:
		validLen = intMin(len(cv1), len(cv2))
	case ok1:
		validLen = len(cv1)
	case ok2:
		validLen = len(cv2)
	default:
		return
	}
	length, ok := constFold(c.mi, strncmp.Arguments[2]).(constant.IntValue)
	if !ok {
		return
	}
	if int(length) != validLen {
		c.ctxt.Report(strncmp.Arguments[2], linter.LevelWarning, "badCall",
			"expected length arg to be %d, got %d", validLen, length)
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
