package main

import (
	"flag"
	"io"

	"github.com/mkobetic/coin"
)

func init() {
	(&cmdExample{}).newCommand("example", "ex")
}

type cmdExample struct {
	*flag.FlagSet
	ledger     bool
	begin, end coin.Date
}

func (_ *cmdExample) newCommand(names ...string) command {
	var cmd cmdExample
	cmd.FlagSet = newCommand(&cmd, names...)
	cmd.BoolVar(&cmd.ledger, "ledger", false, "use ledger compatible format")
	cmd.Var(&cmd.begin, "b", "from this date")
	cmd.Var(&cmd.end, "e", "to this date")
	return &cmd
}

func (cmd *cmdExample) init() {
	if cmd.begin.IsZero() {
		cmd.begin.Set("2000")
	}
	if cmd.end.IsZero() {
		cmd.end.Time = cmd.begin.AddDate(10, 0, 0)
	}
}

func (cmd *cmdExample) execute(f io.Writer) {
	commodities := cmd.generateCommodities()
	accounts := cmd.generateAccounts(commodities)
	transactions := cmd.generateTransactions(accounts)
	for _, c := range commodities {
		c.Write(f, cmd.ledger)
		for p := range c.Prices {
			p.Write(f, cmd.ledger)
		}
	}
	for _, a := range accounts {
		a.Write(f, cmd.ledger)
	}
	for _, t := range transactions {
		t.Write(f, cmd.ledger)
	}
}

func (cmd *cmdExample) generateCommodities() []*coin.Commodity {
	return []*coin.Commodity{coin.DefaultCommodity()}
}
