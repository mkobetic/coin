package main

import (
	"strings"
	"testing"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/assert"
)

var rules *Rules

func init() {
	coin.DefaultCommodityId = "USD"

	r := strings.NewReader(`
commodity USD
commodity CAD
commodity VXF
  format 1.000 VXF
  symbol VXF
commodity VAB
  format 1 VAB
commodity VCN
  format 1 VCN
commodity BND
  format 1 BND
commodity XBAL
  format 1 XBAL
account Assets:Investments
  csv_acctid XXX
account Assets:Investments:VXF
  commodity VXF
account Assets:Investments:BBB
  csv_acctid BBB
account Assets:Investments:BBB:CAD
  commodity CAD
account Assets:Investments:BBB:VAB
  commodity VAB
account Assets:Investments:BBB:BND
  commodity BND
account Assets:Investments:BBB:VCN
  commodity VCN
account Assets:Investments:BBB:XBAL
  commodity XBAL
account Expenses:Fees
account Income:Dividends
account Income:Interest
  commodity CAD
`)
	coin.Load(r, "")
	coin.ResolveAccounts()

	coin.AccountsDo(func(a *coin.Account) {
		if a.CSVAcctId != "" {
			accountsByCSVId[a.CSVAcctId] = a
		}
	})

	rules = ReadRules(strings.NewReader(`src 1
  account 0
  description 4
  date 1
  amount 7
  amount 3 "-$1" VALUE = \s+([\d\.\-]+)
  symbol 2
  quantity 5
  note 3
bbb 3
  account "BBB"
  activity 1
  date 0 "$1/$2/$3" (\d\d\d\d)-(\d\d)-(\d\d)
  amount 7
  currency 8
  symbol_ref 3
  quantity_ref 4
  description "${activity} ${quantity_ref} ${symbol_ref} in ${currency}"
  symbol "${symbol_ref}" activity Buy|Sell
  quantity "${quantity_ref}" activity Buy|Sell
---
XXX Assets:Investments
  Expenses:Fees Fee|HST
  Income:Dividends DRIP
  Assets:Investments Sold|Bought
BBB Assets:Investments:BBB
  Income:Interest Interest
  Income:Dividends Dividend
  Assets:Investments:BBB Sell|Buy

`))
}

const sample1 = `
Account number,Trade date,Symbol,Description,Operation,Quantity,Price,Net amount
XXX,2019/09/10,VXF,"blah blah VALUE =      1630.59",DRIP,123.999,,
XXX,2019/10/10,VXF,whatever,Sold,5148.219,13.150,67689.08
XXX,2019/10/30,,MGMT FEE,Fee,0.00,,-12.40
`

func Test_Sample1(t *testing.T) {
	r := strings.NewReader(sample1)
	txs := readTransactions(r, rules.sources["src"], rules)
	for i, exp := range []string{
		// symbol=VXF, quantity=123.999, amount=1630.59
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

const sample2 = `
Transaction Type=All,Product Type=All,Symbol=,From=2021-12-01,To=2022-01-01
Date,Activity Description,Description,Symbol,Quantity,Price,Currency,Total Amount,Currency
----------------,---------------,--------------------,-----------,------,--------,-----,--------,------------
2021-12-16,Buy,ISHARES CORE BAL ETF PORT INDE,XBAL,151.,27.60,CDN,-4181.00,CAD
2021-12-08,Interest,VANGUARD CDN AGGREGATE BD INDE,VAB,230.,,,12.57,CAD
2021-12-09,Sell,VANGUARD TOTAL BD MKT ETF,BND,-50,,,4269.56,USD
2021-12-09,Sell,VANGUARD FTSE CDA ALL CAP INDE,VCN,-55,42.91,CDN,2360.05,CAD
2021-12-06,Dividend,VANGUARD TOTAL BD MKT ETF,BND,500,,,67.46,USD
`

func Test_Sample2(t *testing.T) {
	r := strings.NewReader(sample2)
	txs := readTransactions(r, rules.sources["bbb"], rules)
	for i, exp := range []string{
		`2021/12/16 Buy 151. XBAL in CAD
  Assets:Investments:BBB:XBAL       151 XBAL
  Assets:Investments:BBB:CAD   -4181.00 CAD
`,
		`2021/12/08 Interest 230. VAB in CAD
  Assets:Investments:BBB:CAD   12.57 CAD
  Income:Interest             -12.57 CAD
`,
		`2021/12/09 Sell -50 BND in USD
  Assets:Investments:BBB      4269.56 USD
  Assets:Investments:BBB:BND      -50 BND
`,
		`2021/12/09 Sell -55 VCN in CAD
  Assets:Investments:BBB:CAD  2360.05 CAD
  Assets:Investments:BBB:VCN      -55 VCN
`,
		`2021/12/06 Dividend 500 BND in USD
  Assets:Investments:BBB   67.46 USD
  Income:Dividends        -67.46 USD
`,
	} {
		got := txs[i].String()
		assert.Equal(t, got, exp)
	}
}
