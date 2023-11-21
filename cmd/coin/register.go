package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/check"
)

func init() {
	(&cmdRegister{}).newCommand("register", "reg", "r")
}

type cmdRegister struct {
	flagsWithUsage
	verbose           bool
	recurse           bool
	begin, end        coin.Date
	weekly, monthly   bool
	quarterly, yearly bool
	top               int
	cumulative        bool
	maxLabelWidth     int
	location          bool
	output            string
}

func (*cmdRegister) newCommand(names ...string) command {
	var cmd cmdRegister
	cmd.FlagSet = newCommand(&cmd, names...)
	setUsage(cmd.FlagSet, `(register|reg|r) [flags] account

Lists or aggregate postings from the specified account.`)
	cmd.BoolVar(&cmd.verbose, "v", false, "log debug info to stderr")
	cmd.BoolVar(&cmd.recurse, "r", false, "include subaccount postings in parent accounts")
	cmd.Var(&cmd.begin, "b", "begin register from this date")
	cmd.Var(&cmd.end, "e", "end register on this date")
	// aggregation options
	cmd.BoolVar(&cmd.weekly, "w", false, "aggregate postings by week")
	cmd.BoolVar(&cmd.monthly, "m", false, "aggregate postings by month")
	cmd.BoolVar(&cmd.quarterly, "q", false, "aggregate postings by quarter")
	cmd.BoolVar(&cmd.yearly, "y", false, "aggregate postings by year")
	cmd.IntVar(&cmd.top, "t", 5, "include this many largest subaccounts in aggregate results")
	cmd.BoolVar(&cmd.cumulative, "c", false, "aggregate cumulatively across time")
	// output options
	cmd.IntVar(&cmd.maxLabelWidth, "l", 12, "maximum width of a column label")
	cmd.BoolVar(&cmd.location, "f", false, "include file location on postings in non-aggregated results")
	cmd.StringVar(&cmd.output, "o", "text", "output format for aggregated results: text, json, csv, chart")
	return &cmd
}

func (cmd *cmdRegister) init() {
	check.If(cmd.NArg() > 0, "account filter is required")
	coin.LoadAll()
}

func (cmd *cmdRegister) execute(f io.Writer) {
	pattern := cmd.Arg(0)
	acc := coin.MustFindAccount(pattern)
	if cmd.output == "text" {
		fmt.Fprintln(f, acc.FullName, acc.Commodity.Id)
	}
	if by := cmd.period(); by != nil {
		if cmd.recurse {
			cmd.recursiveAggregatedRegister(f, acc, by)
		} else {
			cmd.flatAggregatedRegister(f, acc, by)
		}
	} else {
		var opts = options{prefix: acc.FullName, maxAcct: cmd.maxLabelWidth, location: cmd.location, commodity: acc.Commodity}
		if cmd.recurse {
			var ps postings
			acc.WithChildrenDo(func(a *coin.Account) {
				ps = append(ps, cmd.trim(a.Postings)...)
			})
			sort.SliceStable(ps, func(i, j int) bool {
				return ps[i].Transaction.Posted.Before(ps[j].Transaction.Posted)
			})
			ps.printLong(f, &opts)
		} else {
			cmd.trim(acc.Postings).print(f, &opts)
		}
	}
}

func (cmd *cmdRegister) flatAggregatedRegister(f io.Writer, acc *coin.Account, by *reducer) {
	totals := accountTotals{}
	acc.WithChildrenDo(func(a *coin.Account) {
		ts := totals.newTotals(a, by)
		for _, p := range cmd.trim(a.Postings) {
			ts.add(p.Transaction.Posted, p.Quantity)
		}
	})
	var accounts []*coin.Account
	totals, accounts = totals.top(cmd.top)
	top := totals[accounts[0]]
	for _, ts := range totals {
		top.mergeTime(ts)
	}
	totals.mergeTime(top)
	if cmd.cumulative {
		totals.makeCumulative()
	}
	label := func(a *coin.Account) string {
		switch a {
		case nil:
			return "Other"
		case acc:
			return acc.Name
		default:
			n := strings.TrimPrefix(a.FullName, acc.FullName)
			return coin.ShortenAccountName(n, cmd.maxLabelWidth)
		}
	}
	totals.output(f, accounts, label, cmd.output)
}

func (cmd *cmdRegister) recursiveAggregatedRegister(f io.Writer, acc *coin.Account, by *reducer) {
	totals := accountTotals{}
	acc.WithChildrenDo(func(a *coin.Account) {
		ts := totals.newTotals(a, by)
		for _, p := range cmd.trim(a.Postings) {
			ts.add(p.Transaction.Posted, p.Quantity)
		}
	})
	if cmd.recurse {
		acc.FirstWithChildrenDo(func(a *coin.Account) {
			child := totals[a]
			parent := totals[a.Parent]
			if parent != nil {
				parent.merge(child)
			}
		})
	}
	totals.sanitize()
	accTotals := totals[acc]
	check.If(accTotals != nil, "root account totals shouldn't be empty\n")
	delete(totals, acc)
	var accounts []*coin.Account
	totals, accounts = totals.top(cmd.top)
	totals.mergeTime(accTotals)
	totals[acc] = accTotals
	accounts = append(accounts, acc)
	if cmd.cumulative {
		totals.makeCumulative()
	}
	label := func(a *coin.Account) string {
		switch a {
		case nil:
			return "Other"
		case acc:
			return "Totals"
		default:
			n := strings.TrimPrefix(a.FullName, acc.FullName)
			return coin.ShortenAccountName(n, cmd.maxLabelWidth)
		}
	}
	totals.output(f, accounts, label, cmd.output)
}

func (cmd *cmdRegister) period() *reducer {
	switch {
	case cmd.weekly:
		return &week
	case cmd.monthly:
		return &month
	case cmd.quarterly:
		return &quarter
	case cmd.yearly:
		return &year
	}
	return nil
}

func (cmd *cmdRegister) trim(ps []*coin.Posting) postings {
	return postings(trim(ps, cmd.begin, cmd.end))
}

func (cmd *cmdRegister) debugf(format string, args ...interface{}) {
	if !cmd.verbose {
		return
	}
	fmt.Fprintf(os.Stderr, format, args...)
}
