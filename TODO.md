### Features

* detect duplicate transactions between Asset accounts to help with ofx imports (in stats?)
* add opened/closed clauses to accounts (ditch the 0 balance filtering)
* allow date without year (interpret as closest to now)
* add -b/-e options to coin
* add weekly/monthly/yearly totals (to register?)
* add support for lots/costs
* sorting by quantity to aid finding largest transactions
* how to match :Acct when there's :XAcct as well?
* shortened full account names, drop letters from the left down to first on leftmost elements
* backfill prices from transactions
* support for tags?
* account/commodity renames?
* language server?

### Implementation

* clean up parsing, i.e. ditch the hobo parser and use PEG or something
* consider allowing multiple commodities in an account
