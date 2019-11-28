Coin is a heavily simplified offshoot of ledger-cli.org. The idea of plain text accounting is brilliant, and ledger implements it beautifully. However ledger makes certain fundamental tradeoffs that have implications that some may find undesirable. For example its extreme flexibility in how amounts and commodities can be written (prefix/postfix/symbolic etc) forces commodities that include numbers to be quoted. That gets annoying when your ledger includes a lot of mutual fund names. Coin sacrifices this flexibility to avoid quoting.


## Adding new transactions

Instead of editing the file directly in place, following process can be easier
* import/write new transactions into a new file `new.coin`
* replace all `Unbalanced` references with existing accounts
* use `coin stats` to verify final balances and fix what's wrong
* move the target file, e.g. `2019.coin`, to drop the coin extension, e.g. `2019`
* append `new.coin` to it `cat new.coin >>2019`
* run the combined file through `coin format`
  `coin fmt 2019 >2019.coin`
* check the diffs in the target file, e.g. `git diff 2019.coin`
* move `new.coin` to `new` and check stats again `coin stats`
* delete `new` and `2019` and commit the changes

## Ledger Differences

* no prefix commodities (i.e. $10)
* stricter naming restrictions for commodities (no whitespace, etc) => no need to quote
* single commodity accounts
* commodity symbol directive
* account check commodity == directive
* multi-file structure (*.coin, *.prices files) => COINDB directory
* no account inference => accounts.coin
* no commodity inference => commodities.coin
* account selection expressions
* commodities quotes command (yahoo)
* gc2coin: gnucash import (XML v2 database only)
* ofx2coin: ofx/qfx import
* ofx.rules file

## TODO

### Features

* detect duplicate transactions between Asset accounts to help with ofx imports (in stats?)
* add -b/-e options to coin
* add weekly/monthly/yearly totals (to register?)
* sorting by quantity to aid finding largest transactions
* how to match :Acct when there's :XAcct as well?
* shortened full account names, drop letters from the left down to first on leftmost elements
* backfill prices from transactions
* account/commodity renames?

#### VSCode integration

* vscode snippets for updates?
* links to transactions from editor problem reports?
* language server?


### Implementation

* clean up parsing, i.e. ditch the hobo parser and use PEG or something
* consider allowing multiple commodities in an account
* track transaction position in files


## Implementation Notes

* Amount is implemented as big.Int plus number of decimal places. Consequently computations are truncated to the specified number of decimal places at every step.
