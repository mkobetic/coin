### Features

* allow date without year (interpret as closest to now)
* allow relative dates (-5m)
* add -b/-e options to coin
* add weekly/monthly/yearly totals (to register?)
* add opened/closed clauses to accounts (ditch the 0 balance filtering)
* sorting by quantity to aid finding largest transactions
* how to match :Acct when there's :XAcct as well?
* shortened full account names, drop letters from the left down to first on leftmost elements
* backfill prices from transactions

### Implementation

* replace regexp usage with rex

### Maybe

* support for tags?
* account/commodity renames?
* language server?
* lots/costs
* multiple commodities in single account