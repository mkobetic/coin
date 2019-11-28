Converts OFX/QFX files into coin transactions

* loads account `$COINDB/accounts.coin` and commodities `$COINDB/commodities.coin`
* loads classification rules `$COINDB/ofx.rules`
* loads OFX/QFX files specified as cmd line arguments
* converts OFX bank/credit card transactions to coin transactions
  using the provided rules to match the transaction description/payees to target accounts.
* if match is not found the target account is set to `Unbalanced` and needs to be corrected manually
* outputs all transactions sorted by date

## Importing Steps

1. `export COINDB=~/coindb`
2. `ofx2coin *.qfx >new.coin`
2. edit `new.coin`
    * fix classification errors (update `$COINDB/ofx.rules` as necessary)
    * delete duplicate transactions for transfers between imported accounts
    (should be sorted close to each other assuming their dates are close)
    * add transaction comments (e.g. what was bought for larger items)
3. `coin format new.coin >clean.coin`
4. append `clean.coin` to the current transactions file
   `cat clean.coin >>$COINDB/<YYYY>.coin`

## Notes

If you get error `Invalid message set: BANKMSGSETV1` (observed with QFX files from Bank of Montreal) try the `-bmo` flag.
(See bmo.go for more details.)