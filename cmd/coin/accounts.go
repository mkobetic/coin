package main

import (
	"fmt"
	"io"
	"regexp"

	"github.com/mkobetic/coin"
)

var (
	cmdAccounts *command
)

func init() {
	cmdAccounts = newCommand(accounts_load, accounts_list, "accounts", "acc", "a")
}

func accounts_load() {
	coin.LoadFile(coin.AccountsFile)
}

func accounts_list(f io.Writer) {
	var pattern *regexp.Regexp
	if cmdAccounts.NArg() > 0 {
		pattern = coin.ToRegex(cmdAccounts.Arg(0))
	}
	var max int
	coin.AccountsDo(func(a *coin.Account) {
		if pattern != nil && !pattern.MatchString(a.FullName) {
			return
		}
		if l := len(a.FullName); l > max {
			max = l
		}
	})
	coin.AccountsDo(func(a *coin.Account) {
		if pattern != nil && !pattern.MatchString(a.FullName) {
			return
		}
		fmt.Fprintf(f, "%-*s | %-10s | %s\n", max, a.FullName, a.CommodityId, a.Description)
	})
}
