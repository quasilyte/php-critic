package main

import (
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/name"
	"github.com/z7zmey/php-parser/walker"
)

type blockChecker struct {
	ctxt linter.BlockContext
}

func (c *blockChecker) AfterEnterNode(w walker.Walkable)  {}
func (c *blockChecker) BeforeLeaveNode(w walker.Walkable) {}
func (c *blockChecker) AfterLeaveNode(w walker.Walkable)  {}

func (c *blockChecker) BeforeEnterNode(w walker.Walkable) {
	switch n := w.(type) {
	case *expr.FunctionCall:
		c.handleFunctionCall(n)
	}
}

func (c *blockChecker) handleFunctionCall(call *expr.FunctionCall) {
	name, ok := call.Function.(*name.Name)
	if !ok {
		return
	}

	if meta.NameEquals(name, "define") {
		c.checkDefine(call)
	}
}

func (c *blockChecker) checkDefine(define *expr.FunctionCall) {
	if len(define.Arguments) != 2 {
		c.ctxt.Report(define.Arguments[2], linter.LevelWarning, "don't use case_insensitive argument")
	}
}
