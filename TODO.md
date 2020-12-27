### Features

* stats: check for commodity mismatches in accounts
* register: mark cleared postings (postings with balance assertions)
* balance: csv, json, chart output
* test: support for generating test data
* better usage messages
* make sure all errors include Location
* register: sorting by quantity to aid finding largest transactions
* register: more advanced filtering options

### Issues

* how to match :Acct when there's :XAcct as well?

### Maybe

* backfill prices from transactions
* add opened/closed clauses to accounts (ditch the 0 balance filtering)
* support for tags?
* account/commodity renames?
* language server?
* lots/costs
* multiple commodities in single account