Converts GnuCash XML database (v2) to a coin/ledger style database. Either `$GNUCASHDB` env var
or `-gnucashdb` option must be set and point to the GnuCash database file.

If `$COINDB` is set the output is split into separate files for commodities, accounts, prices and transactions
in that directory. Otherwise everything goes to stdout.

If `-y` is used, price and transaction files are further split by year.

The '-l' option forces a ledger friendly output, which primarily means quoting commodities that ledger requires to be quoted.