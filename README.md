Coin is a heavily simplified offshoot of ledger-cli.org. The idea of plain text accounting is brilliant, and ledger implements it beautifully. However ledger makes certain fundamental tradeoffs that have implications that some may find undesirable. For example its extreme flexibility in how amounts and commodities can be written (prefix/postfix/symbolic etc) forces commodities that include numbers to be quoted. That gets annoying when your ledger includes a lot of mutual fund names. Coin sacrifices this flexibility to avoid quoting.


## Ledger Differences
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

### coin
* commodities quotes command (yahoo)

### gc2coin
* gnucash import (XML v2 database only)

### ofx2coin
* ofx/qfx import
* ofx.rules file


## Implementation Notes

* Amount is implemented as big.Int plus number of decimal places. Consequently computations are truncated to the specified number of decimal places at every step.
* Amount includes the Commodity
