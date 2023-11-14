package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/check"
)

func init() {
	(&cmdFormat{}).newCommand("format", "fmt", "f")
}

type cmdFormat struct {
	*flag.FlagSet
	ledger  bool
	replace bool
}

func (*cmdFormat) newCommand(names ...string) command {
	var cmd cmdFormat
	cmd.FlagSet = newCommand(&cmd, names...)
	cmd.BoolVar(&cmd.ledger, "ledger", false, "use ledger compatible format")
	cmd.BoolVar(&cmd.replace, "replace", false, "format files in place")
	return &cmd
}

func (cmd *cmdFormat) init() {
	coin.LoadFile(coin.CommoditiesFile)
	coin.LoadFile(coin.AccountsFile)
	coin.ResolveAccounts()
}

func (cmd *cmdFormat) execute(f io.Writer) {
	for _, fn := range cmd.Args() {
		var err error
		var tf *os.File
		coin.LoadFile(fn)
		coin.ResolveTransactions(false)
		if cmd.replace {
			tf, err = os.CreateTemp(path.Dir(fn), path.Base(fn))
			check.NoError(err, "creating temp file")
			f = tf
		}
		for _, t := range coin.Transactions {
			t.Write(f, cmd.ledger)
			fmt.Fprintln(f)
		}
		if cmd.replace {
			err = os.Remove(fn)
			check.NoError(err, "deleting old file")
			err = os.Rename(tf.Name(), fn)
			check.NoError(err, "renaming temp file")
		}
		// Note that this doesn't properly get rid of transactions,
		// postings are still referenced through the accounts,
		// but we don't care in this case.
		coin.Transactions = nil
	}
}
