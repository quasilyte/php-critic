package main

import (
	"github.com/VKCOM/noverify/src/cmd"
	"github.com/VKCOM/noverify/src/linter"
)

func init() {
	linter.RegisterBlockChecker(func(ctxt linter.BlockContext) linter.BlockChecker {
		return &blockChecker{
			ctxt: ctxt,
		}
	})
}

func main() {
	cmd.Main()
}
