package main

import (
	"strings"
	"testing"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/assert"
)

var rules1 *Rules

func init() {
	coin.DefaultCommodityId = "USD"

	r := strings.NewReader(`
commodity USD
commodity VXF
  format 1.000 VXF
  symbol VXF

account Assets:Investments
  csv_acctid XXX
account Assets:Investments:VXF
  check commodity == VXF
account Expenses:Fees
account Income:Dividends
`)
	coin.Load(r, "")
	coin.ResolveAccounts()

	coin.AccountsDo(func(a *coin.Account) {
		if a.CSVAcctId != "" {
			accountsByCSVId[a.CSVAcctId] = a
		}
	})

	rules1 = ReadRules(strings.NewReader(`src
  account 0
  description 4
  date 1
  amount 7
  amount 3 -$1 VALUE = \s+([\d\.\-]+)
  symbol 2
  quantity 5
  note 3
---
XXX Assets:Investments
  Expenses:Fees Fee|HST
  Income:Dividends DRIP
`))
}
func Test_All(t *testing.T) {
	r := strings.NewReader(sample)
	txs := readTransactions(r, rules1.sources["src"].fields, rules1)
	for i, exp := range []string{
		`2019/09/10 DRIP ; blah blah VALUE =      1630.59
  Assets:Investments:VXF   123.999 VXF
  Income:Dividends        -1630.59 USD
`,
	} {
		got := txs[i].String()
		assert.Equal(t, got, exp)
	}
}

const sample = `
Account number,Trade date,Symbol,Description,Operation,Quantity,Price,Net amount
XXX,2019/09/10,VXF,"blah blah VALUE =      1630.59",DRIP,123.999,,
`
