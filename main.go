package main

import (
	"github.com/VKCOM/noverify/src/cmd"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/z7zmey/php-parser/node"
)

func init() {
	mi := &metainfoExt{
		constValue: map[string]node.Node{},
		st:         &meta.ClassParseState{},
	}
	linter.RegisterBlockChecker(func(ctxt linter.BlockContext) linter.BlockChecker {
		mi.ctxt = ctxt
		return mi
	})
	linter.RegisterBlockChecker(func(ctxt linter.BlockContext) linter.BlockChecker {
		return &blockChecker{
			ctxt: ctxt,
			mi:   mi,
		}
	})
}

func main() {
	cmd.Main()
}
