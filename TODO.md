### Features

* thousand separator in amounts
* check for account/cc numbers in transactions
* balance: last reconciled posting date
* register: sorting by quantity to aid finding largest transactions
* register: show account balances with begin/end
* stats: check closed accounts have 0 balance
* register: show posting commodity (not just total commodity)
* register: recursive prints transactions within the parent tree twice
* register: recursive totals are useless
* balance: csv, json, chart output
* register/balance: markdown output
* register: more advanced filtering options
* stats: aggregate transaction/price stats by time (-y, -q, -m) and begin/end

#### ofx2coin

* duplicate elimination too aggressive with identical transactions (e.g. 2x ROGERS top up for cell phones)
  ? duplicate transactions from the same source/file should be kept?
* sanitize sensitive information, account/cc numbers
* commodity mismatches (USD vs CAD)
* use ofxid for deduping (need tags?)

### Issues

* how to match :Acct when there's :XAcct as well?

### Maybe

* backfill prices from transactions
* filter out closed accounts where it makes sense (ditch the 0 balance filtering)
* commodity renames?
* language server?
* lots/costs
* multiple commodities in single account?
* query language