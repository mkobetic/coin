**NB:** This is still a work in progress, a lot of information is missing here and will be filled in later.

Coin is a heavily simplified offshoot of [ledger-cli.org](https://www.ledger-cli.org/). The idea of [plain text accounting](https://plaintextaccounting.org/) is brilliant, and ledger implements it beautifully. However ledger makes certain fundamental tradeoffs that have implications that some may find undesirable. For example its extreme flexibility in how amounts and commodities can be written (prefix/postfix/symbolic etc) forces commodities that include numbers to be quoted. That gets annoying when your ledger includes a lot of mutual fund names. Coin sacrifices this flexibility to avoid quoting.


## COINDB

Coin is also written with the intent of maintaining a ledger split across number of files, this makes it easier to navigate a larger ledger with common editors and is more friendly with version control systems (e.g. git) which are also designed to manage multiple files.

`COINDB` is simply an environment variable pointing to a directory where `coin` expects to find the ledger files (`.coin`). The obvious organization schemes are splitting the ledger by year or month. While coin allows mixing any types of entries in the files (just like ledger) it looks for two special file names, `accounts.coin` and `commodities.coin`, and reads those first in order to satisfy the requirement that any commodities and accounts are defined upfront.

Optionally coin also supports reading prices from `prices.coin` or a set of files with `.prices` extension. The latter allows organizing prices records along the same structure as the ledger (e.g. by year or month) but separate from the transaction records.

Finally the rest of the `.coin` files is read and `coin` resolves and resorts everything once everything is read in. Coin ignores any other files in the directory (e.g. the .git directory if you use git to version the ledger)

For illustration here's how a sample COINDB directory could look like

```bash
2017.coin     2018.prices   accounts.coin       qfx/
2017.prices   2019.coin     commodities.coin
2018.coin     2019.prices   .git/
```


## Commands

Coin includes several binaries besides the main `coin` command.


### coin

`coin` is the main command mimicing the leger cli with a number of subcommands. The subcommands include the usual suspects like `balance` and `register`, but also `accounts`, `commodities` and `test`. For more details see `cmd/coin/README.md`.


### gc2coin

gnucash import (XML v2 database only), see [`cmd/gc2coin/README.md`](https://github.com/mkobetic/coin/blob/master/cmd/gc2coin/README.md)


### ofx2coin

ofx/qfx import, see [`cmd/ofx2coin/README.md`](https://github.com/mkobetic/coin/blob/master/cmd/gc2coin/README.md)

* ofx.rules file


## Assorted Ledger Differences
(besides vastly reduced set of commands/options and capabilities)

* no prefix commodities (i.e. $10)
* stricter naming restrictions for commodities (no whitespace, etc) => no need to quote
* single commodity accounts
* commodity symbol directive
* account check commodity == directive
* multi-file structure (*.coin, *.prices files) => COINDB directory
* all entities (accounts, prices, transactions,...) remember their position in the file (`Location()`) to aid tooling to provide quick access to them.
* no account inference => accounts.coin
* no commodity inference => commodities.coin
* account selection expressions


## Implementation Notes

* Amount is implemented as big.Int plus number of decimal places. Consequently computations are truncated to the specified number of decimal places at every step.
* Amount includes the Commodity
