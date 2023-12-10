This is the main coin command with subcommands modeled after ledger CLI.
Use `-h` for detailed option descriptions.

## balance

* print account balances
* select time range to total (begin/end)
* selecting postings by payee or tag name or name:value (regex)
* zero balance and closed account suppression (optional)
* filtering to top N levels of accounts for display

## register

* flat and recursive (including sub-accounts) posting listings
* aggregated amounts by week/month/quarter/year
* recursive and cumulative aggregation
* top n sub-account aggregations (the rest as Other)
* selecting postings in a time range (begin/end)
* selecting postings by payee or tag name or name:value (regex)
* text, json, csv and chart output formats

## accounts

* list accounts and commodities
* suppress closed accounts (optional)

## commodities

* list commodities
* -p print price stats
* -q to fetch current commodity quotes (yahoo)

## format

* reformat input file
* output ledger compatible format

## modify

* move postings to different account
* filtering by payee

## tags

* list all tags and optionally tag values

## stats

* print ledger stats
* duplicate transaction check
* unbalanced transaction check
* selecting transactions in a time range (-b/-e)

## test

* read a coin file and execute any test clauses found in it (see tests/ directory)

## version

* print version info