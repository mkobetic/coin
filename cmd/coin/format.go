package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/mkobetic/coin"
)

func init() {
	(&cmdFormat{}).newCommand("format", "fmt", "f")
}

type cmdFormat struct {
	*flag.FlagSet
	ledger bool
}

func (_ *cmdFormat) newCommand(names ...string) command {
	var cmd cmdFormat
	cmd.FlagSet = newCommand(&cmd, names...)
	cmd.BoolVar(&cmd.ledger, "ledger", false, "use ledger compatible format")
	return &cmd
}

func (cmd *cmdFormat) init() {
	coin.LoadFile(coin.CommoditiesFile)
	coin.LoadFile(coin.AccountsFile)
	coin.ResolveAccounts()
	for _, fn := range cmd.Args() {
		coin.LoadFile(fn)
	}
	coin.ResolveTransactions(false)
}

func (cmd *cmdFormat) execute(f io.Writer) {
	for _, t := range coin.Transactions {
		t.Write(f, cmd.ledger)
		fmt.Fprintln(f)
	}
}
