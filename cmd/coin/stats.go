package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/mkobetic/coin"
)

var (
	dupes      bool
	unbalanced bool
)

func init() {
	cmdStats := newCommand(coin.LoadAll, stats, "stats", "s")
	cmdStats.BoolVar(&dupes, "d", false, "check for duplicate transactions")
	cmdStats.BoolVar(&unbalanced, "u", false, "check for unbalanced transactions")
	cmdStats.Var(&begin, "b", "begin register from this date")
	cmdStats.Var(&end, "e", "end register on this date")
}

func stats(f io.Writer) {

	transactions := sliceTransactions(begin.Time, end.Time)
	if dupes {
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

	if unbalanced {
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

func sliceTransactions(begin, end time.Time) []*coin.Transaction {
	transactions := coin.Transactions
	if !begin.IsZero() {
		from := sort.Search(len(transactions), func(i int) bool {
			return !transactions[i].Posted.Before(begin)
		})
		if from == len(transactions) {
			return nil
		}
		transactions = transactions[from:]
	}
	if !end.IsZero() {
		to := sort.Search(len(transactions), func(i int) bool {
			return !transactions[i].Posted.Before(end)
		})
		if to == len(transactions) {
			return transactions
		}
		transactions = transactions[:to]
	}
	return transactions
}
