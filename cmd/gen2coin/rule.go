package main

import (
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/check"
)

/*

Every rule is intended to generate a specific type of transaction.
The date generator is invoked to produce a sequence of dates and a transaction
is created for each of those with a randomly picked payee from the rule's list.

Transaction postings are generated separately after all the transactions have been sorted.
This allows generating posting amounts based on the running balances of the involved accounts
Accounts are again picked randomly from the rule's list.

	commodity CAD
		default

	account Assets:Bank:Checking
	account Liabilities:Credit:Card1
	account Liabilities:Credit:Card2
	account Expenses:Groceries
	account Expenses:Dining

	<dateGen> FOOD MART|WENDY'S|BURGERKING
		Groceries|Dining <30,200>
		Checking|Card1|Card2
*/

const (
	// these are special min values (large enough to not interfere with useful numeric values)
	constants    = iota + 1<<31
	FROM_BALANCE // use the 'from' account balance
	TO_BALANCE   // use the 'to' account balance
)

type rule struct {
	dates    dateGen         // date generator
	payees   []string        // list of possible payees
	from, to []*coin.Account // list of possible from and to accounts
	min, max int             // min and max amount value (see constants above)
}

// sample bundles a rule and a generated transaction together.
// this allows using the same rule later to generate postings.
type sample struct {
	*rule
	*coin.Transaction
}

// sortable sample list
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
	check.NoError(to.Balance().AddIn(amt), "failed to update balance")
	check.NoError(from.Balance().AddIn(amt.Negated()), "failed to update balance")
}

// generate amount given min/max and from/to accounts.
func amtBetween(a, b int, from, to *coin.Account) *coin.Amount {
	if a > constants {
		return amtFromBalance(a, b, from, to)
	}
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

// Compute the amount from the balance of one of the accounts.
// The divisor can be used to compute a percentage of the balance (e.g. 2% => divisor=50),
// this can be used to generate interest transactions.
func amtFromBalance(account, divisor int, from, to *coin.Account) *coin.Amount {
	var amt *coin.Amount
	switch account {
	case FROM_BALANCE:
		amt = from.Balance().Negated()
	case TO_BALANCE:
		amt = to.Balance().Negated()
	default:
		panic(fmt.Errorf("invalid constant: %d", account))
	}
	amt.Div(amt.Int, big.NewInt(int64(divisor)))
	return amt
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
