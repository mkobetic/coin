### Features

* csv and json output for `balance`
* better usage messages
* accounts: change `check commodity ==` clause to just `commodity`
* commodities: add `default` clause
* add -b/-e options to `balance`
* fix the fixed amount widths in `balance`
* make sure all errors include Location
* sorting by quantity to aid finding largest transactions
* how to match :Acct when there's :XAcct as well?
* shortened full account names, drop letters from the left down to first on leftmost elements

### Maybe

* backfill prices from transactions
* add opened/closed clauses to accounts (ditch the 0 balance filtering)
* support for tags?
* account/commodity renames?
* language server?
* lots/costs
* multiple commodities in single account