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
		rules = ReadRules(file)
	}

	if *dumpRules {
		rules.Write(os.Stdout)
		return
	}

	var src *Source
	if *fields != "" {
		src = &Source{fields: parseFields(*fields)}
	} else if *source != "" {
		src = rules.sources[*source]
		check.If(src != nil, "Unknown source %s", *source)
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
		_, err = r.Read()
		check.NoError(err, "Failed to read header from %s", fileName)

		batch, err := readTransactions(r, src.fields, rules)
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

func readTransactions(r *csv.Reader, fields map[string]Fields, rules *Rules) (transactions []*coin.Transaction, err error) {
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transactionFrom(rec, fields, rules))
	}
	return transactions, nil
}

// transactionFrom builds a transaction from a csv row.
// columns list field indexes in the following order:
// account_id, description, posted, amount, symbol, quantity, note
func transactionFrom(row []string, fields map[string]Fields, rules *Rules) *coin.Transaction {
	valueFor := func(name string) string {
		check.Includes(labels, name, "Invalid field name")
		return fields[name].Value(name, row)
	}
	acctId := valueFor("account")
	description := valueFor("description")
	account, found := accountsByCSVId[acctId]
	check.OK(found, "Can't find account with id %s", acctId)
	toAccount := rules.AccountRulesFor(acctId).AccountFor(description)
	if toAccount == nil {
		toAccount = coin.Unbalanced
	}

	date := valueFor("date")
	posted, err := time.Parse(coin.DateFormat, date)
	check.NoError(err, "Parsing date %s", date)

	t := &coin.Transaction{
		Description: description,
		Posted:      posted,
	}

	t.Note = valueFor("note")

	var amount *coin.Amount
	if amt := valueFor("amount"); amt != "" {
		amtf, err := strconv.ParseFloat(amt, 64)
		check.NoError(err, "Parsing amount %s", amt)
		if math.Abs(amtf) > 0.0001 {
			amount = account.Commodity.NewAmountFloat(amtf)
		}
	}
	symbol := valueFor("symbol")
	if symbol == "" {
		check.If(amount != nil, "wat? %s", valueFor("amount"))
		t.Post(account, toAccount, amount, nil)
		return t
	}
	commodity, found := coin.CommoditiesBySymbol[symbol]
	check.OK(found, "Could not find commodity for symbol %s", symbol)
	if account.Commodity != commodity {
		account.WithChildrenDo(func(a *coin.Account) {
			if a.Commodity == commodity {
				account = a
			}
		})
	}

	var quantity *coin.Amount
	if amt := valueFor("quantity"); amt != "" {
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

var labels = []string{
	"account",     //target account ID
	"description", // transaction description
	"date",        // date of the transaction
	"amount",      // the cost of the transaction
	"symbol",      // symbol of the commodity that was traded
	"quantity",    // quantity of the commodity that was traded
	"note",        // optional note associated with the transaction
}

func parseFields(list string) map[string]Fields {
	idxs := strings.Split(list, ",")
	check.If(len(idxs) == len(labels),
		"%d fields must be specified:\n%v\n", len(labels), labels)
	fields := make(map[string]Fields)
	for i, s := range idxs {
		c, err := strconv.Atoi(s)
		check.NoError(err, "%s is not a valid column index", i)
		fields[labels[i]] = Fields{&Field{c, nil}}
	}
	return fields
}
