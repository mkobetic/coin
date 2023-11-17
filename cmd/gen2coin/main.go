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
	begin, end coin.Date
	rnd        = rand.New(rand.NewSource(time.Now().Unix()))
)

func init() {
	flag.Var(&begin, "b", "begin ledger on or after this date")
	flag.Var(&end, "e", "end ledger on or before this date")
}

func main() {
	flag.Parse()
	var dir string
	var err error
	os.Args = os.Args[1:]
	if len(os.Args) > 0 {
		dir = os.Args[0]
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
	w := os.Stdout
	if dir != "" {
		w, err = os.Create(filepath.Join(dir, "commodities.coin"))
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
		w, err = os.Create(filepath.Join(dir, "transactions.coin"))
		check.NoError(err, "opening transaction file")
		defer w.Close()
	}
	for _, t := range transactions {
		t.Write(w, false)
		fmt.Fprintln(w)
	}
}

func mustParse(s string) coin.Item {
	p := coin.NewParser(strings.NewReader(s))
	i, err := p.Next("")
	if err != nil {
		panic(err)
	}
	return i
}
