package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/check"
)

var (
	fields = flag.String("fields", "", "ordered list of column indexes to use as transaction fields")

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

	var transactions []*coin.Transaction
	for _, fileName := range flag.Args() {
		file, err := os.Open(fileName)
		check.NoError(err, "Failed to open %s", fileName)
		defer file.Close()

		r := csv.NewReader(file)
		header, err := r.Read()
		check.NoError(err, "Failed to read header from %s", fileName)

		var columns []int
		for _, i := range strings.Split(*fields, ",") {
			c, err := strconv.Atoi(i)
			check.NoError(err, "%s is not a valid column index", i)
			columns = append(columns, c)
		}
		if len(columns) != len(labels) {
			fmt.Fprintf(os.Stderr, "%d fields must be specified:\n%v\n", len(labels), labels)
			os.Exit(1)
		}
		for i, h := range header {
			if label := toLabel(columns, i); label != "" {
				fmt.Fprintf(os.Stderr, "%s => %s\n", h, label)
			} else {
				fmt.Fprintf(os.Stderr, "%s\n", h)
			}
		}
		batch, err := readTransactions(r, columns)
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

func readTransactions(r *csv.Reader, columns []int) (transactions []*coin.Transaction, err error) {
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transactionFrom(rec, columns))
	}
	return transactions, nil
}

var labels = []string{"account_id", "description", "posted", "amount", "symbol", "quantity"}

// transactionFrom builds a transaction from a csv row.
// columns list field indexes in the following order:
// account_id, description, posted, amount, symbol, quantity
func transactionFrom(row []string, columns []int) *coin.Transaction {
	account, found := accountsByCSVId[row[columns[0]]]
	check.OK(found, "Can't find account with id %s", row[columns[0]])
	description := row[columns[1]]
	posted, err := time.Parse(coin.DateFormat, row[columns[2]])
	check.NoError(err, "Parsing date %s", row[columns[2]])

	t := &coin.Transaction{
		Description: description,
		Posted:      posted,
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
		t.Post(account, coin.Unbalanced, amount, nil)
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
		t.Post(account, coin.Unbalanced, quantity, nil)
		return t
	}

	t.PostConversion(account, amount, nil, coin.Unbalanced, quantity, nil)
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
