package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/check"
)

var (
	fields      = flag.String("fields", "", "ordered list of column indexes to use as transaction fields")
	source      = flag.String("source", "", "which source rules to use to read the files")
	dumpAcctIDs = flag.Bool("ids", false, "dump accounts with known csv ids")
	dumpRules   = flag.Bool("rules", false, "dump the loaded account rules (useful for formatting)")

	accountsByCSVId = map[string]*coin.Account{}
)

func main() {
	flag.Parse()

	coin.LoadFile(coin.CommoditiesFile)
	coin.LoadFile(coin.AccountsFile)
	coin.ResolveAccounts()

	coin.AccountsDo(func(a *coin.Account) {
		if a.CSVAcctId != "" {
			accountsByCSVId[a.CSVAcctId] = a
		}
	})

	if *dumpAcctIDs {
		coin.AccountsDo(func(a *coin.Account) {
			if a.CSVAcctId != "" {
				fmt.Printf("%s %s\n", a.CSVAcctId, a.FullName)
			}
		})
		return
	}

	var rules *Rules
	fn := filepath.Join(coin.DB, "csv.rules")
	if _, err := os.Stat(fn); !os.IsNotExist(err) {
		file, err := os.Open(fn)
		check.NoError(err, "Failed to open %s", fn)
		defer file.Close()
		rules, err = ReadRules(file)
		check.NoError(err, "Failed to parse %s", fn)
	}

	if *dumpRules {
		rules.Write(os.Stdout)
		return
	}

	var columns []int
	if *fields != "" {
		columns = parseFields(*fields)
	} else if *source != "" {
		columns = rules.fields[*source]
		check.If(columns != nil, "Unknown source %s", *source)
	} else {
		fmt.Fprintf(os.Stderr, "One of -source or -fields must be specified\n")
		os.Exit(1)
	}

	var transactions []*coin.Transaction
	for _, fileName := range flag.Args() {
		file, err := os.Open(fileName)
		check.NoError(err, "Failed to open %s", fileName)
		defer file.Close()

		r := csv.NewReader(file)
		header, err := r.Read()
		check.NoError(err, "Failed to read header from %s", fileName)

		for i, h := range header {
			if label := toLabel(columns, i); label != "" {
				fmt.Fprintf(os.Stderr, "%s => %s\n", h, label)
			} else {
				fmt.Fprintf(os.Stderr, "%s\n", h)
			}
		}
		batch, err := readTransactions(r, columns, rules)
		check.NoError(err, "Cannot parse file %s", fileName)
		transactions = append(transactions, batch...)
	}

	// write transactions
	sort.SliceStable(transactions, func(i, j int) bool {
		return transactions[i].Posted.Before(transactions[j].Posted)
	})

	for _, t := range transactions {
		t.Write(os.Stdout, false)
		fmt.Fprintln(os.Stdout)
	}
}

func readTransactions(r *csv.Reader, columns []int, rules *Rules) (transactions []*coin.Transaction, err error) {
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transactionFrom(rec, columns, rules))
	}
	return transactions, nil
}

var labels = []string{"account_id", "description", "posted", "amount", "symbol", "quantity", "note"}

// transactionFrom builds a transaction from a csv row.
// columns list field indexes in the following order:
// account_id, description, posted, amount, symbol, quantity, note
func transactionFrom(row []string, columns []int, rules *Rules) *coin.Transaction {
	acctId := row[columns[0]]
	description := row[columns[1]]
	account, found := accountsByCSVId[acctId]
	check.OK(found, "Can't find account with id %s", acctId)
	toAccount := rules.AccountRulesFor(acctId).AccountFor(description)
	if toAccount == nil {
		toAccount = coin.Unbalanced
	}

	posted, err := time.Parse(coin.DateFormat, row[columns[2]])
	check.NoError(err, "Parsing date %s", row[columns[2]])

	t := &coin.Transaction{
		Description: description,
		Posted:      posted,
	}

	if columns[6] >= 0 {
		t.Note = row[columns[6]]
	}
	var amount *coin.Amount
	if amt := row[columns[3]]; amt != "" {
		amtf, err := strconv.ParseFloat(amt, 64)
		check.NoError(err, "Parsing amount %s", amt)
		if math.Abs(amtf) > 0.0001 {
			amount = account.Commodity.NewAmountFloat(amtf)
		}
	}
	symbol := row[columns[4]]
	if symbol == "" {
		if amount == nil {
			panic("wat? " + row[columns[3]])
		}
		t.Post(account, toAccount, amount, nil)
		return t
	}
	commodity, found := coin.CommoditiesBySymbol[symbol]
	check.OK(found, "Could not find commodity for symbol %s", symbol)

	var quantity *coin.Amount
	if amt := row[columns[5]]; amt != "" {
		amtf, err := strconv.ParseFloat(amt, 64)
		check.NoError(err, "Parsing quantity %s", amt)
		if math.Abs(amtf) > 0.0001 {
			quantity = commodity.NewAmountFloat(amtf)
		}
	}

	if amount == nil {
		t.Post(account, toAccount, quantity, nil)
		return t
	}

	t.PostConversion(account, amount, nil, toAccount, quantity, nil)
	return t
}

func isZero(f float64) bool {
	return f < math.SmallestNonzeroFloat64
}

func toLabel(columns []int, field int) string {
	for i, label := range columns {
		if field == label {
			return labels[i]
		}
	}
	return ""
}

func parseFields(fields string) (columns []int) {
	for _, i := range strings.Split(fields, ",") {
		c, err := strconv.Atoi(i)
		check.NoError(err, "%s is not a valid column index", i)
		columns = append(columns, c)
	}
	check.If(len(columns) == len(labels),
		"%d fields must be specified:\n%v\n", len(labels), labels)
	return columns
}
