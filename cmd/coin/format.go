package main

import (
	"fmt"
	"io"

	"github.com/mkobetic/coin"
)

var (
	cmdFormat *command
	ledger    bool
)

func init() {
	cmdFormat = newCommand(format_load, format, "format", "fmt", "f")
	cmdFormat.BoolVar(&ledger, "ledger", false, "use ledger compatible format")
}

func format_load() {
	coin.LoadFile(coin.CommoditiesFile)
	coin.LoadFile(coin.AccountsFile)
	coin.ResolveAccounts()
	for _, fn := range cmdFormat.Args() {
		coin.LoadFile(fn)
	}
	coin.ResolveTransactions(false)
}

func format(f io.Writer) {
	for _, t := range coin.Transactions {
		t.Write(f, ledger)
		fmt.Fprintln(f)
	}
}
