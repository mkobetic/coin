package main

import (
	"flag"
	"math/rand"
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
		begin = end.AddDate(-1, 0, 0)
	}

}
