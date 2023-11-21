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

var ()

func init() {
	(&cmdCommodities{}).newCommand("commodities", "com", "c")
}

type cmdCommodities struct {
	flagsWithUsage
	getQuotes bool
	prices    bool
}

func (*cmdCommodities) newCommand(names ...string) command {
	var cmd cmdCommodities
	cmd.FlagSet = newCommand(&cmd, names...)
	setUsage(cmd.FlagSet, `(commodities|com|c) [flags]

Lists commodities and prices.`)
	cmd.BoolVar(&cmd.getQuotes, "q", false, "get current quotes for all commodities")
	cmd.BoolVar(&cmd.prices, "p", false, "print commodity price stats")
	return &cmd
}

func (cmd *cmdCommodities) init() {
	if cmd.prices {
		coin.LoadPrices()
		coin.ResolvePrices()
	} else {
		coin.LoadFile(coin.CommoditiesFile)
	}
}

func (cmd *cmdCommodities) execute(f io.Writer) {
	if cmd.getQuotes {
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
		if cmd.prices {
			fmt.Fprintln(f, c.String())
		} else {
			sym := c.Symbol
			dl := ' '
			if len(sym) > 0 && !c.NoMarket {
				dl = 'Q'
			}
			fmt.Fprintf(f, "%10s | %10s | %c | %s\n", c.Id, sym, dl, c.Name)
		}
	})
}
