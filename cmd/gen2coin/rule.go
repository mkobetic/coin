package main

import (
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/mkobetic/coin"
)

/*
	commodity CAD

	account Assets:Bank:Checking
	  commodity CAD
	account Liabilities:Credit:Card1
	  commodity CAD
	account Liabilities:Credit:Card2
	  commodity CAD
	account Expenses:Groceries
	  commodity CAD
	account Expenses:Dining
	  commodity CAD

	<-2,8> FOOD MART|WENDY'S|BURGERKING
		Groceries|Dining <30,200> CAD
		Checking|Card1|Card2
*/

const (
	constants = iota + 1<<31
	FROM_BALANCE
	TO_BALANCE
)

type rule struct {
	from, to []*coin.Account
	dates    dateGen
	min, max int
	payees   []string
}

type sample struct {
	*rule
	*coin.Transaction
}

type samples []sample

func (transactions samples) Len() int { return len(transactions) }
func (transactions samples) Swap(i, j int) {
	transactions[i], transactions[j] = transactions[j], transactions[i]
}
func (transactions samples) Less(i, j int) bool {
	return transactions[i].Posted.Before(transactions[j].Posted)
}

func newRule(dates dateGen, payees, from, to string, min, max int) *rule {
	return &rule{
		dates:  dates,
		payees: strings.Split(payees, "|"),
		from:   toAccounts(strings.Split(from, "|")),
		to:     toAccounts(strings.Split(to, "|")),
		min:    min,
		max:    max,
	}
}

func (r *rule) generateTransactions(begin, end time.Time) samples {
	var transactions samples
	for _, posted := range r.dates(begin, end) {
		payee := r.payees[rand.Intn(len(r.payees))]
		t := &coin.Transaction{
			Posted:      posted,
			Description: payee}
		transactions = append(transactions, sample{r, t})
	}
	return transactions
}

func (r *rule) generatePostings(t *coin.Transaction) {
	from := r.from[rnd.Intn(len(r.from))]
	to := r.to[rnd.Intn(len(r.to))]
	amt := amtBetween(r.min, r.max, from, to)
	t.Post(to, from, amt, nil)
}

func amtBetween(a, b int, from, to *coin.Account) *coin.Amount {
	if a > b {
		a, b = b, a
	}
	amt := a
	if b != a {
		amt = a + rnd.Intn(b-a)
	}
	return &coin.Amount{
		big.NewInt(int64(amt) * pow(10, from.Commodity.Decimals)),
		from.Commodity,
	}
}

func toAccounts(patterns []string) (accounts []*coin.Account) {
	for _, p := range patterns {
		accounts = append(accounts, coin.MustFindAccount(p))
	}
	return accounts
}

func pow(base, exp int) (res int64) {
	bitMask := 1
	pow := int64(base)
	res = 1
	for bitMask <= exp {
		if bitMask&exp != 0 {
			res *= pow
		}
		pow *= pow
		bitMask <<= 1
	}
	return res
}
