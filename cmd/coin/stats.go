package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/mkobetic/coin"
)

func init() {
	(&cmdStats{}).newCommand("stats", "s")
}

type cmdStats struct {
	*flag.FlagSet
	dupes      bool
	unbalanced bool
	begin, end coin.Date
}

func (_ *cmdStats) newCommand(names ...string) command {
	var cmd cmdStats
	cmd.FlagSet = newCommand(&cmd, names...)
	cmd.BoolVar(&cmd.dupes, "d", false, "check for duplicate transactions")
	cmd.BoolVar(&cmd.unbalanced, "u", false, "check for unbalanced transactions")
	cmd.Var(&cmd.begin, "b", "begin register from this date")
	cmd.Var(&cmd.end, "e", "end register on this date")
	return &cmd
}

func (cmd *cmdStats) init() {
	coin.LoadAll()
}

func (cmd *cmdStats) execute(f io.Writer) {

	transactions := cmd.transactions()
	if cmd.dupes {
		var day []*coin.Transaction
		for _, t := range transactions {
			if len(day) == 0 {
				day = append(day, t)
				continue
			}
			if !t.Posted.Equal(day[0].Posted) {
				day = append(day[:0], t)
				continue
			}
			for _, t2 := range day {
				if t.IsEqual(t2) {
					fmt.Fprintf(os.Stderr,
						"DUPLICATE TRANSACTION?\n%s\n%s\n%s\n%s\n",
						t2.Location(), t2,
						t.Location(), t)
				}
			}
			day = append(day, t)
		}
		return
	}

	if cmd.unbalanced {
		for _, t := range transactions {
			for _, p := range t.Postings {
				if p.Account == coin.Unbalanced {
					fmt.Fprintf(os.Stderr,
						"UNBALANCED TRANSACTION!\n%s\n%s\n",
						t.Location(),
						t)
				}
			}
		}
		return
	}

	fmt.Fprintln(f, "Commodities:", len(coin.Commodities))
	fmt.Fprintln(f, "Prices:", len(coin.Prices))
	fmt.Fprintln(f, "Accounts:", len(coin.AccountsByName))
	fmt.Fprintln(f, "Transactions:", len(transactions))
}

func (cmd *cmdStats) transactions() []*coin.Transaction {
	transactions := coin.Transactions
	if !cmd.begin.IsZero() {
		from := sort.Search(len(transactions), func(i int) bool {
			return !transactions[i].Posted.Before(cmd.begin.Time)
		})
		if from == len(transactions) {
			return nil
		}
		transactions = transactions[from:]
	}
	if !cmd.end.IsZero() {
		to := sort.Search(len(transactions), func(i int) bool {
			return !transactions[i].Posted.Before(cmd.end.Time)
		})
		if to == len(transactions) {
			return transactions
		}
		transactions = transactions[:to]
	}
	return transactions
}
