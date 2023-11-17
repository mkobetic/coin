package main

import (
	"strings"

	"github.com/mkobetic/coin"
)

const sample1setup = `
commodity CAD
  default

account Assets:Bank:Checking
account Assets:Bank:Savings
account Liabilities:MasterCard
account Liabilities:Amex
account Income:Salary
account Income:Interest
account Expenses:Groceries
account Expenses:Heat
account Expenses:Electricity
account Expenses:Water
`

func sample1() []*rule {
	r := strings.NewReader(sample1setup)
	coin.Load(r, "")
	coin.ResolveAll()

	return []*rule{
		newRule(weekly(1, 2, weekday...), "Costco",
			"Checking|MasterCard|Amex",
			"Groceries",
			100, 1000),
		newRule(weekly(2, 1, anyday...),
			"FARM BOY|FRESHCO|LOBLAWS|LOEB|SOBEY'S|NO FRILLS",
			"Checking|MasterCard|Amex",
			"Groceries",
			30, 200),
		newRule(monthly(2, 1, 15, -1), "ACME Inc", "Salary", "Checking", 2000, 2000),
		newRule(monthly(1, 1, -1, -2, -3), "repay balance", "Checking", "MasterCard", TO_BALANCE, 1),
		newRule(monthly(1, 1, -1, -2, -3), "repay balance", "Checking", "Amex", TO_BALANCE, 1),
		newRule(monthly(1, 1, 1), "Enbridge Gas", "Checking", "Heat", 100, 200),
		newRule(monthly(1, 1, 1), "Hydro One", "Checking", "Electricity", 50, 100),
		newRule(monthly(1, 2, -15), "City Water&Sewer", "Checking", "Water", 100, 150),
		newRule(monthly(1, 1, 1), "monthly savings", "Checking", "Savings", 300, 500),
		newRule(monthly(1, 1, -1), "monthly interest %2/12", "Interest", "Savings", TO_BALANCE, -50*12),
	}
}
