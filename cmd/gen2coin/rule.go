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

type rule struct {
	from, to []*coin.Account
	dates    dateGen
	min, max *coin.Amount
	payees   []string
}

func newRule(dates dateGen, payees, from, to string, min, max int, commodity string) *rule {
	return &rule{
		dates:  dates,
		payees: strings.Split(payees, "|"),
		from:   toAccounts(strings.Split(from, "|")),
		to:     toAccounts(strings.Split(to, "|")),
		min:    toAmount(min, commodity),
		max:    toAmount(max, commodity),
	}
}

func (r *rule) generate(begin, end time.Time) []*coin.Transaction {
	var transactions []*coin.Transaction
	for _, posted := range r.dates(begin, end) {
		payee := r.payees[rand.Intn(len(r.payees))]
		from := r.from[rnd.Intn(len(r.from))]
		to := r.to[rnd.Intn(len(r.to))]
		amt := amtBetween(r.min, r.max)
		t := &coin.Transaction{
			Posted:      posted,
			Description: payee}
		t.Post(to, from, amt, nil)
		transactions = append(transactions, t)
	}
	return transactions
}

func amtBetween(a, b *coin.Amount) *coin.Amount {
	if a.IsBigger(b) {
		a, b = b, a
	}
	amt := new(big.Int).Sub(b.Int, a.Int)
	diff := new(big.Int).Rand(rnd, amt)
	return &coin.Amount{
		diff.Add(diff, a.Int),
		a.Commodity,
	}
}

func toAccounts(patterns []string) (accounts []*coin.Account) {
	for _, p := range patterns {
		accounts = append(accounts, coin.MustFindAccount(p))
	}
	return accounts
}

func toAmount(amt int, com string) *coin.Amount {
	c := coin.MustFindCommodity(com, "")
	a := big.NewInt(int64(amt) * pow(10, c.Decimals))
	return &coin.Amount{a, c}
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
