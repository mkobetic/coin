package main

import (
	"fmt"
	"io"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/check"
)

var (
	cmdRegister *command
	recurse     bool
	begin, end  coin.Date
)

func init() {
	cmdRegister = newCommand(register_load, register, "register", "reg", "r")
	cmdRegister.BoolVar(&recurse, "r", false, "include children account postings")
	cmdRegister.Var(&begin, "b", "begin register from this date")
	cmdRegister.Var(&end, "e", "end register on this date")
}

func register_load() {
	check.If(cmdRegister.NArg() > 0, "account filter is required")
	coin.LoadAll()
}
func register(f io.Writer) {
	pattern := cmdRegister.Arg(0)
	acc := coin.MustFindAccount(pattern)
	fmt.Fprintln(f, acc.FullName, acc.Commodity.Id)
	if recurse {
		recursiveRegister(f, acc, begin.Time, end.Time)
	} else {
		flatRegister(f, acc, begin.Time, end.Time)
	}
}

func flatRegister(f io.Writer, acc *coin.Account, begin, end time.Time) {
	var desc, acct int
	debugf("flat register from %s to %s\n", begin, end)
	postings := slice(acc.Postings, begin, end)
	for _, s := range postings {
		desc = max(desc, len(s.Transaction.Description))
		acct = max(acct, len(s.Transaction.Other(s).Account.FullName))
	}
	var total = coin.NewAmount(big.NewInt(0), acc.Commodity)
	for _, s := range postings {
		total.AddIn(s.Quantity)
		fmt.Fprintf(f, "%s | %*s | %*s | %10a | %10a \n",
			s.Transaction.Posted.Format(coin.DateFormat),
			min(desc, 50),
			s.Transaction.Description,
			min(acct, 50),
			s.Transaction.Other(s).Account.FullName,
			s.Quantity,
			total,
		)
	}
}

func recursiveRegister(f io.Writer, acc *coin.Account, begin, end time.Time) {
	var postings []*coin.Posting
	var desc, acct, from int
	prefix := acc.FullName
	acc.WithChildrenDo(func(a *coin.Account) {
		if l := len(a.FullName) - len(acc.FullName); l > from {
			from = l
		}
		for _, s := range slice(a.Postings, begin, end) {
			desc = max(desc, len(s.Transaction.Description))
			acct = max(acct, len(strings.TrimPrefix(s.Transaction.Other(s).Account.FullName, prefix)))
			postings = append(postings, s)
		}
	})
	// sort all postings by time
	sort.SliceStable(postings, func(i, j int) bool {
		return postings[i].Transaction.Posted.Before(postings[j].Transaction.Posted)
	})
	for _, s := range postings {
		fmt.Fprintf(f, "%s | %*s | %*s | %*s | %10a %s\n",
			s.Transaction.Posted.Format(coin.DateFormat),
			min(desc, 50),
			s.Transaction.Description,
			min(from, 50),
			strings.TrimPrefix(s.Account.FullName, prefix),
			min(acct, 50),
			strings.TrimPrefix(s.Transaction.Other(s).Account.FullName, prefix),
			s.Quantity,
			s.Account.CommodityId,
		)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func slice(postings []*coin.Posting, begin, end time.Time) []*coin.Posting {
	if !begin.IsZero() {
		from := sort.Search(len(postings), func(i int) bool {
			return !postings[i].Transaction.Posted.Before(begin)
		})
		if from == len(postings) {
			return nil
		}
		postings = postings[from:]
	}
	if !end.IsZero() {
		to := sort.Search(len(postings), func(i int) bool {
			return !postings[i].Transaction.Posted.Before(end)
		})
		if to == len(postings) {
			return postings
		}
		postings = postings[:to]
	}
	return postings
}
