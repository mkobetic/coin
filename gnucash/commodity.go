package gnucash

import (
	"encoding/xml"

	"github.com/mkobetic/coin"
)

/*
Commodity = element gnc:commodity {
  attribute version { "2.0.0" },
  ( ( element cmdty:space { "ISO4217" },    # catégorie (monnaies)
      element cmdty:id { text }    # dénomination
    )
  | ( element cmdty:space { text },
      element cmdty:id { text },
      element cmdty:name { text }?,
      element cmdty:xcode { text }?,
      element cmdty:fraction { text }
    )
  ),
  ( element cmdty:get_quotes { empty },
    element cmdty:quote_source { text }?,
    element cmdty:quote_tz { text | empty }?
  )?,
  element cmdty:slots { KvpSlot+ }?
}
*/
type Commodity struct {
	XMLName     xml.Name   `xml:"commodity"`
	Space       string     `xml:"space"`
	Id          string     `xml:"id"`
	Name        string     `xml:"name,omitempty"`
	Code        string     `xml:"xcode,omitempty"`
	Fraction    int64      `xml:"fraction"`
	QuoteSource string     `xml:"quote_source,omitempty"`
	QuoteTz     string     `xml:"quote_tz,omitempty"`
	Slots       []*KvpSlot `xml:"slots>slot"`
}

func CommodityFrom(c *Commodity) *coin.Commodity {
	cc := &coin.Commodity{
		Id:     c.Id,
		Name:   c.Name,
		Code:   c.Code,
		Prices: make(map[*coin.Commodity][]*coin.Price),
	}
	cc.SetFraction(c.Fraction)
	return cc
}

func resolveCommodities(commodities []*Commodity) {
	for _, gc := range commodities {
		c := CommodityFrom(gc)
		coin.Commodities[c.Id] = c
	}
}
