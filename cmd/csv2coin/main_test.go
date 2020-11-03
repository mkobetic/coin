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
  commodity VXF
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

	rules1 = ReadRules(strings.NewReader(`src 1
  account 0
  description 4
  date 1
  amount 7
  amount 3 "-$1" VALUE = \s+([\d\.\-]+)
  symbol 2
  quantity 5
  note 3
---
XXX Assets:Investments
  Expenses:Fees Fee|HST
  Income:Dividends DRIP
  Assets:Investments Sold|Bought
`))
}
func Test_All(t *testing.T) {
	r := strings.NewReader(sample)
	txs := readTransactions(r, rules1.sources["src"], rules1)
	for i, exp := range []string{
		`2019/09/10 DRIP ; blah blah VALUE =      1630.59
  Assets:Investments:VXF   123.999 VXF
  Income:Dividends        -1630.59 USD
`,
		`2019/10/10 Sold ; whatever
  Assets:Investments       67689.08 USD
  Assets:Investments:VXF  -5148.219 VXF
`,
		`2019/10/30 Fee ; MGMT FEE
  Expenses:Fees        12.40 USD
  Assets:Investments  -12.40 USD
`,
	} {
		got := txs[i].String()
		assert.Equal(t, got, exp)
	}
}

const sample = `
Account number,Trade date,Symbol,Description,Operation,Quantity,Price,Net amount
XXX,2019/09/10,VXF,"blah blah VALUE =      1630.59",DRIP,123.999,,
XXX,2019/10/10,VXF,whatever,Sold,5148.219,13.150,67689.08
XXX,2019/10/30,,MGMT FEE,Fee,0.00,,-12.40
`
