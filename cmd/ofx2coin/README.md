Converts OFX/QFX files into coin transactions

* loads account `$COINDB/accounts.coin` and commodities `$COINDB/commodities.coin`
* loads classification rules `$COINDB/ofx.rules`
* loads OFX/QFX files specified as cmd line arguments
* converts OFX bank/credit card transactions to coin transactions
  using the provided rules to match the transaction description/payees to target accounts.
* if match is not found the target account is set to `Unbalanced` and needs to be corrected manually
* attaches balance to the last imported transaction
* performs basic duplicate detection and removes duplicate transactions unless told not to (removals are reported)
* outputs all transactions sorted by date

## ofx.rules

The file contains groupes of rules, one rule per line. The groups are associated either with a specific account or with a label that can be used to include that group in other groups to allow sharing of rules between accounts.

When importing transactions for given account the tool will apply the rule group associated with that account. The account is matched through the account ID associated with the transactions. The same ID must also be associated with an account through the `ofx_acctid` directive.

Each rule group starts with a line containing either a label, or an account ID and full account name. This is followed by lines starting with whitespace containing either a group reference or a rule.

A group reference is simply a group name prefixed with `@`. Referencing a group includes all the rules of the referenced group in the referencing group.

A rule is a full account name followed by a list of regular expressions separated with `|`. The rules and regular expressions are matched against transaction descriptions in the order in which they are listed. The search stops on the first match and the corresponing account is used as the transaction counterpart of the imported account.

```
common
  Expenses:Groceries       FRESHCO|COSTCO WHOLESALE|FARM BOY|LOBLAWS
  Expenses:Auto:Gas        COSTCO GAS|PETROCAN|SHELL
389249328477983 Assets:Bank:Savings
  @common
  Income:Interest     Interest
392843029797099 Assets:Bank:Checking
  @common
  Income:Salary       ACME PAY 
```

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