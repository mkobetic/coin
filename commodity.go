package coin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"
	"unicode"

	"github.com/mkobetic/coin/rex"
)

var (
	// Commodities by Id
	Commodities         = map[string]*Commodity{}
	CommoditiesBySymbol = map[string]*Commodity{}
)

type Commodity struct {
	Id       string
	Name     string
	Code     string
	Decimals int    // how many decimal places to use
	NoMarket bool   // Don't download prices
	Symbol   string // symbol to use for quotes

	// price lists by currency
	Prices map[*Commodity][]*Price

	// Id quoted if required by ledger
	quotedId string

	line uint
	file string
}

func (c *Commodity) AddPrice(p *Price) {
	if c.Prices == nil {
		c.Prices = make(map[*Commodity][]*Price)
	}
	c.Prices[p.Currency] = append(c.Prices[p.Currency], p)
}

func (c *Commodity) Currencies() (currencies []*Commodity) {
	for cur := range c.Prices {
		currencies = append(currencies, cur)
	}
	return currencies
}

/*
Ledger requires quoting the commodity id if it contains any of the following:
Any kind of white-space
Numerical digits
Punctuation: .,;:?!
Mathematical and logical operators: -+*^/&|=
Bracketing characters: <>[](){}
The at symbol: @
*/
const disallowed = " \t0123456789.,;:?!-+*^/&|=<>[](){}@"

func (c *Commodity) SafeId(ledger bool) string {
	if ledger {
		if c.quotedId == "" {
			c.quotedId = c.Id
			if strings.ContainsAny(c.Id, disallowed) {
				c.quotedId = `"` + c.Id + `"`
			}
		}
		return c.quotedId
	}
	return c.Id
}

func (c *Commodity) String() string {
	if len(c.Prices) == 0 {
		return c.Id
	}
	var prices []string
	for c, p := range c.Prices {
		format := "%s %a %s [%d]"
		if c.Decimals > 0 {
			format = "%s %." + strconv.Itoa(c.Decimals) + "a %s [%d]"
		}
		prices = append(prices,
			fmt.Sprintf(format,
				p[0].Time.Format(DateFormat),
				p[0].Value,
				c.Id,
				len(p)))
	}
	return fmt.Sprintf("%s: %s",
		c.Id,
		strings.Join(prices, ", "))
}

func (c *Commodity) Location() string {
	return fmt.Sprintf("%s:%d", c.file, c.line)
}

/*
commodity USD

	note American Dollars
	format 1000.00 USD
	nomarket
	default
*/
func (c *Commodity) Write(w io.Writer, ledger bool) error {
	format := "1"
	if c.Decimals > 0 {
		format = "1." + strings.Repeat("0", c.Decimals)
	}
	lines := []string{"commodity ", c.SafeId(ledger), "\n"}
	if c.Name != "" {
		lines = append(lines, "  note ", c.Name, "\n")
	}
	lines = append(lines, "  format ", format, " ", c.SafeId(ledger), "\n")
	if c.NoMarket {
		lines = append(lines, "  nomarket\n")
	}
	for _, line := range lines {
		_, err := io.WriteString(w, line)
		if err != nil {
			return err
		}
	}
	return nil
}

var CommodityREX = rex.MustCompile(`(?P<commodity>[A-Za-z][\w]*)`)
var commodityHeadREX = rex.MustCompile(`commodity\s+%s`, CommodityREX)
var commodityBodyREX = rex.MustCompile(``+
	`(\s+note\s+(?P<note>.+))|`+
	`(\s+format\s+(?P<format>%s))|`+
	`(\s+(?P<nomarket>nomarket)\s*)|`+
	`(\s+symbol\s+(?P<symbol>[\w\.]+))|`+
	`(\s+(?P<default>default)\s*)`,
	AmountREX)

func (p *Parser) parseCommodity(fn string) (*Commodity, error) {
	c := &Commodity{Decimals: 2, line: p.lineNr, file: fn}
	match := commodityHeadREX.Match(p.Bytes())
	c.Id = match["commodity"]
	for p.Scan() {
		line := p.Bytes()
		if len(bytes.TrimSpace(line)) == 0 || !unicode.IsSpace(rune(line[0])) {
			return c, nil
		}
		match = commodityBodyREX.Match(line)
		if match == nil {
			return c, fmt.Errorf("unrecognized commodity line: %s", p.Text())
		}
		if n := match["note"]; n != "" {
			c.Name = n
		} else if match["amount"] != "" {
			if f := match["decimals"]; len(f) == 0 {
				c.Decimals = 0
			} else {
				c.Decimals = len(f) - 1
			}
		} else if match["nomarket"] != "" {
			c.NoMarket = true
		} else if s := match["symbol"]; s != "" {
			c.Symbol = s
		} else if match["default"] != "" {
			DefaultCommodityId = c.Id
		} else {
			return c, fmt.Errorf("%s - failed to match commodity line: %s", c.Location(), p.Text())
		}
	}
	return c, p.Err()
}

// Convert the amount from c2 commodity to amount in c commodity using known
// commodity prices. Try to find a conversion path through known intermediate
// commodities as well.
func (c *Commodity) Convert(amount *Amount, c2 *Commodity) (*Amount, error) {
	return c.convert(amount, c2, nil)
}

func (c *Commodity) includedIn(list []*Commodity) bool {
	for _, c2 := range list {
		if c == c2 {
			return true
		}
	}
	return false
}

func (c *Commodity) convert(amount *Amount, c2 *Commodity, previous []*Commodity) (*Amount, error) {
	if c == c2 {
		// Nothing to convert
		return amount, nil
	}
	// Does c2 have prices in c currency?
	prices := c2.Prices[c]
	if prices != nil {
		p := prices[0]
		val := amount.Times(p.Value)
		return val, nil
	}
	// Otherwise try to follow each c2 price currency
	for c3, prices := range c2.Prices {
		// Check if we tried this currency before to avoid cycles
		if c3.includedIn(previous) {
			continue
		}
		p := prices[0]
		val2 := amount.Times(p.Value)
		val3, err := c.convert(val2, c3, append(previous, c2))
		if err == nil {
			return val3, nil
		}
	}
	// Didn't find any path that leads to c
	return nil, fmt.Errorf("cannot convert %s => %s", c2.Id, c.Id)
}

func (c *Commodity) NewAmountFloat(f float64) *Amount {
	return NewAmount(
		// FIXME: the int64 conversion can overflow
		big.NewInt(int64(f*float64(pow10(c.Decimals)))),
		c,
	)
}

func (c *Commodity) SetFraction(frac int64) {
	c.Decimals = log10(frac)
}

func (c *Commodity) MarshalJSON() ([]byte, error) {
	value := map[string]interface{}{
		"id":       c.Id,
		"name":     c.Name,
		"decimals": c.Decimals,
	}
	if c.Code != "" {
		value["code"] = c.Code
	}
	return json.MarshalIndent(value, "", "\t")
}
