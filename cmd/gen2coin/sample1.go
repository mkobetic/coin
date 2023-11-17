package main

import (
	"strings"

	"github.com/mkobetic/coin"
)

const sample1setup = `
commodity CAD
  default

account Assets:Bank:Checking
account Liabilities:MasterCard
account Liabilities:Amex
account Income:Salary
account Expenses:Groceries
`

func sample1() []*rule {
	r := strings.NewReader(sample1setup)
	coin.Load(r, "")
	coin.ResolveAll()

	return []*rule{
		newRule(weekly(1, 2, weekday...),
			"Costco",
			"Checking|MasterCard|Amex",
			"Groceries",
			100, 1000,
		),
		newRule(weekly(2, 1, anyday...),
			"FARM BOY|FRESHCO|LOBLAWS|LOEB|SOBEY'S|NO FRILLS",
			"Checking|MasterCard|Amex",
			"Groceries",
			30, 200,
		),
		newRule(monthly(2, 1, 15, -1),
			"ACME Inc",
			"Salary",
			"Checking",
			2000, 2000,
		),
		newRule(monthly(1, 1, -1, -2, -3),
			"repay",
			"Checking",
			"MasterCard",
			TO_BALANCE, TO_BALANCE,
		),
		newRule(monthly(1, 1, -1, -2, -3),
			"repay",
			"Checking",
			"Amex",
			TO_BALANCE, TO_BALANCE,
		),
	}
}
