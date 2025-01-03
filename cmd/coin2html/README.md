Converts the coin database into an HTML doc with an embedded single page, interactive JS app for viewing. No server needed, just open the file in any browser.
Check out [examples/yearly](https://mkobetic.github.io/coin/).

![Balances](https://github.com/user-attachments/assets/b423cd88-13f7-448c-8021-8cc5f2c237e5)

Account hierarchy nav on the left, segregated by account type (`Assets`, `Income`, `Expenses`, `Liabilities`, `Equity`).

Clicking a specific account restricts the detail view on the right to that account's hierarchy.
Currently selected account is shown as a heading above the details view. Individual segments of the account name in the heading can be clicked to select the parent account represented by that name, this allows climbing up the account parent chain conveniently.

Details of the selected account are shown on the right with different presentation options (`Balances`, `Register`, `Balances - Chart`, ...).
The details show account postings/balances restricted to selected time range, controlled by the `From/To` inputs.
The `Closed Accounts` checkbox controls whether closed accounts are excluded (i.e. accounts closed before the `To` date).

![Account View Options](https://github.com/user-attachments/assets/0f31c9c4-121c-4c48-bffa-bae92dbb5266)

# Balances View

Lists the balances of the selected account and its sub-accounts as of the `To` date. `Balance` is the balance of the account itself, `Total` sums up balances of the account and its sub-accounts.

# Register View

Shows the account transaction details in tabular form. `Account` column shows the other side of the transaction. `Balance` is the balance of the account as of that posting. `Cum.Total` shows the running total of the listed transaction amounts.

![Screenshot 2025-01-03 at 15 00 22](https://github.com/user-attachments/assets/7cd0770d-420f-47d6-91e3-78ee9d5f8b4b)

### SubAccounts

When un-checked only transactions for the selected account are shown.
When checked transactions of the account and any sub-accounts are shown. Sub-account is indicated in the additional `SubAccount` column.

![Register Full With SubAccounts and Location](https://github.com/user-attachments/assets/1a7ea9a0-bca5-4058-812b-6cd329f57f51)

### Show Notes

When `Show Notes` is checked, each transaction is displayed with an additional row containing the transaction notes.

### Show Location

When `Show Location` is checked, each transaction is displayed with an additional column showing the file location of the transaction in the path:lineNr format, e.g. examples/yearly/2010.coin:17; supported by some editors for file navigation.

## Aggregation

When `Aggregate` is not `None`, the transactions are aggregated by the selected aggregation period (`Weekly`, `Monthly`, `Quarterly`, `Yearly`).

![Aggregated Monthly Flows](https://github.com/user-attachments/assets/310b63c6-97eb-4444-b40b-a7dc65163f6f)

### Aggregation Style

When transactions are being aggregated, the aggregation is performed using one of two styles. `Flows` style sums the incoming/outgoing amounts for the period, and is generally interesting for Income and Expenses accounts. `Balances` style shows the final account balance at the end of the period, and is generally useful for Assets and Liabilities accounts.

![Aggregated Yearly Balances](https://github.com/user-attachments/assets/345150af-6ecc-48e5-b107-6ea2907b88f7)

### Aggregation Max SubAccounts

When aggregating with sub-accounts, the `SubAccount Max` option controls how many "top" sub-accounts should be shown; the rest of the sub-accounts are combined into the "Other" column. Top here means the sub-accounts with the highest average transaction value across the time range.

![Register Aggregated Yearly With SubAccounts](https://github.com/user-attachments/assets/f219077f-c00e-41a8-96ee-8f31ba0a385e)

The `Total` column sums up the amounts in the row. The `Cum.Total` is the running total of the rows.

### Aggregation Details

When aggregating each row represents a group of postings. If aggregating with subaccounts each column in the row represents different group of postings. Clicking the amount field opens a `Details` view that lists top 20 postings by the absolute value of their amounts. This can be useful to find any outliers.

![Aggregation Posting Group Details](https://github.com/user-attachments/assets/181a939a-ec25-4b02-a68f-86cebae93063)

# Balanaces - Chart View

Balances chart can be useful with a larger account hierarchy. It provides visual representation of the proportions of the individual account.

![Balances Chart](https://github.com/user-attachments/assets/2d042ca5-d112-4297-b2c5-3779ac67556c)

The `Depth` controls how deep down the account hierarchy it should go.

Clicking a rectangle makes the corresponding account the selected account, this can be useful when exploring larger account structure.
A hover tooltip shows the full balance details when the text in the chart is clipped due to the rectangle layout.

# Aggregated Register - Chart View

This chart is a visual analog of the tabular Register view when aggregating with sub-accounts. The meaning of the available options is the same as for the Register aggregations.

![Aggregated Register Chart Monthly](https://github.com/user-attachments/assets/6eee81a7-3fd2-4a78-89e3-515b54019375)

A hover tooltip shows the amount that each rectangle represents. As with aggregated Register view, clicking a rectangle shows the top 20 posting details

![Register Chart with Posting Details](https://github.com/user-attachments/assets/0fd171c9-b018-44cc-8d55-f65c806884df)



