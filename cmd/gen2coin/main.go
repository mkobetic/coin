package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/mkobetic/coin"
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
	end := end.Time
	if end.IsZero() {
		end = time.Now()
	}
	begin := begin.Time
	if begin.IsZero() {
		begin = end.AddDate(0, -3, 0)
	}
	var transactions coin.TransactionsByTime
	for _, r := range sample1() {
		transactions = append(transactions, r.generate(begin, end)...)
	}
	sort.Stable(transactions)
	for _, t := range transactions {
		t.Write(os.Stdout, false)
		fmt.Fprintln(os.Stdout)
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
