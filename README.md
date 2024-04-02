![tests](https://github.com/mkobetic/coin/.github/workflows/test.yml/badge.svg)

**NB:** This is still a work in progress, a lot of information is missing and is filled in gradually. There isn't much in terms of user documentation, but most features follow the same pattern as ledger-cli. You can learn most things from ledger-cli's excellent documentation.

Coin is a heavily simplified offshoot of [ledger-cli.org](https://www.ledger-cli.org/). The idea of [plain text accounting](https://plaintextaccounting.org/) is brilliant, and ledger implements it beautifully. However ledger makes certain fundamental tradeoffs that have implications that some may find undesirable. For example its extreme flexibility in how amounts and commodities can be written (prefix/postfix/symbolic etc) forces commodities that include numbers to be quoted. That gets annoying when your ledger includes a lot of mutual fund names. Coin sacrifices this flexibility to avoid quoting.

Coin is written with the intent of maintaining a ledger split across number of files. This makes it easier to navigate a larger ledger with common editors and is more friendly with version control systems (e.g. git) which are also designed to manage multiple files.

![coin example in vscode](https://github.com/mkobetic/coin/assets/871693/a9998caf-7bd5-4d04-9990-0f319e06ff87)

## COINDB

`COINDB` is simply an environment variable pointing to a directory where `coin` expects to find the ledger files (`.coin`). The obvious organization schemes are splitting the ledger by year, quarter or month. While coin allows mixing any types of entries in the files (just like ledger) it looks for two special file names, `accounts.coin` and `commodities.coin`, and reads those first in order to satisfy the strict requirement that any commodities and accounts are defined upfront.

Optionally coin also supports reading prices from `prices.coin` or a set of files with `.prices` extension. The latter allows organizing price records along the same structure as the ledger (e.g. by year or month) but separate from the transaction records.

Finally the rest of the `.coin` files are read and `coin` resolves and resorts everything once all of the ledger is read in. Coin ignores any other files in the directory (e.g. the `.git` directory if you use git to version you ledger)

For illustration here's how a sample COINDB directory could look like

```bash
2017.coin     2018.prices   accounts.coin       qfx/
2017.prices   2019.coin     commodities.coin    csv/
2018.coin     2019.prices   .git/
```


## Commands

Coin includes several executables.


### coin

`coin` is the main command mimicking the leger cli with a number of subcommands. The subcommands include the usual suspects like `balance` and `register`, but also `accounts`, `commodities` and `test`. For more details see [`cmd/coin/README.md`](https://github.com/mkobetic/coin/blob/master/cmd/coin/README.md).


### gc2coin

gnucash import (XML v2 database only), see [`cmd/gc2coin/README.md`](https://github.com/mkobetic/coin/blob/master/cmd/gc2coin/README.md)


### ofx2coin

ofx/qfx import, see [`cmd/ofx2coin/README.md`](https://github.com/mkobetic/coin/blob/master/cmd/ofx2coin/README.md)


### csv2coin

csv import, see [`cmd/csv2coin/README.md`](https://github.com/mkobetic/coin/blob/master/cmd/csv2coin/README.md)


### gen2coin

generates ledger samples for testing or demos, see [`cmd/gen2coin/README.md`](https://github.com/mkobetic/coin/blob/master/cmd/gen2coin/README.md)

### coin2html

converts the ledger into a single-page html viewer, see [`cmd/coin2html/README.md`](https://github.com/mkobetic/coin/blob/master/cmd/coin2html/README.md)

![Register Full](https://github.com/mkobetic/coin/assets/871693/8bc5704e-cd03-42c0-b4ca-828c51f8fec8)

![Chart Monthly](https://github.com/mkobetic/coin/assets/871693/d640b4a4-1cd2-4faf-8fbd-0c10f3cc90b3)

## Assorted Ledger Differences
(besides vastly reduced set of commands/options and capabilities)

* multi-file structure (*.coin, *.prices files) => $COINDB directory
* all entities (accounts, prices, transactions,...) remember their position in the file (`Location()`) to aid tooling to provide quick access to them.

### Date Entry

Date entry applies to both specifying dates as command line option values or as the input in the coin files.
Note that using entries that are relative to today in coin files is only intended for initial entry into the ledger,
and should be promptly reformatted with `coin format` to turn them into absolute entries, otherwise your ledger can become corrupted fairly quickly.

* Dates are entered as [[YY]YY/]M/D, where year can be 2 or 4 digits
  * if year is 2 digits, it is the year closest to today (e.g. in 2020, 92 is 1992 and 55 is 2055).
  * if year is omitted its the date closest to today (e.g. on 2020/03/05, 6/22 is 2020/06/22 and 10/22 is 2019/10/22)
* Date can also be entered as YYYY[/M]
  * if month is omitted the date is set to Jan 1st, if month is present it's the first day of that month
* Date can also include a suffix specifying offset of +/- number of days,weeks,months or years from that date (e.g. -50d)
  * if only offset is specified, it is offset from today (e.g. on 2020/03/05, +2m is 2020/05/05 and -2d is 2020/03/03)

### Account Entry

Accounts can be entered as full account name path or as shortened versions of the same where some parts of the path are elided as long as the shortened path matches an existing account unambiguously. A full name path match always overrides any partial path matches. Double colon `::` matches any number of path elements.

For example the following expressions could match account `Assets:Investments:Broker:RRSP:Joe:VGRO`:
* `Joe:VGRO`
* `A:I::Joe`
* `Broker::VGRO`

### Amount differences

* an amount is always associated with a commodity
* amount precision is dictated by the associated commodity, there's no inference of precision from the amount values

### Commodity differences

* no prefix commodities (i.e. $10)
* stricter naming restrictions for commodities (no whitespace, etc) => no need to quote
* commodity symbol directive - used for transaction and price imports
* default directive - used to identify the default account commodity
* no commodity inference => commodities.coin

### Account differences

* single commodity accounts
* account commodity directive
* account selection expressions (see Account Entry above)
* no account inference => accounts.coin

### Transaction differences

* only date, code, description/payee, and note/comment is recognized in transaction header
* only account, quantity and optional balance is recognized in any transaction posting
* posting note/comment is supported as well
* any combination of 'short notes' (appended at the end of the transaction or posting line)
  and 'long notes' on separate lines following the transaction or posting line is possible
* tags are parsed out of notes, simple tag #key or value tags #key: some value, (value terminated by a comma or EOL) are supported

### Other types of ledger entries

* Include entry is supported and can be used to inject content of other files in place of the include entry

## Implementation Notes

* Amount is implemented as big.Int plus number of decimal places. Computations are truncated to the specified number of decimal places at every step.
* Amount always includes Commodity
* Everything is loaded into memory on start, so there is a theoretical limit on the total size of data.
* Trying to keep dependencies to a minimum
