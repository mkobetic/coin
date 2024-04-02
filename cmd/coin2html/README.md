Converts the coin database into an HTML doc with a single page, interactive JS app for viewing.

Account hierarchy nav on the left, segregated by account type (Assets, Income, Expenses, Liabilities, Equity).
Selecting a specific account restricts the detail view to that account's hierarchy.

Details of selected accounts on the right with different presentation possibilities (Table, Chart, ...).
Displayed transactions are restricted to selected time range.

# Register View

Shows the transaction details in tabular form

![Register Full](https://github.com/mkobetic/coin/assets/871693/d25a6cd8-9775-4261-a601-3d2173ec8a6c)

## SubAccounts

When un-checked only transactions for the selected account are shown.
When checked transaction of the account and any subaccounts are shown.

![Register Aggregated Monthly](https://github.com/mkobetic/coin/assets/871693/ca4897e1-54f3-4d94-93c7-c054b925f566)

## Aggregate

When None, the individual transactions are shown.
When not None, the transactions are aggregated by the selected aggregation period (Weekly, Monthly, Quarterly, Yearly).
When aggregated with subaccounts, the SubAccount Max option controls how many "top" subaccounts should be shown; the rest of the subaccounts are combined into an "Other" column. Top means the subaccounts with the highest average transaction value across the time range.

# Chart View

Chart shows aggregated transactions (including subaccounts) by the selected aggregation period as a bar chart. The meaning of the available options is the same as for the Register aggregations.

![Chart Monthly](https://github.com/mkobetic/coin/assets/871693/7e265e93-131b-4a9e-b1db-3b201a53092b)
