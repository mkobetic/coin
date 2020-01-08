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

		batch := readTransactions(file, src, rules)
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

func readTransactions(in io.Reader, src *Source, rules *Rules) (transactions []*coin.Transaction) {
	r := csv.NewReader(in)
	for i := 0; i < src.skip; i++ {
		row, err := r.Read()
		if err, ok := err.(*csv.ParseError); ok && err.Err == csv.ErrFieldCount {
			r.FieldsPerRecord = len(row)
			continue
		}
		check.NoError(err, "Failed to read header line %d", i)
	}
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		check.NoError(err, "reading transaction")
		transactions = append(transactions, transactionFrom(rec, src.fields, rules))
	}
	return transactions
}

// transactionFrom builds a transaction from a csv row.
// columns list field indexes in the following order:
// account_id, description, posted, amount, symbol, quantity, note
func transactionFrom(row []string, fields map[string]Fields, rules *Rules) *coin.Transaction {
	valueFor := func(name string) string {
		check.Includes(labels, name, "Invalid field name")
		return fields[name].Value(row, fields)
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
		amount = coin.MustParseAmount(amt, toAccount.Commodity)
	}
	check.If(amount != nil, "amount not found!")
	symbol := valueFor("symbol")
	if symbol == "" {
		if account.Commodity != amount.Commodity {
			account.WithChildrenDo(func(a *coin.Account) {
				if a.Commodity == amount.Commodity {
					account = a
				}
			})
		}
		t.Post(account, toAccount, amount, nil)
		return t
	}
	commodity, found := coin.Commodities[symbol]
	if !found {
		commodity, found = coin.CommoditiesBySymbol[symbol]
	}
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
		quantity = coin.MustParseAmount(amt, commodity)
	}

	if quantity == nil {
		t.Post(account, toAccount, amount, nil)
		return t
	}
	// Quantity and amount cannot be both positive or negative,
	// if they are amount wins, make quantity the opposite.
	if quantity.Sign()*amount.Sign() > 0 {
		quantity = quantity.Negated()
	}
	t.PostConversion(account, quantity, nil, toAccount, amount, nil)
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
		fields[labels[i]] = Fields{&Field{c, "", nil}}
	}
	return fields
}
