package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"regexp"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/check"
)

func init() {
	(&cmdModify{}).newCommand("modify", "mod", "m")
}

type cmdModify struct {
	flagsWithUsage
	// flags
	fPayee      string
	fAccount    string
	fSetAccount string

	// internal
	from, to *coin.Account
	payee    *regexp.Regexp
}

func (*cmdModify) newCommand(names ...string) command {
	var cmd cmdModify
	cmd.FlagSet = newCommand(&cmd, names...)
	setUsage(cmd.FlagSet, `(modify|mod|m) [flags] files...

Modify transactions or postings in specified files. Rewrites the files in place.`)
	cmd.StringVar(&cmd.fPayee, "p", "", "modify transactions with matching payee (regex)")
	cmd.StringVar(&cmd.fAccount, "a", "", "modify transaction or posting associated with given account")
	cmd.StringVar(&cmd.fSetAccount, "a:", "", "move posting matching -a to given account")
	return &cmd
}

func (cmd *cmdModify) init() {
	coin.LoadFile(coin.CommoditiesFile)
	coin.LoadFile(coin.AccountsFile)
	coin.ResolveAccounts()
}

func (cmd *cmdModify) execute(_ io.Writer) {
	check.If(cmd.NArg() > 1, "both from and to account must be specified")
	if len(cmd.fAccount) > 0 {
		cmd.from = coin.MustFindAccount(cmd.fAccount)
	}
	if len(cmd.fSetAccount) > 0 {
		cmd.to = coin.MustFindAccount(cmd.fSetAccount)
	}
	if len(cmd.fPayee) > 0 {
		cmd.payee = regexp.MustCompile(cmd.fPayee)
	}
	for _, fn := range cmd.Args() {
		coin.LoadFile(fn)
		coin.ResolveTransactions(false)
		tf, err := os.CreateTemp(path.Dir(fn), path.Base(fn))
		check.NoError(err, "creating temp file")
		var count int
		for _, t := range coin.Transactions {
			if cmd.modify(t) {
				count++
			}
			t.Write(tf, false)
			fmt.Fprintln(tf)
		}
		err = os.Remove(fn)
		check.NoError(err, "deleting old file")
		err = os.Rename(tf.Name(), fn)
		check.NoError(err, "renaming temp file")
		fmt.Fprintf(os.Stderr, "Updated %d transactions in %s\n", count, fn)
		coin.DropTransactions()
	}
}

func (cmd *cmdModify) modify(t *coin.Transaction) (modified bool) {
	if cmd.payee != nil && !cmd.payee.Match([]byte(t.Description)) {
		return modified
	}
	var hasPostingsMatchingAccount bool
	for _, p := range t.Postings {
		if p.Account == cmd.from {
			hasPostingsMatchingAccount = true
			if cmd.to != nil {
				p.MoveTo(cmd.to)
				modified = true
			}
		}
	}
	if cmd.from != nil && !hasPostingsMatchingAccount {
		return modified
	}
	return modified
}
