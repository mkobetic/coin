package main

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mkobetic/coin"
)

//go:generate npm --prefix js run build

//go:embed js/src/index.js
var jsSrc string

//go:embed js/styles.css
var cssSrc string

//go:embed js/head.html
var htmlHead string

//go:embed js/body.html
var htmlBody string

func main() {
	coin.LoadAll()
	var f = os.Stdout
	var encoder = json.NewEncoder(f)
	encoder.SetIndent("", "\t")
	fmt.Fprint(f, htmlHead)
	fmt.Fprintln(f, "\n\t<style>")
	fmt.Fprint(f, cssSrc)
	fmt.Fprintln(f, "\n\t</style>")
	fmt.Fprintln(f, "</head>\n<body>")
	fmt.Fprint(f, htmlBody)
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

	// skip imports from the compiled index.js file,
	// imports are managed explicitly in the html head
	js := bufio.NewReader(strings.NewReader(jsSrc))
	line := "import - fake import that will be skipped"
	for strings.HasPrefix(line, "import") {
		line, _ = js.ReadString('\n')
	}
	fmt.Fprintln(f, `<script type="text/javascript" id="code">`)
	fmt.Fprintln(f, line) // write first non-import line
	io.Copy(f, js)        // copy the rest from the reader
	fmt.Fprintln(f, "\n</script>")
	fmt.Fprintln(f, "</body>\n</html>")
	f.Close()
}
