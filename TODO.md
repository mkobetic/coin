### Features

* register: cache balance on posting, show global totals with begin/end
* stats: check closed accounts have 0 balance
* register: show posting commodity (not just total commodity)
* register: show description/notes
* register: recursive register prints transactions within the parent tree twice
* register: recurisve register totals are useless
* better usage messages
* balance: csv, json, chart output
* test: support for generating test data
* register: sorting by quantity to aid finding largest transactions
* register: more advanced filtering options

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