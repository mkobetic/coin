package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/mkobetic/coin"
	finance "github.com/piquette/finance-go"
	"github.com/piquette/finance-go/forex"
	"github.com/piquette/finance-go/quote"
)

var (
	cmdCommodities *command
	getQuotes      bool
	prices         bool
)

func init() {
	cmdCommodities = newCommand(commodities_load, commodities, "commodities", "com", "c")
	cmdCommodities.BoolVar(&getQuotes, "q", false, "get current quotes for all commodities")
	cmdCommodities.BoolVar(&prices, "p", false, "print commodity price stats")
}

func commodities_load() {
	if prices {
		coin.LoadPrices()
		coin.ResolvePrices()
	} else {
		coin.LoadFile(coin.CommoditiesFile)
	}
}

func commodities(f io.Writer) {
	if getQuotes {
		coin.CommoditiesDo(func(c *coin.Commodity) {
			if !(c.NoMarket || c.Id == coin.DefaultCommodityId) {
				var q *finance.Quote
				var err error
				if c.Symbol != "" {
					q, err = quote.Get(c.Symbol)
				} else {
					var fx *finance.ForexPair
					fx, err = forex.Get(c.Id + coin.DefaultCommodityId + "=X")
					if err == nil {
						q = &fx.Quote
					}
				}
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s: %s\n", c.Id, err)
					return
				}
				cur := coin.Commodities[q.CurrencyID]
				if cur == nil {
					fmt.Fprintf(os.Stderr, "%s: no commodity for %s\n", c.Id, q.CurrencyID)
					return
				}
				amt := cur.NewAmountFloat(q.RegularMarketPrice)
				var b strings.Builder
				amt.Write(&b, false)
				fmt.Fprintf(f, "P %s %s %s\n", time.Now().Format(coin.DateFormat), c.Id, b.String())
			}
		})
		return
	}
	coin.CommoditiesDo(func(c *coin.Commodity) {
		if prices {
			fmt.Fprintln(f, c.String())
		} else {
			fmt.Fprintf(f, "%10s | %s\n", c.Id, c.Name)
		}
	})
}
