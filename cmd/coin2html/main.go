package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mkobetic/coin"
)

//go:generate npm --prefix js run build

//go:embed js/dist/head.html
var htmlHead string

//go:embed js/dist/body.html
var htmlBody string

func main() {
	coin.LoadAll()
	var f = os.Stdout
	var encoder = json.NewEncoder(f)
	encoder.SetIndent("", "\t")
	fmt.Fprint(f, htmlHead)
	fmt.Fprintln(f, "</head>\n<body>")
	fmt.Fprintf(f, `<script type="application/json" id="importedCommodities">`)
	if err := encoder.Encode(coin.Commodities); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
	}
	fmt.Fprintln(f, "\n</script>")
	fmt.Fprintf(os.Stderr, "Commodities: %d\n", len(coin.Commodities))
	fmt.Fprintf(f, `<script type="application/json" id="importedPrices">`)
	if err := encoder.Encode(coin.Prices); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
	}
	fmt.Fprintln(f, "\n</script>")
	fmt.Fprintf(os.Stderr, "Prices: %d\n", len(coin.Prices))
	fmt.Fprintf(f, `<script type="application/json" id="importedAccounts">`)
	if err := encoder.Encode(coin.AccountsByName); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
	}
	fmt.Fprintln(f, "\n</script>")
	fmt.Fprintf(os.Stderr, "Accounts: %d\n", len(coin.AccountsByName))
	fmt.Fprintf(f, `<script type="application/json" id="importedTransactions">`)
	if err := encoder.Encode(coin.Transactions); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
	}
	fmt.Fprintln(f, "\n</script>")
	fmt.Fprintf(os.Stderr, "Transactions: %d\n", len(coin.Transactions))

	fmt.Fprint(f, htmlBody)
	fmt.Fprintln(f, "</body>\n</html>")
	f.Close()
}
