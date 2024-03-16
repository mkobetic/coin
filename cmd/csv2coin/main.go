package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/check"
)

const usage = `Usage: csv2coin [flags] files...

Converts CSV files to coin transactions based on a set of rules (see README).

Flags:`

var (
	fields    = flag.String("fields", "", "ordered list of column indexes to use as transaction fields")
	source    = flag.String("source", "", "which source rules to use to read the files")
	dumpRules = flag.Bool("rules", false, "dump the loaded account rules (useful for formatting)")
)

func init() {
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintln(w, usage)
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	coin.LoadFile(coin.CommoditiesFile)
	coin.LoadFile(coin.AccountsFile)
	coin.ResolveAccounts()

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

	var transactions coin.TransactionsByTime
	for _, fileName := range flag.Args() {
		file, err := os.Open(fileName)
		check.NoError(err, "Failed to open %s", fileName)
		defer file.Close()

		batch := readTransactions(file, src, rules)
		transactions = append(transactions, batch...)
	}

	// write transactions
	sort.Stable(transactions)

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
		if nt := transactionFrom(rec, src.fields, rules); nt != nil {
			transactions = append(transactions, nt)
		}
	}
	return transactions
}

// transactionFrom builds a transaction from a csv row.
func transactionFrom(row []string, fields map[string]Fields, allRules *Rules) *coin.Transaction {
	valueFor := func(name string) string {
		check.Includes(labels, name, "Invalid field name")
		return fields[name].Value(row, fields)
	}
	acctId := valueFor("account")
	description := trim(valueFor("description"))
	rules := allRules.AccountRulesFor(acctId)
	check.If(rules != nil, "Can't find rules for account id %s", acctId)
	account := rules.Account
	toAccount := coin.Unbalanced
	var notes []string
	if rule := rules.RuleFor(description); rule != nil {
		if rule.Account == nil {
			// drop transaction
			return nil
		}
		toAccount = rule.Account
		notes = rule.Notes
	}

	date := valueFor("date")
	posted, err := time.Parse(coin.DateFormat, date)
	check.NoError(err, "Parsing date %s", date)

	t := &coin.Transaction{
		Description: description,
		Posted:      posted,
		Notes:       notes,
	}

	if n := trim(valueFor("note")); len(n) > 0 {
		t.Notes = []string{n}
	}

	var amount *coin.Amount
	if amt := valueFor("amount"); amt != "" {
		commodity := toAccount.Commodity
		if currency := valueFor("currency"); currency != "" {
			var found bool
			commodity, found = coin.Commodities[currency]
			if !found {
				commodity, found = coin.CommoditiesBySymbol[currency]
			}
			check.If(found, "unknown currency %s", currency)
		}
		amount = coin.MustParseAmount(amt, commodity)
	}
	check.If(amount != nil, "amount not found!")
	// refine toAccount based on the amount commodity
	toAccount = findAccountForCommodity(amount.Commodity, toAccount)

	symbol := valueFor("symbol")
	if symbol == "" {
		fromAccount := findAccountForCommodity(amount.Commodity, account)
		t.Post(fromAccount, toAccount, amount, nil)
		return t
	}

	commodity, found := coin.Commodities[symbol]
	if !found {
		commodity, found = coin.CommoditiesBySymbol[symbol]
	}
	check.OK(found, "Could not find commodity for symbol %s", symbol)

	var quantity *coin.Amount
	if amt := valueFor("quantity"); amt != "" {
		quantity = coin.MustParseAmount(amt, commodity)
	}

	if quantity == nil {
		fromAccount := findAccountForCommodity(amount.Commodity, account)
		t.Post(fromAccount, toAccount, amount, nil)
		return t
	}

	// Quantity and amount cannot be both positive or negative,
	// if they are amount wins, make quantity the opposite.
	if quantity.Sign()*amount.Sign() > 0 {
		quantity = quantity.Negated()
	}

	// refine fromAccount based on the quantity commodity
	fromAccount := findAccountForCommodity(quantity.Commodity, account)
	t.PostConversion(fromAccount, quantity, nil, toAccount, amount, nil)
	return t
}

func findAccountForCommodity(c *coin.Commodity, root *coin.Account) *coin.Account {
	account := coin.Unbalanced
	root.FirstWithChildrenDo(func(a *coin.Account) {
		if a.Commodity == c && account == coin.Unbalanced {
			account = a
		}
	})
	return account
}

var labels = []string{
	"account",     //target account ID
	"description", // transaction description
	"date",        // date of the transaction
	"amount",      // the cost of the transaction
	"currency",    // optional currency of the transaction cost
	"symbol",      // optional symbol of the commodity that was traded
	"quantity",    // optional quantity of the commodity that was traded
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
		fields[labels[i]] = Fields{&Field{&c, "", "", nil}}
	}
	return fields
}
