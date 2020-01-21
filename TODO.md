### Features

* balance: csv, json, chart output
* balance: add -b/-e options
* stats: check for commodity mismatches in accounts
* accounts: change `check commodity ==` clause to just `commodity`
* commodities: add `default` clause
* test: support for generating test data
* better usage messages
* make sure all errors include Location
* register: sorting by quantity to aid finding largest transactions
* register: more advanced filtering options
* how to match :Acct when there's :XAcct as well?
* document account selection expressions

### Maybe

* backfill prices from transactions
* add opened/closed clauses to accounts (ditch the 0 balance filtering)
* support for tags?
* account/commodity renames?
* language server?
* lots/costs
* multiple commodities in single account