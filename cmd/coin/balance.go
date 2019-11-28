package main

import (
	"fmt"
	"io"

	"github.com/mkobetic/coin"
)

var (
	cmdBalance  *command
	zeroBalance bool
	level       int
)

func init() {
	cmdBalance = newCommand(coin.LoadAll, balance, "balance", "bal", "b")
	cmdBalance.BoolVar(&zeroBalance, "z", false, "list accounts with zero total balance")
	cmdBalance.IntVar(&level, "l", 0, "print accounts up to this level, 0 means all")
}

func balance(f io.Writer) {
	account := coin.Root
	if cmdBalance.NArg() > 0 {
		account = coin.MustFindAccount(cmdBalance.Arg(0))
	}
	printAccounts(f, account, zeroBalance, level)
}

func printAccounts(f io.Writer, a *coin.Account, zeroBalance bool, level int) {
	if a != coin.Root && !zeroBalance {
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
		printAccounts(f, c, zeroBalance, level)
	}
}
