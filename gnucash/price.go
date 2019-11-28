package gnucash

import (
	"sort"

	"github.com/mkobetic/coin"
)

/*
Price = element price {
  element price:id { attribute type { "guid" }, GUID },
  element price:commodity {
    element cmdty:space { text },
    element cmdty:id { text }
  },
  element price:currency {
    element cmdty:space { text },
    element cmdty:id { text }
  },
  element price:time { TimeSpec },
  element price:source { text }?,
  element price:type { "bid" | "ask" | "last" | "nav" | "transaction" | "unknown" }?,
  element price:value { GncNumeric }
}
*/

type Price struct {
	Guid           string `xml:"id"`
	CommoditySpace string `xml:"commodity>space"`
	CommodityId    string `xml:"commodity>id"`
	CurrencySpace  string `xml:"currency>space"`
	CurrencyId     string `xml:"currency>id"`
	Date           string `xml:"time>date"`
	Source         string `xml:"source,omitempty"`
	ValueFraction  string `xml:"value"`
	Type           string `xml:"type,omitempty"`
}

func resolvePrices(prices []*Price) {
	for _, gp := range prices {
		p := &coin.Price{}
		p.Commodity = coin.Commodities[gp.CommodityId]
		p.Currency = coin.Commodities[gp.CurrencyId]
		p.Value = mustParseAmount(gp.ValueFraction, p.Currency)
		p.Time = mustParseTimeStamp(gp.Date)
		p.Commodity.Prices[p.Currency] =
			append(p.Commodity.Prices[p.Currency], p)
		coin.Prices = append(coin.Prices, p)
	}
	// Sort commodity prices.
	for _, c := range coin.Commodities {
		for _, p := range c.Prices {
			sort.Slice(p, func(i, j int) bool {
				return p[i].Time.After(p[j].Time)
			})
		}
	}
	sort.Slice(coin.Prices, func(i, j int) bool {
		return coin.Prices[i].Time.Before(coin.Prices[j].Time)
	})
}
