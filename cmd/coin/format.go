package main

import (
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
	flagsWithUsage
	ledger      bool
	replace     bool
	trimWS      bool
	thousandSep bool
}

func (*cmdFormat) newCommand(names ...string) command {
	var cmd cmdFormat
	cmd.FlagSet = newCommand(&cmd, names...)
	setUsage(cmd.FlagSet, `(format|fmt|f) [flags] files...

Output transactions from specified files sorted and in the standard format.`)
	cmd.BoolVar(&cmd.ledger, "ledger", false, "use ledger compatible format")
	cmd.BoolVar(&cmd.replace, "i", false, "format files in-place")
	cmd.BoolVar(&cmd.trimWS, "t", false, "trim excessive whitespace from descriptions")
	cmd.BoolVar(&cmd.thousandSep, "ts", false, "print amounts with thousand separators")
	return &cmd
}

func (cmd *cmdFormat) init() {
	coin.LoadFile(coin.CommoditiesFile)
	coin.LoadFile(coin.AccountsFile)
	coin.ResolveAccounts()
}

func (cmd *cmdFormat) execute(f io.Writer) {
	if len(cmd.Args()) == 0 { // for testing
		cmd.writeTransactions(f)
		return
	}
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
		cmd.writeTransactions(f)
		if cmd.replace {
			err = os.Remove(fn)
			check.NoError(err, "deleting old file")
			err = os.Rename(tf.Name(), fn)
			check.NoError(err, "renaming temp file")
		}
		coin.DropTransactions()
	}
}

func (cmd *cmdFormat) writeTransactions(f io.Writer) {
	for _, t := range coin.Transactions {
		if cmd.trimWS {
			t.Description = trimWS(t.Description)[0]
			t.Notes = trimWS(t.Notes...)
		}
		t.Write(f, cmd.ledger, cmd.thousandSep)
		fmt.Fprintln(f)
	}
}
