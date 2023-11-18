Generates a ledger sample based on a specified time range and a set of internal rules (see rules_*.go files).

```
% gen2coin -h  
Usage: gen2coin [flags] [directory path]

Generates a ledger sample based on internally defined rules.
If directory path is absent, output transactions to stdout.
Otherwise generates accounts, commodities and transactions files as directed.

Flags:
  -b value
        begin ledger on or after this date (default: -3 months)
  -e value
        end ledger on or before this date (default: today)
  -m    split ledger into multiple files by month
  -y    split ledger into multiple files by year
```
