### Features

* detect duplicate transactions between Asset accounts to help with ofx imports (in stats?)
* add -b/-e options to coin
* add weekly/monthly/yearly totals (to register?)
* sorting by quantity to aid finding largest transactions
* how to match :Acct when there's :XAcct as well?
* shortened full account names, drop letters from the left down to first on leftmost elements
* backfill prices from transactions
* account/commodity renames?
* language server?

### Implementation

* clean up parsing, i.e. ditch the hobo parser and use PEG or something
* consider allowing multiple commodities in an account
