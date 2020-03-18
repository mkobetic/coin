// Package gnucash allows reading GnuCash v2 XML database and converting it to a coin ledger.
package gnucash

import (
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/mkobetic/coin"
)

// Relax NG schema:
// https://github.com/Gnucash/gnucash/blob/maint/libgnucash/doc/xml/gnucash-v2.rnc

// Gnucash is the root element of the Gnucash DB
type Gnucash struct {
	XMLName xml.Name `xml:"gnc-v2"`
	Book    Book     `xml:"book"`
}

type Book struct {
	Accounts     []*Account     `xml:"account"`
	Commodities  []*Commodity   `xml:"commodity"`
	Prices       []*Price       `xml:"pricedb>price"`
	Transactions []*Transaction `xml:"transaction"`
}

// Resolve relinks all the book objects after unmarshaling from XML
func (book *Book) Resolve() {
	resolveCommodities(book.Commodities)
	resolvePrices(book.Prices)
	resolveAccounts(book.Accounts)
	resolveTransactions(book.Transactions)
}

func Load(fn string) *Book {
	var db Gnucash
	file, err := os.Open(fn)
	if err != nil {
		panic(fmt.Errorf("%s: %s", fn, err))
	}
	defer file.Close()

	r, err := gzip.NewReader(file)
	if err != nil {
		panic(fmt.Errorf("%s: %s", fn, err))
	}

	d := xml.NewDecoder(r)
	err = d.Decode(&db)
	if err != nil {
		panic(err)
	}
	db.Book.Resolve()
	return &(db.Book)
}

// TimeStamp is GnuCash time format
const TimeStamp = "2006-01-02 15:04:05 -0700"

func mustParseTimeStamp(ts string) time.Time {
	t, err := time.Parse(TimeStamp, ts)
	if err != nil {
		panic(err)
	}
	return t
}

// GncNumeric is a fraction
func mustParseAmount(f string, c *coin.Commodity) *coin.Amount {
	values := strings.Split(f, "/")
	if len(values) != 2 {
		panic("invalid fraction: " + f)
	}
	num := new(big.Int)
	num.SetString(values[0], 10)
	den := new(big.Int)
	den.SetString(values[1], 10)
	if den.Sign() <= 0 {
		panic("invalid denominator: " + den.String())
	}
	return coin.NewAmountFrac(num, den, c)
}

/*
KvpSlot = element slot {
  element slot:key { text },
  KvpValue
}

KvpValue = ( element slot:value { attribute type { "integer" }, xsd:int }
           | element slot:value { attribute type { "double" }, xsd:double }
           | element slot:value { attribute type { "numeric" }, GncNumeric }
           | element slot:value { attribute type { "string" }, text }
           | element slot:value { attribute type { "guid" }, GUID }
           | element slot:value { attribute type { "timespec" }, TimeSpec }
           | element slot:value { attribute type { "gdate" }, GDate }
           | element slot:value { attribute type { "binary" }, xsd:string { pattern = "[0-9a-f]*" }}
           | element slot:value { attribute type { "list" }, KvpValue* }
           | element slot:value { attribute type { "frame" }, KvpSlot* }
           )
*/

type KvpValue struct {
	XMLName xml.Name `xml:"value"`
	Value   string   `xml:",innerxml"`
	Type    string   `xml:"type,attr"`
}

type KvpSlot struct {
	XMLName xml.Name `xml:"slot"`
	Key     string   `xml:"key"`
	Value   KvpValue `xml:"value"`
}
