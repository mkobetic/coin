package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/mkobetic/coin"
	"github.com/mkobetic/coin/gnucash"
)

const usage = `Usage: gc2coin [flags]

Converts GnuCash XML database (v2) to a coin transactions.
`

var (
	gnucashDB = flag.String("gnucashdb", os.Getenv("GNUCASHDB"), "path to the database")
	yearly    = flag.Bool("y", false, "split transactions into separate files by year (requires COINDB directory)")
	ledger    = flag.Bool("l", false, "write ledger friendly format")
)

func init() {
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintln(w, usage)
		flag.PrintDefaults()
	}
}

func next(fn string) (*os.File, func()) {
	if coin.DB == "" {
		return os.Stdout, func() {}
	}
	fn = filepath.Join(coin.DB, fn)
	f, err := os.Create(fn)
	if err != nil {
		panic(err)
	}
	return f, func() { f.Close() }
}

func main() {
	flag.Parse()
	if len(*gnucashDB) == 0 {
		fmt.Println("missing gnucash db filename")
		flag.Usage()
		os.Exit(1)
	}
	if *yearly && coin.DB == "" {
		fmt.Println("-y requires $COINDB set")
		flag.Usage()
		os.Exit(1)
	}
	gnucash.Load(*gnucashDB)
	f, done := next(coin.CommoditiesFilename)
	defer done()
	coin.CommoditiesDo(func(c *coin.Commodity) {
		if c.Name == "template" {
			return
		}
		err := c.Write(f, *ledger)
		if err != nil {
			panic(err)
		}
		_, err = io.WriteString(f, "\n")
		if err != nil {
			panic(err)
		}
	})

	// Assume prices are sorted by date
	var year int
	if *yearly {
		year := coin.Prices[0].Time.Year()
		f, done = next(strconv.Itoa(year) + coin.PricesExtension)
		defer done()
	} else {
		f, done = next(coin.PricesFilename)
		defer done()
	}
	for _, p := range coin.Prices {
		if *yearly && p.Time.Year() != year {
			year = p.Time.Year()
			f, done = next(strconv.Itoa(year) + coin.PricesExtension)
			defer done()
		}
		err := p.Write(f, *ledger)
		if err != nil {
			panic(err)
		}
	}

	f, done = next(coin.AccountsFilename)
	defer done()
	coin.AccountsDo(func(a *coin.Account) {
		err := a.Write(f, *ledger)
		if err != nil {
			panic(err)
		}
		_, err = io.WriteString(f, "\n")
		if err != nil {
			panic(err)
		}
	})

	if *yearly {
		year := coin.Transactions[0].Posted.Year()
		f, done = next(strconv.Itoa(year) + coin.TransactionsExtension)
		defer done()
	} else {
		f, done = next(coin.TransactionsFilename)
		defer done()
	}
	for _, p := range coin.Transactions {
		if *yearly && p.Posted.Year() != year {
			year = p.Posted.Year()
			f, done = next(strconv.Itoa(year) + coin.TransactionsExtension)
			defer done()
		}
		err := p.Write(f, *ledger)
		if err != nil {
			panic(err)
		}
		_, err = io.WriteString(f, "\n")
		if err != nil {
			panic(err)
		}
	}

}
