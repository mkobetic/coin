Converts the coin database into an HTML doc with a single page, interactive JS app for viewing.
Check out the [examples/yearly](https://mkobetic.github.io/coin/).

Account hierarchy nav on the left, segregated by account type (`Assets`, `Income`, `Expenses`, `Liabilities`, `Equity`).
Clicking a specific account restricts the detail view on the right to that account's hierarchy.
Currently selected account is shown as a heading above the details view. The individual names in the heading can be clicked to select the account represented by the name (this allows climbing up the account parent chain conveniently).

Details of the selected account are shown on the right with different presentation options (`Register`, `Chart`, ...).
The details show account transactions restricted to selected time range, controlled by the `From/To` inputs.
The `Closed Accounts` checkbox controls whether closed accounts are excluded from the account list.

# Register View

Shows the transaction details in tabular form

![Register Full](https://github.com/mkobetic/coin/assets/871693/d25a6cd8-9775-4261-a601-3d2173ec8a6c)

## SubAccounts

When un-checked only transactions for the selected account are shown.
When checked transactions of the account and any sub-accounts are shown.

![Register Full With SubAccounts](https://github.com/mkobetic/coin/assets/871693/011f46e4-2f1d-4566-ac6a-58f7b4b8d66f)

## Aggregate

When `None`, the individual transactions are shown.
When not `None`, the transactions are aggregated by the selected aggregation period (`Weekly`, `Monthly`, `Quarterly`, `Yearly`).
When aggregated with sub-accounts, the `SubAccount Max` option controls how many "top" sub-accounts should be shown; the rest of the sub-accounts are combined into the "Other" column. Top here means the sub-accounts with the highest average transaction value across the time range.

![Register Aggregated Monthly](https://github.com/mkobetic/coin/assets/871693/ca4897e1-54f3-4d94-93c7-c054b925f566)

### Aggregation Style

When transactions are being aggregated, the aggregation is performed using one of two styles. `Flows` style sums the incoming/outgoing amounts for the period, and is generally useful for Income and Expenses accounts. `Balances` style shows the final account balance at the end of the period, and is generally useful for Assets and Liabilities accounts.

## Show Notes

When Aggregate is set to None, and Show Notes is checked, each transaction is displayed with an additional row containing the transaction notes.

## Show Location

When Aggregate is set to None, and Show Location is checked, each transaction is displayed with an additional column showing the file location of the transaction in the path:lineNr format, e.g. examples/yearly/2010.coin:17, supported by some editors for navigation.

# Chart View

Chart shows aggregated transactions (including sub-accounts) by the selected aggregation period as a bar chart. The meaning of the available options is the same as for the Register aggregations.

![Chart Monthly](https://github.com/mkobetic/coin/assets/871693/7e265e93-131b-4a9e-b1db-3b201a53092b)
