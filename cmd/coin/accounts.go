package main

import (
	"flag"
	"fmt"
	"io"
	"regexp"

	"github.com/mkobetic/coin"
)

func init() {
	(&cmdAccounts{}).newCommand("accounts", "acc", "a")
}

type cmdAccounts struct {
	*flag.FlagSet
}

func (_ *cmdAccounts) newCommand(names ...string) command {
	var cmd cmdAccounts
	cmd.FlagSet = newCommand(&cmd, names...)
	return &cmd
}

func (cmd *cmdAccounts) init() {
	coin.LoadFile(coin.AccountsFile)
}

func (cmd *cmdAccounts) execute(f io.Writer) {
	var pattern *regexp.Regexp
	if cmd.NArg() > 0 {
		pattern = coin.ToRegex(cmd.Arg(0))
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
