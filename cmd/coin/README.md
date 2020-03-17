This is the main coin command with subcommands modeled after ledger CLI.
Use `-h` for detailed option descriptions.

## balance

* print account balances

## register

* flat and recursive (including sub-accounts) posting listings
* aggregated amounts by week/month/quarter/year
* recursive and cumulative aggregation
* top n sub-account aggregations (the rest as Other)
* selecting postings in a time range (begin/end)
* text, json, csv and chart output formats

## accounts

* list accounts and commodities

## commodities

* list commodities
* -p print price stats
* -q to fetch current commodity quotes (yahoo)

## format

* reformat input file
* output ledger compatible format

## stats

* print ledger stats
* duplicate transaction check
* unbalanced transaction check
* selecting transactions in a time range (-b/-e)

## test

* read a coin file and execute any test clauses found in it (see tests/ directory)

## version

* print version info