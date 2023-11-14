### Features

* thousand separator in amounts
* check for account/cc numbers in transactions
* fmt: remove extra whitespace from payee
* move postings between accounts/rename account
* balance: last reconciled posting date
* register: filter by payee
* register: sorting by quantity to aid finding largest transactions
* register: cache balance on posting, show global totals with begin/end
* stats: check closed accounts have 0 balance
* register: show posting commodity (not just total commodity)
* register: show description/notes
* register: recursive prints transactions within the parent tree twice
* register: recursive totals are useless
* better usage messages
* balance: csv, json, chart output
* test: support for generating test data
* register: more advanced filtering options
* stats: aggregate transaction/price stats by time (-y, -q, -m) and begin/end

#### ofx2coin

* duplicate elimination too aggressive with identical transactions (e.g. 2x ROGERS top up for cell phones)
  ? duplicate transactions from the same source/file should be kept?
* sanitise sensistive information, account/cc numbers
* commodity mismatches (USD vs CAD)

### Issues

* how to match :Acct when there's :XAcct as well?

### Maybe

* backfill prices from transactions
* filter out closed accounts where it makes sense (ditch the 0 balance filtering)
* support for tags?
* account/commodity renames?
* language server?
* lots/costs
* multiple commodities in single account?
* query language