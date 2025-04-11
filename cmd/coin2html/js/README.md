This is a Typescript project powering the coin2html single page viewer.

The overarching goal is to have everything packaged in a single html file, the data, the style sheet, the code, everything. It needs to run in the browser without being backed by any server infrastructure, just open the HTML file and that's it.

For the longest time I tried to avoid using a JS builder, just piecing the final HTML file together from the various bits here. However the un-modularized Typescript code was difficult to test. I had hopes for ES module support in the browser, but it turns out that inlined modules are useless, cannot be imported (without resorting to some ugly hacks), so eventually I gave up and introduced webpack. I picked webpack because of its HtmlBundlerPlugin that can produce a single HTML page bundling everything.

D3 is the main workhorse of the app, it builds all the visual elements beyond the bare bones dom skeleton defined in body.html. That include not just the charts but the html tables as well. I'd like to stick with D3 as much as possible to keep dependencies to a minimum. I might be tempted to switch to Observable/Plot to reduce the amount of code given that at least some of the charts are pretty standard stock, we'll see, at the moment I'm having too much fun with D3.

As it is the final html file for examples/yearly is ~7MB. This is after stopping importing D3 wholesale and importing just the bits that are used. It does include sourcemaps though to make debugging convenient. There's no minification yet either, which would presumably help as well.

# Project structure

The main pieces are

- head.html - defines the HTML head part and pulls in styles.css
- body.html - lays out the basic/static page structure and pulls in the typescript modules
- styles.css - minimal styling for the page elements
- src/ - contains all the typescript modules

The bundler plugin pulls all of the above together producing two files in dist/

- body.html
- head.html

coin2html looks for these two files to combine them with the JSON data and produce the final HTML page. This is all orchestrated by the `webpack.config.js`. The `npm build` script executes the process.

# Testing

`jest` is used as its the defacto standard these days. Use `npm test` to run the test.

# Data Format

The ledger/journal data can potentially come from any source as long as it conforms to the following specification.

## Commodities

All commodities must be listed in a script element with `type` and `id` as follows

```html
<script type="application/json" id="importedCommodities">
```

The contents of the element must be a JSON object with a property for each commodity, where the property key is the ID of the commodity and the value is its definition.
The `location` property is optional.

Example:

```json
	"CAD": {
    "decimals": 2,
		"id": "CAD",
		"location": "examples/yearly/commodities.coin:1",
		"name": ""
	}
```

## Prices

Prices are optional, however if the ledger uses multiple commodities, there should be at least a single price point for each of the non-default commodities so that conversions can work at all.

If provided, prices must be listed in a script element with `type` and `id` as follows

```html
<script type="application/json" id="importedPrices">
```

The contents of the element must be a JSON array where each element represents individual price entry.
The `location` property is optional.

Example:

```json
{
  "commodity": "USD",
  "currency": "CAD",
  "location": "examples/yearly/prices.coin:116",
  "time": "2008/11/26",
  "value": "1.23 CAD"
}
```

## Accounts

All accounts must be listed in a script element with `type` and `id` as follows.

```html
<script type="application/json" id="importedAccounts">
```

The contents of the element must be a JSON object with a property for each account, where the property key is the full name of the account and the value is its definition.
Parent accounts must be listed before the child accounts. The `parent` property should be omitted for root accounts, e.g `Assets`, `Income`, ...
The `location` property is optional.

Example:

```json
	"Assets:Bank:Checking": {
		"commodity": "CAD",
		"fullName": "Assets:Bank:Checking",
		"location": "examples/yearly/accounts.coin:10",
		"name": "Checking",
		"parent": "Assets:Bank"
	},
```

## Transactions

Transactions must be listed in a script element with type and id as follows.

```html
<script type="application/json" id="importedTransactions">
```

The contents of the element must be a JSON array where each element represents individual transaction.
`notes`, `tags`, `location` and `balance_asserted` properties are optional.

Example:

```json
{
  "description": "Costco",
  "location": "examples/yearly/2010.coin:1",
  "posted": "2010/01/01",
  "notes": ["lots of good stuff #key:value"],
  "tags": {
    "key": "value"
  },
  "postings": [
    {
      "account": "Expenses:Groceries",
      "balance": "764.00 CAD",
      "balance_asserted": false,
      "quantity": "764.00 CAD",
      "notes": ["lots of good stuff #key:value"],
      "tags": {
        "key": "value"
      }
    },
    {
      "account": "Assets:Bank:Checking",
      "balance": "-764.00 CAD",
      "balance_asserted": false,
      "quantity": "-764.00 CAD"
    }
  ]
}
```
