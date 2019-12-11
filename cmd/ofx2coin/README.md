Converts OFX/QFX files into coin transactions

* loads account `$COINDB/accounts.coin` and commodities `$COINDB/commodities.coin`
* loads classification rules `$COINDB/ofx.rules`
* loads OFX/QFX files specified as cmd line arguments
* converts OFX bank/credit card transactions to coin transactions
  using the provided rules to match the transaction description/payees to target accounts.
* if match is not found the target account is set to `Unbalanced` and needs to be corrected manually
* outputs all transactions sorted by date


## Sugested Import Procedure

* assuming we're working in the directory where the coin files are located
    `cd $COINDB`
* create a coin file from the ofx files
    `ofx2coin *.qfx >new.coin`
* clean up `new.coin`
    * delete duplicate transactions
        `coin stats -d`
    * use `coin stats` to verify final balances and fix what's wrong
    * replace all `Unbalanced` references with existing accounts
        `coin stats -u`
    * fix classification errors (update `$COINDB/ofx.rules` as necessary)
    * add transaction comments (e.g. what was bought for larger items)
* move the target file to drop the coin extension, e.g.
    `mv 2019.coin 2019`
* append `new.coin` to it
    `cat new.coin >>2019`
* run the combined file through `coin format`
    `coin fmt 2019 >2019.coin`
* check the diffs in the target file, e.g.
    `git diff 2019.coin`
* move `new.coin` to `new` and check stats again `coin stats`
* delete `new` and `2019` and commit the changes


## Notes

If you get error `Invalid message set: BANKMSGSETV1` (observed with QFX files from Bank of Montreal) try the `-bmo` flag.
(See bmo.go for more details.)