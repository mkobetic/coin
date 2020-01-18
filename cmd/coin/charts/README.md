The `chart` output option for aggregated results produces an html document with embedded result data and d3 code that turns it into an SVG chart when the document is opened in a recent web browser.

The document is produced from the corresponding `.html` and `.js` files in this directory and the results from coin. The names of the files tie them to the specific report/chart type, e.g. `totals.html` and `totals.js` are used to output a chart from the `totals` result type (i.e. aggregated register results)

The `.js` and `.hmtl` files get embedded into the `coin` binary using the `embed.go` script that is normally invoked through `go:generate`. The resulting file is `charts.go` in the `cmd/coin` directory.