package main

import (
	"github.com/VKCOM/noverify/src/linter"
	"github.com/z7zmey/php-parser/walker"
)

type blockChecker struct {
	ctxt linter.BlockContext
}

func (c *blockChecker) AfterEnterNode(w walker.Walkable)  {}
func (c *blockChecker) BeforeLeaveNode(w walker.Walkable) {}
func (c *blockChecker) AfterLeaveNode(w walker.Walkable)  {}

func (c *blockChecker) BeforeEnterNode(w walker.Walkable) {}
