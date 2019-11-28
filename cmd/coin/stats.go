package main

import (
	"fmt"
	"io"
	"os"

	"github.com/mkobetic/coin"
)

var dupes bool

func init() {
	cmdStats := newCommand(coin.LoadAll, stats, "stats", "s")
	cmdStats.BoolVar(&dupes, "d", false, "check for duplicate transactions")
}

func stats(f io.Writer) {
	if dupes {
		var day []*coin.Transaction
		for _, t := range coin.Transactions {
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
					fmt.Fprintf(os.Stderr, "DUPLICATE TRANSACTION?\n%s\n%s\n", t2, t)
				}
			}
		}
		return
	}

	fmt.Fprintln(f, "Commodities:", len(coin.Commodities))
	fmt.Fprintln(f, "Prices:", len(coin.Prices))
	fmt.Fprintln(f, "Accounts:", len(coin.AccountsByName))
	fmt.Fprintln(f, "Transactions:", len(coin.Transactions))
}
