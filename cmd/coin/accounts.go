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
	closed bool
}

func (_ *cmdAccounts) newCommand(names ...string) command {
	var cmd cmdAccounts
	cmd.FlagSet = newCommand(&cmd, names...)
	cmd.BoolVar(&cmd.closed, "c", false, "show closed accounts")
	return &cmd
}

func (cmd *cmdAccounts) init() {
	coin.LoadFile(coin.CommoditiesFile)
	coin.LoadFile(coin.AccountsFile)
	coin.ResolveAccounts()
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
		if !cmd.closed && a.IsClosed() {
			return
		}
		if pattern != nil && !pattern.MatchString(a.FullName) {
			return
		}
		fmt.Fprintf(f, "%-*s | %-10s | %s\n", max, a.FullName, a.CommodityId, a.Description)
	})
}
