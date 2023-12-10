package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strings"

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
	fTTag       string
	fPTag       string
	fSetAccount string
	fSetTTag    string
	fSetPTag    string

	// internal
	from, to         *coin.Account
	payee            *regexp.Regexp
	ttag, ptag       *coin.TagMatcher
	setTTag, setPTag string
}

func (*cmdModify) newCommand(names ...string) command {
	var cmd cmdModify
	cmd.FlagSet = newCommand(&cmd, names...)
	setUsage(cmd.FlagSet, `(modify|mod|m) [flags] files...

Modify transactions or postings in specified files. Rewrites the files in place.`)
	cmd.StringVar(&cmd.fPayee, "p", "", "modify transactions with matching payee (regex)")
	cmd.StringVar(&cmd.fTTag, "tt", "", "modify transactions with matching tag (regex)")
	cmd.StringVar(&cmd.fPTag, "pt", "", "modify posting with matching tag (regex)")
	cmd.StringVar(&cmd.fAccount, "a", "", "modify transaction or posting associated with given account")
	cmd.StringVar(&cmd.fSetAccount, "a:", "", "move posting matching -a to given account")
	cmd.StringVar(&cmd.fSetPTag, "pt:", "", "add tag to posting matching -a")
	cmd.StringVar(&cmd.fSetTTag, "tt:", "", "add tag to transaction")
	return &cmd
}

func (cmd *cmdModify) init() {
	coin.LoadFile(coin.CommoditiesFile)
	coin.LoadFile(coin.AccountsFile)
	coin.ResolveAccounts()
}

func (cmd *cmdModify) execute(f io.Writer) {
	if len(cmd.fAccount) > 0 {
		cmd.from = coin.MustFindAccount(cmd.fAccount)
	}
	if len(cmd.fSetAccount) > 0 {
		cmd.to = coin.MustFindAccount(cmd.fSetAccount)
	}
	if len(cmd.fPayee) > 0 {
		cmd.payee = regexp.MustCompile(cmd.fPayee)
	}
	if len(cmd.fTTag) > 0 {
		cmd.ttag = coin.NewTagMatcher(cmd.fTTag)
	}
	if len(cmd.fPTag) > 0 {
		cmd.ptag = coin.NewTagMatcher(cmd.fPTag)
	}
	if len(cmd.fSetPTag) > 0 {
		cmd.setPTag = mustParseTags(cmd.fSetPTag)
	}
	if len(cmd.fSetTTag) > 0 {
		cmd.setTTag = mustParseTags(cmd.fSetTTag)
	}
	if cmd.NArg() == 0 { // for testing
		for _, t := range coin.Transactions {
			cmd.modify(t)
			t.Write(f, false)
			fmt.Fprintln(f)
		}
		return
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
		return false
	}
	if cmd.ttag != nil && !cmd.ttag.Match(t.Tags) {
		return false
	}
	var hasPostingsMatchingAccount bool
	for _, p := range t.Postings {
		if p.Account == cmd.from {
			hasPostingsMatchingAccount = true
			modified = modified || cmd.modifyPosting(p)
		}
		if cmd.ptag != nil && cmd.ptag.Match(p.Tags) {
			modified = modified || cmd.modifyPosting(p)
		}
	}

	if cmd.from != nil && !hasPostingsMatchingAccount {
		return modified
	}
	if len(cmd.setTTag) > 0 {
		t.Notes = addTagLine(t.Notes, cmd.setTTag)
	}
	return modified
}

func (cmd *cmdModify) modifyPosting(p *coin.Posting) (modified bool) {
	if cmd.to != nil {
		p.MoveTo(cmd.to)
		modified = true
	}
	if len(cmd.setPTag) > 0 {
		p.Notes = addTagLine(p.Notes, cmd.setPTag)
		modified = true
	}
	return modified
}

func mustParseTags(val string) string {
	tags := coin.ParseTags(val)
	check.If(len(tags) > 0, "cannot parse tag value %s", val)
	var parts []string
	for _, k := range tags.Keys() {
		v := tags[k]
		if len(v) > 0 {
			v = ":" + v
		}
		parts = append(parts, fmt.Sprintf("#%s%s", k, v))
	}
	return strings.Join(parts, ", ")
}

func addTagLine(notes []string, tagLine string) []string {
	if len(notes) == 0 {
		return append(notes, tagLine)
	}
	if len(notes[0])+len(tagLine) < coin.TRANSACTION_LINE_MAX {
		notes[0] += " " + tagLine
		return notes
	}
	return append(notes, tagLine)
}
