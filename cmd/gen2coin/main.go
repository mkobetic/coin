package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/check"
)

var (
	begin, end      coin.Date
	byYear, byMonth bool
	rnd             = rand.New(rand.NewSource(time.Now().Unix()))
)

func init() {
	flag.Var(&begin, "b", "begin ledger on or after this date")
	flag.Var(&end, "e", "end ledger on or before this date")
	flag.BoolVar(&byYear, "y", false, "split ledger by year")
	flag.BoolVar(&byMonth, "m", false, "split ledger by month")
}

func main() {
	flag.Parse()
	var dir string
	if len(flag.Args()) > 0 {
		dir = flag.Arg(0)
		fi, err := os.Stat(dir)
		check.NoError(err, "stating path %s", dir)
		check.If(fi.IsDir(), "%s is not a directory", dir)
	}
	end := end.Time
	if end.IsZero() {
		end = time.Now()
	}
	begin := begin.Time
	if begin.IsZero() {
		begin = end.AddDate(0, -3, 0)
	}
	var transactions samples
	for _, r := range sample1() {
		transactions = append(transactions, r.generateTransactions(begin, end)...)
	}
	sort.Stable(transactions)
	for _, s := range transactions {
		s.generatePostings(s.Transaction)
	}
	if dir == "" { // just dump everything into stdout
		for _, t := range transactions {
			t.Write(os.Stdout, false)
			fmt.Fprintln(os.Stdout)
		}
		return
	}

	w, err := os.Create(filepath.Join(dir, "commodities.coin"))
	check.NoError(err, "opening commodities file")
	for n, c := range coin.Commodities {
		check.NoError(c.Write(w, false), "writing commodity %s", n)
		_, err = fmt.Fprintln(w, "")
		check.NoError(err, "writing newline")
	}
	check.NoError(w.Close(), "closing commodities file")
	w, err = os.Create(filepath.Join(dir, "accounts.coin"))
	check.NoError(err, "opening accounts file")
	for n, a := range coin.AccountsByName {
		check.NoError(a.Write(w, false), "writing account %s", n)
		_, err = fmt.Fprintln(w, "")
		check.NoError(err, "writing newline")
	}
	check.NoError(w.Close(), "closing accounts file")

	labeler := func(t time.Time) string { return "transactions" }
	if byYear {
		labeler = func(t time.Time) string { return t.Format("2006") }
	}
	if byMonth {
		labeler = func(t time.Time) string { return t.Format("2006-01") }
	}

	for label, batch := range getBatches(transactions, labeler) {
		w, err = os.Create(filepath.Join(dir, label+".coin"))
		check.NoError(err, "opening transaction file %s.coin", label)
		for _, t := range batch {
			t.Write(w, false)
			fmt.Fprintln(w)
		}
		check.NoError(w.Close(), "closing transaction file %s.coin", label)
	}
}

func getBatches(transactions samples, label func(t time.Time) string) map[string]samples {
	batches := make(map[string]samples)
	if len(transactions) == 0 {
		return batches
	}
	var batch samples
	previous := label(transactions[0].Posted)
	for _, t := range transactions {
		next := label(t.Posted)
		if previous == next {
			batch = append(batch, t)
		} else {
			batches[previous] = batch
			previous = next
			batch = samples{t}
		}
	}
	// add the last batch
	batches[previous] = batch
	return batches
}

func mustParse(s string) coin.Item {
	p := coin.NewParser(strings.NewReader(s))
	i, err := p.Next("")
	if err != nil {
		panic(err)
	}
	return i
}
