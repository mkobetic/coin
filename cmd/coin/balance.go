package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/mkobetic/coin"
)

func init() {
	(&cmdBalance{}).newCommand("balance", "bal", "b")
}

type cmdBalance struct {
	*flag.FlagSet
	zeroBalance bool
	level       int
}

func (_ *cmdBalance) newCommand(names ...string) command {
	var cmd cmdBalance
	cmd.FlagSet = newCommand(&cmd, names...)
	cmd.BoolVar(&cmd.zeroBalance, "z", false, "list accounts with zero total balance")
	cmd.IntVar(&cmd.level, "l", 0, "print accounts up to this level, 0 means all")
	return &cmd
}

func (cmd *cmdBalance) init() {
	coin.LoadAll()
}

func (cmd *cmdBalance) execute(f io.Writer) {
	account := coin.Root
	if cmd.NArg() > 0 {
		account = coin.MustFindAccount(cmd.Arg(0))
	}
	cmd.printAccounts(f, account, cmd.level)
}

func (cmd *cmdBalance) printAccounts(f io.Writer, a *coin.Account, level int) {
	if a != coin.Root && !cmd.zeroBalance {
		bal, err := a.CumulativeBalance()
		if err != nil {
			fmt.Fprintln(f, a.FullName, err)
			return
		}
		if bal.IsZero() {
			return
		}
	}
	fmt.Fprintln(f, a)
	if a.Parent != nil {
		level--
		if level == 0 {
			return
		}
	}

	for _, c := range a.Children {
		cmd.printAccounts(f, c, level)
	}
}
