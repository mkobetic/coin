package main

import (
	"math/big"
	"math/rand"
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
		t.Post(from, to, amt, nil)
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
		a.Add(a.Int, diff),
		a.Commodity,
	}
}
