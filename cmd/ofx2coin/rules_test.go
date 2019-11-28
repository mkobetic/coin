package main

import (
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/assert"
)

func init() {
	r := strings.NewReader(`
commodity CAD
  format 1.00 CAD

commodity USD
  format 1.00 USD

account Assets:Bank:Checking

account Assets:Bank:Savings

account Expenses:Groceries

account Expenses:Auto

account Expenses:Auto:Gas

account Expenses:Miscellaneous

account Income:Salary

account Income:Interest

account Liabilities:Credit:MC
`)
	coin.Load(r)
	coin.ResolveAccounts()
}

var sample = `
common
  Expenses:Groceries       FRESHCO|COSTCO WHOLESALE|FARM BOY|LOBLAWS
  Expenses:Auto:Gas        COSTCO GAS|PETROCAN|SHELL
389249328477983 Assets:Bank:Savings
  Income:Interest     Interest
392843029797099 Assets:Bank:Checking
    @common
	Income:Salary       ACME PAY 
479347938749398 Liabilities:Credit:MC
  Expenses:Auto            HUYNDAI|TOYOTA
  @common
  Expenses:Miscellaneous   
`

func Test_ReadRules(t *testing.T) {
	r := strings.NewReader(sample)
	rules, err := ReadRules(r)
	assert.NoError(t, err)
	assert.Equal(t, len(rules.Accounts), 3)
	mc := rules.Accounts["479347938749398"]
	assert.NotNil(t, mc)
	assert.Equal(t, len(mc.Rules), 3)
}

func Test_Classification(t *testing.T) {
	r := strings.NewReader(sample)
	rules, err := ReadRules(r)
	assert.NoError(t, err)
	date, _ := time.Parse("06/01/02", "18/10/20")
	for _, fix := range []struct {
		from  string
		payee string
		to    string
	}{
		{"479347938749398", "[TR] COSTCO WHOLESALE #9239", "Expenses:Groceries"},
		{"479347938749398", "JOE'S DINER", "Expenses:Miscellaneous"},
	} {
		tran := rules.Accounts[fix.from].Transaction(date, fix.payee, *big.NewRat(-10000, 100), nil)
		if account := tran.Postings[0].Account.FullName; account != fix.to {
			t.Errorf("mismatched\nexp: %s\ngot: %s\n", fix.to, account)
		}
	}
}
